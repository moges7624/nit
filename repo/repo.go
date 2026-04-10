package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Repository struct {
	workTreePath string
	nitPath      string
}

func NewRepository(path string) *Repository {
	return &Repository{
		workTreePath: path,
		nitPath:      path + "/.git",
	}
}

func (r *Repository) WorkTreePath() string {
	return r.workTreePath
}

func (r *Repository) NitPath() string {
	return r.nitPath
}

func (r *Repository) Init() error {
	dirs := []string{
		".git",
		filepath.Join(r.nitPath, "objects"),
		filepath.Join(r.nitPath, "objects", "info"),
		filepath.Join(r.nitPath, "objects", "pack"),
		filepath.Join(r.nitPath, "refs"),
		filepath.Join(r.nitPath, "refs", "heads"),
		filepath.Join(r.nitPath, "refs", "tags"),
		filepath.Join(r.nitPath, "info"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}

	headPath := filepath.Join(r.nitPath, "HEAD")
	err := os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0o644)
	if err != nil {
		return fmt.Errorf("failed writing to HEAD %w", err)
	}

	return nil
}

// Open returns an existing repository
// by searching upward from the given path
func Open(startPath string) (*Repository, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return nil, err
	}

	current := absPath
	for {
		nitPath := filepath.Join(current, ".git")
		if fi, err := os.Stat(nitPath); err == nil && fi.IsDir() {
			return &Repository{
				workTreePath: current,
				nitPath:      nitPath,
			}, nil
		}

		parent := filepath.Dir(current)
		if parent == current { // reached filesystem root
			break
		}
		current = parent
	}

	return nil, errors.New("not a mygit repository (or any of the parent directories): .git")
}

func (r *Repository) ListFiles() ([]os.DirEntry, error) {
	files, err := os.ReadDir(r.workTreePath)
	if err != nil {
		return nil, fmt.Errorf("error listing files in the working directory: %w", err)
	}

	ignore := map[string]bool{
		".git": true,
		".":    true,
		"..":   true,
	}

	var validFiles []os.DirEntry
	for _, file := range files {
		if ignore[file.Name()] {
			continue
		}

		validFiles = append(validFiles, file)

	}

	return validFiles, nil
}
