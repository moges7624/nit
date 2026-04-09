package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moges7624/nit/repo"
)

func Init(args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	absPath, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(absPath, 0o755); err != nil {
		return err
	}

	repo := repo.NewRepository(absPath)
	if err := repo.Init(); err != nil {
		return err
	}

	fmt.Printf("Initialized empty nit repository in %s/.git/\n", repo.WorkTreePath())
	return nil
}
