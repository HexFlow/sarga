package sdht

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/sakshamsharma/sarga/common/dht"
	"github.com/sakshamsharma/sarga/common/iface"
)

// SDHT is a minimal implementation of a DHT (dht.DHT) to be used with sarga.
type SDHT struct {
	id      ID
	addr    iface.Address
	buckets buckets
	store   Storage
	alive   map[ID]int

	shutdown chan bool
}

var _ dht.DHT = &SDHT{}

// TODO: This should not remain global if we want to allow multiple instances
// of SDHT.
var network iface.Net

func (d *SDHT) Init(addr iface.Address, seeds []iface.Address, net iface.Net) error {
	d.id = genID()
	d.store = make(Storage)
	d.alive = map[ID]int{}
	d.addr = addr
	network = net

	log.Println(d.id, "starting init")

	for _, seed := range seeds {
		root := &Peer{ID{}, seed}
		if err := root.Ping(); err != nil {
			continue
		}
		// If ping was successful, root.ID should now be filled.
		log.Println(d.id, "realized about", root.ID)

		d.buckets.insert(d.id, *root)
		d.findClosestPeers(marshalID(d.id), true)

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
		fmt.Println(marshalID(d.id), "was asked about FindValue for", req.Key)
		d.setAliveTime(req.ID)
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
		d.setAlive(req.Asker)
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
		d.setAliveTime(req.ID)
		fmt.Println(marshalID(d.id), "is storing key", req.Key)
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
	fmt.Println(marshalID(d.id), "Sending StoreValue", key, string(data))
	peers, err := d.findClosestPeers(key, false)
	if err != nil {
		return err
	}

	for _, p := range peers {
		err := p.SendStore(d.id, key, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *SDHT) FindValue(key string) ([]byte, error) {
	//fmt.Println("FindValue", marshalID(d.id), key)
	fmt.Println(marshalID(d.id), "wants key", key)
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
			dataP, peersP, err := p.FindValue(d.id, key)
			if dataP != nil {
				return dataP, nil
			}
			if err != nil {
				log.Println("Error contacting peer:", err)
			}
			hopPeers = append(hopPeers, peersP...)
		}

		sort.Slice(hopPeers, func(i, j int) bool {
			return isBetter(keyID, hopPeers[i], hopPeers[j])
		})

		if !isBetterSlice(keyID, hopPeers, peers) {
			return nil, nil
		}

		peers = hopPeers[:min(len(hopPeers), dhtK)]
	}
}

func (d *SDHT) getPeer() Peer {
	return Peer{
		ID:   d.id,
		Addr: d.addr,
	}
}

func (d *SDHT) findClosestPeers(key string, insert bool) ([]Peer, error) {
	keyID, _ := unmarshalID(key)
	peers, err := d.findNode(key)
	if err != nil {
		return nil, err
	}

	sort.Slice(peers, func(i, j int) bool {
		return isBetter(keyID, peers[i], peers[j])
	})
	log.Println(d.id, "has peers", peers)

	var peersUniq map[ID]Peer

	for {
		peersUniq = map[ID]Peer{}
		for _, p := range peers {
			peersUniq[p.ID] = p
		}

		hopPeers := []Peer{}
		for _, p := range peersUniq {
			if insert && p.ID != d.id {
				log.Println(d.id, "added peer", p.ID)
				d.buckets.insert(d.id, p)
			}
			peersP, err := p.FindNode(d.getPeer(), key)
			if err != nil {
				log.Println("Error contacting peer:", err)
			}
			hopPeers = append(hopPeers, peersP...)
		}
		for _, p := range hopPeers {
			peersUniq[p.ID] = p
		}

		hopPeers = []Peer{}
		for _, p := range peersUniq {
			hopPeers = append(hopPeers, p)
		}

		sort.Slice(hopPeers, func(i, j int) bool {
			return isBetter(keyID, hopPeers[i], hopPeers[j])
		})
		log.Println(d.id, "got hops", hopPeers)

		if !isBetterSlice(keyID, hopPeers, peers) {
			break
		}

		peers = hopPeers[:min(len(hopPeers), dhtK)]
	}

	return peers[:min(len(peers), dhtK)], nil
}

func (d *SDHT) findValue(key string) ([]byte, []Peer, error) {
	//fmt.Println("findValue", marshalID(d.id), key)
	if val, err := d.store.Get(key); err == nil {
		fmt.Println(marshalID(d.id), "GOT THE VALUE FOR", key)
		fmt.Println("IT WAS", string(val))
		return val, nil, nil
	}
	fmt.Println(marshalID(d.id), "DID NOT GET THE VALUE FOR", key)
	peers, err := d.findNode(key)
	if err == nil {
		return nil, peers, nil
	}
	return nil, nil, err
}

func (d *SDHT) findNode(key string) ([]Peer, error) {
	//fmt.Println("findNode", marshalID(d.id), key)
	buckets := d.buckets.bs
	newBuckets := []Peer{}

	// TODO(pallavag): Remove unsafe unmarshals.
	keyID, _ := unmarshalID(key)
	for _, bs := range buckets {
		for _, b := range bs.peers {
			newBuckets = append(newBuckets, b)
		}
	}

	//fmt.Println("Here", newBuckets)
	sort.Slice(newBuckets, func(i, j int) bool {
		return isBetter(keyID, newBuckets[i], newBuckets[j])
	})

	return newBuckets[:min(len(newBuckets), dhtK)], nil
}

func (d *SDHT) setAlive(peer Peer) {
	d.alive[peer.ID] = int(time.Now().Unix())
	d.buckets.insert(d.id, peer)
}

func (d *SDHT) setAliveTime(id ID) {
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
