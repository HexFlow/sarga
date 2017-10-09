package sdht

import "github.com/sakshamsharma/sarga/common/iface"

const k = 20

// ID is the representation of the type used for the key in the DHT.
type ID string

// Node represents each running peer in sarga, with the ID and address.
type Node struct {
	ID   ID
	Addr iface.Address
}

// bucket struct handles the list of nodes stored in each bucket.
type bucket struct {
	nodes []Node
}

func (b *bucket) insert(node Node) {
	b.nodes = append(b.nodes, node)
}

// buckets is the underlying struct which handles the creation and deletion of buckets.
type buckets struct {
	bs []bucket
}

func (b *buckets) insert(node Node) {
}
