package net

import (
	"github.com/sakshamsharma/sarga/common/iface"
)

// net is a minimal implementation of a Network (network.Network) to be used
// with sarga.
type Net struct {
}

var _ iface.Network = &Net{}

func (n *Net) SendMessage(addr iface.Address, data []byte) error {
	return nil
}

func (n *Net) Listen(addr iface.Address) error {
	return nil
}
