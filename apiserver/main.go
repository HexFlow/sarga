package apiserver

import (
	"fmt"
	"math/rand"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/sakshamsharma/sarga/common/iface"
	"github.com/sakshamsharma/sarga/impl/httpnet"
	"github.com/sakshamsharma/sarga/impl/sdht"
	"github.com/sakshamsharma/sarga/impl/slog"
)

type ServerArgs struct {
	iface.CommonArgs

	Seeds          []string
	RandomDHTCount int

	DHTLogLevel string
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

	if args.DHTLogLevel != "" {
		sdht.SetLog(slog.GetLevelFromString(args.DHTLogLevel))
	}

	dhtInst := &sdht.SDHT{}
	if err = dhtInst.Init(iface.Address{IP: "0.0.0.0", Port: 8080},
		seeds, &httpnet.HTTPNet{}); err != nil {
		return err
	}

	if args.RandomDHTCount > 0 {
		time.Sleep(2 * time.Second)

		ports := []int{8080}

		for i := 1; i <= args.RandomDHTCount; i++ {
			nodeDHT := &sdht.SDHT{}
			addr := iface.Address{IP: "0.0.0.0", Port: rand.Intn(3000) + 4000}
			ports = append(ports, addr.Port)
			nodeDHT.Init(addr,
				[]iface.Address{{IP: "0.0.0.0", Port: ports[rand.Intn(i)]}},
				&httpnet.HTTPNet{})
		}
		time.Sleep(2 * time.Second)
	}

	StartAPIServer(args.CommonArgs, dhtInst)

	return nil
}
