package screens

import "github.com/M-Xue/grove/app"

// fakeApp satisfies changeApp. Its methods are no-ops so screens can be
// constructed and driven in tests without a real app.
type fakeApp struct{}

func (fakeApp) RequestSubmitSelectedPath(string) app.Command    { return nil }
func (fakeApp) RemoveWorktree(string) app.Command               { return nil }
func (fakeApp) ForceRemoveWorktree(string) app.Command          { return nil }
func (fakeApp) PruneWorktrees() app.Command                     { return nil }
func (fakeApp) Quit() app.Command                               { return nil }
func (fakeApp) RequestAddWorktree(string, string) app.Command   { return nil }
func (fakeApp) CreateBranchWorktree(string, string) app.Command { return nil }
