package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moges7624/nit/index"
	"github.com/moges7624/nit/objects"
)

func Add(args []string) {
	if len(args) < 1 {
		fmt.Println("Nothing specified, nothing added.")
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting working directory :%v", err.Error())
		return
	}

	index := index.NewIndex(filepath.Join(wd, ".git/index"))
	err = index.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading index: %s", err.Error())
		return
	}

	for _, arg := range args {
		data, err := os.ReadFile(filepath.Join(wd, arg))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading file: %v", err.Error())
			return
		}

		blob := objects.NewBlob(data)
		blobHash, _ := blob.Hash()

		f, err := os.OpenFile(filepath.Join(wd, arg), os.O_RDONLY|os.O_EXCL, 0o444)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening file: %v", err.Error())
			return
		}
		defer f.Close()

		stat, _ := f.Stat()

		err = index.Add(arg, blobHash, stat)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error adding blob to index: %v", err.Error())
			return
		}
	}

	index.Write()
}
