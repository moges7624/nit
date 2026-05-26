package commands

import (
	"fmt"
	"path/filepath"

	"github.com/moges7624/nit/internal/index"
	"github.com/moges7624/nit/internal/repo"
	"github.com/moges7624/nit/internal/status"
)

type FileStatus struct {
	name   string
	status string
}

func Status(args []string) (string, error) {
	repo, err := repo.Open(".")
	if err != nil {
		return "", fmt.Errorf("error opening repo: %s", err.Error())
	}

	index := index.NewIndex(filepath.Join(repo.NitPath(), "index"))
	if err = index.Load(); err != nil {
		return "", fmt.Errorf("error loading index: %s", err.Error())
	}

	status, _ := status.GetStatus(repo, index)

	var resp string
	if len(args) > 0 && args[0] == "--porcelain" {
		resp = status.FormatStatusPorcelain()
	} else {
		resp = status.FormatStatus()
	}

	return resp, nil
}
