package worktree

import (
	"errors"
	"reflect"
	"testing"
)

type commandCall struct {
	name string
	args []string
}

type commandResult struct {
	output []byte
	err    error
}

type stubRunner struct {
	results map[string]commandResult
	calls   []commandCall
}

func (s *stubRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	call := commandCall{name: name, args: append([]string(nil), args...)}
	s.calls = append(s.calls, call)

	key := name + "\x00" + joinArgs(args)
	result, ok := s.results[key]
	if !ok {
		return nil, errors.New("unexpected command")
	}
	return result.output, result.err
}

func joinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	joined := args[0]
	for i := 1; i < len(args); i++ {
		joined += "\x00" + args[i]
	}
	return joined
}

func commandKey(name string, args ...string) string {
	return name + "\x00" + joinArgs(args)
}

func TestManagerAddRunsGitWorktreeAdd(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "add", "../feature-auth", "feature/auth"): {},
		},
	}
	manager := NewServiceWithRunner(runner)

	err := manager.Add("../feature-auth", "feature/auth")
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if len(runner.calls) != 1 {
		t.Fatalf("expected 1 command call, got %d", len(runner.calls))
	}

	want := commandCall{name: "git", args: []string{"worktree", "add", "../feature-auth", "feature/auth"}}
	if !reflect.DeepEqual(runner.calls[0], want) {
		t.Fatalf("unexpected call: got %+v want %+v", runner.calls[0], want)
	}
}

func TestManagerAddRequiresPath(t *testing.T) {
	runner := &stubRunner{}
	manager := NewServiceWithRunner(runner)

	err := manager.Add("   ", "feature/auth")
	if err == nil {
		t.Fatal("expected error for missing path")
	}
	if !errors.Is(err, ErrWorktreePathRequired) {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runner.calls) != 0 {
		t.Fatal("runner should not be called for invalid input")
	}
}

func TestManagerAddRequiresBranch(t *testing.T) {
	runner := &stubRunner{}
	manager := NewServiceWithRunner(runner)

	err := manager.Add("../feature-auth", " ")
	if err == nil {
		t.Fatal("expected error for missing branch")
	}
	if !errors.Is(err, ErrBranchNameRequired) {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runner.calls) != 0 {
		t.Fatal("runner should not be called for invalid input")
	}
}

func TestManagerAddReturnsRunnerError(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "add", "../feature-auth", "feature/auth"): {err: errors.New("git failed")},
		},
	}
	manager := NewServiceWithRunner(runner)

	err := manager.Add("../feature-auth", "feature/auth")
	if err == nil {
		t.Fatal("expected runner error")
	}
	if err.Error() != "git failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagerAddNewBranchRunsGitWorktreeAddWithBranchCreation(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "add", "-b", "feature/auth", "../feature-auth"): {},
		},
	}
	manager := NewServiceWithRunner(runner)

	err := manager.AddNewBranch("../feature-auth", "feature/auth")
	if err != nil {
		t.Fatalf("AddNewBranch returned error: %v", err)
	}

	want := commandCall{name: "git", args: []string{"worktree", "add", "-b", "feature/auth", "../feature-auth"}}
	if len(runner.calls) != 1 || !reflect.DeepEqual(runner.calls[0], want) {
		t.Fatalf("unexpected call: got %+v want %+v", runner.calls, want)
	}
}

func TestManagerRemoveRunsGitWorktreeRemove(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "remove", "../feature-auth"): {},
		},
	}
	manager := NewServiceWithRunner(runner)

	err := manager.Remove("../feature-auth")
	if err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}

	want := commandCall{name: "git", args: []string{"worktree", "remove", "../feature-auth"}}
	if len(runner.calls) != 1 || !reflect.DeepEqual(runner.calls[0], want) {
		t.Fatalf("unexpected call: got %+v want %+v", runner.calls, want)
	}
}

