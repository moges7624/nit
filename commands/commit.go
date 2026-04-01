package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moges7624/nit/objects"
	"github.com/moges7624/nit/repo"
)

func Commit(args []string) {
	wd, err := os.Getwd()
	if err != nil {
		return
	}

	repo := repo.NewRepository(wd)

	files, err := repo.ListFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting list of files: %v", err.Error())
		return
	}

	var treeEntries []objects.Entry
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(repo.NitDir, file.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading a file: %v", err.Error())
			return
		}

		blob := objects.NewBlob(data)
		hash, err := objects.Store(repo, blob)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating file: %v", err.Error())
			return
		}

		treeEntries = append(treeEntries, objects.Entry{
			Name: file.Name(),
			Hash: hash,
		})
	}

	tree := &objects.Tree{
		Entries: treeEntries,
	}

	_, err = objects.Store(repo, tree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing tree to a disk: %v", err.Error())
		return
	}
}
