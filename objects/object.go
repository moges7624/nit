package objects

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moges7624/nit/repo"
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
