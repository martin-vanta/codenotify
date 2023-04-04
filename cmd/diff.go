package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type Differ interface {
	Diff() ([]string, error)
}

type gitDiffer struct {
	fromRef string
}

func NewGitDiffer(fromRef string) Differ {
	return gitDiffer{
		fromRef: fromRef,
	}
}

func (d gitDiffer) Diff() ([]string, error) {
	cmd := exec.Command("git", "diff", fmt.Sprintf("%s..", d.fromRef), "--name-only")
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error executing '%s'\n%s", strings.Join(cmd.Args, " "), stderr.String())
	}
	return strings.Fields(stdout.String()), nil
}
