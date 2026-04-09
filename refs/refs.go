package refs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Ref struct {
	nitDir string
	head   string
}

func NewRef(dir string) *Ref {
	return &Ref{
		nitDir: dir,
	}
}

func (r Ref) UpdateHead(hashHex string) error {
	headPath, err := r.GetHeadPath()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(
		filepath.Join(r.nitDir, headPath),
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0o644,
	)
	if err != nil {
		return fmt.Errorf("error updating head: %s", err.Error())
	}

	defer f.Close()

	if _, err := f.WriteString(hashHex); err != nil {
		return fmt.Errorf("error updating head: %s", err.Error())
	}

	return nil
}

func (r *Ref) GetHeadCommit() (string, error) {
	headPath, err := r.GetHeadPath()
	if err != nil {
		return "", nil
	}

	h, err := os.ReadFile(filepath.Join(r.nitDir, headPath))
	if err != nil {
		if os.IsNotExist(err) {
			return "", err
		}
		return "", fmt.Errorf("error getting head: %s", err.Error())
	}

	return string(h), err
}

func (r *Ref) GetHeadPath() (string, error) {
	if r.head != "" {
		return r.head, nil
	}

	f, err := os.ReadFile(filepath.Join(r.nitDir, "HEAD"))
	if err != nil {
		return "", fmt.Errorf("error reading head: %s", err.Error())
	}

	h := strings.Fields(string(f))
	r.head = string(h[1])

	return r.head, nil
}
