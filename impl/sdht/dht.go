package sdht

import (
	"encoding/json"
	"log"
	"sort"
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

// TODO: This should not remain global if we want to allow multiple instances of SDHT.
var network iface.Net

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (d *SDHT) Init(seeds []iface.Address, net iface.Net) error {
	d.id = genId()
	d.store = make(Storage)
	d.alive = map[ID]int{}
	network = net

	for _, seed := range seeds {
		root := &Peer{ID{}, seed}
		if err := root.Ping(); err != nil {
			continue
		}

		nodes, err := root.FindNode(marshalID(d.id))
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
	//return errors.New("no provided seed completed initial connection")
	go d.serve()
	return nil
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
		peers, err := d.findNode(req.Key)
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

func (d *SDHT) findClosestPeers(key string) ([]Peer, error) {
	keyID, _ := unmarshalID(key)
	peers, err := d.findNode(key)
	if err != nil {
		return nil, err
	}

	sort.Slice(peers, func(i, j int) bool {
		return isBetter(keyID, peers[i], peers[j])
	})

	for {
		hopPeers := []Peer{}
		for _, p := range peers {
			peersP, err := p.FindNode(key)
			if err != nil {
				log.Println("Error contacting peer:", err)
			}
			hopPeers = append(hopPeers, peersP...)
		}

		sort.Slice(hopPeers, func(i, j int) bool {
			return isBetter(keyID, hopPeers[i], hopPeers[j])
		})

		// TODO: This is wrong.
		if !isBetter(keyID, hopPeers[0], peers[0]) {
			return nil, nil
		}

		peers = hopPeers[:min(len(hopPeers), dhtK)]
	}
}

func (d *SDHT) StoreValue(key string, data []byte) error {
	// TODO: fill this function.
	return nil
}

func (d *SDHT) FindValue(key string) ([]byte, error) {
	data, peers, err := d.findValue(key)
	if err != nil {
		return nil, err
	}
	if data != nil {
		return data, nil
	}

	keyID, _ := unmarshalID(key)
	sort.Slice(peers, func(i, j int) bool {
		return isBetter(keyID, peers[i], peers[j])
	})

	for {
		hopPeers := []Peer{}
		for _, p := range peers {
			dataP, peersP, err := p.FindValue(key)
			if dataP != nil {
				return data, nil
			}
			if err != nil {
				log.Println("Error contacting peer:", err)
			}
			hopPeers = append(hopPeers, peersP...)
		}

		sort.Slice(hopPeers, func(i, j int) bool {
			return isBetter(keyID, hopPeers[i], hopPeers[j])
		})

		if !isBetter(keyID, hopPeers[0], peers[0]) {
			return nil, nil
		}

		peers = hopPeers[:min(len(hopPeers), dhtK)]
	}

	return nil, nil
}

// isBetter returns true if peer1 is closer to key than peer2.
func isBetter(key ID, peer1, peer2 Peer) bool {
	return dist(peer1.ID, key) < dist(peer2.ID, key)
}

// dist returns the distance between two keys.
func dist(id1, id2 ID) int {
	id1bits := id1.toBitString()
	id2bits := id2.toBitString()
	for i := range id1bits {
		if id1bits[i] != id2bits[i] {
			return numBuckets - i
		}
	}
	return 0
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
	buckets := d.buckets.bs
	newBuckets := []Peer{}

	// TODO(pallavag): Remove unsafe unmarshals.
	keyID, _ := unmarshalID(key)
	for _, b := range buckets {
		newBuckets = append(newBuckets, b...)
	}

	sort.Slice(newBuckets, func(i, j int) bool {
		return isBetter(keyID, newBuckets[i], newBuckets[j])
	})

	return newBuckets[:min(len(newBuckets), dhtK)], nil
}

func (d *SDHT) setAlive(id ID) {
	d.alive[id] = int(time.Now().Unix())
}

func (d *SDHT) recordExit(id ID) {
	delete(d.alive, id)
	d.buckets.replace(d.id, id)
}

// TODO: Move this to apiserver.
func (d *SDHT) serve() error {
	return network.Listen(iface.Address{
		IP:   "0.0.0.0",
		Port: 6779,
		//Proto: iface.TCP,
	}, d.Respond, d.shutdown)
}
