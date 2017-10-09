package sdht

import (
	"fmt"
	"log"
	"net"

	"github.com/sakshamsharma/sarga/common/iface"
)

// SDHT is a minimal implementation of a DHT (dht.DHT) to be used with sarga.
type SDHT struct {
	peers []Peer

	buckets buckets

	listen net.Listener

	shutdown chan bool
}

func (d *SDHT) Init(id ID, seeds []iface.Address) error {
	for _, seed := range seeds {
		root := Peer{seed}
		if err := root.Ping(); err != nil {
			continue
		}

		nodes, err := root.FindNode(id)
		if err != nil {
			continue
		}

		for _, node := range nodes {
			d.buckets.insert(node)
		}

		// TODO: Somehow insert the root node into buckets as well.
		// We do not knows its ID.

		d.shutdown = make(chan bool)
		go d.serve()
		return nil
	}
	return fmt.Errorf("no provided seed completed initial connection")
}

func (d *SDHT) FindValue(key string) ([]byte, error) {
	return nil, nil
}

func (d *SDHT) StoreValue(key string, data []byte) error {
	return nil
}

func (d *SDHT) ShutDown() {
	if d.listen == nil {
		return
	}
	d.shutdown <- true
	d.listen.Close()
}

func (d *SDHT) serve() {
	listen, err := net.Listen("tcp", "0.0.0.0:6779")
	if err != nil {
		log.Fatalf("Failed to open listening socket: %s", err)
	}
	d.listen = listen

	for {
		conn, err := d.listen.Accept()
		if err != nil {
			select {
			case <-d.shutdown:
				return
			default:
				log.Printf("Accept failed: %v", err)
			}
		}
		err = d.respond(conn)
		if err != nil {
			log.Printf("Responding to connection failed: %v", err)
		}
	}
}

func (d *SDHT) respond(conn net.Conn) error {
	return nil
}
