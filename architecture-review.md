# Grove Architecture Review & Remediation Plan

> Full-codebase review, 2026-09-27. Revised the same day after the scope
> reduction that removed the branch screen (and the `branch` package) and
> merged the add screen into the change screen as a dialog. Findings that
> lived entirely in deleted code (the branch-scope data race, the `-a` empty
> screen bug, the bulk-delete `-D` mismatch, the stateful branch service) are
> resolved by deletion and no longer listed. Everything below was re-verified
> against the current tree.

## Overall verdict

The layering (`app` as a framework-free state machine, `ui` as the only Bubble
Tea layer, the worktree service behind an injectable runner) is real, not
aspirational: `app` never imports Bubble Tea, and the change screen consumes
`app` through a narrow interface it defines itself. The multi-screen plumbing
(ScreenID routing, Registry/Mode, OnMessage delivery) is deliberately retained
even though only one screen remains. The remaining weaknesses are in the
seams: the app↔ui synchronization protocol does redundant work, the "pure
state machine" claim is undermined by slice aliasing, and the git layer has an
N+1 subprocess problem plus a parsing-robustness flaw. Two genuine defects
remain (Part 1).

---

## Part 1 — Concrete defects (fix regardless of any refactor)

### D1. Add-worktree pre-check defeats git's remote DWIM
- `worktree/service.go:132-148` (`BranchExists`) checks only `refs/heads`. For
  a branch that exists only on a remote, plain `git worktree add` would create
  a tracking branch; grove instead reports "branch does not exist" and the add
  dialog offers to create a *new unrelated branch at HEAD*.
- **Fix:** extend the existence check to remotes and add a third semantic
  outcome ("exists on remote") → offer "create tracking branch".

### D2. `EnsureInRepo` misreports every failure
- `repo/repo.go:21-30` maps *any* error — including git not installed — to
  "current directory is not a git repository".
- **Fix:** distinguish exec-not-found (and other runner errors) from the
  actual not-a-repo case.

---

## Part 2 — Structural / system-level findings

### S1. The message loop has one delivery model too many
Messages flow through three mechanisms: `HandleMessage` (domain state),
`OnMessage` (screen presentation reactions), and `Sync` (blanket pull-based
reconciliation after every event — `ui/model.go`).

- `Sync` runs 2–3 times per keystroke, including from `View()` — a side effect
  in the render path.
- It rebuilds the change screen's item slice on every call.
- The screen must keep `Sync` idempotent and hand-preserve ephemeral state.
- Exactly one `OnMessage` reaction exists (`BranchAbsentMessage` → the
  create-branch confirm dialog); the machinery is generic, the usage singular.

**Direction:** commit to one model. Either push (the screen reacts to specific
messages; no blanket Sync) or pull (drop `OnMessage`; derive dialog-opening
from state). Don't keep paying for both. Related watch item: `deliverMessage`
routes by the *currently active* screen, which is trivially safe with one
screen but becomes a latent trap the moment a second screen returns — if
`OnMessage` survives, route by correlation (the LoadingID machinery already
provides unique IDs) rather than by focus.

### S2. Nothing guards against out-of-order list loads
The branch feature's `branchCommitSeq` supersession guard was deleted with it,
so no message family has one. `WorktreesLoadedMessage` is triggered from
several flows (init, add completes, remove, prune), all non-blocking, so an
out-of-order completion can overwrite a newer list with an older one — and
then persist the stale list to the disk cache via the saver hook
(`app/handle.go:26-29`).

**Direction:** a seq (or generation) guard on worktree loads, ideally
generalized into the loading-entry machinery rather than reinvented per
message type.

### S3. `State` is copied by value but its slices alias
`State()` returns a copy, yet `DismissCompletedLoading` (`app/app.go:78`) and
`clearLoadingEntry` (`app/state.go:91`) filter with the `s[:0]` in-place
trick, mutating the backing array that any previously-taken `State` snapshot
still points into. Benign today (single-threaded main loop, no one holds old
snapshots), but it quietly breaks `app`'s core promise of inspectable,
snapshot-able state.

**Direction:** copy-on-filter (allocate a fresh slice), or document that a
`State` value is only valid until the next `HandleMessage`.

### S4. Streaming add is a second concurrency protocol with a special case
`WorktreeAddStartedMessage` is documented as "never reaches HandleMessage"
(`app/message.go`) and is special-cased in the UI loop (`ui/model.go`). One
exception is tolerable; a second streaming operation should force a
generalization (every Command yields a message *sequence*, single-element in
the common case) rather than another bypass.

---

## Part 3 — Subsystem notes

### app
- `LoadingEntry.Active` is always true — dead field (`app/types.go`, only read
  at `app/app.go:80`).
- Status IDs derive from `len(statuses)+1` (`app/status.go:32`) while loading
  IDs use a monotonic counter — inconsistent; use the counter pattern.
- `WorktreesLoadedMessage` clearing `SubmittedPath` (`app/handle.go:15`) is an
  unexplained coupling; document or remove.

### ui
- **Best idea in the layer:** the key registry (`ui/screens/common.go`) —
  footer and dispatch derived from one ordered source, per mode. The add
  dialog's `ModeAdd` slots into it cleanly, and the footer hints stay accurate
  by construction. Keep it.
- Dialogs are now fully compositional: `dialog` is chrome only (`Frame`, built
  on `panel.Render`), the button row is its own widget
  (`ui/components/buttongroup`), and the confirm and add dialogs both compose
  chrome + widgets in the screen layer. Dialog text is unstyled so it matches
  the app's default-foreground text. The right shape for any future dialog.
