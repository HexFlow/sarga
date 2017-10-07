package network

type Protocol int

const (
	TCP Protocol = 1 + iota
	UDP
	HTTP
)

type Address struct {
	IP       string
	Port     int
	Protocol Protocol
}

type Network interface {
	SendMessage(addr Address, data []byte) error
	Listen(addr Address) error
}
