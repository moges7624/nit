package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		files map[string]struct {
			content string
			hash    string
		}
	}{
		{
			name: "emtpy file",
			args: []string{"file.txt"},
			files: map[string]struct {
				content string
				hash    string
			}{
				"file.txt": {
					content: "",
					hash:    "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
				},
			},
		},
		{
			name: "single file with content",
			args: []string{"file.txt"},
			files: map[string]struct {
				content string
				hash    string
			}{
				"file.txt": {
					content: "hello\n",
					hash:    "ce013625030ba8dba906f756967f9e9ca394464a",
				},
			},
		},
		{
			name: "multiple files",
			args: []string{"file.txt", "readme.txt"},
			files: map[string]struct {
				content string
				hash    string
			}{
				"file.txt": {
					content: "hello\n",
					hash:    "ce013625030ba8dba906f756967f9e9ca394464a",
				},
				"readme.txt": {
					content: "",
					hash:    "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
				},
			},
		},
		{
			name: "file in nested directory",
			args: []string{"file.txt", "internal/readme.txt"},
			files: map[string]struct {
				content string
				hash    string
			}{
				"file.txt": {
					content: "hello\n",
					hash:    "ce013625030ba8dba906f756967f9e9ca394464a",
				},
				"internal/readme.txt": {
					content: "",
					hash:    "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			originalWd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			if err = os.Chdir(tempDir); err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() {
				_ = os.Chdir(originalWd)
			})

			err = Init([]string{})
			if err != nil {
				t.Fatalf("error inititializing the repo: %v", err)
			}

			for k, v := range tt.files {
				err := os.MkdirAll(filepath.Dir(k), 0o755)
				if err != nil {
					t.Fatal(err)
				}

				err = os.WriteFile(k, []byte(v.content), 0o644)
				if err != nil {
					t.Fatal(err)
				}
			}

			Add(tt.args)

			for k, v := range tt.files {
				fullPath := filepath.Join(tempDir, ".git/objects", v.hash[:2], v.hash[2:])
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Errorf("file %s not written to .git/objects", k)
				}
			}
		})
	}
}
