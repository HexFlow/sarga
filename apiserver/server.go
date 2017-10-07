package apiserver

import (
	"fmt"

	arg "github.com/alexflint/go-arg"
	"github.com/sakshamsharma/sarga/common/cli"
)

type ServerArgs struct {
	cli.CommonArgs

	Seeds []string
}

func StartServer() {
	var args ServerArgs
	arg.MustParse(&args)

	for _, seed := range args.Seeds {
		fmt.Println(peer)
	}
}
