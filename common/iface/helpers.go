package iface

import (
	"fmt"
	"strconv"
	"strings"
)

func GetAddress(ip string, port int) Address {
	return Address{
		IP:   ip,
		Port: port,
	}
}

func ParseAddress(addr string) (Address, error) {
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
		IP:   ip,
		Port: port,
	}, nil
}

func ParseAddresses(addrs []string) ([]Address, error) {
	result := []Address{}
	for _, addr := range addrs {
		parsed, err := ParseAddress(addr)
		if err != nil {
			return nil, err
		}
		result = append(result, parsed)
	}
	return result, nil
}
