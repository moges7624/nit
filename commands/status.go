package commands

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/moges7624/nit/index"
	"github.com/moges7624/nit/objects"
	"github.com/moges7624/nit/refs"
	"github.com/moges7624/nit/repo"
)

type FileStatus struct {
	name   string
	status string
}

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

	var stagedChanges []FileStatus
	headTreeHash, err := getHeadTreeHash(repo)
	if err != nil {
		return "", fmt.Errorf("error getting head tree hash: %w", err)
	} else {
		stagedChanges = findStagedChanges(repo, index, headTreeHash)
	}

	for _, entry := range stagedChanges {
		if entry.status == "new" {
			fmt.Fprintf(&res, "A %s\n", entry.name)
		} else if entry.status == "modified" {
			fmt.Fprintf(&res, "AM %s\n", entry.name)
		}
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

func findStagedChanges(repo *repo.Repository,
	idx *index.Index,
	headTreeHash string,
) []FileStatus {
	var changes []FileStatus

	// no commit on the repo yet
	if headTreeHash == "" {
		for _, entry := range idx.Entries {
			changes = append(changes, FileStatus{
				name:   entry.Name,
				status: "new",
			})
		}

		return changes
	}

	headTreeObj, err := objects.Read(repo, headTreeHash)
	if err != nil {
		for _, entry := range idx.Entries {
			changes = append(changes, FileStatus{
				name:   entry.Name,
				status: "new",
			})
		}

		return changes
	}

	headTree, ok := headTreeObj.(*objects.Tree)
	if !ok {
		return changes
	}

	// convert head tree into flat map: path -> blob hash
	headFiles := make(map[string]string)
	flattenTree(repo, headTree, "", headFiles)

	// build index map
	indexFiles := make(map[string]string)
	for _, entry := range idx.Entries {
		indexFiles[entry.Name] = entry.ObjHash
	}

	for path, idxHash := range indexFiles {
		headHash, exists := headFiles[path]
		if !exists {
			changes = append(changes, FileStatus{
				name:   path,
				status: "new",
			})
		} else if idxHash != headHash {
			changes = append(changes, FileStatus{
				name:   path,
				status: "modified",
			})
		}
	}

	for path := range headFiles {
		if _, exists := indexFiles[path]; !exists {
			changes = append(changes, FileStatus{
				name:   path,
				status: "deleted",
			})
		}
	}

	return changes
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

func flattenTree(repo *repo.Repository,
	tree *objects.Tree,
	prefix string,
	result map[string]string,
) {
	for _, entry := range tree.Entries {
		fullPath := entry.Name
		if prefix != "" {
			fullPath = prefix + "/" + entry.Name
		}

		if entry.Mode == "040000" {
			subTreeObj, err := objects.Read(repo, entry.Hash)
			if err != nil {
				continue
			}

			if subTree, ok := subTreeObj.(*objects.Tree); ok {
				flattenTree(repo, subTree, fullPath, result)
			}
		} else {
			result[fullPath] = entry.Hash
		}
	}
}

func getHeadTreeHash(repo *repo.Repository) (string, error) {
	refs := refs.NewRef(repo.NitPath())
	headCommitHash, err := refs.GetHeadCommit()
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("error getting head commit hash: %w", err)
	}

	headCommitObj, err := objects.Read(repo, headCommitHash)
	if err != nil {
		return "", fmt.Errorf("error getting head commit: %w", err)
	}

	headCommit, ok := headCommitObj.(*objects.Commit)
	if !ok {
		return "", fmt.Errorf("invalid head commit object: %w", err)
	}

	headTreeObj, err := objects.Read(repo, headCommit.Tree)
	if !ok {
		return "", fmt.Errorf("error getting head tree obj: %w", err)
	}

	headTree, ok := headTreeObj.(*objects.Tree)
	if !ok {
		return "", fmt.Errorf("invalid head tree obj: %w", err)
	}

	headTreeHash, _ := headTree.Hash()

	return headTreeHash, nil
}
