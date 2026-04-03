package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moges7624/nit/objects"
	"github.com/moges7624/nit/refs"
	"github.com/moges7624/nit/repo"
)

var dirIgnoreList = map[string]bool{
	".git": true,
	".":    true,
	"..":   true,
}

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

	treeHash, _ := commitDir(*repo, repo.WorkDir)

	// emtpy dir
	if treeHash == "" {
		fmt.Println("nothing to commit")
		return
	}

	commit := objects.NewCommit(
		treeHash,
		"john <john@mail.com>",
		message,
	)

	ref := refs.NewRef(filepath.Join(wd, ".git/"))
	par, err := ref.GetHeadCommit()
	if err != nil {
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

	fmt.Fprintf(&buf, "%s] %s\n", commitHash[:7], message)

	fmt.Println(buf.String())
}

func commitBlob(repo *repo.Repository, data []byte) (string, error) {
	blob := objects.NewBlob(data)
	hash, err := objects.Store(repo, blob)
	if err != nil {
		return "", fmt.Errorf("error writing blob to a disk: %s", err.Error())
	}

	return hash, nil
}

func commitDir(repo repo.Repository, dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("error reading dir files: %s", err.Error())
	}

	var treeEntries []objects.Entry
	for _, file := range files {
		if file.IsDir() {
			if dirIgnoreList[file.Name()] {
				continue
			}

			treeHash, err := commitDir(repo, filepath.Join(dir, file.Name()))
			if err != nil {
				return "", fmt.Errorf("error commiting directory: %s", err.Error())
			}

			// emtpy dir
			if treeHash == "" {
				continue
			}

			treeEntries = append(treeEntries, objects.Entry{
				Name: file.Name(),
				Hash: treeHash,
				Mode: objects.Directory,
			})

		} else {

			data, err := os.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading a file: %v", err.Error())
				return "", fmt.Errorf("error reading file: %s", err.Error())
			}

			blobHash, err := commitBlob(&repo, data)
			if err != nil {
				return "", err
			}

			fm := objects.Regular

			fi, err := file.Info()
			if err != nil {
				return "", fmt.Errorf("error getting file info: %s", err.Error())
			}

			if fi.Mode()&0o111 != 0 {
				fm = objects.Executable
			}

			treeEntries = append(treeEntries, objects.Entry{
				Name: file.Name(),
				Hash: blobHash,
				Mode: fm,
			})
		}
	}

	// emtpy directory
	if len(treeEntries) == 0 {
		return "", nil
	}

	tree := &objects.Tree{
		Entries: treeEntries,
	}

	treeHash, err := objects.Store(&repo, tree)
	if err != nil {
		return "", fmt.Errorf("error writing tree to disk: %s", err.Error())
	}

	return treeHash, nil
}
