package main

import (
	"fmt"
	"os"

	"github.com/moges7624/nit/commands"
)

func main() {
	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {

	case "init":
		commands.Init(args)

	case "commit":
		commands.Commit(args)

	default:
		fmt.Println(cmd)
		fmt.Printf("nit: '%s' is not a nit command.\n", cmd)
	}
}
