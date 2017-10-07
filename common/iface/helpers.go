package iface

import (
	"fmt"
	"strconv"
	"strings"
)

func GetAddress(ip string, port int, proto Proto) Address {
	return Address{
		IP:    ip,
		Port:  port,
		Proto: proto,
	}
}

func ParseAddress(addr string, proto Proto) (Address, error) {
	chunks := strings.SplitN(addr, ":", 2)

	var ip = chunks[0]

	var port int
	if len(chunks) == 1 {
		port = DefaultPort
	} else {
		var err error
		port, err = strconv.Atoi(chunks[1])
		if err != nil {
			return Address{}, fmt.Errorf("error while parsing port number %q: %v", chunks[1], err)
		}
	}

	return Address{
		IP:    ip,
		Port:  port,
		Proto: proto,
	}, nil
}

func ParseAddresses(addrs []string, proto Proto) ([]Address, error) {
	result := []Address{}
	for _, addr := range addrs {
		parsed, err := ParseAddress(addr, proto)
		if err != nil {
			return nil, err
		}
		result = append(result, parsed)
	}
	return result, nil
}
