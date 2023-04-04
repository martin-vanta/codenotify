package codeowners

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestMatcherMatch(t *testing.T) {
	fs := afero.NewMemMapFs()

	fs.MkdirAll("a/b", 0755)
	afero.WriteFile(fs, "OWNERS", []byte(`
		[required]
		root.go @root
		a/a.go @root_overridden
		/root_slash.go @root_slash
		.//./root_slash_unnormalized.go @root_slash_unnormalized

		a/**/*.s @doublestar
		b/*.s @singlestar
		**/doublestar_prefix.s @doublestar_prefix

		^[optional]
		root_optional.go @root_optional
		`), 0644)
	afero.WriteFile(fs, "a/OWNERS", []byte(`
		[required]
		a.go @a_overridden
		a.go @a
		a_both.go @a
		/a_slash.go @a_slash

		^[optional]
		a_optional.go @a_optional
		a_both.go @a_optional
		`), 0644)

	matcher := newMatcherWithFs("OWNERS", fs)

	tests := []struct {
		filePath string
		expected Match
	}{
		{filePath: "", expected: Match{}},
		{filePath: ".", expected: Match{}},
		{filePath: "does_not_exist.go", expected: Match{}},
		{filePath: "a/does_not_exist.go", expected: Match{}},
		{filePath: "a/b/c/d/does_not_exist.go", expected: Match{}},

		{filePath: "root.go", expected: Match{RequiredOwners: []string{"@root"}}},
		{filePath: "root_optional.go", expected: Match{OptionalOwners: []string{"@root_optional"}}},
		{filePath: "root_slash.go", expected: Match{RequiredOwners: []string{"@root_slash"}}},
		{filePath: "root_slash_unnormalized.go", expected: Match{RequiredOwners: []string{"@root_slash_unnormalized"}}},

		{filePath: "a/a.go", expected: Match{RequiredOwners: []string{"@a"}}},
		{filePath: "a/a_optional.go", expected: Match{OptionalOwners: []string{"@a_optional"}}},
		{filePath: "a/a_both.go", expected: Match{RequiredOwners: []string{"@a"}, OptionalOwners: []string{"@a_optional"}}},
		{filePath: "a/a_slash.go", expected: Match{RequiredOwners: []string{"@a_slash"}}},

		{filePath: "doublestar_prefix.s", expected: Match{RequiredOwners: []string{"@doublestar_prefix"}}},
		{filePath: "a/doublestar_prefix.s", expected: Match{RequiredOwners: []string{"@doublestar_prefix"}}},
		{filePath: "a/b/c/d/doublestar_prefix.s", expected: Match{RequiredOwners: []string{"@doublestar_prefix"}}},

		{filePath: "a/doublestar.s", expected: Match{RequiredOwners: []string{"@doublestar"}}},
		{filePath: "a/b/c/d/doublestar.s", expected: Match{RequiredOwners: []string{"@doublestar"}}},

		{filePath: "b/singlestar.s", expected: Match{RequiredOwners: []string{"@singlestar"}}},
		{filePath: "b/c/d/singlestar.s", expected: Match{}},
	}
	for _, test := range tests {
		got, err := matcher.Match(test.filePath)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, got, "file: %s", test.filePath)
	}
}

func TestMatcherLoad(t *testing.T) {
	fs := afero.NewMemMapFs()

	fs.MkdirAll("a/b", 0755)
	afero.WriteFile(fs, "OWNERS", []byte("root.go @root"), 0644)
	afero.WriteFile(fs, "a/OWNERS", []byte("a.go @a"), 0644)

	matcher := newMatcherWithFs("OWNERS", fs)

	// Loads root OWNERS file with empty argument.
	ownersFile, err := matcher.Load("")
	assert.NoError(t, err)
	assert.Equal(t, &OwnersFile{Sections: []*Section{
		{Name: defaultSectionName, Approvals: 1, Rules: []*Rule{
			{Pattern: "root.go", Owners: []string{"@root"}},
		}},
	}}, ownersFile)

	// Loads root OWNERS file with . argument.
	ownersFile, err = matcher.Load(".")
	assert.NoError(t, err)
	assert.Equal(t, &OwnersFile{Sections: []*Section{
		{Name: defaultSectionName, Approvals: 1, Rules: []*Rule{
			{Pattern: "root.go", Owners: []string{"@root"}},
		}},
	}}, ownersFile)

	// Loads a/OWNERS file.
	ownersFile, err = matcher.Load("a")
	assert.NoError(t, err)
	assert.Equal(t, &OwnersFile{Sections: []*Section{
		{Name: defaultSectionName, Approvals: 1, Rules: []*Rule{
			{Pattern: "a.go", Owners: []string{"@a"}},
		}},
	}}, ownersFile)

	// Loads proxy empty file for directory without OWNERS file.
	ownersFile, err = matcher.Load("a/b")
	assert.NoError(t, err)
	assert.Equal(t, &OwnersFile{}, ownersFile)

	// Loads proxy empty file for non existant directory.
	ownersFile, err = matcher.Load("a/b/c/d")
	assert.NoError(t, err)
	assert.Equal(t, &OwnersFile{}, ownersFile)
}
