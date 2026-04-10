package objects

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/moges7624/nit/index"
	"github.com/moges7624/nit/repo"
)

type FileMode string

var (
	Directory  FileMode = "40000"
	Executable FileMode = "100755"
	Regular    FileMode = "100644"
)

type Entry struct {
	Name string
	Hash string
	Mode FileMode
}

type Tree struct {
	Entries []Entry
	hash    string
}

func (t *Tree) Type() string {
	return "tree"
}

func (t *Tree) Serialize() ([]byte, error) {
	sort.Slice(t.Entries, func(i, j int) bool {
		return t.Entries[i].Name < t.Entries[j].Name
	})

	var buf bytes.Buffer
	for _, entry := range t.Entries {
		fmt.Fprintf(&buf, "%s %s\x00", entry.Mode, entry.Name)
		hashBytes, _ := hex.DecodeString(entry.Hash)
		buf.Write(hashBytes)
	}

	return buf.Bytes(), nil
}

func (t *Tree) Hash() (string, error) {
	if t.hash != "" {
		return t.hash, nil
	}

	data, _ := t.Serialize()
	header := fmt.Sprintf("tree %d\x00", len(data))
	fullContent := append([]byte(header), data...)

	hash := sha1.Sum(fullContent)
	hashHex := fmt.Sprintf("%x", hash)

	t.hash = hashHex
	return hashHex, nil
}

func BuildFromIndex(repo *repo.Repository, idx index.Index) (string, error) {
	if len(idx.Entries) == 0 {
		emptyTree := &Tree{Entries: []Entry{}}
		return Store(repo, emptyTree)
	}

	idxEntries := idx.Entries
	treeMap := make(map[string][]Entry)

	for _, idxEntry := range idxEntries {
		dir, name := splitPath(idxEntry.Name)

		treeEntry := Entry{
			Name: name,
			Hash: idxEntry.ObjHash,
			Mode: FileMode(fmt.Sprintf("%06o", idxEntry.Mode)),
		}

		treeMap[dir] = append(treeMap[dir], treeEntry)
	}

	return buildTreeRecursive(repo, "", treeMap)
}

// getChildDirs returns list of direct child directories
// ex.
// treeMap = [
//
//		"": [file.txt],
//		"internal" : [readme.md]
//		"internal/utils" : [help.go]
//		"src" : [main.go]
//	]
//
// getChildDirs(treeMap, "") => [internal, src]
// getChildDirs(treeMap, "internal") => [utils]
// getChildDirs(treeMap, "src") => []
func getChildDirs(treeMap map[string][]Entry, dir string) []string {
	seen := make(map[string]bool)

	prefix := dir
	if prefix != "" {
		prefix += "/"
	}

	for k := range treeMap {
		if !strings.HasPrefix(k, prefix) {
			continue
		}

		rest := strings.TrimPrefix(k, prefix)
		if rest == "" {
			continue
		}

		parts := strings.Split(rest, "/")
		child := parts[0]

		fullPath := child
		if dir != "" {
			fullPath = dir + "/" + child
		}

		seen[fullPath] = true
	}

	var result []string
	for k := range seen {
		result = append(result, k)
	}

	return result
}

func buildTreeRecursive(
	repo *repo.Repository,
	dir string,
	treeMap map[string][]Entry,
) (string, error) {
	entries := treeMap[dir]

	children := getChildDirs(treeMap, dir)
	for _, child := range children {
		hash, err := buildTreeRecursive(repo, child, treeMap)
		if err != nil {
			return "", err
		}

		entries = append(entries, Entry{
			Mode: Directory,
			Name: filepath.Base(child),
			Hash: hash,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	tree := &Tree{Entries: entries}
	return Store(repo, tree)
}

func splitPath(fullPath string) (dir, name string) {
	lastSlashIdx := strings.LastIndex(fullPath, "/")
	if lastSlashIdx == -1 {
		return "", fullPath
	}

	return fullPath[:lastSlashIdx], fullPath[lastSlashIdx+1:]
}
