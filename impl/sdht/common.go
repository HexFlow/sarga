package sdht

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const dhtK = 20
const numBuckets = 160

// ID is the representation of the type used for the key in the DHT.
type ID [20]byte

func (id ID) String() string {
	return marshalID(id)
}

func unmarshalID(id string) (ID, error) {
	idDecoded, err := hex.DecodeString(id)
	if err != nil {
		return ID{}, err
	}
	if len(idDecoded) != 20 {
		return ID{}, fmt.Errorf(
			"invalid ID, expected length 20, got: %d", len(idDecoded))
	}

	result := ID{}
	for i, b := range idDecoded {
		result[i] = b
	}
	return result, nil
}

func marshalID(id ID) string {
	return hex.EncodeToString(id[:])
}

func marshal(data interface{}) []byte {
	bytes, _ := json.Marshal(data)
	return bytes
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// genId generates a 160 bit random ID for the node.
func genId() ID {
	ret := ID{}
	for i, _ := range ret {
		ret[i] = byte(rand.Int() % 256)
	}
	return ret
}

// bucket struct handles the list of nodes stored in each bucket.
type bucket map[ID]Peer

func (b *bucket) insert(node Peer) {
	if *b == nil {
		*b = make(bucket)
	}
	(*b)[node.ID] = node
}

func (b *bucket) del(id ID) {
	if *b == nil {
		*b = make(bucket)
	}
	delete(*b, id)
}

func (id ID) toBitString() string {
	arr := []string{}
	for _, b := range id {
		arr = append(arr, fmt.Sprintf("%.8b", b))
	}
	return strings.Join(arr, "")
}

// buckets is the underlying struct which handles the creation and deletion of buckets.
type buckets struct {
	bs    [numBuckets]bucket
	count int

	// Use a pointer so it is possible to check it against nil.
	replacement *Peer
}

func (b *buckets) insert(owner ID, node Peer) {
	obits := owner.toBitString()
	nbits := node.ID.toBitString()

	for i := 0; i < numBuckets; i++ {
		if nbits[i] != obits[i] {
			b.bs[i].insert(node)
			return
		}
	}
}

// TODO(pallavag): Add locks around non atomic operations.
func (b *buckets) replace(owner ID, id ID) {
	obits := owner.toBitString()
	ibits := id.toBitString()
	for i := 0; i < numBuckets; i++ {
		if ibits[i] != obits[i] {
			b.bs[i].del(id)
			break
		}
	}
	if b.replacement != nil {
		replacement := *b.replacement
		b.replacement = nil
		b.insert(owner, replacement)
	}
}
