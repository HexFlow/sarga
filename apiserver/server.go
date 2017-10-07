package apiserver

import (
	"fmt"

	"github.com/sakshamsharma/sarga/common/dht"
	"github.com/sakshamsharma/sarga/common/iface"
)

func StartAPIServer(args ServerArgs, dht dht.DHT, netw iface.Network) {
	err := netw.Listen(iface.GetAddress(args.IP, args.Port, args.Proto))
	if err != nil {
		fmt.Println(err)
	}
}
