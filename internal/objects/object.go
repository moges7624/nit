package objects

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/moges7624/nit/internal/repo"
)

type Object interface {
	Type() string
	Serialize() ([]byte, error)
	Hash() (string, error)
}

func Store(repo *repo.Repository, obj Object) (string, error) {
	content, err := obj.Serialize()
	if err != nil {
		return "", err
	}

	header := fmt.Sprintf("%s %d\x00", obj.Type(), len(content))
	store := append([]byte(header), content...)

	hash, _ := obj.Hash()

	dir := filepath.Join(repo.WorkTreePath(), ".git/objects/", hash[:2])

	if err = os.MkdirAll(dir, 0o755); err != nil {
		return "", nil
	}

	path := filepath.Join(dir, hash[2:])
	if _, err = os.Stat(path); err == nil {
		return hash, nil
	}

	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)

	if _, err = w.Write(store); err != nil {
		return "", err
	}

	w.Close()

	f, err := os.OpenFile(
		path,
		os.O_RDWR|os.O_CREATE,
		0o4444,
	)
	if err != nil {
		return "", err
	}

	defer f.Close()

	f.Write(buf.Bytes())
	return hash, nil
}

func Read(repo *repo.Repository, hash string) (Object, error) {
	if len(hash) != 40 {
		return nil, fmt.Errorf("invalid object hash: %s", hash)
	}

	path := repo.LooseObjectPath(hash)
	if path == "" {
		return nil, fmt.Errorf("invalid hash format")
	}

	compressed, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object not found: %s", hash)
		}
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	zr, err := zlib.NewReader(bytes.NewBuffer(compressed))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer zr.Close()

	raw, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress file: %w", err)
	}

	headerByte, content, ok := bytes.Cut(raw, []byte{0})
	if !ok {
		return nil, fmt.Errorf("corrupt object: no null byte found")
	}

	header := string(headerByte)
	// content := after

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid object header")
	}

	objType := parts[0]

	switch objType {
	case "blob":
		return &Blob{Data: content}, nil
	case "tree":
		return parseTree(content)
	case "commit":
		return parseCommit(content)
	default:
		return nil, fmt.Errorf("unknown object type; %s", objType)
	}
}

func parseTree(data []byte) (*Tree, error) {
	tree := &Tree{}
	offset := 0

	for offset < len(data) {
		spaceIdx := bytes.IndexByte(data[offset:], ' ')
		if spaceIdx == -1 {
			break
		}

		modeStr := string(data[offset : offset+spaceIdx])
		offset += spaceIdx + 1

		nullIdx := bytes.IndexByte(data[offset:], 0)
		if nullIdx == -1 {
			break
		}

		name := string(data[offset : offset+nullIdx])
		offset += nullIdx + 1

		if offset+20 > len(data) {
			break
		}

		hashBytes := data[offset : offset+20]
		hashHex := hex.EncodeToString(hashBytes)
		offset += 20

		tree.Entries = append(tree.Entries, Entry{
			Name: name,
			Mode: FileMode(modeStr),
			Hash: hashHex,
		})

	}

	return tree, nil
}

func parseCommit(data []byte) (*Commit, error) {
	c := &Commit{}
	lines := strings.Split(string(data), "\n")

	for i, line := range lines {
		if line == "" {
			if i+1 < len(lines) {
				c.Message = strings.Join(lines[i+1:], "n")
			}
			break
		}

		if strings.HasPrefix(line, "tree ") {
			c.Tree = strings.TrimPrefix(line, "tree ")
		} else if strings.HasPrefix(line, "parent") {
			c.parent = strings.TrimPrefix(line, "parent ")
		} else if strings.HasPrefix(line, "author") {
			c.Author = strings.TrimPrefix(line, "author ")
		} else if strings.HasPrefix(line, "committer ") {
			c.Committer = strings.TrimPrefix(line, "committer ")
		}
	}

	return c, nil
}
