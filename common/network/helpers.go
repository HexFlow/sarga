package network

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sakshamsharma/sarga/common"
)

func GetAddress(ip string, port int, protocol Protocol) Address {
	return Address{
		IP:       ip,
		Port:     port,
		Protocol: protocol,
	}
}

func ParseAddress(addr string, protocol Protocol) (Address, error) {
	chunks := strings.SplitN(addr, ":", 2)

	var ip = chunks[0]

	var port int
	if len(chunks) == 1 {
		port = common.DefaultPort
	} else {
		var err error
		port, err = strconv.Atoi(chunks[1])
		if err != nil {
			return Address{}, fmt.Errorf("error while parsing port number %q: %v", chunks[1], err)
		}
	}

	return Address{
		IP:       ip,
		Port:     port,
		Protocol: protocol,
	}, nil
}

func ParseAddresses(addrs []string, protocol Protocol) ([]Address, error) {
	result := []Address{}
	for _, addr := range addrs {
		parsed, err := ParseAddress(addr, protocol)
		if err != nil {
			return nil, err
		}
		result = append(result, parsed)
	}
	return result, nil
}
