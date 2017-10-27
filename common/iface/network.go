package iface

import "strconv"

type Proto int

const (
	TCP Proto = iota
	UDP
	HTTP
)

func (p *Proto) ToString() string {
	switch *p {
	case TCP:
		return "tcp"
	case UDP:
		return "udp"
	case HTTP:
		return "http"
	default:
		return "unknown"
	}
}

type Address struct {
	IP   string
	Port int
}

func (a *Address) ToString() string {
	return a.IP + ":" + strconv.Itoa(a.Port)
}

type Net interface {
	Get(addr Address, path string) ([]byte, error)
	Put(addr Address, path string, data []byte) error
	Post(addr Address, path string, data []byte) ([]byte, error)
	Listen(addr Address, handler func(string, []byte) []byte, shutdown chan bool) error
}
