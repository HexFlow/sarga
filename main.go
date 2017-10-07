package main

import (
	"fmt"
	"log"

	"github.com/alexflint/go-arg"

	"github.com/sakshamsharma/sarga/apiserver"
	"github.com/sakshamsharma/sarga/common/iface"
)

func main() {
	var runType iface.CommonArgs
	arg.Parse(&runType)

	var err error
	switch runType.Type {
	case "":
		fmt.Printf("missing --type=<TYPE>, please select a type among server, daemon, proxy, CLI\n")
	case "server":
		err = apiserver.Init()
	default:
		fmt.Printf("invalid type used to initialize sargo: %q\n", runType.Type)
	}

	if err != nil {
		log.Fatal(err)
	}
}
