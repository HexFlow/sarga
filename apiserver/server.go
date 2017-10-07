package apiserver

import (
	"fmt"

	"github.com/sakshamsharma/sarga/common/dht"
	"github.com/sakshamsharma/sarga/common/network"
)

func StartAPIServer(args ServerArgs, dht dht.DHT, net network.Network) {
	err := net.Listen(network.GetAddress(args.IP, args.Port, args.Protocol))
	if err != nil {
		fmt.Println(err)
	}
}
