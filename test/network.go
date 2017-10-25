package test

import (
	"github.com/sakshamsharma/sarga/common/dht"
	"github.com/sakshamsharma/sarga/common/iface"
)

// TestNet is used for unit tests of DHT.
// TODO: Errors are not handled well yet.
type TestNet struct {
	dhts map[iface.Address]dht.DHT
}

var _ iface.Net = &TestNet{}

func (n *TestNet) Get(addr iface.Address, path string) ([]byte, error) {
	return n.dhts[addr].Respond(path, nil), nil
}

func (n *TestNet) Put(addr iface.Address, path string, data []byte) error {
	n.dhts[addr].Respond(path, data)
	return nil
}

func (n *TestNet) Post(addr iface.Address, path string, data []byte) ([]byte, error) {
	return n.dhts[addr].Respond(path, data), nil
}

// Listen simply blocks till shutdown. Since we control the network, we will directly
// call the member functions during the unit tests.
func (n *TestNet) Listen(_ iface.Address, _ func(string, []byte) []byte, shutdown chan bool) error {
	for {
		select {
		case <-shutdown:
			return nil
		}
	}
}
