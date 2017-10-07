package main

import (
	"fmt"

	"github.com/alexflint/go-arg"

	"github.com/sakshamsharma/sarga/apiserver"
	"github.com/sakshamsharma/sarga/common/cli"
)

func main() {
	var runType cli.CommonArgs
	arg.Parse(&runType)

	switch runType.Type {
	case "":
		fmt.Printf("missing --type=<TYPE>, please select a type among daemon, proxy, CLI\n")
	case "server":
		apiserver.StartServer()
	default:
		fmt.Printf("invalid type used to initialize sargo: %q\n", runType.Type)
	}
}
