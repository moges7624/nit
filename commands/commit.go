package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moges7624/nit/index"
	"github.com/moges7624/nit/objects"
	"github.com/moges7624/nit/refs"
	"github.com/moges7624/nit/repo"
)

func Commit(args []string) error {
	if len(args) < 1 || args[0] != "-m" || args[1] == "" {
		return fmt.Errorf("Usage: nit commit -m <message>\n")
	}

	message := args[1]

	repo, err := repo.Open(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return fmt.Errorf("error opening repo: %s", err.Error())
	}

	index := index.NewIndex(filepath.Join(repo.NitPath(), "index"))
	if err = index.Load(); err != nil {
		return fmt.Errorf("error loading index: %s", err.Error())
	}

	if len(index.Entries) == 0 {
		fmt.Println("nothing to commit, working tree clean")
		return nil
	}

	treeHash, err := objects.BuildFromIndex(repo, *index)
	if err != nil {
		return fmt.Errorf("error building tree from index: %s", err.Error())
	}

	commit := objects.NewCommit(
		treeHash,
		"john <john@mail.com>",
		message,
	)

	ref := refs.NewRef(repo.NitPath())
	par, err := ref.GetHeadCommit()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error getting head commit: %s", err.Error())
	}

	if par != "" {
		commit.SetParent(par)
	}

	commitHash, err := objects.Store(repo, commit)
	if err != nil {
		return fmt.Errorf("error writing commit to a disk: %v", err.Error())
	}

	err = ref.UpdateHead(commitHash)
	if err != nil {
		return fmt.Errorf("error updating head: %v", err.Error())
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "[main ")

	if par == "" {
		fmt.Fprintf(&buf, "(root-commit) ")
	}

	fmt.Fprintf(&buf, "%s] %s", commitHash[:7], message)

	fmt.Println(buf.String())

	return nil
}
