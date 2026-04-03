package objects

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"time"
)

type Commit struct {
	Tree      string
	parent    string
	Author    string
	Committer string
	Message   string
	hash      string
}

func NewCommit(treeHash, author, message string) *Commit {
	return &Commit{
		Tree:      treeHash,
		Author:    author,
		Committer: author,
		Message:   message,
	}
}

func (c *Commit) SetParent(parent string) {
	c.parent = parent
}

func (c Commit) Type() string {
	return "commit"
}

func (c Commit) Serialize() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "tree %s\n", c.Tree)

	timeStamp := time.Now().Unix()
	timeZone := time.Now().Format("-0700")

	if c.parent != "" {
		fmt.Fprintf(&buf, "parent %s\n", c.parent)
	}

	fmt.Fprintf(&buf, "author %s %d %s\n", c.Author, timeStamp, timeZone)
	fmt.Fprintf(&buf, "committer %s %d %s\n", c.Committer, timeStamp, timeZone)

	buf.WriteString("\n")
	buf.WriteString(c.Message)
	buf.WriteString("\n")

	return buf.Bytes(), nil
}

func (c *Commit) Hash() (string, error) {
	if c.hash != "" {
		return c.hash, nil
	}

	data, _ := c.Serialize()
	header := fmt.Sprintf("commit %d\x00", len(data))
	fullContent := append([]byte(header), data...)

	hash := sha1.Sum(fullContent)
	hashHex := fmt.Sprintf("%x", hash)

	return hashHex, nil
}
