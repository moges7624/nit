package repo

import (
	"fmt"
	"os"
)

type Repository struct {
	WorkDir string
}

func NewRepository(path string) *Repository {
	return &Repository{
		WorkDir: path,
	}
}

func (r *Repository) ListFiles() ([]os.DirEntry, error) {
	files, err := os.ReadDir(r.WorkDir)
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
