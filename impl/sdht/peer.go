package sdht

import "github.com/sakshamsharma/sarga/common/iface"

// Peer wraps interactions with the peers of a DHT.
type Peer struct {
	Addr iface.Address
}

func (p *Peer) Ping() error {
	return nil
}

func (p *Peer) SendStore(key string, data []byte) error {
	return nil
}

func (p *Peer) FindNode(id ID) ([]Node, error) {
	return nil, nil
}

func (p *Peer) FindValue(key string) ([]byte, error) {
	return nil, nil
}

func (p *Peer) AnnounceExit() error {
	return nil
}
