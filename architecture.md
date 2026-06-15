# Architecture

## Overview

`grove` is a terminal UI for managing Git worktrees and branches.

Core user actions:

- browse existing worktrees
- filter and select a worktree
- add or remove worktrees
- browse local or remote-tracking branches
- switch, fetch, delete, and bulk-delete branches

The most important process-level rule is the shell-wrapper contract:

- success with a selection: print the selected path to `stdout`
- success with no selection or cancel: print nothing
- failure: print an error to `stderr` and exit non-zero

`grove` never changes the parent shell directory directly.

## Current Package Structure

### `main.go`

- parses startup flags
- verifies the current directory is inside a Git repo
- constructs services and `app.App`
- runs Bubble Tea
- prints the final selected path on success

### `app/`

- owns application state and screen flow
- owns loading, status, and confirmation semantics
- defines typed effects and typed async results
- contains no Bubble Tea dependencies

### `ui/`

- adapts `app.App` to Bubble Tea
- renders screens and shared UI components
- owns local widget state such as search text, selection, and dialog focus
- executes `app` effects via services and feeds results back into `app`

### `worktree/`

- runs Git worktree commands
- parses `git worktree list --porcelain`
- validates add/remove inputs
- enriches worktree data with latest commit subject and tracked-dirty state

### `branch/`

- lists local or remote-tracking branches
- tracks scope switching
- switches branches
- fetches branches
- deletes one or many local branches
- loads recent commits for the selected branch

## Runtime Flow

Startup flow:

1. `main.go` parses flags.
2. `main.go` creates services and checks `worktree.Service.InRepo()`.
3. `ui.New(app.New(...))` creates the Bubble Tea model.
4. `ui.Model.Init()` runs the initial `app` effect.

From there the app follows the same loop for all screens:

1. the UI sends user intent into `app`
2. `app` updates state and returns an effect value
3. `ui` executes that effect against a service
4. the service result is turned back into an `app.Result`
5. `app.HandleResult()` updates state and may return another effect

That effect/result loop is the core architectural pattern in the codebase.

## Screens

### Change Screen

Default screen.

Capabilities:

- load worktrees on startup
- filter by path substring
- move selection with `up`, `down`, `tab`, and `shift+tab`
- press `enter` to submit the selected worktree path and quit
- press `ctrl+a` to open add mode
- press `ctrl+b` to open branch mode
- press `ctrl+d` to remove the selected worktree

### Add Screen

Two-field form:

- relative path
- branch name

Submission flow:

1. validate required fields
2. check whether the branch already exists
3. add a worktree from the existing branch, or confirm creation of a new branch
4. reload worktrees after success

### Branch Screen

Branch-management screen.

Capabilities:

- browse local or remote-tracking branches
- filter branches by substring
- preview recent commits for the selected branch
- switch to the selected branch
- fetch branches
- delete one local branch
- delete all deletable local branches
- return to the change screen

## Services and Commands

### `worktree.Service`

Public surface:

- `Add(path, branch)`
- `AddNewBranch(path, branch)`
- `BranchExists(branch)`
- `InRepo()`
- `List()`
- `Remove(path)`

Git commands used include:

- `git rev-parse --is-inside-work-tree`
- `git worktree list --porcelain`
- `git worktree add`
- `git worktree remove`
- `git for-each-ref`
- `git -C <path> log -1 --pretty=%s`
- `git -C <path> status --porcelain`

`List()` enriches each worktree with:

- normalized branch name
- latest commit subject
- whether tracked file changes are present

### `branch.Service`

Public surface:

- `List()`
- `RecentCommits(name, limit)`
- `ToggleScope()`
- `Checkout(name)`
- `Delete(name)`
- `DeleteAllLocal()`
- `Fetch()`

Branch ordering and status are derived from Git refs, current branch state, worktree state, and reflog recency.

## UI Notes

The UI is one top-level Bubble Tea model with focused screen adapters and small reusable components:

- `textinput`
- `selectlist`
- `dialog`
- `loading`
- `status`

Rendering is manual string composition with `lipgloss` sizing and help/footer rendering.

## Testing

The repo is covered mostly by unit tests:

- `worktree/` tests validate parsing, Git command construction, and dirty-state enrichment
- `branch/` tests validate branch listing, ordering, scope switching, checkout, fetch, and deletion behavior
- `app/` tests validate state transitions, effects, loading, and dialog behavior
- `ui/` tests validate screen-level rendering and interaction logic

## Summary

The current split is:

- `main.go` for process concerns
- `app/` for business flow and state
- `ui/` for Bubble Tea integration and rendering
- `worktree/` and `branch/` for Git-facing behavior

That matches the code as it exists today and keeps the app small, testable, and easy to evolve.
