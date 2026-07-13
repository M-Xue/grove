package worktree

import (
	"errors"
	"io/fs"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// fakeFileInfo is a minimal fs.FileInfo carrying only a mod time, used to
// drive worktree creation-time ordering in tests.
type fakeFileInfo struct {
	modTime time.Time
}

func (f fakeFileInfo) Name() string       { return "gitdir" }
func (f fakeFileInfo) Size() int64        { return 0 }
func (f fakeFileInfo) Mode() fs.FileMode  { return 0 }
func (f fakeFileInfo) ModTime() time.Time { return f.modTime }
func (f fakeFileInfo) IsDir() bool        { return false }
func (f fakeFileInfo) Sys() any           { return nil }

// existingStat builds a stat func for List tests where every worktree directory
// is present on disk. modTimes supplies mod times for specific administrative
// paths (used to exercise creation-time ordering); any other path is reported
// as an existing file with the zero mod time, so List's on-disk existence check
// treats the fixture worktrees as reachable. Tests that need a path to read as
// missing supply their own stat instead.
func existingStat(modTimes map[string]time.Time) func(name string) (fs.FileInfo, error) {
	return func(name string) (fs.FileInfo, error) {
		if t, ok := modTimes[name]; ok {
			return fakeFileInfo{modTime: t}, nil
		}
		return fakeFileInfo{}, nil
	}
}

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

func TestServiceAddRunsGitWorktreeAdd(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "add", "../feature-auth", "feature/auth"): {},
		},
	}
	service := NewService(runner)

	err := service.Add("../feature-auth", "feature/auth")
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

func TestServiceAddRequiresPath(t *testing.T) {
	runner := &stubRunner{}
	service := NewService(runner)

	err := service.Add("   ", "feature/auth")
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

func TestServiceAddRequiresBranch(t *testing.T) {
	runner := &stubRunner{}
	service := NewService(runner)

	err := service.Add("../feature-auth", " ")
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

func TestServiceAddReturnsRunnerError(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "add", "../feature-auth", "feature/auth"): {err: errors.New("git failed")},
		},
	}
	service := NewService(runner)

	err := service.Add("../feature-auth", "feature/auth")
	if err == nil {
		t.Fatal("expected runner error")
	}
	if err.Error() != "git failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceAddNewBranchRunsGitWorktreeAddWithBranchCreation(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "add", "-b", "feature/auth", "../feature-auth"): {},
		},
	}
	service := NewService(runner)

	err := service.AddWithNewBranch("../feature-auth", "feature/auth")
	if err != nil {
		t.Fatalf("AddNewBranch returned error: %v", err)
	}

	want := commandCall{name: "git", args: []string{"worktree", "add", "-b", "feature/auth", "../feature-auth"}}
	if len(runner.calls) != 1 || !reflect.DeepEqual(runner.calls[0], want) {
		t.Fatalf("unexpected call: got %+v want %+v", runner.calls, want)
	}
}

func TestServiceRemoveRunsGitWorktreeRemove(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "remove", "../feature-auth"): {},
		},
	}
	service := NewService(runner)

	err := service.Remove("../feature-auth")
	if err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}

	want := commandCall{name: "git", args: []string{"worktree", "remove", "../feature-auth"}}
	if len(runner.calls) != 1 || !reflect.DeepEqual(runner.calls[0], want) {
		t.Fatalf("unexpected call: got %+v want %+v", runner.calls, want)
	}
}

func TestServiceRemoveRequiresPath(t *testing.T) {
	runner := &stubRunner{}
	service := NewService(runner)

	err := service.Remove(" ")
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

func TestServiceRemoveReturnsRunnerError(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "remove", "../feature-auth"): {err: errors.New("git failed")},
		},
	}
	service := NewService(runner)

	err := service.Remove("../feature-auth")
	if err == nil {
		t.Fatal("expected runner error")
	}
	if err.Error() != "git failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServicePruneRunsGitWorktreePrune(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "prune"): {},
		},
	}
	service := NewService(runner)

	if err := service.Prune(); err != nil {
		t.Fatalf("Prune returned error: %v", err)
	}

	want := commandCall{name: "git", args: []string{"worktree", "prune"}}
	if len(runner.calls) != 1 || !reflect.DeepEqual(runner.calls[0], want) {
		t.Fatalf("unexpected call: got %+v want %+v", runner.calls, want)
	}
}

