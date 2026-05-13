package commands

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/moges7624/nit/index"
	"github.com/moges7624/nit/repo"
)

// TODO: if files in a given directory are not tracked, the directory
// should be listed as untracked rather than listing each files in
// the directory as untracked
func Status() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting working directory: %w", err)
	}

	repo := repo.NewRepository(wd)
	index := index.NewIndex(filepath.Join(repo.NitPath(), "index"))

	if err = index.Load(); err != nil {
		return "", fmt.Errorf("error loading index: %w", err)
	}

	var res bytes.Buffer

	stagedChanges := findStagedChanges(index)
	for _, file := range stagedChanges {
		fmt.Fprintf(&res, "A %s\n", file)
	}

	untrackedFiles, err := findUntrackedFiles(repo, index)
	if err != nil {
		return "", err
	}

	for _, file := range untrackedFiles {
		fmt.Fprintf(&res, "?? %s\n", file)
	}

	return res.String(), nil
}

func findStagedChanges(idx *index.Index) []string {
	var staged []string
	for _, entry := range idx.Entries {
		staged = append(staged, entry.Name)
	}

	return staged
}

func findUntrackedFiles(repo *repo.Repository, idx *index.Index) ([]string, error) {
	var untracked []string

	err := filepath.Walk(repo.WorkTreePath(), func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(repo.NitPath(), path)
		cleanRelPath := strings.TrimPrefix(relPath, "../")

		if _, exists := idx.Entries[cleanRelPath]; exists {
			return nil
		}

		untracked = append(untracked, cleanRelPath)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return untracked, nil
}