- `keys.Normalize` (`ui/keys/keys.go`) restates `msg.String()` case by case;
  `Key(msg.String())` plus the constants is behaviorally identical with zero
  maintenance. Unused constants (`KeyJ`, `KeyK`, `KeyQ`, `KeyLeft`,
  `KeyRight`, `KeyBackspace`) show drift already.
- Layout is hand-rolled string math (`fitLine`, `overlayLine`, manual padding
  in `ui/model.go`). Works, but it's the highest-maintenance code in the repo;
  lipgloss `JoinHorizontal`/`Place` would eliminate most of it.
- **Theming is partially centralized:** `ui/theme` now holds the chrome colors
  (active/inactive borders, title tab) shared by panel and dialog, but the
  component-local accents are still scattered magic numbers — `lipgloss.Color`
  values in textinput/buttongroup/loading/progressbar/status and raw
  `\x1b[38;5;…m` literals in selectlist, status, and the change screen's stale
  dimming — with no fallback for non-256-color terminals. Finish the migration
  into `ui/theme`.
- `textinput.filterASCII` (`ui/components/textinput/model.go:121-129`)
  silently discards non-ASCII input — unicode branch names can't be typed —
  and the byte-indexed cursor math is safe *only because* of that filter (an
  invisible coupling). If deliberate, comment it loudly; otherwise move to
  rune-indexed editing with proper width handling.
- Completed loading entries show `[DONE]` until the next keypress — if the
  user never types, it stays forever. Consider a timed dismiss.

### services / git layer
- Parsing split into a pure, table-tested `parse.go` is exemplary; the
  stale/locked worktree handling and its comments in `worktree/service.go` are
  the kind of code that survives maintainers.
- **N+1 subprocesses:** `worktree.List()` runs `log`, `status --porcelain`,
  and `rev-parse` per worktree, sequentially (`worktree/service.go:182-200`) —
  O(3n) process spawns; `status --porcelain` alone can be seconds per worktree
  on a big repo. The SWR cache treats the symptom. Options: run the
  per-worktree probes concurrently (they're independent), and/or degrade
  gracefully (paint paths/branches first, fill dirty-state later).
- **`CombinedOutput` feeds parsers** (`command/command.go:44-62`): stdout and
  stderr are merged, so any git warning/advice line (ownership warnings,
  advice hints) lands in the input of `parseWorktreeList` and `BranchExists`'s
  string equality. Machine-parsed commands should capture stdout alone,
  keeping stderr for the error message.
- Vestigial API surface: `worktree.Service.Add` / `AddWithNewBranch` are used
  only by tests (production uses `AddWithProgress` only —
  `app/operations.go`); `selectlist.VisibleItems` has no callers at all. Trim.

### cache
- Well designed: non-load-bearing by construction, atomic rename writes,
  hashed keys, honest tests.
- One watch-item: the seed includes *decision inputs*
  (`HasUncommittedChanges` drives plain vs. force delete in
  `ui/screens/change.go`), not just paint data. Today's failure modes
  self-correct (git refuses a non-force delete of a dirty tree), but any
  future action trusting seeded flags without git re-checking becomes a
  stale-data hazard. Consider stripping dirty flags from what's persisted, or
  blocking destructive actions until the first live load.

### testing
- Services and app: well covered. The add-dialog tests
  (`ui/screens/adddialog_test.go`) assert real intents through a recording
  fake — the right pattern; the change screen's other actions are still
  exercised only against an all-no-op `fakeApp`, verifying they don't crash
  rather than that keystrokes produce the right intents.
- **Highest-value missing test:** an integration loop — real `app` with a fake
  worktree service, drive `ui.Model.Update` with key events, assert on
  rendered output. That's exactly the layer where S1–S2 live, and it has no
  coverage. (The existing `ui` tests cover helpers, footers, and registry
  dispatch, but not the full update loop.)

---

## Part 4 — Prioritized plan

Each step should leave `go build ./... && go vet ./... && go test ./...` green.

### Phase 1 — Defects (small, independent; do first)
1. **D2** Distinguish git-missing from not-a-repo.
2. **D1** Remote-aware branch existence check + "create tracking branch" flow.

### Phase 2 — Structural fixes with best payoff/risk
3. **Runner split** (git layer): add a stdout-only runner method for
   machine-parsed commands; keep stderr for error text. Migrate
   `worktree.List` and `BranchExists` to it.
4. **S2** Supersession guard for worktree-list loads (protects the cache from
   stale writes too).

### Phase 3 — Tests before the seam rework
5. Write the ui↔app integration tests (fake service, key events in, rendered
   frames out). These are the net under everything in Phase 4.

### Phase 4 — Seam and polish
6. **S1** Consolidate the sync/OnMessage protocol (pick push or pull;
   correlation-based routing if OnMessage survives).
7. **S3** Copy-on-filter for `State` slices (or document the invariant).
8. `ui/theme` package; replace scattered ANSI literals and magic numbers.
9. Simplify `keys.Normalize`; delete unused key constants.
10. Trim vestigial surface: `worktree.Add`/`AddWithNewBranch`,
    `selectlist.VisibleItems`, `LoadingEntry.Active`; counter-based status
    IDs.
11. Optional: parallelize per-worktree probes in `worktree.List()`;
    rune-based textinput editing; timed dismiss for `[DONE]` entries.
