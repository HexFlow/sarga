package sdht

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/sakshamsharma/sarga/common/dht"
	"github.com/sakshamsharma/sarga/common/iface"
)

// SDHT is a minimal implementation of a DHT (dht.DHT) to be used with sarga.
type SDHT struct {
	id      ID
	buckets buckets
	store   Storage
	alive   map[ID]int

	shutdown chan bool
}

var _ dht.DHT = &SDHT{}

var network iface.Net

func (d *SDHT) Init(seeds []iface.Address, net iface.Net) error {
	d.id = genId()
	d.store = make(Storage)
	network = net

	for _, seed := range seeds {
		root := &Peer{ID{}, seed}
		if err := root.Ping(); err != nil {
			continue
		}

		nodes, err := root.FindNode(d.id)
		if err != nil {
			continue
		}

		for _, node := range nodes {
			d.buckets.insert(d.id, node)
		}

		// TODO: Somehow insert the root node into buckets as well.
		// We do not know its ID.

		d.shutdown = make(chan bool)
		go d.serve()
		return nil
	}
	return errors.New("no provided seed completed initial connection")
}

func (d *SDHT) Shutdown() {
	d.shutdown <- true
}

func (d *SDHT) Respond(action string, data []byte) []byte {
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
		peers, err := d.findNode(marshalID(req.FindID))
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
		d.store.Set(req.Key, []byte(req.Data))

	case "exit":
		req := exitReq{}
		if err := json.Unmarshal(data, &req); err != nil {
			log.Println(err)
			return nil
		}
		d.recordExit(req.ID)

	default:
		log.Println("Request not recognized.")
	}
	return nil
}

func (d *SDHT) StoreValue(key string, data []byte) error {
	return d.store.Set(key, data)
}

func (d *SDHT) FindValue(key string) ([]byte, error) {
	// TODO
	return nil, nil
}

func (d *SDHT) findValue(key string) ([]byte, []Peer, error) {
	if val, err := d.store.Get(key); err == nil {
		return val, nil, nil
	}
	peers, err := d.findNode(key)
	if err == nil {
		return nil, peers, nil
	}
	return nil, nil, err
}

func (d *SDHT) findNode(key string) ([]Peer, error) {
	// TODO
	return nil, nil
}

func (d *SDHT) setAlive(id ID) {
	d.alive[id] = int(time.Now().Unix())
}

func (d *SDHT) recordExit(id ID) {
	delete(d.alive, id)
	d.buckets.replace(d.id, id)
}

// TODO: Move this to apiserver
func (d *SDHT) serve() error {
	return network.Listen(iface.Address{
		IP:    "0.0.0.0",
		Port:  6779,
		Proto: iface.TCP,
	}, d.Respond, d.shutdown)

	// listen, err := net.Listen("tcp", "0.0.0.0:6779")
	// if err != nil {
	// 	log.Printf("Failed to open listening socket: %s", err)
	// 	return
	// }
	// d.listen = listen

	// for {
	// 	conn, err := d.listen.Accept()
	// 	if err != nil {
	// 		select {
	// 		case <-d.shutdown:
	// 			return
	// 		default:
	// 			log.Printf("Accept failed: %v", err)
	// 		}
	// 	}
	// 	// TODO: Make this work
	// 	// err = d.respond(conn)
	// 	// if err != nil {
	// 	// 	log.Printf("Responding to connection failed: %v", err)
	// 	// }
	// }
}