func TestManagerRemoveRequiresPath(t *testing.T) {
	runner := &stubRunner{}
	manager := NewServiceWithRunner(runner)

	err := manager.Remove(" ")
	if err == nil {
		t.Fatal("expected error for missing path")
	}
	if !errors.Is(err, ErrWorktreePathRequired) {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runner.calls) != 0 {
		t.Fatal("runner should not be called for invalid input")
	}
}

func TestManagerRemoveReturnsRunnerError(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "remove", "../feature-auth"): {err: errors.New("git failed")},
		},
	}
	manager := NewServiceWithRunner(runner)

	err := manager.Remove("../feature-auth")
	if err == nil {
		t.Fatal("expected runner error")
	}
	if err.Error() != "git failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagerBranchExistsReturnsTrueForExistingBranch(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "for-each-ref", "--format=%(refname:short)", "refs/heads/feature/auth"): {
				output: []byte("feature/auth\n"),
			},
		},
	}
	manager := NewServiceWithRunner(runner)

	exists, err := manager.BranchExists("feature/auth")
	if err != nil {
		t.Fatalf("BranchExists returned error: %v", err)
	}
	if !exists {
		t.Fatal("expected branch to exist")
	}
}

func TestManagerBranchExistsReturnsFalseForMissingBranch(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "for-each-ref", "--format=%(refname:short)", "refs/heads/feature/auth"): {
				output: []byte(""),
			},
		},
	}
	manager := NewServiceWithRunner(runner)

	exists, err := manager.BranchExists("feature/auth")
	if err != nil {
		t.Fatalf("BranchExists returned error: %v", err)
	}
	if exists {
		t.Fatal("expected branch to be missing")
	}
}

func TestManagerBranchExistsRequiresBranch(t *testing.T) {
	runner := &stubRunner{}
	manager := NewServiceWithRunner(runner)

	_, err := manager.BranchExists(" ")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, ErrBranchNameRequired) {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runner.calls) != 0 {
		t.Fatal("runner should not be called for invalid input")
	}
}

func TestManagerListReturnsStructuredWorktrees(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "list", "--porcelain"): {
				output: []byte("worktree /repo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /repo-feature\nHEAD def456\nbranch refs/heads/feature/auth\n"),
			},
			commandKey("git", "-C", "/repo", "log", "-1", "--pretty=%s"): {
				output: []byte("Initial commit\n"),
			},
			commandKey("git", "-C", "/repo", "status", "--porcelain"): {
				output: []byte(""),
			},
			commandKey("git", "-C", "/repo-feature", "log", "-1", "--pretty=%s"): {
				output: []byte("Add auth flow\n"),
			},
			commandKey("git", "-C", "/repo-feature", "status", "--porcelain"): {
				output: []byte(" M main.go\n"),
			},
		},
	}

	manager := NewServiceWithRunner(runner)
	got, err := manager.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	want := []WorktreeInfo{
		{
			Path:                  "/repo",
			Branch:                "main",
			CommitLabel:           "Initial commit",
			CommitHash:            "abc123",
			HasUncommittedChanges: false,
		},
		{
			Path:                  "/repo-feature",
			Branch:                "feature/auth",
			CommitLabel:           "Add auth flow",
			CommitHash:            "def456",
			HasUncommittedChanges: true,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected worktrees: got %#v want %#v", got, want)
	}
}

func TestManagerListSupportsDetachedHead(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "list", "--porcelain"): {
				output: []byte("worktree /repo\nHEAD abc123\ndetached\n"),
			},
			commandKey("git", "-C", "/repo", "log", "-1", "--pretty=%s"): {
				output: []byte("Detached commit\n"),
			},
			commandKey("git", "-C", "/repo", "status", "--porcelain"): {
				output: []byte(""),
			},
		},
	}

	manager := NewServiceWithRunner(runner)
	got, err := manager.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(got))
	}
	if got[0].Branch != "detached" {
		t.Fatalf("expected detached branch label, got %q", got[0].Branch)
	}
}
