package apiserver

import (
	"fmt"

	arg "github.com/alexflint/go-arg"
	"github.com/sakshamsharma/sarga/common/cli"
	"github.com/sakshamsharma/sarga/common/network"
	"github.com/sakshamsharma/sarga/net"
	"github.com/sakshamsharma/sarga/sdht"
)

type ServerArgs struct {
	cli.CommonArgs

	Seeds []string
}

func Init() {
	var args ServerArgs
	arg.MustParse(&args)

	for _, seed := range args.Seeds {
		fmt.Println(seed)
	}

	seeds, err := network.ParseAddresses(args.Seeds, network.UDP)
	if err != nil {
		fmt.Println(err)
		return
	}

	dht := &sdht.SDHT{}
	err = dht.Init(seeds)
	if err != nil {
		fmt.Println(err)
		return
	}

	StartAPIServer(args, dht, &net.Net{})
}
