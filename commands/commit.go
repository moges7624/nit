package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moges7624/nit/objects"
	"github.com/moges7624/nit/refs"
	"github.com/moges7624/nit/repo"
)

func Commit(args []string) {
	if len(args) < 1 || args[0] != "-m" || args[1] == "" {
		fmt.Fprintf(os.Stderr, "Usage: nit commit -m <message>\n")
		return
	}

	message := args[1]

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

	treeHash, err := objects.Store(repo, tree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing tree to a disk: %v", err.Error())
		return
	}

	commit := &objects.Commit{
		Tree:      treeHash,
		Author:    "john <john@mail.com>",
		Committer: "john <john@mail.com>",
		Message:   message,
	}

	commitHash, err := objects.Store(repo, commit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing commit to a disk: %v", err.Error())
		return
	}

	ref := refs.NewRef(filepath.Join(wd, ".git/"))
	err = ref.UpdateHead(commitHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error updating head: %v", err.Error())
		return
	}

	fmt.Printf("[main %x] %s\n", commitHash[:7], message)
}
