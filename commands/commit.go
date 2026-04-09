package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/moges7624/nit/index"
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
	index := index.NewIndex(filepath.Join(wd, ".git/index"))
	if err = index.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "error loading index: %s", err.Error())
		return
	}

	tree := objects.Tree{}
	for _, entry := range index.Entries {
		treeEntry := objects.Entry{
			Name: entry.Name,
			Hash: entry.ObjHash,
			Mode: objects.FileMode(strconv.FormatUint(uint64(entry.Mode), 8)),
		}

		tree.Entries = append(tree.Entries, treeEntry)
	}

	treeHash, err := objects.Store(repo, &tree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error storing tree: %s", err.Error())
		return
	}

	commit := objects.NewCommit(
		treeHash,
		"john <john@mail.com>",
		message,
	)

	ref := refs.NewRef(filepath.Join(wd, ".git/"))
	par, err := ref.GetHeadCommit()
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%v", err.Error())
		return
	}

	if par != "" {
		commit.SetParent(par)
	}

	commitHash, err := objects.Store(repo, commit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing commit to a disk: %v", err.Error())
		return
	}

	err = ref.UpdateHead(commitHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error updating head: %v", err.Error())
		return
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "[main ")

	if par == "" {
		fmt.Fprintf(&buf, "(root-commit) ")
	}

	fmt.Fprintf(&buf, "%s] %s", commitHash[:7], message)

	fmt.Println(buf.String())
}
