package worktree

import (
	"errors"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// twoWorktreeResults scripts the git calls List makes for a main worktree at
// /repo and a linked one at /repo-feature, with statusArgs as the status
// invocation (so tests can script the lock-free variant).
func twoWorktreeResults(statusPrefix ...string) map[string]commandResult {
	status := func(path string) string {
		args := append(append([]string{}, statusPrefix...), "-C", path, "status", "--porcelain")
		return commandKey("git", args...)
	}
	return map[string]commandResult{
		commandKey("git", "worktree", "list", "--porcelain"): {
			output: []byte("worktree /repo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /repo-feature\nHEAD def456\nbranch refs/heads/feature\n"),
		},
		commandKey("git", "-C", "/repo", "log", "-1", "--pretty=%s"):                {output: []byte("Initial commit\n")},
		commandKey("git", "-C", "/repo", "rev-parse", "--absolute-git-dir"):         {output: []byte("/repo/.git\n")},
		commandKey("git", "-C", "/repo-feature", "log", "-1", "--pretty=%s"):        {output: []byte("Feature work\n")},
		commandKey("git", "-C", "/repo-feature", "rev-parse", "--absolute-git-dir"): {output: []byte("/repo/.git/worktrees/repo-feature\n")},
		status("/repo"):         {output: []byte("")},
		status("/repo-feature"): {output: []byte(" M main.go\n")},
	}
}

func TestServiceListBackgroundRunsStatusWithoutOptionalLocks(t *testing.T) {
	runner := &stubRunner{results: twoWorktreeResults("--no-optional-locks")}
	svc := service{runner: runner, stat: existingStat(nil)}

	got, err := svc.List(ListOptions{Background: true})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(got) != 2 || !got[1].HasUncommittedChanges {
		t.Fatalf("expected the lock-free status to be parsed normally, got %#v", got)
	}
	for _, call := range runner.calls {
		if len(call.args) > 0 && call.args[0] == "--no-optional-locks" {
			// --no-optional-locks is a global option: it must precede -C.
			if call.args[1] != "-C" || !strings.Contains(strings.Join(call.args, " "), "status --porcelain") {
				t.Fatalf("expected --no-optional-locks -C <path> status --porcelain, got %v", call.args)
			}
			return
		}
	}
	t.Fatal("expected a background listing to run git status with --no-optional-locks")
}

func TestServiceListForegroundRunsStatusWithLocks(t *testing.T) {
	runner := &stubRunner{results: twoWorktreeResults()}
	svc := service{runner: runner, stat: existingStat(nil)}

	if _, err := svc.List(ListOptions{}); err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	for _, call := range runner.calls {
		for _, arg := range call.args {
			if arg == "--no-optional-locks" {
				t.Fatalf("a foreground listing must let git status refresh the index, got %v", call.args)
			}
		}
	}
}

func TestServiceListProbesWorktreesConcurrently(t *testing.T) {
	// Three reachable worktrees whose status calls all block until every one
	// of them has started. Sequential probing would deadlock here, so the
	// barrier being released proves the probes overlap. The result must still
	// come back in git's order.
	results := map[string]commandResult{
		commandKey("git", "worktree", "list", "--porcelain"): {
			output: []byte("worktree /a\nHEAD 1\nbranch refs/heads/a\n\nworktree /b\nHEAD 2\nbranch refs/heads/b\n\nworktree /c\nHEAD 3\nbranch refs/heads/c\n"),
		},
	}
	for _, path := range []string{"/a", "/b", "/c"} {
		results[commandKey("git", "-C", path, "log", "-1", "--pretty=%s")] = commandResult{output: []byte(path + "\n")}
		results[commandKey("git", "-C", path, "status", "--porcelain")] = commandResult{}
		results[commandKey("git", "-C", path, "rev-parse", "--absolute-git-dir")] = commandResult{output: []byte(path + "/.git\n")}
	}
	arrived := make(chan struct{}, 3)
	release := make(chan struct{})
	runner := &stubRunner{
		results: results,
		onCall: func(_ string, args []string) {
			if len(args) >= 3 && args[2] == "status" {
				arrived <- struct{}{}
				<-release
			}
		},
	}
	svc := service{runner: runner, stat: existingStat(nil)}

	type outcome struct {
		list []Info
		err  error
	}
	done := make(chan outcome, 1)
	go func() {
		list, err := svc.List(ListOptions{})
		done <- outcome{list, err}
	}()

	timeout := time.After(2 * time.Second)
	for i := 0; i < 3; i++ {
		select {
		case <-arrived:
		case <-timeout:
			t.Fatalf("only %d of 3 status probes started within 2s: probes are not running concurrently", i)
		}
	}
	close(release)

	result := <-done
	if result.err != nil {
		t.Fatalf("List returned error: %v", result.err)
	}
	var paths []string
	for _, w := range result.list {
		paths = append(paths, w.Path)
	}
	if want := []string{"/a", "/b", "/c"}; !reflect.DeepEqual(paths, want) {
		t.Fatalf("expected git's order %v preserved, got %v", want, paths)
	}
}

func TestServiceListReportsProbeFailure(t *testing.T) {
	results := twoWorktreeResults()
	results[commandKey("git", "-C", "/repo-feature", "status", "--porcelain")] = commandResult{err: errors.New("index corrupt")}
	runner := &stubRunner{results: results}
	svc := service{runner: runner, stat: existingStat(nil)}

	if _, err := svc.List(ListOptions{}); err == nil || !strings.Contains(err.Error(), "index corrupt") {
		t.Fatalf("expected the failing probe's error, got %v", err)
	}
}

func TestServiceListMainWorktreeWithoutGitdirFileHasZeroCreatedAt(t *testing.T) {
	// The main worktree's git directory has no `gitdir` file. Its directory
	// mod time must not be used in its place: that moves on every index
	// write, which would make the list look changed on every poll.
	runner := &stubRunner{results: twoWorktreeResults()}
	svc := service{
		runner: runner,
		stat: func(name string) (fs.FileInfo, error) {
			switch name {
			case filepath.Join("/repo/.git", "gitdir"):
				return nil, errors.New("no such file or directory")
			case "/repo/.git":
				return fakeFileInfo{modTime: time.Date(2026, 10, 27, 10, 35, 0, 0, time.UTC)}, nil
			case filepath.Join("/repo/.git/worktrees/repo-feature", "gitdir"):
				return fakeFileInfo{modTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}, nil
			}
			return fakeFileInfo{}, nil
		},
	}

	got, err := svc.List(ListOptions{})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if !got[0].CreatedAt.IsZero() {
		t.Fatalf("expected the main worktree's CreatedAt to be zero, got %v", got[0].CreatedAt)
	}
	if got[1].CreatedAt.IsZero() {
		t.Fatal("expected the linked worktree's CreatedAt from its gitdir file")
	}
}

func TestInfoEqualIgnoresTimeLocation(t *testing.T) {
	instant := time.Date(2026, 8, 18, 10, 0, 0, 0, time.UTC)
	fresh := Info{Path: "/repo", Branch: "main", CreatedAt: instant.In(time.Local)}
	cached := Info{Path: "/repo", Branch: "main", CreatedAt: instant.In(time.FixedZone("AEDT", 11*3600))}
	if !fresh.Equal(cached) {
		t.Fatal("expected Infos differing only in time.Location to be Equal")
	}
	if !InfosEqual([]Info{fresh}, []Info{cached}) {
		t.Fatal("expected InfosEqual to use Info.Equal")
	}
	changed := cached
	changed.HasUntrackedFiles = true
	if InfosEqual([]Info{fresh}, []Info{changed}) {
		t.Fatal("expected a changed flag to be detected")
	}
	if InfosEqual([]Info{fresh}, nil) {
		t.Fatal("expected lists of different length to differ")
	}
}
