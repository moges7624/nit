package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/moges7624/nit/internal/objects"
	"github.com/moges7624/nit/internal/refs"
	"github.com/moges7624/nit/internal/repo"
)

func Log(args []string) error {
	repo, err := repo.Open(".")
	if err != nil {
		return fmt.Errorf("error opening repo: %w", err)
	}

	ref := refs.NewRef(repo.NitPath())
	headCommitHash, err := ref.GetHeadCommit()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("fatal: your current branch 'main' does not have any commits yet")
			return nil
		}

		return fmt.Errorf("error getting head commit hash: %w", err)
	}

	if headCommitHash == "" {
		fmt.Println("fatal: your current branch 'main' does not have any commits yet")
		return nil
	}

	currHash := headCommitHash

	var output strings.Builder
	for currHash != "" {
		obj, err := objects.Read(repo, currHash)
		if err != nil {
			return fmt.Errorf("failed to read commit %s: %w", currHash, err)
		}

		commit, ok := obj.(*objects.Commit)
		if !ok {
			return fmt.Errorf("object %s is not a commit", currHash)
		}

		output.WriteString(formatCommit(currHash, commit))

		currHash = commit.GetParent()
	}

	return displayWithPager(output.String())
}

func formatCommit(hash string, commit *objects.Commit) string {
	var sb strings.Builder

	authorStrArr := strings.Split(commit.Author, " ")
	commitDateInt, _ := strconv.ParseInt(authorStrArr[2], 10, 64)
	commitDate := time.Unix(commitDateInt, 0).Format("Mon Jan 02 15:04:05 2006 -0700")

	fmt.Fprintf(&sb, "\033[33mcommit %s\033[0m\n", hash)
	fmt.Fprintf(&sb, "Author: %s\n", authorStrArr[0]+authorStrArr[1])
	fmt.Fprintf(&sb, "Date: %s\n\n", commitDate)
	fmt.Fprintf(&sb, "    %s\n", commit.Message)

	return sb.String()
}

func displayWithPager(content string) error {
	if len(strings.Split(content, "\n")) <= 25 {
		fmt.Print(content)
		return nil
	}

	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less"
	}

	var cmd *exec.Cmd

	switch {
	case strings.Contains(pager, "less"):
		cmd = exec.Command("less", "-RFX")
	case strings.Contains(pager, "more"):
		cmd = exec.Command("more")
	default:
		cmd = exec.Command(pager)
	}

	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err == nil {
		return nil
	}

	// fallback to simple pager if system pager fails
	return simpleEnterPager(content)
}

func simpleEnterPager(content string) error {
	lines := strings.Split(content, "\n")
	const pageSize = 20
	scanner := bufio.NewScanner(os.Stdin)

	for i := 0; i < len(lines); i += pageSize {
		end := i + pageSize
		if end > len(lines) {
			end = len(lines)
		}

		for j := i; j < end; j++ {
			fmt.Println(lines[j])
		}

		if end < len(lines) {
			fmt.Print("\n-- Press Enter to continue --")
			scanner.Scan()
		}
	}
	return nil
}
