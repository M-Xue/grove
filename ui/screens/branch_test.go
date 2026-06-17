package screens

import (
	"strings"
	"testing"

	"github.com/M-Xue/grove/app"
	branchsvc "github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/worktree"
)

func TestBranchSwitchWarningShownForDirtyCurrentBranch(t *testing.T) {
	state := app.State{
		Branches: []branchsvc.Info{{Name: "feature/a", CheckedOutHere: true}},
		Worktrees: []worktree.Info{{
			Branch:                "feature/a",
			HasUncommittedChanges: true,
		}},
	}

	warning := branchSwitchWarning(state)
	if warning == "" {
		t.Fatal("expected warning for dirty current branch")
	}
	if want := "uncommitted files may prevent switching branches"; !strings.Contains(warning, want) {
		t.Fatalf("expected warning to contain %q, got %q", want, warning)
	}
}

func TestBranchSwitchWarningHiddenWhenCurrentBranchClean(t *testing.T) {
	state := app.State{
		Branches: []branchsvc.Info{{Name: "feature/a", CheckedOutHere: true}},
		Worktrees: []worktree.Info{{
			Branch:                "feature/a",
			HasUncommittedChanges: false,
		}},
	}

	if warning := branchSwitchWarning(state); warning != "" {
		t.Fatalf("expected no warning, got %q", warning)
	}
}
