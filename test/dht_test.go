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

// genId generates a 160 bit random ID for the node.
func genId() sdht.ID {
	ret := sdht.ID{}
	for i, _ := range ret {
		ret[i] = byte(rand.Int() % 256)
	}
	return ret
}

func marshalID(id sdht.ID) string {
	return hex.EncodeToString(id[:])
}

const dhtCount = 200

func TestDHT(t *testing.T) {
	network := InitTestNet()

	rand.Seed(time.Now().UTC().UnixNano())

	nodeDHT := sdht.SDHT{}
	addr := iface.Address{"0", 0}
	nodeDHT.Init(addr, []iface.Address{}, network)
	network.dhts[addr] = &nodeDHT

	for i := 1; i < dhtCount; i++ {
		nodeDHT := sdht.SDHT{}
		addr := iface.Address{strconv.Itoa(i), 0}
		nodeDHT.Init(addr, []iface.Address{{strconv.Itoa(rand.Intn(i)), 0}}, network)
		network.dhts[addr] = &nodeDHT
	}

	ii := marshalID(genId())
	nodeDHT.StoreValue(ii, []byte("hi"))
	v, err := nodeDHT.FindValue(ii)
	if err != nil {
		t.Fatal(err)
	}
	if string(v) != "hi" {
		t.Fatal("Nope", string(v))
	}
	t.Fatal("COOL")
}
