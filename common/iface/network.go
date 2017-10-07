package iface

type Proto int

const (
	TCP Proto = iota
	UDP
	HTTP
)

type Address struct {
	IP    string
	Port  int
	Proto Proto
}

type Network interface {
	SendMessage(addr Address, data []byte) error
	Listen(addr Address) error
}