func TestServicePruneReturnsRunnerError(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "prune"): {err: errors.New("git failed")},
		},
	}
	service := NewService(runner)

	if err := service.Prune(); err == nil {
		t.Fatal("expected runner error")
	}
}

func TestServiceBranchExistsReturnsTrueForExistingBranch(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "for-each-ref", "--format=%(refname:short)", "refs/heads/feature/auth"): {
				output: []byte("feature/auth\n"),
			},
		},
	}
	service := NewService(runner)

	exists, err := service.BranchExists("feature/auth")
	if err != nil {
		t.Fatalf("BranchExists returned error: %v", err)
	}
	if !exists {
		t.Fatal("expected branch to exist")
	}
}

func TestServiceBranchExistsReturnsFalseForMissingBranch(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "for-each-ref", "--format=%(refname:short)", "refs/heads/feature/auth"): {
				output: []byte(""),
			},
		},
	}
	service := NewService(runner)

	exists, err := service.BranchExists("feature/auth")
	if err != nil {
		t.Fatalf("BranchExists returned error: %v", err)
	}
	if exists {
		t.Fatal("expected branch to be missing")
	}
}

func TestServiceBranchExistsRequiresBranch(t *testing.T) {
	runner := &stubRunner{}
	service := NewService(runner)

	_, err := service.BranchExists(" ")
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

func TestServiceListReturnsStructuredWorktrees(t *testing.T) {
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

	svc := service{runner: runner, stat: existingStat(nil)}
	got, err := svc.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	want := []Info{
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

func TestServiceListPinsMainWorktreeThenOrdersRestOldestFirst(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			// git always lists the main worktree (/main) first, even though it
			// was created before the linked worktrees below it.
			commandKey("git", "worktree", "list", "--porcelain"): {
				output: []byte("worktree /main\nHEAD m\nbranch refs/heads/main\n\nworktree /new\nHEAD n\nbranch refs/heads/new\n\nworktree /old\nHEAD o\nbranch refs/heads/old\n"),
			},
			commandKey("git", "-C", "/main", "log", "-1", "--pretty=%s"):        {output: []byte("main\n")},
			commandKey("git", "-C", "/main", "status", "--porcelain"):           {output: []byte("")},
			commandKey("git", "-C", "/main", "rev-parse", "--absolute-git-dir"): {output: []byte("/main/.git\n")},
			commandKey("git", "-C", "/new", "log", "-1", "--pretty=%s"):         {output: []byte("new\n")},
			commandKey("git", "-C", "/new", "status", "--porcelain"):            {output: []byte("")},
			commandKey("git", "-C", "/new", "rev-parse", "--absolute-git-dir"):  {output: []byte("/new/.git\n")},
			commandKey("git", "-C", "/old", "log", "-1", "--pretty=%s"):         {output: []byte("old\n")},
			commandKey("git", "-C", "/old", "status", "--porcelain"):            {output: []byte("")},
			commandKey("git", "-C", "/old", "rev-parse", "--absolute-git-dir"):  {output: []byte("/old/.git\n")},
		},
	}

	// /main is the newest by mod time, yet must still be pinned to the top.
	modTimes := map[string]time.Time{
		filepath.Join("/main/.git", "gitdir"): time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		filepath.Join("/new/.git", "gitdir"):  time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		filepath.Join("/old/.git", "gitdir"):  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	svc := service{runner: runner, stat: existingStat(modTimes)}

	got, err := svc.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	gotPaths := []string{}
	for _, w := range got {
		gotPaths = append(gotPaths, w.Path)
	}
	want := []string{"/main", "/old", "/new"}
	if !reflect.DeepEqual(gotPaths, want) {
		t.Fatalf("expected order %v (main pinned, then oldest-first), got %v", want, gotPaths)
	}
}

func TestServiceListSortsUndeterminableCreationTimesLast(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			// /main pinned first, /dated has a known creation time, and /stale
			// is prunable so its time is undeterminable (zero).
			commandKey("git", "worktree", "list", "--porcelain"): {
				output: []byte("worktree /main\nHEAD m\nbranch refs/heads/main\n\nworktree /stale\nHEAD s\nbranch refs/heads/stale\nprunable gitdir gone\n\nworktree /dated\nHEAD d\nbranch refs/heads/dated\n"),
			},
			commandKey("git", "-C", "/main", "log", "-1", "--pretty=%s"):         {output: []byte("main\n")},
			commandKey("git", "-C", "/main", "status", "--porcelain"):            {output: []byte("")},
			commandKey("git", "-C", "/main", "rev-parse", "--absolute-git-dir"):  {output: []byte("/main/.git\n")},
			commandKey("git", "-C", "/dated", "log", "-1", "--pretty=%s"):        {output: []byte("dated\n")},
			commandKey("git", "-C", "/dated", "status", "--porcelain"):           {output: []byte("")},
			commandKey("git", "-C", "/dated", "rev-parse", "--absolute-git-dir"): {output: []byte("/dated/.git\n")},
		},
	}

	modTimes := map[string]time.Time{
		filepath.Join("/main/.git", "gitdir"):  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		filepath.Join("/dated/.git", "gitdir"): time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	svc := service{runner: runner, stat: existingStat(modTimes)}

	got, err := svc.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	gotPaths := []string{}
	for _, w := range got {
		gotPaths = append(gotPaths, w.Path)
	}
	// Main pinned, the dated worktree next, and the stale (zero-time) one last.
	want := []string{"/main", "/dated", "/stale"}
	if !reflect.DeepEqual(gotPaths, want) {
		t.Fatalf("expected order %v, got %v", want, gotPaths)
	}
}

func TestServiceListSupportsDetachedHead(t *testing.T) {
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

	svc := service{runner: runner, stat: existingStat(nil)}
	got, err := svc.List()
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

func TestServiceListMarksStaleWorktreesWithoutInspectingPath(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "list", "--porcelain"): {
				output: []byte("worktree /repo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /repo-gone\nHEAD def456\nbranch refs/heads/feature/old\nprunable gitdir file points to non-existent location\n"),
			},
			commandKey("git", "-C", "/repo", "log", "-1", "--pretty=%s"): {
				output: []byte("Initial commit\n"),
			},
			commandKey("git", "-C", "/repo", "status", "--porcelain"): {
				output: []byte(""),
			},
		},
	}

	svc := service{runner: runner, stat: existingStat(nil)}
	got, err := svc.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	want := []Info{
		{
			Path:        "/repo",
			Branch:      "main",
			CommitLabel: "Initial commit",
			CommitHash:  "abc123",
		},
		{
			Path:       "/repo-gone",
			Branch:     "feature/old",
			CommitHash: "def456",
			Stale:      true,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected worktrees: got %#v want %#v", got, want)
	}

	// The stale worktree's path must never be inspected with git -C.
	for _, call := range runner.calls {
		for i, arg := range call.args {
			if arg == "-C" && i+1 < len(call.args) && call.args[i+1] == "/repo-gone" {
				t.Fatalf("unexpected git command against stale worktree path: %+v", call)
			}
		}
	}
}

