package test

import (
	"encoding/hex"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/sakshamsharma/sarga/common/iface"
	"github.com/sakshamsharma/sarga/impl/sdht"
)

// genID generates a 160 bit random ID for the node.
func genID() sdht.ID {
	ret := sdht.ID{}
	for i := range ret {
		ret[i] = byte(rand.Int() % 256)
	}
	return ret
}

func marshalID(id sdht.ID) string {
	return hex.EncodeToString(id[:])
}

const (
	dhtCount    = 6
	dataToStore = "hi-this*is*a#test#string"
)

func TestDHT(t *testing.T) {
	network := InitTestNet()

	rand.Seed(time.Now().UTC().UnixNano())

	nodeDHT := sdht.SDHT{}
	addr := iface.Address{"0", 0}
	network.dhts[addr] = &nodeDHT
	nodeDHT.Init(addr, []iface.Address{}, network)

	for i := 1; i < dhtCount; i++ {
		nodeDHT := sdht.SDHT{}
		addr := iface.Address{strconv.Itoa(i), 0}
		network.dhts[addr] = &nodeDHT
		nodeDHT.Init(addr, []iface.Address{{strconv.Itoa(rand.Intn(i)), 0}}, network)
	}

	ii := marshalID(genID())

	err := nodeDHT.StoreValue(ii, []byte(dataToStore))
	if err != nil {
		t.Fatalf("error while storing file in DHT: %v", err)
	}

	v, err := nodeDHT.FindValue(ii)
	if err != nil {
		t.Fatalf("error while fetching file from DHT: %v", err)
	}

	if string(v) != dataToStore {
		t.Fatalf("invalid data receieved from DHT, expected %q, got %q", dataToStore, v)
	}
}
