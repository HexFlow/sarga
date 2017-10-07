package main

import (
	"fmt"

	"github.com/alexflint/go-arg"

	"github.com/sakshamsharma/sarga/apiserver"
	"github.com/sakshamsharma/sarga/common/iface"
)

func main() {
	var runType iface.CommonArgs
	arg.Parse(&runType)

	switch runType.Type {
	case "":
		fmt.Printf("missing --type=<TYPE>, please select a type among daemon, proxy, CLI\n")
	case "server":
		apiserver.Init()
	default:
		fmt.Printf("invalid type used to initialize sargo: %q\n", runType.Type)
	}
}
