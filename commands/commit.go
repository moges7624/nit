package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moges7624/nit/internal/index"
	"github.com/moges7624/nit/internal/objects"
	"github.com/moges7624/nit/internal/refs"
	"github.com/moges7624/nit/internal/repo"
	"github.com/moges7624/nit/internal/status"
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

	status, _ := status.GetStatus(repo, index)

	if len(status.Staged) == 0 {
		if len(status.Modified) > 0 {
			fmt.Println(status.FormatStatus())
			fmt.Println(`no changes added to commit (use "nit add")`)
			return nil
		}
		if len(status.Untracked) > 0 {
			fmt.Println(status.FormatStatus())
			fmt.Println(`nothing added to commit but untracked files present (use "nit add" to track)`)
			return nil
		}
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

	stat := status.Stat()

	fileTxt := "file"
	if len(status.Staged) > 1 {
		fileTxt = "files"
	}

	deletionTxt := "deletion"
	if stat.Deletions > 1 {
		deletionTxt = "deletions"
	}

	insertionTxt := "insertion"
	if stat.Insertions > 1 {
		deletionTxt = "insertions"
	}
	fmt.Fprintf(&buf, "%s] %s\n", commitHash[:7], message)
	fmt.Fprintf(&buf, " %d %s changed", stat.FilesChanged, fileTxt)

	if stat.Insertions > 0 {
		fmt.Fprintf(&buf, ", %d %s(+)", stat.Insertions, insertionTxt)
	}

	if stat.Deletions > 0 {
		fmt.Fprintf(&buf, ", %d %s(+)", stat.Deletions, deletionTxt)
	}

	fmt.Println(buf.String())

	return nil
}
