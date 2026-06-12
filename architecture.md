# Architecture

## Overview

`grove` is a terminal UI for browsing and managing Git worktrees.

Its main job is to let a user:

- view existing worktrees
- filter and select a worktree
- remove a worktree
- add a worktree from an existing branch or a new branch
- read `git worktree` documentation without leaving the app

The most important behavioral detail is that selecting a worktree does not directly change the user's shell directory. The Go process cannot mutate its parent shell, so `grove` uses a wrapper-friendly CLI contract:

- success with a selection: print the selected path to `stdout`
- success with no selection or cancel: print nothing
- failure: print an error to `stderr` and exit non-zero

Shell wrappers in `zsh`, `bash`, or PowerShell can capture `stdout` and run `cd` or `Set-Location` themselves.

## Design Guidance

These are the architectural decisions that should continue to shape changes to the codebase:

- keep UI interaction state inside the Bubble Tea model
- keep worktree domain operations, parsing, and input validation in `worktree/`
- allow UI-only process calls in `ui/` when they do not mutate repository state
- let `Update` interpret user intent instead of embedding side effects directly
- use `tea.Cmd` plus typed message types for Git and OS side effects
- keep rendering side-effect free
- preserve the `stdout` contract: selected path only on success, nothing on cancel, errors on `stderr`
- treat `worktree.WorktreeInfo` as canonical repository data and UI lists as derived state
- preserve selection by worktree path across reloads when possible
- keep one top-level Bubble Tea model with focused mode-specific sub-state
- keep package seams easy to unit test
- prefer a small number of clear packages over extra abstraction layers

Some additional guidance from the original design still applies:

- avoid over-decoupling simple local code
- do not introduce multiple interface layers unless they solve a real testing or ownership problem
- do not move trivial UI-only helpers out of `ui/` just for separation's sake
- avoid duplicating state across UI and domain layers

For new features, add repository operations to `worktree/` first and then expose them to the TUI through commands and messages. Only introduce a new package when there is a real ownership boundary.

The goal is a clear boundary around side effects and domain behavior, not maximum abstraction.

## Runtime Flow

At startup, `main.go` creates a `worktree.Manager`, verifies the current directory is inside a Git repository, and starts Bubble Tea in the alternate screen.

The app then follows this flow:

1. `ui.Model.Init()` returns a command that loads worktrees.
2. The command calls `Manager.List()`.
3. The `worktree` package shells out to Git, parses `git worktree list --porcelain`, and enriches each entry with branch, latest commit subject, and dirty state.
4. The UI stores the resulting `[]worktree.WorktreeInfo` and renders the list.
5. User input is handled by mode-specific update functions.
6. Mutating operations return `tea.Cmd` values, which perform the Git call and send a message back into the Bubble Tea loop.
7. When the user presses `enter` in change mode, the selected worktree path is stored in UI state and the program quits.
8. After Bubble Tea exits, `main.go` reads that stored path and prints it to `stdout`.

`main.go` stays intentionally thin. It is only responsible for process-level concerns and the final `stdout` handoff.

## Repository Structure

### `main.go`

- process entry point
- repository check via `Manager.InRepo()`
- Bubble Tea program startup
- final selected-path output

### `ui/`

- Bubble Tea model and state
- key handling
- async command helpers
- per-mode state transitions
- screen rendering

### `worktree/`

- Git subprocess execution
- parsing `git worktree` output
- validating add/remove inputs
- producing structured worktree data for the UI

### `dist/`

- prebuilt binaries for supported platform and architecture pairs

## UI Architecture

The UI is a single Bubble Tea model split across focused files.

### Model State

`ui.Model` contains shared application state plus three mode-specific sub-states:

- `ChangeState`: search query, selection, scroll, loaded worktrees, filtered items, submitted path
- `AddState`: focused field, path input, branch input, pending operation state, branch-creation confirmation state
- `DocsState`: rendered documentation lines and scroll position
- `StatusState`: either a status message or an error

The active screen is tracked by `Mode`:

- `ModeChange`
- `ModeAdd`
- `ModeDocs`

### Screen Modes

#### Change Worktree

This is the default mode.

Capabilities:

- load all worktrees on startup
- filter by substring against worktree paths
- move selection with `up` and `down` or `tab` and `shift+tab`
- press `enter` to submit the selected path and quit
- press `ctrl+d` to remove the selected worktree
- press `ctrl+a` to enter add mode
- press `?` to open docs mode

The list is rendered from `ChangeState.filtered`, while display labels are enriched with branch names from `ChangeState.worktrees`.

#### Add Worktree

Add mode is a two-field form:

- relative path
- branch name

Submission flow:

1. Validate that both fields are non-empty.
2. Check whether the branch already exists with `Manager.BranchExists()`.
3. If the branch exists, run `Manager.Add(path, branch)`.
4. If the branch does not exist, ask for confirmation.
5. On confirmation, run `Manager.AddNewBranch(path, branch)`.
6. Reload the worktree list after a successful add.

This mode is intentionally small and keeps the confirmation state inside `AddState` rather than introducing a separate modal abstraction.

#### Docs

Docs mode displays the output of:

```bash
git --no-pager help worktree
```

The output is cleaned before rendering:

