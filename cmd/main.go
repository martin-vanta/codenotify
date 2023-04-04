package main

import (
	"fmt"

	"github.com/sourcegraph/codenotify/codeowners"
)

func main() {
	differ := NewGitDiffer("origin/main")
	diffs, err := differ.Diff()
	if err != nil {
		panic(err)
	}

	matcher := codeowners.NewMatcher()
	for _, filePath := range diffs {
		match, err := matcher.Match(filePath)
		if err != nil {
			fmt.Println(filePath, err)
			continue
		}
		fmt.Println(filePath, match.RequiredOwners, match.OptionalOwners)
	}
}
