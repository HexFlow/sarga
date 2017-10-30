package sdht

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/sakshamsharma/sarga/common/iface"
	"github.com/sakshamsharma/sarga/impl/testnet"
)

const (
	dhtCount    = 5
	dataToStore = "hi-this*is*a#test#string"
)

func TestDHT(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	network := testnet.InitTestNet()

	rand.Seed(time.Now().UTC().UnixNano())

	nodeDHT := SDHT{}
	addr := iface.Address{"0", 0}
	network.DHTs[addr] = &nodeDHT
	nodeDHT.Init(addr, []iface.Address{}, network)

	for i := 1; i < dhtCount; i++ {
		nodeDHT := SDHT{}
		addr := iface.Address{strconv.Itoa(i), 0}
		network.DHTs[addr] = &nodeDHT
		nodeDHT.Init(addr, []iface.Address{{strconv.Itoa(rand.Intn(i)), 0}}, network)
	}

	fmt.Println("**** INIT FINISHED **")

	ii := marshalID(genID())

	err := nodeDHT.StoreValue(ii, []byte(dataToStore))
	if err != nil {
		t.Fatalf("error while storing file in DHT: %v", err)
	}

	reqData := findValueReq{
		ID:  nodeDHT.id,
		Key: ii,
	}
	v, err := network.Post(iface.Address{strconv.Itoa(rand.Intn(dhtCount)), 0}, "find_value", marshal(reqData))
	if err != nil {
		t.Fatalf("error while fetching file from DHT: %v", err)
	}
	resp := findValueResp{}
	if err := json.Unmarshal(v, &resp); err != nil {
		t.Fatalf("invalid JSON received as response to find_value: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("node returned error in response to find_value: %v", resp.Error)
	}

	if string(resp.Data) != dataToStore {
		t.Fatalf("invalid data receieved from DHT, expected %q, got %q", dataToStore, v)
	}

	fmt.Println("ASKING for info now")
	v, err = network.Get(iface.Address{"0", 0}, "info")
	if err != nil {
		t.Fatalf("error while fetching file from DHT: %v", err)
	}
	fmt.Println(string(v))
}
