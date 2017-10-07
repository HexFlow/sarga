package apiserver

import (
	"fmt"

	arg "github.com/alexflint/go-arg"
	"github.com/sakshamsharma/sarga/common/iface"
	"github.com/sakshamsharma/sarga/impl/net"
	"github.com/sakshamsharma/sarga/impl/sdht"
)

type ServerArgs struct {
	iface.CommonArgs

	Seeds []string
}

func Init() {
	var args ServerArgs
	arg.MustParse(&args)

	for _, seed := range args.Seeds {
		fmt.Println(seed)
	}

	seeds, err := iface.ParseAddresses(args.Seeds, iface.UDP)
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
