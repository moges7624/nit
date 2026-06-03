package main

import (
	"fmt"
	"os"

	"github.com/moges7624/nit/commands"
)

func main() {
	if len(os.Args) == 1 {
		commands.Help()
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {

	case "help":
		commands.Help()

	case "init":
		if err := commands.Init(args); err != nil {
			fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		}

	case "add":
		commands.Add(args)

	case "commit":
		if err := commands.Commit(args); err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}

	case "status":
		res, err := commands.Status(args)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}

		fmt.Print(res)

	case "diff":
		res, err := commands.Diff(args)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}

		fmt.Print(res)

	default:
		fmt.Printf("nit: '%s' is not a nit command.\n", cmd)
	}
}
