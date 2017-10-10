package sdht

import (
	"encoding/json"
	"errors"
	"log"
	"net"

	"github.com/sakshamsharma/sarga/common/iface"
)

// SDHT is a minimal implementation of a DHT (dht.DHT) to be used with sarga.
type SDHT struct {
	id      ID
	buckets buckets
	store   Storage

	listen *net.Listener

	shutdown chan bool
}

var network iface.Net

func (d *SDHT) Init(seeds []iface.Address, net iface.Net) error {
	d.id = genId()
	d.store = make(Storage)
	network = net
	d.buckets.ownerID = d.id

	for _, seed := range seeds {
		root := &Peer{ID{}, seed}
		if err := root.Ping(); err != nil {
			continue
		}

		nodes, err := root.FindNode(id)
		if err != nil {
			continue
		}

		for _, node := range nodes {
			d.buckets.insert(node)
		}

		// TODO: Somehow insert the root node into buckets as well.
		// We do not know its ID.

		d.shutdown = make(chan bool)
		go d.serve()
		return nil
	}
	return errors.New("no provided seed completed initial connection")
}

func (d *SDHT) ShutDown() {
	if d.listen == nil {
		return
	}
	d.shutdown <- true
	d.listen.Close()
}

func (d *SDHT) Respond(action string, data []byte) []byte {
	var out []byte
	var err error
	switch action {
	case "ping":
		return marshal(pingResp{ID: d.id})

	case "find_value":
		req := findValueReq{}
		if err := json.Unmarshal(data, &req); err != nil {
			return marshal(findValueResp{Error: err})
		}
		d.setAlive(req.ID)
		out, peers, err := d.findValue(req.Key)
		if err != nil {
			return marshal(findValueResp{Error: err})
		}
		return marshal(findValueResp{Data: out, Peers: peers})

	case "find_node":
		req := findNodeReq{}
		if err := json.Unmarshal(data, &req); err != nil {
			return marshal(findNodeResp{Error: err})
		}
		d.setAlive(req.ID)
		peers, err := d.findNode(req.FindID)
		if err != nil {
			return marshal(findNodeResp{Error: err})
		}
		return marshal(findNodeResp{Peers: peers})

	case "store":
		req := storeReq{}
		if err := json.Unmarshal(data, &req); err != nil {
			log.Println(err)
			return nil
		}
		d.setAlive(req.ID)
		d.store(req.Key, req.Data)

	case "exit":
		req := exitReq{}
		if err := json.Unmarshal(data, &req); err != nil {
			log.Println(err)
			return nil
		}
		d.Exit(req.ID)

	default:
		log.Println("Request not recognized.")
	}
	return nil
}

func (d *SDHT) findValue(key string) ([]byte, []Peer, error) {
	if val, err := d.store.Get(key); err == nil {
		return val, nil, nil
	}
	if peers, err := d.findNode(key); err == nil {
		return nil, peers, nil
	} else {
		return nil, nil, err
	}
}

func (d *SDHT) FindValue(key string) ([]byte, error) {
	return d.store.Get(key)
}

func (d *SDHT) StoreValue(key string, data []byte) error {
	return d.store.Set(key, data)
}

// TODO: Move this to apiserver
func (d *SDHT) serve() {
	listen, err := net.Listen("tcp", "0.0.0.0:6779")
	if err != nil {
		log.Printf("Failed to open listening socket: %s", err)
		return
	}
	d.listen = listen

	for {
		conn, err := d.listen.Accept()
		if err != nil {
			select {
			case <-d.shutdown:
				return
			default:
				log.Printf("Accept failed: %v", err)
			}
		}
		err = d.respond(conn)
		if err != nil {
			log.Printf("Responding to connection failed: %v", err)
		}
	}
}
