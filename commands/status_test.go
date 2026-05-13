package commands

import (
	"os"
	"testing"
)

func setupRepo(t *testing.T) {
	t.Helper()

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
		t.Fatalf("error initializing the repo: %v", err)
	}
}

func TestStatus_UntrackedFiles(t *testing.T) {
	t.Run("untracked files in a repo with empty index", func(t *testing.T) {
		setupRepo(t)
		if err := os.WriteFile("test.txt", []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile("file.txt", []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		res, err := Status()
		if err != nil {
			t.Fatal(err)
		}

		expected := "?? file.txt\n?? test.txt\n"
		if res != expected {
			t.Errorf("expected: \n%s got: \n%s", expected, res)
		}
	})

	t.Run("repo with tracked files", func(t *testing.T) {
		setupRepo(t)

		if err := os.WriteFile("test.txt", []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile("file.txt", []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		Add([]string{"test.txt", "file.txt"})
		// if err != nil {
		// 	t.Errorf("error adding a file: %w", err)
		// }
		if err := os.WriteFile("file.txt", []byte("modifying file"), 0o644); err != nil {
			t.Fatal(err)
		}

		res, err := Status()
		if err != nil {
			t.Fatal(err)
		}

		expected := "A file.txt\nA test.txt\n"
		if res != expected {
			t.Errorf("expected: \n%s got: \n%s", expected, res)
		}
	})

	t.Run("repo with nested directory", func(t *testing.T) {
		setupRepo(t)

		err := os.Mkdir("internal", 0o755)
		if err != nil {
			t.Fatal(err)
		}

		if err = os.WriteFile("internal/test.txt", []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		if err = os.WriteFile("file.txt", []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		Add([]string{"internal/test.txt"})
		res, err := Status()
		if err != nil {
			t.Fatal(err)
		}

		expected := "A internal/test.txt\n?? file.txt\n"
		if res != expected {
			t.Errorf("expected: \n%s got: \n%s", expected, res)
		}
	})
}
