package sdht

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const dhtK = 20
const numBuckets = 160

// ID is the representation of the type used for the key in the DHT.
type ID [20]byte

func (id ID) String() string {
	return id.toBitString()[:10]
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

// genID generates a 160 bit random ID for the node.
func genID() ID {
	ret := ID{}
	for i := range ret {
		ret[i] = byte(rand.Int() % 256)
	}
	return ret
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (id ID) toBitString() string {
	arr := []string{}
	for _, b := range id {
		arr = append(arr, fmt.Sprintf("%.8b", b))
	}
	return strings.Join(arr, "")
}

// isBetter returns true if peer1 is closer to key than peer2.
func isBetter(key ID, peer1, peer2 Peer) bool {
	return dist(peer1.ID, key) < dist(peer2.ID, key)
}

func isBetterSlice(key ID, peer1, peer2 []Peer) bool {
	l := min(len(peer1), len(peer2))
	for i := 0; i < l; i++ {
		if dist(peer1[i].ID, key) < dist(peer2[i].ID, key) {
			return true
		}
		if dist(peer1[i].ID, key) > dist(peer2[i].ID, key) {
			return false
		}
	}
	return (len(peer1) > len(peer2)) && (len(peer2) < dhtK)
}

// dist returns the distance between two keys.
func dist(id1, id2 ID) int {
	id1bits := id1.toBitString()
	id2bits := id2.toBitString()
	for i := range id1bits {
		if id1bits[i] != id2bits[i] {
			return numBuckets - i
		}
	}
	return 0
}

// bucket struct handles the list of nodes stored in each bucket.
type bucket struct {
	peers       map[ID]Peer
	replacement *Peer

	lock *sync.RWMutex
}

func (b *bucket) insert(owner ID, node Peer) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(b.peers) < dhtK {
		if _, ok := b.peers[node.ID]; !ok {
			log.Println(owner, "added peer", node.ID)
			b.peers[node.ID] = node
		}
	} else {
		b.replacement = &node
	}
}

func (b *bucket) del(id ID) {
	b.lock.Lock()
	defer b.lock.Unlock()

	delete(b.peers, id)
}

func (b *bucket) replace(owner ID, id ID) {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.del(id)
	if b.replacement != nil {
		replacement := *b.replacement
		b.replacement = nil
		b.insert(owner, replacement)
	}
}

func (b *bucket) Marshal() string {
	b.lock.RLock()
	defer b.lock.RUnlock()

	tmpMap := map[string]string{}
	for key, val := range b.peers {
		tmpMap[marshalID(key)] = val.Addr.String()
	}
	return string(marshal(tmpMap))
}

// buckets is the underlying struct which handles the creation and deletion of buckets.
type buckets struct {
	bs   [numBuckets]bucket
	lock *sync.RWMutex
}

func (b *buckets) insert(owner ID, node Peer) {
	b.lock.Lock()
	defer b.lock.Unlock()

	obits := owner.toBitString()
	nbits := node.ID.toBitString()

	for i := 0; i < numBuckets; i++ {
		if nbits[i] != obits[i] {
			b.bs[i].insert(owner, node)
			return
		}
	}
}

// TODO(pallavag): Add locks around non atomic operations.
func (b *buckets) replace(owner ID, id ID) {
	b.lock.Lock()
	defer b.lock.Unlock()

	obits := owner.toBitString()
	ibits := id.toBitString()
	for i := 0; i < numBuckets; i++ {
		if ibits[i] != obits[i] {
			b.bs[i].replace(owner, id)
			break
		}
	}
}

func (b *buckets) Marshal() string {
	b.lock.RLock()
	defer b.lock.RUnlock()

	tmpList := []string{}
	for _, buck := range b.bs {
		v := buck.Marshal()
		if v != "" && v != "{}" {
			tmpList = append(tmpList, buck.Marshal())
		}
	}
	return string(marshal(tmpList))
}

func initBuckets() buckets {
	bs := [numBuckets]bucket{}
	for i := 0; i < numBuckets; i++ {
		bs[i] = bucket{
			lock:  &sync.RWMutex{},
			peers: make(map[ID]Peer),
		}
	}
	return buckets{
		lock: &sync.RWMutex{},
		bs:   bs,
	}
}
