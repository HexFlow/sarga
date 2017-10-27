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
	resp, err := network.Get(p.Addr, "ping")
	if err != nil {
		return err
	}
	ret := pingResp{}
	if err := json.Unmarshal(resp, &ret); err != nil {
		return err
	}

	p.ID = ret.ID
	return nil
}

func (p *Peer) SendStore(key string, data []byte) error {
	// TODO: Validate key
	keyValue := storeReq{p.ID, key, string(data)}
	bytes, err := json.Marshal(keyValue)
	if err != nil {
		return err
	}
	// TODO: Errors will be ignored. Handle errors for PUT.
	return network.Put(p.Addr, "store", bytes)
}

// TODO(pallavag): Make all calls async.
func (p *Peer) FindNode(key string) ([]Peer, error) {
	req := findNodeReq{p.ID, key}
	bytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := network.Post(p.Addr, "find_node", bytes)
	if err != nil {
		return nil, err
	}
	ret := findNodeResp{}
	if err := json.Unmarshal(resp, &ret); err != nil {
		return nil, err
	}
	if ret.Error != nil {
		return nil, ret.Error
	}
	return ret.Peers, nil
}

func (p *Peer) FindValue(key string) ([]byte, []Peer, error) {
	req := findValueReq{p.ID, key}
	bytes, err := json.Marshal(req)
	if err != nil {
		return nil, nil, err
	}

	resp, err := network.Post(p.Addr, "find_value", bytes)
	if err != nil {
		return nil, nil, err
	}
	ret := findValueResp{}
	if err := json.Unmarshal(resp, &ret); err != nil {
		return nil, nil, err
	}
	if ret.Error != nil {
		return nil, nil, ret.Error
	}
	return ret.Data, ret.Peers, nil
}

func (p *Peer) AnnounceExit() error {
	req := exitReq{p.ID}
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return network.Put(p.Addr, "exit", bytes)
}
