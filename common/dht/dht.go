package dht

import (
	"fmt"

	"github.com/sakshamsharma/sarga/common/iface"
)

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

type FakeDHT struct {
	data map[string][]byte
}

var _ DHT = &FakeDHT{}

func (f *FakeDHT) Init(addr iface.Address, seeds []iface.Address, net iface.Net) error {
	f.data = map[string][]byte{}
	return nil
}

func (f *FakeDHT) FindValue(key string) ([]byte, error) {
	if val, ok := f.data[key]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("Key %q not found in FakeDHT", key)
}

func (f *FakeDHT) StoreValue(key string, data []byte) error {
	f.data[key] = data
	return nil
}

func (f *FakeDHT) Shutdown() {}

func (f *FakeDHT) Respond(string, []byte) []byte {
	return nil
}
