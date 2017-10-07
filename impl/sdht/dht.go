package sdht

import (
	"github.com/sakshamsharma/sarga/common/dht"
	"github.com/sakshamsharma/sarga/common/iface"
)

// SDHT is a minimal implementation of a DHT (dht.DHT) to be used with sarga.
type SDHT struct {
}

var _ dht.DHT = &SDHT{}

func (d *SDHT) Init(seeds []iface.Address) error {
	return nil
}

func (d *SDHT) FindValue(key string) ([]byte, error) {
	return nil, nil
}

func (d *SDHT) StoreValue(key string, data []byte) error {
	return nil
}

func (d *SDHT) ShutDown() error {
	return nil
}
