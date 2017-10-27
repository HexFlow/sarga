package dht

import "github.com/sakshamsharma/sarga/common/iface"

// DHT is a common interface to be satisfied by
// all implementations to be used with sarga.
type DHT interface {
	Init(addr iface.Address, seeds []iface.Address, net iface.Net) error
	FindValue(key string) ([]byte, error)
	StoreValue(key string, data []byte) error
	Shutdown()

	// Respond consumes a path and data, and returns the serialized response.
	// Helpful for unit tests.
	Respond(string, []byte) []byte
}
