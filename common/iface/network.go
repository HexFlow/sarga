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

type Net interface {
	Get(addr Address, path string) ([]byte, error)
	Put(addr Address, path string, data []byte) error
	Post(addr Address, path string, data []byte) ([]byte, error)
	SendMessage(addr Address, data []byte) error
	Listen(addr Address) error
}
