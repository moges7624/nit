package objects

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
)

type Entry struct {
	Name string
	Hash string
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
		fmt.Fprintf(&buf, "100644 %s\x00", entry.Name)
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
