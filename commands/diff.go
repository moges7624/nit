package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moges7624/nit/internal/index"
	"github.com/moges7624/nit/internal/objects"
	"github.com/moges7624/nit/internal/repo"
	"github.com/moges7624/nit/internal/status"
	"github.com/moges7624/nit/lib/diff"
)

func Diff(args []string) (string, error) {
	repo, err := repo.Open(".")
	if err != nil {
		return "", fmt.Errorf("error opening repo: %w", err)
	}

	index := index.NewIndex(filepath.Join(repo.NitPath(), "index"))
	if err = index.Load(); err != nil {
		return "", fmt.Errorf("error loading index: %w", err)
	}

	status, err := status.GetStatus(repo, index)
	if err != nil {
		return "", fmt.Errorf("error getting status: %w", err)
	}

	var buf bytes.Buffer
	for file := range status.Modified {
		fmt.Fprintf(&buf, "diff --nit a/%s b/%[1]s\n", file)
		fileHash := index.Entries[file].ObjHash

		blobObj, err := objects.Read(repo, fileHash)
		if err != nil {
			return "", err
		}

		blob, ok := blobObj.(*objects.Blob)
		if !ok {
			return "", fmt.Errorf("invalid blob object")
		}

		currFilepath := filepath.Join(repo.WorkTreePath(), file)
		currFile, err := os.ReadFile(currFilepath)
		if err != nil {
			return "", err
		}

		curBlob := objects.NewBlob(currFile)

		diff.PrintDiffContent(blob, curBlob, file, index.Entries[file].Mode)
	}

	return buf.String(), nil
}
