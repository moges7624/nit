package repo

import (
	"fmt"
	"os"
)

type Repository struct {
	NitDir string
}

func NewRepository(path string) *Repository {
	return &Repository{
		NitDir: path,
	}
}

func (r *Repository) ListFiles() ([]os.DirEntry, error) {
	files, err := os.ReadDir(r.NitDir)
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