func TestServiceListMarksLockedWorktreeLockedNotStale(t *testing.T) {
	// A locked worktree whose working directory is gone is reported by git as
	// "locked" but never "prunable" (locking protects it from pruning). It must
	// be surfaced as locked, not stale, and never inspected with git -C, which
	// would fail with exit status 128 and abort the whole listing.
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "list", "--porcelain"): {
				output: []byte("worktree /repo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /repo-gone\nHEAD def456\nbranch refs/heads/feature/locked\nlocked\n"),
			},
			commandKey("git", "-C", "/repo", "log", "-1", "--pretty=%s"): {
				output: []byte("Initial commit\n"),
			},
			commandKey("git", "-C", "/repo", "status", "--porcelain"): {
				output: []byte(""),
			},
			commandKey("git", "-C", "/repo", "rev-parse", "--absolute-git-dir"): {
				output: []byte("/repo/.git\n"),
			},
		},
	}

	svc := service{
		runner: runner,
		stat: func(name string) (fs.FileInfo, error) {
			if name == "/repo-gone" {
				return nil, errors.New("no such file or directory")
			}
			return fakeFileInfo{modTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}, nil
		},
	}

	got, err := svc.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	want := []Info{
		{
			Path:        "/repo",
			Branch:      "main",
			CommitLabel: "Initial commit",
			CommitHash:  "abc123",
			CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			Path:       "/repo-gone",
			Branch:     "feature/locked",
			CommitHash: "def456",
			Locked:     true,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected worktrees: got %#v want %#v", got, want)
	}

	// The locked worktree's path must never be inspected with git -C.
	for _, call := range runner.calls {
		for i, arg := range call.args {
			if arg == "-C" && i+1 < len(call.args) && call.args[i+1] == "/repo-gone" {
				t.Fatalf("unexpected git command against locked worktree path: %+v", call)
			}
		}
	}
}

