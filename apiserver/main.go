package apiserver

import (
	"fmt"

	arg "github.com/alexflint/go-arg"
	"github.com/sakshamsharma/sarga/common/iface"
	"github.com/sakshamsharma/sarga/impl/sdht"
)

type ServerArgs struct {
	iface.CommonArgs

	Seeds []string
}

func Init() error {
	var args ServerArgs
	arg.MustParse(&args)

	for _, seed := range args.Seeds {
		fmt.Println(seed)
	}

	seeds, err := iface.ParseAddresses(args.Seeds, iface.UDP)
	if err != nil {
		return err
	}

	dhtInst := &sdht.SDHT{}
	if err = dhtInst.Init(seeds); err != nil {
		return err
	}
	StartAPIServer(args, dhtInst)

	return nil
}
