package commands

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func Help() {
	entries := map[string]string{
		"help":   "Show this help message",
		"init":   "Create an empty Nit repository",
		"add":    "Add file contents to the index",
		"commit": "Record changes to the repository",
		"status": "Show the working tree status",
		"diff":   "Show changes between commit and working tree",
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	fmt.Fprintf(w, "These are common Nit commands used in various situations:\n\n")

	for k, v := range entries {
		fmt.Fprintf(w, "\t%s\t%s\n", k, v)
	}

	w.Flush()
}
