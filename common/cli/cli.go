package cli

import "github.com/sakshamsharma/sarga/common/network"

type CommonArgs struct {
	Type string

	Port     int
	IP       string
	Protocol network.Protocol
}
