package sdht

import (
	"encoding/json"

	"github.com/sakshamsharma/sarga/common/iface"
)

// Peer wraps interactions with the peers of a DHT.
type Peer struct {
	ID   ID
	Addr iface.Address
}

func (p *Peer) Ping() error {
	id, err := network.Get(p.addr, "ping")
	if err != nil {
		return err
	}

	parsedID, err = unmarshalID(id)
	if err != nil {
		return err
	}

	p.id = parsedID
	return nil
}

func (p *Peer) SendStore(key string, data []byte) error {
	// TODO: Validate key
	keyValue := storeReq{p.id, key, string(data)}
	bytes, err := json.Marshal(keyValue)
	if err != nil {
		return err
	}
	return network.Put(p.addr, "store", bytes)
}

func (p *Peer) FindNode(id ID) ([]Peer, error) {
	req := findNodeReq{p.id, id}
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := network.Post(p.addr, "find_node", bytes)
	ret := findNodeResp{}
	if err := json.Unmarshal(resp, &ret); err != nil {
		return nil, err
	}
	if ret.Error != nil {
		return nil, ret.Error
	}
	return ret.Peers, nil
}

func (p *Peer) FindValue(key string) ([]byte, error) {
	req := findValueReq{p.id, key}
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := network.Post(p.addr, "find_value", bytes)
	ret := findValueResp{}
	if err := json.Unmarshal(resp, &ret); err != nil {
		return nil, err
	}
	if ret.Error != nil {
		return nil, ret.Error
	}
	return ret.Data, nil
}

func (p *Peer) AnnounceExit() error {
	req := exitReq{p.id}
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return network.Put(p.addr, "exit", bytes)
}