- CRLF and CR are normalized to LF
- backspace overstrike sequences are stripped so manpage formatting does not render as garbage in the TUI

Docs can be closed with `esc`, `q`, or `?`.

## Command and Message Pattern

The app uses the normal Bubble Tea side-effect pattern:

- `loadWorktreesCmd`
- `checkBranchExistsCmd`
- `addWorktreeCmd`
- `removeWorktreeCmd`
- `openWorktreeDocsCmd`

Those commands return message structs:

- `worktreesLoadedMsg`
- `branchCheckedMsg`
- `worktreeAddedMsg`
- `worktreeRemovedMsg`
- `worktreeDocsLoadedMsg`

This keeps event handling readable:

- `Update` routes by message type or by current mode
- subprocess work is done outside the render path
- results are fed back into state synchronously through messages

Most Git behavior lives in `worktree/`. The one exception is docs loading in `ui/commands.go`, which directly runs `git help worktree` because it is UI-only text retrieval rather than worktree domain logic.

## Worktree Package

The `worktree` package is the Git-facing service layer.

### Public Surface

`Manager` exposes:

- `Add(path, branch)`
- `AddNewBranch(path, branch)`
- `BranchExists(branch)`
- `InRepo()`
- `List()`
- `Remove(path)`

`WorktreeInfo` is the domain object returned to the UI:

- `Path`
- `Branch`
- `CommitLabel`
- `CommitHash`
- `HasUncommittedChanges`

### Git Commands Used

`grove` currently depends on these Git commands:

- `git rev-parse --is-inside-work-tree`
- `git worktree list --porcelain`
- `git worktree add <path> <branch>`
- `git worktree add -b <branch> <path>`
- `git worktree remove <path>`
- `git for-each-ref --format=%(refname:short) refs/heads/<branch>`
- `git -C <path> log -1 --pretty=%s`
- `git -C <path> status --porcelain`

### Listing and Enrichment

`Manager.List()` is more than a raw passthrough.

It:

1. reads `git worktree list --porcelain`
2. parses each block into an internal `listEntry`
3. normalizes branch names by trimming `refs/heads/`
4. labels detached HEAD worktrees as `detached`
5. queries each worktree for its latest commit subject
6. queries each worktree for uncommitted changes

That means the UI renders richer information than Git's base listing alone.

### Subprocess Boundary

The package uses a small `Runner` interface with one method:

```go
CombinedOutput(name string, args ...string) ([]byte, error)
```

Production code uses `commandRunner`, which wraps `os/exec`. Tests use stub runners. Failed commands return errors that include trimmed subprocess output when available, which makes Git failures easier to surface in the UI.

## Rendering

Rendering is manual string composition rather than a heavy component tree.

Notable details:

- the app uses fixed horizontal and vertical padding
- `lipgloss.Width` is used to fit lines correctly
- ANSI escape sequences are used for placeholder and selection colors
- Charm's `help` component renders the footer key hints
- header, body, footer, and status lines are composed separately

The current visual layout is intentionally simple and text-first.

## State and Selection Behavior

Selection and filtering in change mode are path-based, not index-based alone.

Important behaviors:

- filtering is case-insensitive substring matching
- `selectedItem` preserves the selected path across list reloads when possible
- scroll position is clamped to visible rows
- when no items match, the UI renders `No matches`
- successful reloads clear any previously submitted change-path so stale output is not printed later

## Error Handling

There are two main error channels:

- fatal startup or Bubble Tea errors go to `stderr` in `main.go` and exit the process
- in-app operation errors are stored in `StatusState.err` and rendered in the footer area

The repo check is intentionally strict. If `git rev-parse --is-inside-work-tree` fails or returns anything other than `true`, `main.go` exits with `ErrNotGitRepo`.

## Testing Coverage

The test suite is mostly unit-level and focuses on behavior at package seams.

### `worktree/` tests

- command construction for add, add-with-new-branch, remove, repo check, and branch existence
- parser behavior for porcelain output
- detached HEAD normalization
- structured list output including commit subject and dirty-state enrichment

### `ui/` tests

- pressing `enter` in change mode submits the selected path
- submitting with no selection reports a status message
- worktree reload clears submitted output
- `?` opens docs mode
- docs mode returns to change mode and scrolls correctly

### `main` test

- verifies that the selected-path export helper returns an empty string for a fresh model

There are no end-to-end terminal tests yet, so the integration between Bubble Tea, real Git repositories, and shell wrappers is still validated primarily by manual use plus package-level tests.

## Operational Constraints

The app is built around a few non-negotiable constraints:

- the standalone binary cannot change the parent shell directory
- `stdout` must stay clean because wrappers depend on it
- Git subprocesses are the source of truth for repository state
- the app must be run inside a Git working tree

Because of that, the selected-path output contract is part of the architecture, not just a convenience detail.

## Summary

The current architecture is straightforward:

- `main.go` handles process startup, repo validation, and final output
- `ui/` owns Bubble Tea state, modes, rendering, and user interaction
- `worktree/` owns Git worktree operations, parsing, and validation
- the shell wrapper is responsible for converting selected-path `stdout` into an actual directory change

That split matches the code as it exists today and keeps the app small, testable, and easy to extend.
