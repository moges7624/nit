package status

import (
	"bytes"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/moges7624/nit/internal/index"
	"github.com/moges7624/nit/internal/objects"
	"github.com/moges7624/nit/internal/refs"
	"github.com/moges7624/nit/internal/repo"
)

type Status struct {
	Staged    map[string]string
	Modified  map[string]string
	Untracked []string
}

type FileStatus struct {
	name   string
	status string
}

func GetStatus(repo *repo.Repository, index *index.Index) (*Status, error) {
	headTreeHash, err := getHeadTreeHash(repo)
	if err != nil {
		return nil, fmt.Errorf("error getting head tree hash: %w", err)
	}

	stagedChanges := findStagedChanges(repo, index, headTreeHash)
	modifiedChanges := findModifiedFiles(repo, index)

	untrackedFiles, err := findUntrackedFiles(repo, index)
	if err != nil {
		return nil, err
	}

	status := &Status{
		Staged:    stagedChanges,
		Modified:  modifiedChanges,
		Untracked: untrackedFiles,
	}
	return status, nil
}

func findStagedChanges(repo *repo.Repository,
	idx *index.Index,
	headTreeHash string,
) map[string]string {
	changes := make(map[string]string)

	// no commit on the repo yet
	if headTreeHash == "" {
		for _, entry := range idx.Entries {
			changes[entry.Name] = "A"
		}

		return changes
	}

	headTreeObj, err := objects.Read(repo, headTreeHash)
	if err != nil {
		for _, entry := range idx.Entries {
			changes[entry.Name] = "A"
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
			changes[path] = "A"
		} else if idxHash != headHash {
			changes[path] = "M"
		}
	}

	for path := range headFiles {
		if _, exists := indexFiles[path]; !exists {
			changes[path] = "D"
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

		relPath, _ := filepath.Rel(repo.NitPath(), path)
		cleanRelPath := strings.TrimPrefix(relPath, "../")

		if info.IsDir() {
			if !hasStagedFile(idx, cleanRelPath) && cleanRelPath != ".." {
				untracked = append(untracked, cleanRelPath+"/")
				return filepath.SkipDir
			}
			return nil
		}

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

func findModifiedFiles(repo *repo.Repository, idx *index.Index) map[string]string {
	modified := make(map[string]string)

	for _, entry := range idx.Entries {
		fullpath := filepath.Join(repo.WorkTreePath(), entry.Name)

		fi, err := os.Stat(fullpath)
		if err != nil {
			if os.IsNotExist(err) {
				modified[entry.Name] = "D"
			}
			continue
		}

		if uint32(fi.Size()) != entry.Size ||
			uint32(fi.ModTime().Unix()) != entry.MTimeSec {
			// modified = append(modified, entry.Name)
			modified[entry.Name] = "M"
			continue
		}
	}

	return modified
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

		if entry.Mode == "40000" {
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

func hasStagedFile(idx *index.Index, dir string) bool {
	for name := range idx.Entries {
		if strings.HasPrefix(name, dir) {
			return true
		}
	}
	return false
}

func (s *Status) FormatStatus() string {
	var res bytes.Buffer

	statusMap := map[string]string{
		"M": "modified",
		"A": "new file",
		"D": "deleted",
	}

	fmt.Fprintln(&res, "On branch main")

	if len(s.Staged) == 0 &&
		len(s.Untracked) == 0 &&
		len(s.Modified) == 0 {
		fmt.Fprintln(&res, "nothing to commit, working tree clean")

		return res.String()
	}

	if len(s.Staged) > 0 {
		fmt.Fprintln(&res, "Changes to be committed:")
	}

	stagedChangesSlc := slices.Collect(maps.Keys(s.Staged))
	slices.Sort(stagedChangesSlc)

	for _, file := range stagedChangesSlc {
		fmt.Fprintf(&res, "\t\033[32m%s:   %s\033[0m\n", statusMap[s.Staged[file]], file)
	}

	if len(s.Modified) > 0 {
		fmt.Fprintln(&res, "\nChanges not staged for commit:")
		fmt.Fprintln(&res, `  (use "nit add <file>..." to update what will be committed)`)
	}

	modifiedIndexSlc := slices.Collect(maps.Keys(s.Modified))
	slices.Sort(modifiedIndexSlc)

	for _, file := range modifiedIndexSlc {
		fmt.Fprintf(&res, "\t\033[31m%s:   %s\033[0m\n", statusMap[s.Modified[file]], file)
	}

	if len(s.Untracked) > 0 {
		fmt.Fprintln(&res, "\nUntracked files:")
		fmt.Fprintln(&res, `  (use "nit add <file>..." to include in what will be committed)`)
	}

	for _, file := range s.Untracked {
		fmt.Fprintf(&res, "\t\033[31m%s\033[0m\n", file)
	}

	return res.String()
}

func (s Status) FormatStatusPorcelain() string {
	trackedChanges := make(map[string]string)

	maps.Copy(trackedChanges, s.Staged)

	for file, status := range s.Modified {
		if _, exists := trackedChanges[file]; exists {
			trackedChanges[file] = trackedChanges[file] + status
		} else {
			switch status {
			case "M":
				trackedChanges[file] = "unstagedM"
			case "D":
				trackedChanges[file] = "unstagedD"
			}
		}
	}

	var res bytes.Buffer
	alteredFiles := slices.Collect(maps.Keys(trackedChanges))

	slices.Sort(alteredFiles)

	for _, file := range alteredFiles {
		switch trackedChanges[file] {
		case "A":
			fmt.Fprintf(&res, "A  %s\n", file)
		case "M":
			fmt.Fprintf(&res, "M  %s\n", file)
		case "AM":
			fmt.Fprintf(&res, "AM %s\n", file)
		case "MM":
			fmt.Fprintf(&res, "MM %s\n", file)
		case "D":
			fmt.Fprintf(&res, "D %s\n", file)
		case "unstagedM":
			fmt.Fprintf(&res, " M %s\n", file)
		case "unstagedD":
			fmt.Fprintf(&res, " D %s\n", file)
		}
	}

	for _, file := range s.Untracked {
		fmt.Fprintf(&res, "?? %s\n", file)
	}

	return res.String()
}
