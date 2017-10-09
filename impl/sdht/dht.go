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

	listen *net.Listener

	shutdown chan bool
}

var network iface.Net

func (d *SDHT) Init(seeds []iface.Address, net iface.Net) error {
	d.id = genId()
	network = net

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
	case "find_value":
		req := findValueReq{}
		if err := json.Unmarshal(data, &req); err != nil {
			return marshal(findValueResp{Error: err})
		}
		setAlive(req.ID)
		out, err = d.FindValue(req.Key)
		if err != nil {
			return marshal(findValueResp{Error: err})
		}
		return marshal(findValueResp{Data: out})

	case "find_node":
		req := findNodeReq{}
		if err := json.Unmarshal(data, &req); err != nil {
			return marshal(findNodeResp{Error: err})
		}
		setAlive(req.ID)
		peers, err := d.FindNode(req.FindID)
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
		setAlive(req.ID)
		d.Store(req.Key, req.Data)

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

func (d *SDHT) FindValue(key string) ([]byte, error) {
	return nil, nil
}

func (d *SDHT) StoreValue(key string, data []byte) error {
	return nil
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
