package worktree

import (
	"errors"
	"testing"
)

func TestValidateAddInput(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		branch     string
		wantPath   string
		wantBranch string
		wantErr    error
	}{
		{name: "trims surrounding whitespace", path: "  ../feature  ", branch: "  feature/a  ", wantPath: "../feature", wantBranch: "feature/a"},
		{name: "valid unchanged", path: "../feature", branch: "feature/a", wantPath: "../feature", wantBranch: "feature/a"},
		{name: "missing path", path: "   ", branch: "feature/a", wantErr: ErrWorktreePathRequired},
		{name: "missing branch", path: "../feature", branch: " ", wantErr: ErrBranchNameRequired},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path, branch, err := validateAddInput(test.path, test.branch)
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("expected error %v, got %v", test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if path != test.wantPath || branch != test.wantBranch {
				t.Fatalf("validateAddInput(%q, %q) = (%q, %q), want (%q, %q)", test.path, test.branch, path, branch, test.wantPath, test.wantBranch)
			}
		})
	}
}
