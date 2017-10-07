package net

import "github.com/sakshamsharma/sarga/common/network"

// net is a minimal implementation of a Network (network.Network) to be used with sarga.
type Net struct {
}

var _ network.Network = &Net{}

func (n *Net) SendMessage(addr network.Address, data []byte) error {
	return nil
}

func (n *Net) Listen(addr network.Address) error {
	return nil
}
