package screens

import "github.com/M-Xue/grove/app"

// fakeApp satisfies changeApp, addApp, and branchApp. Its methods are no-ops so
// screens can be constructed and driven in tests without a real app.
type fakeApp struct{}

func (fakeApp) RequestSubmitSelectedPath(string) app.Command    { return nil }
func (fakeApp) OpenAdd()                                        {}
func (fakeApp) OpenBranch() app.Command                         { return nil }
func (fakeApp) RemoveWorktree(string) app.Command               { return nil }
func (fakeApp) PruneWorktrees() app.Command                     { return nil }
func (fakeApp) Quit() app.Command                               { return nil }
func (fakeApp) RequestAddWorktree(string, string) app.Command   { return nil }
func (fakeApp) CloseAdd()                                       {}
func (fakeApp) CreateBranchWorktree(string, string) app.Command { return nil }
func (fakeApp) RequestCheckoutBranch(string) app.Command        { return nil }
func (fakeApp) RequestFetchBranches() app.Command               { return nil }
func (fakeApp) RequestToggleBranchScope() app.Command           { return nil }
func (fakeApp) CloseBranch() app.Command                        { return nil }
func (fakeApp) SelectBranch(string) app.Command                 { return nil }
func (fakeApp) DeleteBranch(string) app.Command                 { return nil }
func (fakeApp) CanDeleteAllBranches() bool                      { return false }
func (fakeApp) DeleteAllBranches() app.Command                  { return nil }
