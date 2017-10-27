package apiserver

import (
	"fmt"

	arg "github.com/alexflint/go-arg"
	"github.com/sakshamsharma/sarga/common/iface"
	"github.com/sakshamsharma/sarga/impl/httpnet"
	"github.com/sakshamsharma/sarga/impl/sdht"
)

type ServerArgs struct {
	iface.CommonArgs

	Seeds []string
}

func Init() error {
	var args ServerArgs
	arg.MustParse(&args)

	if args.Port == 0 {
		return fmt.Errorf("port not provided. Please provide a port using --port=<integer>")
	}

	if args.IP == "" {
		args.IP = "127.0.0.1"
	}

	for _, seed := range args.Seeds {
		fmt.Println(seed)
	}

	seeds, err := iface.ParseAddresses(args.Seeds)
	if err != nil {
		return err
	}

	dhtInst := &sdht.SDHT{}
	if err = dhtInst.Init(iface.Address{"0.0.0.0", 8080},
		seeds, &httpnet.HTTPNet{}); err != nil {
		return err
	}
	StartAPIServer(args.CommonArgs, dhtInst)

	return nil
}
