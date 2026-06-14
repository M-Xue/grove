package branch

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

func TestListReturnsLocalBranchesOrderedByReflogRecency(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "for-each-ref", "--sort=-committerdate", "--format=%(refname:short)", "refs/heads"): {
				output: []byte("main\nfeature/a\nfeature/z\n"),
			},
			commandKey("git", "branch", "--show-current"): {
				output: []byte("main\n"),
			},
			commandKey("git", "worktree", "list", "--porcelain"): {
				output: []byte("worktree /repo\nHEAD abc\nbranch refs/heads/main\n\nworktree /repo-feature\nHEAD def\nbranch refs/heads/feature/a\n"),
			},
			commandKey("git", "reflog", "show", "--format=%gs", "HEAD"): {
				output: []byte("checkout: moving from main to feature/a\ncheckout: moving from feature/a to main\n"),
			},
			commandKey("git", "reflog", "show", "--format=%gD", "--all"): {
				output: []byte("refs/heads/feature/a@{0}\nrefs/heads/main@{1}\nrefs/remotes/origin/main@{0}\n"),
			},
		},
	}

	service := NewServiceWithRunner(runner)
	branches, scope, err := service.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if scope != ScopeLocal {
		t.Fatalf("unexpected scope: %q", scope)
	}

	want := []Info{
		{Name: "feature/a", CheckedOutElsewhere: true},
		{Name: "main", CheckedOutHere: true},
		{Name: "feature/z"},
	}
	if !reflect.DeepEqual(branches, want) {
		t.Fatalf("unexpected branches: got %#v want %#v", branches, want)
	}
}

func TestToggleScopeSwitchesToRemoteTrackingBranches(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "for-each-ref", "--sort=-committerdate", "--format=%(refname:short)", "refs/remotes"): {
				output: []byte("origin/main\norigin/HEAD\norigin/feature/a\n"),
			},
		},
	}

	service := NewServiceWithRunner(runner)
	if scope := service.ToggleScope(); scope != ScopeRemoteTracking {
		t.Fatalf("unexpected scope after toggle: %q", scope)
	}

	branches, scope, err := service.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if scope != ScopeRemoteTracking {
		t.Fatalf("unexpected scope: %q", scope)
	}

	want := []Info{{Name: "origin/main"}, {Name: "origin/feature/a"}}
	if !reflect.DeepEqual(branches, want) {
		t.Fatalf("unexpected branches: got %#v want %#v", branches, want)
	}
}

func TestCheckoutUsesScopeSpecificGitCommand(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "switch", "main"):                 {},
			commandKey("git", "switch", "--track", "origin/main"): {},
		},
	}

	service := NewServiceWithRunner(runner)
	if err := service.Checkout("main"); err != nil {
		t.Fatalf("Checkout returned error: %v", err)
	}
	service.ToggleScope()
	if err := service.Checkout("origin/main"); err != nil {
		t.Fatalf("Checkout returned error: %v", err)
	}

	want := []commandCall{
		{name: "git", args: []string{"switch", "main"}},
		{name: "git", args: []string{"switch", "--track", "origin/main"}},
	}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("unexpected calls: got %#v want %#v", runner.calls, want)
	}
}

func TestFetchRunsGitFetch(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "fetch"): {},
		},
	}

	service := NewServiceWithRunner(runner)
	if err := service.Fetch(); err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}
	if len(runner.calls) != 1 || !reflect.DeepEqual(runner.calls[0], commandCall{name: "git", args: []string{"fetch"}}) {
		t.Fatalf("unexpected calls: %#v", runner.calls)
	}
}

func TestRecentCommitsParsesGitLogOutput(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "log", "-n10", "--format=%h%x1f%an%x1f%s", "feature/a"): {
				output: []byte("abc123\x1fMax\x1fAdd auth flow\ndef456\x1fSam\x1fFix tests\n"),
			},
		},
	}

	service := NewServiceWithRunner(runner)
	commits, err := service.RecentCommits("feature/a", 10)
	if err != nil {
		t.Fatalf("RecentCommits returned error: %v", err)
	}
	want := []CommitInfo{
		{Hash: "abc123", Author: "Max", Subject: "Add auth flow"},
		{Hash: "def456", Author: "Sam", Subject: "Fix tests"},
	}
	if !reflect.DeepEqual(commits, want) {
		t.Fatalf("unexpected commits: got %#v want %#v", commits, want)
	}
}

func TestDeleteRunsGitBranchDeleteForLocalScope(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "branch", "-d", "feature/a"): {},
		},
	}

	service := NewServiceWithRunner(runner)
	if err := service.Delete("feature/a"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if len(runner.calls) != 1 || !reflect.DeepEqual(runner.calls[0], commandCall{name: "git", args: []string{"branch", "-d", "feature/a"}}) {
		t.Fatalf("unexpected calls: %#v", runner.calls)
	}
}

func TestDeleteRejectsRemoteTrackingScope(t *testing.T) {
	service := NewServiceWithRunner(&stubRunner{results: map[string]commandResult{}})
	service.ToggleScope()
	if err := service.Delete("origin/feature/a"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteAllLocalDeletesDeletableBranchesAndSkipsCheckedOutOnes(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "for-each-ref", "--sort=-committerdate", "--format=%(refname:short)", "refs/heads"): {
				output: []byte("feature/a\nfeature/b\nmain\n"),
			},
			commandKey("git", "branch", "--show-current"): {
				output: []byte("main\n"),
			},
			commandKey("git", "worktree", "list", "--porcelain"): {
				output: []byte("worktree /repo\nHEAD abc\nbranch refs/heads/main\n\nworktree /repo-feature\nHEAD def\nbranch refs/heads/feature/b\n"),
			},
			commandKey("git", "reflog", "show", "--format=%gs", "HEAD"): {
				output: []byte("checkout: moving from feature/a to feature/b\ncheckout: moving from main to feature/a\n"),
			},
			commandKey("git", "reflog", "show", "--format=%gD", "--all"): {
				output: []byte("refs/heads/feature/b@{0}\nrefs/heads/main@{1}\nrefs/heads/feature/a@{2}\n"),
			},
			commandKey("git", "branch", "-D", "feature/a"): {},
		},
	}

	service := NewServiceWithRunner(runner)
	summary, err := service.DeleteAllLocal()
	if err != nil {
		t.Fatalf("DeleteAllLocal returned error: %v", err)
	}
	want := DeleteAllSummary{
		Deleted: []string{"feature/a"},
		Skipped: []string{"feature/b", "main"},
	}
	if !reflect.DeepEqual(summary, want) {
		t.Fatalf("unexpected summary: got %#v want %#v", summary, want)
	}
}

func TestReflogBranchNameParsesHeadAndRemoteRefs(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "refs/heads/main@{0}", want: "main"},
		{input: "refs/remotes/origin/main@{12}", want: "origin/main"},
		{input: "HEAD@{0}", want: ""},
	}

	for _, test := range tests {
		if got := reflogBranchName(test.input); got != test.want {
			t.Fatalf("reflogBranchName(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestHeadCheckoutBranchNameParsesCheckoutMessages(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "checkout: moving from main to feature/a", want: "feature/a"},
		{input: "checkout: moving from abc123 to main", want: "main"},
		{input: "commit: message", want: ""},
	}

	for _, test := range tests {
		if got := headCheckoutBranchName(test.input); got != test.want {
			t.Fatalf("headCheckoutBranchName(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}
