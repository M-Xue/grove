package app

import "testing"

func TestIsBusyTrueForBlockingLoading(t *testing.T) {
	a := New(Services{})
	a.setBlockingLoading("deleting branch")
	if !a.State().IsBusy() {
		t.Fatal("expected IsBusy true while a blocking op is in flight")
	}
}

func TestIsBusyFalseForPassiveLoading(t *testing.T) {
	a := New(Services{})
	a.setLoading("loading worktrees")
	if a.State().IsBusy() {
		t.Fatal("expected IsBusy false for a passive load")
	}
}

func TestIsBusyTrueForProgressLoading(t *testing.T) {
	a := New(Services{})
	a.setProgressLoading("creating worktree")
	if !a.State().IsBusy() {
		t.Fatal("expected IsBusy true while a progress op is in flight")
	}
}

func TestIsBusyFalseOnceBlockingDone(t *testing.T) {
	a := New(Services{})
	id := a.setBlockingLoading("fetching branches")
	a.markLoadingDone(id)
	if a.State().IsBusy() {
		t.Fatal("expected IsBusy false once the blocking op completes")
	}
}

// TestOperationBlockingClassification pins which operations freeze the UI. It
// drives the public entry points and asserts the Blocking flag on the entry each
// one creates, so a future re-categorization can't silently change the lock.
func TestOperationBlockingClassification(t *testing.T) {
	tests := []struct {
		name     string
		start    func(*App)
		blocking bool
	}{
		{"checkout", func(a *App) { a.RequestCheckoutBranch("feature/a") }, true},
		{"delete branch", func(a *App) { a.DeleteBranch("feature/a") }, true},
		{"delete all branches", func(a *App) { a.DeleteAllBranches() }, true},
		{"fetch branches", func(a *App) { a.RequestFetchBranches() }, true},
		{"remove worktree", func(a *App) { a.RemoveWorktree("/repo") }, true},
		{"select branch (commits)", func(a *App) { a.SelectBranch("feature/a") }, false},
		{"toggle scope", func(a *App) { a.RequestToggleBranchScope() }, false},
	}

	for _, test := range tests {
		a := New(Services{})
		test.start(a)
		loading := a.State().Loading
		if len(loading) != 1 {
			t.Fatalf("%s: expected exactly one loading entry, got %#v", test.name, loading)
		}
		if loading[0].Blocking != test.blocking {
			t.Fatalf("%s: expected Blocking=%v, got %#v", test.name, test.blocking, loading[0])
		}
	}
}
