package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func Init(args []string) {
	rootPath, err := os.Getwd()
	if err != nil {
		fmt.Println("Couldn't get root path")
		os.Exit(1)
	}

	path := "."

	if len(args) > 0 {
		path = args[0]
	}

	nitPath := filepath.Join(rootPath, path, ".git")

	// create .git directory
	err = os.MkdirAll(nitPath, 0o755)
	if err != nil {
		var pathError *os.PathError
		if errors.As(err, &pathError) {
			fmt.Fprintf(os.Stderr, "%s: %s\n", pathError.Path, pathError.Err)
		}

		return
	}

	// create required subdirectories
	dirs := []string{"objects", "refs/heads", "refs/tags"}
	for _, dir := range dirs {
		err = os.MkdirAll(filepath.Join(nitPath, dir), 0o755)
		if err != nil {
			fmt.Printf("Error creating %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	f, err := os.OpenFile(filepath.Join(nitPath, "refs/heads/main"),
		os.O_CREATE,
		0o644,
	)
	if err != nil {
		fmt.Printf("Error creating main: %v\n", err)
		return
	}
	defer f.Close()

	// Write the initial HEAD file
	err = os.WriteFile(filepath.Join(nitPath, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing HEAD: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Initialized empty nit repository in %s/\n", nitPath)
}
