package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		checkFiles []string
	}{
		{
			name:    "init in current directory",
			args:    []string{},
			wantErr: false,
			checkFiles: []string{
				".git",
				".git/HEAD",
				".git/objects",
				".git/refs/heads",
				".git/refs/tags",
			},
		},
		{
			name:    "init with explicit path",
			args:    []string{"myproject"},
			wantErr: false,
			checkFiles: []string{
				".git",
				".git/HEAD",
				".git/objects",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			targetDir := tempDir
			if len(tt.args) > 0 {
				targetDir = filepath.Join(tempDir, tt.args[0])
			}

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

			err = Init(tt.args)
			if err != nil {
				t.Fatalf("Init() failed: %v", err)
			}

			for _, relPath := range tt.checkFiles {
				fullPath := filepath.Join(targetDir, relPath)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Errorf("expected %s to exist after init, but it does not", relPath)
				}
			}

			headPath := filepath.Join(targetDir, ".git", "HEAD")
			if data, err := os.ReadFile(headPath); err == nil {
				content := string(data)
				if content != "ref: refs/heads/main\n" && content != "ref: refs/heads/main" {
					t.Errorf("HEAD has unexpected content %q", content)
				}
			}
		})
	}
}