func TestServiceListMarksMissingUnannotatedWorktreeStale(t *testing.T) {
	// A worktree whose working directory is gone but which git annotates neither
	// "prunable" nor "locked" (e.g. a git version predating the prunable
	// annotation) must still be treated as stale rather than inspected with
	// git -C, which would fail with exit status 128 and abort the whole listing.
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "list", "--porcelain"): {
				output: []byte("worktree /repo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /repo-gone\nHEAD def456\nbranch refs/heads/feature/old\n"),
			},
			commandKey("git", "-C", "/repo", "log", "-1", "--pretty=%s"): {
				output: []byte("Initial commit\n"),
			},
			commandKey("git", "-C", "/repo", "status", "--porcelain"): {
				output: []byte(""),
			},
			commandKey("git", "-C", "/repo", "rev-parse", "--absolute-git-dir"): {
				output: []byte("/repo/.git\n"),
			},
		},
	}

	svc := service{
		runner: runner,
		stat: func(name string) (fs.FileInfo, error) {
			if name == "/repo-gone" {
				return nil, errors.New("no such file or directory")
			}
			return fakeFileInfo{modTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}, nil
		},
	}

	got, err := svc.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	want := []Info{
		{
			Path:        "/repo",
			Branch:      "main",
			CommitLabel: "Initial commit",
			CommitHash:  "abc123",
			CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			Path:       "/repo-gone",
			Branch:     "feature/old",
			CommitHash: "def456",
			Stale:      true,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected worktrees: got %#v want %#v", got, want)
	}

	// The missing worktree's path must never be inspected with git -C.
	for _, call := range runner.calls {
		for i, arg := range call.args {
			if arg == "-C" && i+1 < len(call.args) && call.args[i+1] == "/repo-gone" {
				t.Fatalf("unexpected git command against missing worktree path: %+v", call)
			}
		}
	}
}

func TestServiceListMarksLockedWorktreeWithIntactDirectory(t *testing.T) {
	// A locked worktree whose directory is still present is reachable, so it is
	// inspected normally but still surfaced as locked rather than stale.
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "list", "--porcelain"): {
				output: []byte("worktree /repo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /pinned\nHEAD def456\nbranch refs/heads/pinned\nlocked\n"),
			},
			commandKey("git", "-C", "/repo", "log", "-1", "--pretty=%s"):   {output: []byte("Initial commit\n")},
			commandKey("git", "-C", "/repo", "status", "--porcelain"):      {output: []byte("")},
			commandKey("git", "-C", "/pinned", "log", "-1", "--pretty=%s"): {output: []byte("Pinned work\n")},
			commandKey("git", "-C", "/pinned", "status", "--porcelain"):    {output: []byte("")},
		},
	}

	svc := service{runner: runner, stat: existingStat(nil)}

	got, err := svc.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	want := []Info{
		{
			Path:        "/repo",
			Branch:      "main",
			CommitLabel: "Initial commit",
			CommitHash:  "abc123",
		},
		{
			Path:        "/pinned",
			Branch:      "pinned",
			CommitLabel: "Pinned work",
			CommitHash:  "def456",
			Locked:      true,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected worktrees: got %#v want %#v", got, want)
	}
}

func TestServiceListIgnoresUntrackedFilesForDirtyState(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "worktree", "list", "--porcelain"): {
				output: []byte("worktree /repo\nHEAD abc123\nbranch refs/heads/main\n"),
			},
			commandKey("git", "-C", "/repo", "log", "-1", "--pretty=%s"): {
				output: []byte("Initial commit\n"),
			},
			commandKey("git", "-C", "/repo", "status", "--porcelain"): {
				output: []byte("?? notes.txt\n"),
			},
		},
	}

	svc := service{runner: runner, stat: existingStat(nil)}
	got, err := svc.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(got))
	}
	if got[0].HasUncommittedChanges {
		t.Fatalf("expected untracked files to be ignored, got %#v", got[0])
	}
}
