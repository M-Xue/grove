# Grove Refactor Plan

## Purpose & context (read this first)

`grove` is a keyboard-first terminal UI (Bubble Tea) for managing Git worktrees and
branches. Its architecture is layered:

- **`main.go`** — process entry: flag parsing, repo check, service construction, runs Bubble Tea, prints the selected path (the shell-wrapper contract).
- **`app/`** — a *pure, framework-free state machine*. Today it drives an **effect/result loop**: intent methods mutate state and return a typed `Effect`; `ui` executes it and feeds back a typed `Result`; `HandleResult` updates state and may chain another `Effect`.
- **`ui/`** — Bubble Tea integration: screens, reusable components, rendering, and the only place that touches services/git.
- **`worktree/`, `branch/`** — Git-facing services behind interfaces with an injectable `Runner`.

This refactor has **two goals**:

1. **Pay down structural debt** — thin out `main.go`, remove duplication (command runner), fix naming inconsistencies (`manager` remnants, import-alias collisions), relocate startup-only logic (`InRepo`, shell-init, arg parsing), and split pure helpers into their own tested files.
2. **Re-architect the `app`↔`ui` boundary** — replace the effect/result loop with a thinner, Bubble-Tea-decoupled **command/message** system; introduce a **per-screen key-binding registry**; **invert dialog ownership** to the UI; and wrap `app` behind **per-screen consumer interfaces**.

### Guiding principles (preserve these throughout)

- **`app` never imports Bubble Tea.** Its decoupling and testability are the codebase's biggest asset.
- **`ui` owns presentation and ephemeral state** (search text, selection, focus, dialog representation). `app` owns domain state and operations.
- **All git access goes through one injectable runner.**
- **Every step is independently shippable**: `go build ./...`, `go vet ./...`, and `go test ./...` must be green after each step, and structural steps must not change runtime behavior.

### How to work this plan

Phases **A → B → C** must run in order. Within **Phase A** steps are largely independent. **Phase B** steps are interdependent and must follow the stated order. **Phase C** (UI tests) runs last, after all other refactors. Each step below lists **Goal / Actions / Files / Depends on / Acceptance**.

---

## Phase A — Structural cleanups (low risk, behavior-preserving)

### A1. Extract shell-init into its own package

**Goal:** thin `main.go`; keep shell-init *in the binary* (it must self-locate via `os.Executable()` and stay versioned with the binary — do **not** move it into static script files), but isolate it in a decoupled package.

**Actions:**
- Create package `shellinit` (e.g. `shellinit/shellinit.go`).
- Move `shellInitScript`, `shellQuote`, and `isSupportedShell` (the shell-name validation) out of `main.go` into it.
- Expose a minimal API, e.g. `shellinit.Script(shell, binaryPath string) (string, error)` and `shellinit.IsSupported(shell string) bool`.
- Keep `os.Executable()` resolution in the caller (`cli`/`main`), passing the path in — the package stays pure (no process/global state).

**Files:** new `shellinit/`; edits to `main.go` (later moved behind `cli` in A2).

**Depends on:** none.

**Acceptance:** `grove shell-init bash|zsh|powershell` output is byte-identical to before; `shellinit` is imported only by `cli`/`main`; existing `TestShellInitScript` (moved) passes; no other package imports `shellinit`.

### A2. Extract argument parsing into a `cli` package

**Goal:** remove CLI grammar from `main.go`. Do **not** put it in `app` (keeps `app` free of `flag` and invocation concerns).

**Actions:**
- Create package `cli`.
- Move `parseCommand`, `parseInitialScreen`, `command`/`commandKind` types into it; have it own the `shell-init` subcommand dispatch and call `shellinit` (A1).
- `cli` depends on `app` only for the `app.ScreenID` type it returns.
- `main.go` becomes: `cli.Parse(os.Args[1:])` → switch on command kind → (`shellinit` path | run TUI).

**Files:** new `cli/`; `main.go` slimmed; move `main_test.go` parsing tests into `cli`.

