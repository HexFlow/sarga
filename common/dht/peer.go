package dht

import "github.com/sakshamsharma/sarga/common/network"

// DHTPeer wraps interactions with the peers of a DHT.
type DHTPeer struct {
	Addr network.Address
}

func Ping() error {
	return nil
}

func SendStore(key string, data []byte) error {
	return nil
}

func FindNode() []network.Address {
	return nil
}

func FindValue(key string) ([]byte, error) {
	return nil, nil
}

func AnnounceExit() error {
	return nil
}
