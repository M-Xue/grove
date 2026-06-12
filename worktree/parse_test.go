package worktree

import (
	"reflect"
	"testing"
)

func TestParseWorktreeList(t *testing.T) {
	output := []byte("worktree /repo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /repo-feature\nHEAD def456\nbranch refs/heads/feature/auth\n")

	got, err := parseWorktreeList(output)
	if err != nil {
		t.Fatalf("parseWorktreeList returned error: %v", err)
	}

	want := []listEntry{
		{path: "/repo", branch: "refs/heads/main", commitHash: "abc123"},
		{path: "/repo-feature", branch: "refs/heads/feature/auth", commitHash: "def456"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected entries: got %#v want %#v", got, want)
	}
}

func TestParseWorktreeListRequiresPath(t *testing.T) {
	_, err := parseWorktreeList([]byte("HEAD abc123\nbranch refs/heads/main\n"))
	if err == nil {
		t.Fatal("expected missing path error")
	}
}

func TestParseWorktreeListRequiresCommitHash(t *testing.T) {
	_, err := parseWorktreeList([]byte("worktree /repo\nbranch refs/heads/main\n"))
	if err == nil {
		t.Fatal("expected missing commit hash error")
	}
}

func TestNormalizeBranch(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{name: "head ref", input: "refs/heads/main", output: "main"},
		{name: "feature ref", input: "refs/heads/feature/auth", output: "feature/auth"},
		{name: "detached", input: "", output: "detached"},
	}

	for _, test := range tests {
		if got := normalizeBranch(test.input); got != test.output {
			t.Fatalf("%s: got %q want %q", test.name, got, test.output)
		}
	}
}
