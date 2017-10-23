package dht

import "github.com/sakshamsharma/sarga/common/iface"

// DHT is a common interface to be satisfied by
// all implementations to be used with sarga.
type DHT interface {
	Init(seeds []iface.Address, net iface.Net) error
	FindValue(key string) ([]byte, error)
	StoreValue(key string, data []byte) error
	Shutdown()
}