**Depends on:** A1.

**Acceptance:** `main.go` contains only process wiring (parse → branch → construct → run → print path); all moved tests pass under `cli`; flag/behavior unchanged.

### A3. Resolve the `branch` import-alias collision via variable renames

**Goal:** remove the `branchsvc`/`branchService` alias **without renaming packages** (short singular package names like `branch`/`worktree` are idiomatic; plural or `…svc` package names are not).

**Actions:**
- Keep package names `branch` and `worktree`.
- Rename colliding **local variables** named `branch` (e.g. to `br`, `branchInfo`) so the package can be imported unaliased.
- Apply consistently anywhere a local variable shadows a service package name.

**Files:** `main.go`, and any site looping over branches with a `branch` local (e.g. `app/branch.go`, `app/docs.go`, `ui/screens/branch.go`).

**Depends on:** none (do before/independent of A2's main rewrite to reduce churn).

**Acceptance:** no import aliases for `branch`/`worktree` remain; build/tests green.

### A4. Unify the command runner into one shared, generic, injected package

**Goal:** remove the duplicated `commandRunner` (identical in `worktree/command_runner.go` and `branch/service.go`); one generic runner, injected.

**Actions:**
- Create a shared package (suggest `internal/command`) exporting a generic concrete runner implementing `CombinedOutput(name string, args ...string) ([]byte, error)` (the existing error-wrapping behavior preserved).
- **Add `context.Context` + a timeout** at this single chokepoint (use `exec.CommandContext`) — this is the one place to fix "git can hang forever / no cancellation."
- **Keep the `Runner` interface defined at each consumer** (`worktree` and `branch` each declare their own one-method interface; the shared concrete type satisfies both). This preserves the "interface at the consumer" idiom and keeps services decoupled from the shared package's surface.
- Change constructors to require an injected runner; construct one runner in `main`/`cli` and inject into both services. Keep a test-only constructor that accepts a stub.

**Files:** new `internal/command/`; delete `worktree/command_runner.go`; remove the `commandRunner` block from `branch/service.go`; edit `worktree/manager.go` (→ A6), `branch/service.go`, `main.go`.

**Depends on:** none (foundational; A5 builds on it).

**Acceptance:** exactly one `CombinedOutput` implementation in the tree; both services receive the runner via injection; existing stub-runner tests still pass; a timeout is enforced on real git calls.

### A5. Extract `InRepo` out of the worktree package

**Goal:** `InRepo` is a startup precondition, not a worktree operation; remove it from the `Service` interface.

**Actions:**
- Replace `worktree.Service.InRepo()` with a single standalone function using the shared runner — e.g. `command.EnsureInRepo(runner) error` (or a tiny `repo` helper).
- Call it once at startup in `cli`/`main`; preserve the `ErrNotGitRepo` semantics and the stderr + exit-1 UX.
- Remove `InRepo` from the `worktree.Service` interface.

**Files:** `internal/command/` (or `repo/`), `worktree/manager.go` (→ A6), `main.go`/`cli`.

**Depends on:** A4.

**Acceptance:** `worktree.Service` no longer exposes `InRepo`; startup still errors out cleanly outside a repo; tests cover the standalone check.

### A6. Rename `manager` remnants in `worktree` → `service`

**Goal:** consistency with `branch` (which uses a `service` struct); the public surface is already `Service`/`NewService`.

**Actions:**
- Rename the `manager` struct → `service`.
- Rename files: `manager.go` → `service.go`, `manager_test.go` → `service_test.go`, `manager_repo_test.go` → `service_repo_test.go`; update test names/identifiers referencing "manager".
- Pure cosmetic — **batch with A4/A5** to keep the `worktree` diff coherent.

**Files:** `worktree/*`.

**Depends on:** A4, A5 (same package; avoid overlapping diffs).

**Acceptance:** no `manager` identifiers or filenames remain in `worktree`; behavior unchanged; tests green.

### A7. Split pure helpers into their own tested files

**Goal:** mirror `worktree/parse.go` — isolate pure functions into cohesively-named files with focused tests. **Name files by topic, never a catch-all `util.go`; do not over-fragment.**

**Concrete extractions (the agreed list):**

1. **`branch/parse.go`** ← move the pure ref/reflog parsers out of `branch/service.go`:
   - `normalizeBranch`
   - `headCheckoutBranchName`
   - `reflogBranchName`
   Add **`branch/parse_test.go`** (table tests for reflog/`checkout: moving from … to …` lines, `refs/heads/`–`refs/remotes/` stripping, `@{…}` handling). These are currently only tested indirectly.

2. **`worktree/validate.go`** ← move `validateAddInput` out of the service file, with **`worktree/validate_test.go`** (path/branch required, trimming). *(Minor; include if it reads cleanly.)*

3. **Helper de-duplication pass (ui):** the module targets Go 1.24, so replace the hand-rolled `min`/`max` in `ui/model.go`, `ui/screens/helpers.go`, `ui/components/selectlist/model.go`, `ui/components/dialog/model.go` with the builtins; delete `branchMin`/`branchMax` (`ui/screens/branch.go`) and the **unused** `min` in `ui/components/dialog/model.go`. Consolidate the duplicated `fitLine` where it doesn't introduce an awkward cross-package dependency (otherwise leave as-is and note it). *(Low priority; do last in Phase A.)*

**Note (out of scope here):** `branchSwitchWarning`/`currentBranchName` in `ui/screens/branch.go` encode a *domain rule* living in the UI. Leave them for now; revisit only if B3 makes relocating the rule into `app` natural.

**Files:** as listed.

**Depends on:** A4/A6 (so `branch`/`worktree` files are settled first).

**Acceptance:** extracted functions are unexported, cohesively grouped, and covered by new tests; no `util.go`; build/vet/test green.

---

## Phase B — Architecture (interlocking; keep this order)

These steps form one design ("UI owns presentation + dispatch; `app` owns operations + state"). Design them together; implement in order.

### B1. Replace effect/result with a Bubble-Tea-decoupled command/message system

**Goal:** keep work async/non-blocking, keep `app` framework-free, and cut per-operation boilerplate roughly in half by deleting the outbound `Effect` enumeration and the `runEffect` switch. *(Full design in Appendix — Plan 2.)*

**Actions (summary):**
- Add to `app`: `type Message interface{}` (inspectable result data) and `type Command func() Message` (async unit of work closing over `a.services`). Neither imports Bubble Tea.
- Convert intents to mutate state synchronously and return a `Command` (or `nil`).
- Replace the entire `runEffect` switch with one generic `ui` runner (~5 lines) that runs the thunk on a goroutine and wraps the returned `Message` in a private `appMsg` envelope for the Bubble Tea bus.
- Rename/convert `Result` → `Message`, `HandleResult` → `HandleMessage` (stays an inspectable switch; may return the next `Command`).
- Delete `Effect`, the `Effect` structs, and `runEffect`.
- **Bake in two identity mechanisms now** (shared by B2): per-loading-entry **IDs** carried on the `Command`, and **request-sequencing tokens** so stale messages are dropped (fixes the out-of-order branch-commit preview).
- Keep the one acknowledged exception: terminal control (`tea.Quit`) stays `ui`-side; `app` signals via state or a `QuitRequested` message.

**Invariant:** a `Command` thunk runs on a goroutine and may only read captured args/services and return a `Message` — it must **never** touch `app.State`. All mutation happens in intents and `HandleMessage` on the main loop.

**Depends on:** Phase A complete.

**Acceptance:** `app` imports no Bubble Tea; only one `ui` function references `tea.Cmd`/`tea.Msg`; per-operation touch points drop from four to two; `app` unit tests feed a `Message` and assert state + next `Command`; the held-arrow stale-preview bug no longer reproduces.

### B2. Status & loading: correctness + layout

**Goal:** keep state in `app`; make completion-driven updates automatic; fix concurrent-loading correctness; split rendering left/right.

**Actions:**
- Keep `state.Loading` and `state.Statuses` in `app`. Add explicit public `ClearStatus()` and `ClearLoading()` (formalizing today's `DismissStatuses`/private `clearLoading`).
- **Fix the latent loading bug:** today `markLoadingDone` marks the *last active* entry (LIFO) and `clearLoading` wipes everything, so concurrent tasks clobber each other. Using B1's per-entry **ID**, mark/clear **only** the entry whose `Command` completed — clearing one task must never remove still-pending tasks.
- Rendering: each is its own component (already true). Lay out **loading on the right, status on the left** (a `View`/`overlayNoticeLines` change in `ui/model.go`).
- "View updates when either changes" is automatic under B1 (re-render after every message) — no extra wiring.

**Depends on:** B1 (needs loading-entry IDs).

**Acceptance:** with two concurrent operations, completing one marks only its own entry and leaves the other pending; status renders left, loading right; clear methods are public on `app`.

### B3. Invert dialog ownership to the UI

**Goal:** UI owns dialog *representation* (labels, text, buttons, active flag) and button→action wiring; `app` owns operations and emits semantic outcomes. Removes the presentation strings currently authored in `app`. Reuse the existing stateless `ui/components/dialog` renderer. *(Implementation details are at my discretion per your instruction; the design below is the chosen approach — it must compose with B1 and B4.)*

**Design:**
- **Move dialog state into the screen's ephemeral state.** The active flag lives on the screen Model (this is what B4's `ModeDialog` keys off). Remove `DialogState`/`DialogKind`/`DialogButton` from `app.State` and delete the dialog-construction code from `app` intents (`RequestRemoveWorktree`, `RequestDeleteBranch`, `RequestDeleteAllBranches`) and from `DialogChoose`.
- **`app` exposes plain operations** previously reachable only through `DialogChoose`: e.g. `RemoveWorktree(path)`, `DeleteBranch(name)`, `DeleteAllBranches()`, `CreateBranchWorktree(path, branch)` — each returns a `Command` (B1) and `HandleMessage` does the post-op chaining (reload, status). No UI strings in `app`.
- **User-initiated dialogs** (e.g. ctrl+d → confirm remove): the screen captures the selected path into its dialog ephemeral state, opens its own dialog, and on confirm the button's action calls the `app` operation with that captured value.
- **App-initiated dialogs** (the create-branch-confirm case): `app` cannot open a UI dialog. Instead, the "check branch exists" operation returns a **semantic message** (e.g. `BranchAbsent{path, branch}` vs `BranchExists{…}`). To let the screen react, **B1's message delivery is extended**: in the `appMsg` case, the active screen gets an optional `OnMessage(msg)` hook (UI reactions: open a dialog, clear search) *and* `app.HandleMessage(msg)` runs (state + chaining). The screen reacts to `BranchAbsent` by opening its own confirm dialog; on confirm it calls `app.CreateBranchWorktree(path, branch)`.

**Depends on:** B1 (operations return `Command`; message bus carries the semantic outcomes and the screen-observation hook).

**Acceptance:** `app` contains **no** dialog titles/descriptions/button labels/focus IDs; `app.State` has no `Dialog*`; the create-branch flow works end-to-end via a semantic message + screen-owned dialog; the dialog component remains reusable and stateless.

### B4. Per-screen key-binding registry with modes

**Goal:** each screen owns the single source of truth for both dispatch and help text, with modes (default, dialog, extensible). *(Full design in Appendix — Plan 1.)*

**Actions (summary):**
- `Binding { Keys []Key; Symbol string; Label string; Action Action }` — `Symbol` decoupled from `Keys` so one entry covers multiple triggers and renders one hint (removes today's nil-handler help-only entries).
- A **Mode** = an ordered `[]Binding` plus a derived `map[Key]*Binding` (deterministic footer order + O(1) dispatch, can't drift).
- A screen **Registry** = `map[Mode]ModeBindings`, held as a per-screen instance field; replaces `defaultHandlers`/`confirmHandlers`. Active mode chosen by a pure `activeMode(state) Mode` (dialog active → `ModeDialog`, using B3's screen-owned flag).
- `Action func(*ActionCtx) app.Command` — returns `nil` for ephemeral-only actions (mutate the screen Model, re-render still happens) or the `app` operation's `Command`. Unifies intents and ephemeral mutations.
- Dispatch: normalize key → active mode → run action → else fall through to focused widget (only free-text runes/backspace fall through; commands and navigation are registry entries). Footer renders the active mode's ordered bindings.

**Depends on:** B1 (actions return `app.Command`), B3 (`ModeDialog` + screen-owned dialog flag).

**Acceptance:** no `BindingSet`/dual handler maps remain; footer is derived from the registry (no parallel ordered key lists); every key path (commands, navigation, dialog tab/confirm/cancel) routes through the registry; behavior matches the current key map.

### B5. Wrap `app` behind per-screen consumer interfaces

**Goal:** decouple `ui` from the concrete `*app.App`, improving testability. Define interfaces **at the consumer (per screen)**; `*app.App` satisfies them structurally. **Do this last in Phase B**, after B1's signatures are stable, so the interfaces are defined once.

**Actions:**
- Each screen declares a narrow interface naming only the `app` methods it calls (e.g. a `changeApp`, `branchApp`, `addApp` interface).
- `ScreenContext`/`ActionCtx` carries the relevant interface instead of `*app.App`.
- No big exported `App` interface from `app`.

**Depends on:** B1 (final intent signatures), B3 (final operation set), B4 (actions are the call sites).

**Acceptance:** screens reference per-screen interfaces, not `*app.App`; `app` exports no monolithic interface; build/test green.

---

## Phase C — Testing (final, separate step)

### C1. Add UI tests

**Goal:** the testability unlocked by B5 only pays off with tests. Done as the final step across all refactors, per your instruction.

**Actions:**
- Add table/golden tests for screen rendering at fixed dimensions (the untested `ui/model.go` compositing and `ui/screens/*` views).
- Use fake implementations of the per-screen interfaces (B5) to drive screens deterministically and assert dispatch → action → state-projection behavior.
- Add tests for the registry dispatch (key → action), mode switching, and footer derivation.

**Depends on:** all of Phase B.

**Acceptance:** `ui` package(s) gain meaningful coverage of rendering and dispatch; CI-style `go test ./...` exercises screens without a real terminal or real git.

---

## Appendix — Detailed designs

### Plan 1 — Per-screen key-binding registry with modes

**Goal:** each screen owns a registry that is the single source of truth for dispatch and help, supports modes (default, dialog, extensible), with entries carrying a hint symbol, hint label, and an action that can call an `app` intent, mutate the screen's ephemeral state, or both.

**Structure (three layers):**
- **Binding:** `{ Keys []Key; Symbol string; Label string; Action Action }`. Decoupling `Symbol` from `Keys` lets one entry cover multiple triggers (up/shift+tab) and render a single `↑/↓ move` hint — fixing today's nil-handler help-only duplicate entries.
- **Mode = ordered list of bindings** plus a derived `map[Key]*Binding`. The ordered slice gives deterministic footer order; the derived map gives O(1) dispatch; they cannot drift because both come from one source.
- **Registry = `map[Mode]ModeBindings`**, held as a per-screen instance field (actions close over the screen). Replaces the `defaultHandlers`/`confirmHandlers` pair with a mode-indexed map that scales past two modes.

**Action abstraction:** `type Action func(*ActionCtx) app.Command` (returns `nil` when nothing async is needed). App-affecting actions read ephemeral state and call an operation, returning its `Command`. Ephemeral-only actions mutate the screen Model and return `nil` (re-render still happens). This unifies intents and ephemeral mutations under one type.

**Active mode:** each screen exposes a pure `activeMode(state) Mode` (dialog active → `ModeDialog`), used for both dispatch and footer so they always agree.

**Dispatch:** normalize key → look up in active mode → run action, hand returned `Command` to the runner → else fall through to the focused widget (only free-text runes/backspace fall through; discrete commands and navigation are registry entries).

**Footer:** render the active mode's ordered bindings (`Symbol + Label`, skipping empty labels). Replaces the hand-maintained help-order lists.

### Plan 2 — Async command/message system (decoupled from Bubble Tea)

**Goal:** keep work async/non-blocking, keep `app` free of any Bubble Tea import, and cut per-operation boilerplate roughly in half by deleting the outbound `Effect` enumeration and the `runEffect` switch.

**Two custom types in `app` (neither imports Bubble Tea):**
- `type Message interface{}` — result event, inspectable data (like today's `Result`).
- `type Command func() Message` — a unit of async work that returns a `Message`.

**Flow:**
1. **Intent (app):** mutates state synchronously (set loading *with an ID*), returns a `Command` closing over `a.services`. No `Effect` struct, no `runEffect` case.
2. **Generic runner (ui), the only Bubble Tea bridge (~5 lines):** `run(cmd) tea.Cmd` → `func() tea.Msg { return appMsg{cmd()} }`. Replaces the whole `runEffect` switch.
3. **Completion → `Update` → automatic re-render** (inherent to TEA; nothing to design). In the `appMsg` case, deliver the message to the active screen's optional `OnMessage` hook (UI reactions — see B3) *and* to `app.HandleMessage` (state).
4. **Reducer (app):** `HandleMessage(Message) Command` stays an inspectable switch; mutates state, may return the next `Command` (chaining preserved).

**Boilerplate:** the outbound enumeration + interpreter are deleted; inbound stays as typed structs + one switch (kept deliberately so `app` tests stay pattern-matchable).

**Identity mechanisms baked in (shared with B2):**
- **Per-entry loading IDs** — the `Command` carries the loading-entry ID it set; `HandleMessage` clears only that entry. Fixes the LIFO "wrong spinner / clears pending tasks" bug.
- **Request-sequencing tokens** — drop stale messages (fixes the out-of-order branch-commit preview).

**Invariant:** a `Command` thunk runs on a goroutine; it may only read captured args/services and return a `Message` — never touch `app.State`. All mutation happens in intents and `HandleMessage` on the main loop. This keeps `State` race-free.

**Acknowledged exception:** terminal control (`tea.Quit`) stays ui-side — `app` signals intent via state or a `QuitRequested` message; `ui` translates it.

---

## Dependency graph (quick reference)

```
A1 shellinit ─┐
A2 cli  ──────┘ (A2 needs A1)
A3 var renames        (independent)
A4 unified runner ─┬─ A5 InRepo
                   └─ A6 manager→service
A7 helper splits   (after A4/A6)

B1 command/message  (after Phase A)
        │
        ├─ B2 status/loading
        ├─ B3 dialog inversion
        │        │
        │        └─ B4 registry  (also needs B1)
        └──────────── B5 per-screen interfaces (after B1/B3/B4)

C1 UI tests         (after all of B)
```

## Done-criteria for the whole project

- `main.go` is process-wiring only; `shellinit` and `cli` are isolated and narrowly imported.
- One generic, injected, context-aware command runner; no duplication; `InRepo` is a standalone startup check.
- `worktree` uses `service` naming throughout; pure helpers live in cohesive, tested files.
- `app` imports no Bubble Tea, contains no UI strings, and exposes operations + a command/message API.
- `ui` owns dialogs, key dispatch (registry), and ephemeral state; depends on `app` only through per-screen interfaces.
- Concurrent loading is correct; stale async results are dropped.
- UI has meaningful render/dispatch tests.
- `go build ./...`, `go vet ./...`, `go test ./...` green at every step.
