# Architectural Refactor Plan

## Goals

This refactor introduces a cleaner separation between services, application logic, and Bubble Tea UI while keeping the UI reusable and easy to extend.

Agreed constraints:

- `worktree.Manager` becomes `worktree.Service`
- Git worktree business logic stays in `worktree/`
- docs loading moves into its own `docs.Service`
- a new `app` layer sits between services and UI
- reusable `textinput` and `selectlist` components live in Bubble Tea land under `ui/`
- reusable `loading`, `dialog`, and `status` components also live in Bubble Tea land under `ui/`
- `textinput` owns its text state and focus state, but not submit behavior
- `selectlist` owns selection and scrolling state using stable item IDs
- `loading` should show what operation is currently in progress
- `dialog` should support a variable number of buttons and render as a floating popup
- `status` should support rendering multiple status items at once, and each status item should remain visible until the next keystroke
- screen key handling should be organized as per-screen key handler maps
- screen handlers should be methods on screen structs
- key normalization should be shared in a `ui` key layer rather than scattered `msg.String()` checks

## Target Architecture

The refactor should move the repo toward four layers.

### `main`

Responsibilities:

- construct services
- construct `app.App`
- construct Bubble Tea UI runtime
- preserve the final selected-path `stdout` contract

`main.go` should stay thin.

### `worktree`

Responsibilities:

- Git worktree subprocess execution
- parsing `git worktree list --porcelain`
- worktree add/remove operations
- branch existence checks
- repo validation
- structured worktree data

Planned public surface:

```go
type Service interface {
    Add(path, branch string) error
    AddNewBranch(path, branch string) error
    BranchExists(branch string) (bool, error)
    InRepo() error
    List() ([]WorktreeInfo, error)
    Remove(path string) error
}
```

Rename plan:

- `Manager` -> `Service`
- `NewManager()` -> `NewService()`
- `NewManagerWithRunner()` -> `NewServiceWithRunner()`

The internal `Runner` seam should stay.

### `docs`

Responsibilities:

- load worktree documentation text
- normalize line endings
- strip backspace-overstrike formatting
- return docs content as plain lines or plain text

Planned public surface:

```go
type Service interface {
    WorktreeHelp() ([]string, error)
}
```

This moves docs retrieval out of `ui/commands.go` and makes docs a normal service dependency.

### `app`

Responsibilities:

- hold references to all services
- hold application state and current screen state
- own business rules and screen flow
- own confirmation flow semantics
- own selected-path output state
- return effect requests for async service work
- remain unaware of Bubble Tea types

Planned root types:

```go
type Services struct {
    Worktree worktree.Service
    Docs     docs.Service
}

type App struct {
    services Services
    state    State
}
```

`app` should not import Bubble Tea.

### `ui`

Responsibilities:

- Bubble Tea runtime adapter
- reusable Bubble Tea components
- screen adapters that compose components
- shared key normalization and keymaps
- rendering only
- effect execution bridge between `app` and Bubble Tea commands

The UI should be dumb about business logic, but it is still allowed to own:

- focus management
- component-local editing behavior
- component-local selection behavior
- key routing
- layout and rendering

The UI should not own:

- worktree lifecycle policy
- confirmation semantics
- branch existence policy
- selected-path output contract
- docs retrieval policy

## Package Layout

Recommended target layout:

```text
main.go

app/
  app.go
  services.go
  state.go
  effects.go
  messages.go
  change.go
  add.go
  docs.go

docs/
  service.go
  command_runner.go

ui/
  model.go
  commands.go
  render.go
  screens/
    change.go
    add.go
    docs.go
  components/
    loading/
      model.go
      render.go
    dialog/
      model.go
      render.go
    status/
      model.go
      render.go
    textinput/
      model.go
      render.go
    selectlist/
      model.go
      render.go
  keys/
    keys.go

worktree/
  service.go
  command_runner.go
  parse.go
```

This does not need to happen in one move. The repo can transition incrementally.

## Application Layer Design

The `app` package should be the business and orchestration layer.

### Responsibilities Of `app.App`

- track current screen
- hold status and error state
- hold selected path for final process output
- hold worktree data and docs data
- handle business actions requested by UI screens
- decide what async effect should happen next
- consume async service results

### Suggested State Shape

```go
type ScreenID string

const (
    ScreenChange ScreenID = "change"
    ScreenAdd    ScreenID = "add"
    ScreenDocs   ScreenID = "docs"
)

type State struct {
    Screen         ScreenID
    SubmittedPath  string
    Worktrees      []worktree.WorktreeInfo
    DocsLines      []string
    Loading        LoadingState
    Dialog         DialogState
    Statuses       []StatusEntry

    Change ChangeState
    Add    AddState
    Docs   DocsState
}
```

The exact state shape can evolve, but `SubmittedPath` should move into `app`, not stay in `ui`.

Suggested supporting types:

```go
type LoadingState struct {
    Active  bool
    Message string
}

type DialogState struct {
    Active      bool
    Title       string
    Description string
    Buttons     []DialogButton
    FocusedID   string
    Kind        DialogKind
}

type DialogButton struct {
    ID    string
    Label string
}

type StatusEntry struct {
    ID      string
    Kind    StatusKind
    Message string
}
```

The important ownership split is:

- `app` owns whether loading, dialog, and statuses should exist
- `ui` owns how those are rendered and interacted with

### Suggested Business Actions

The UI should call semantic methods instead of manipulating business state directly.

Examples:

```go
func (a *App) LoadWorktrees() Effect
func (a *App) OpenAdd()
func (a *App) CloseAdd()
func (a *App) OpenDocs() Effect
func (a *App) CloseDocs()
func (a *App) RequestSubmitSelectedPath(path string) Effect
func (a *App) RequestRemoveWorktree(path string)
func (a *App) ConfirmRemoveWorktree() Effect
func (a *App) CancelRemoveWorktree()
func (a *App) RequestAddWorktree(path, branch string) Effect
func (a *App) ConfirmCreateBranch() Effect
func (a *App) CancelCreateBranch()
func (a *App) DismissStatuses()
```

These names can change, but the pattern matters:

- UI asks for business actions
- `app` decides what state changes and effects happen

The `app` layer should also be responsible for:

- setting loading messages before async work starts
- clearing loading state when async work finishes
- opening dialogs for confirmation decisions
- appending status entries after completed work
- clearing all status entries on the next keystroke

### Effects

`app` should return explicit effect values rather than directly running services.

Suggested shape:

```go
type Effect interface{}

type LoadWorktreesEffect struct{}
type LoadDocsEffect struct{}
type CheckBranchExistsEffect struct {
    Path   string
    Branch string
}
type AddWorktreeEffect struct {
    Path         string
    Branch       string
    CreateBranch bool
}
type RemoveWorktreeEffect struct {
    Path string
}
type QuitEffect struct{}
```

This keeps `app` free of Bubble Tea and still allows async work.

### Async Results

The UI runtime should execute effects and send typed results back into `app`.

Suggested result types:

```go
type Result interface{}

type WorktreesLoadedResult struct {
    Worktrees []worktree.WorktreeInfo
    Err       error
}

type BranchCheckedResult struct {
    Path   string
    Branch string
    Exists bool
    Err    error
}

type WorktreeAddedResult struct {
    Err error
}

type WorktreeRemovedResult struct {
    Path string
    Err  error
}

type DocsLoadedResult struct {
    Lines []string
    Err   error
}
```

The UI runtime can wrap these in Bubble Tea messages if needed, but `app` should define the semantic result types.

## Bubble Tea UI Design

The `ui` package becomes a runtime adapter and renderer.

### Top-Level `ui.Model`

The top-level Bubble Tea model should:

- hold a pointer to `app.App`
- hold the active screen adapter
- hold terminal width and height
- run `app` effects as `tea.Cmd`
- route Bubble Tea messages to the active screen or into `app`

Suggested shape:

```go
type Model struct {
    app    *app.App
    width  int
    height int

    change *screens.ChangeScreen
    add    *screens.AddScreen
    docs   *screens.DocsScreen
}
```

The `ui` layer should render from `app` state plus screen-local component state.

At the top level, `ui.Model.Update` should clear status entries on the first keystroke after they appear by calling `app.DismissStatuses()` before normal key routing continues.

## Reusable Bubble Tea Components

The reusable components live in `ui/components/`.

### `loading`

Responsibilities:

- render a shared loading indicator style
- show the current loading message from `app`
- optionally animate if desired later
- not decide when loading starts or stops

Suggested API:

```go
type Model struct {}

func New() Model
func (m Model) View(message string, width int) string
```

This component is mostly presentational. The semantic state lives in `app.LoadingState`.

### `dialog`

Responsibilities:

- render a floating popup dialog
- render a title, description, and a variable number of buttons
- manage focused button state and left/right movement
- return the currently selected button to the parent screen
- not decide what each button means semantically

Suggested API:

```go
type Button struct {
    ID    string
    Label string
}

type Model struct {
    title       string
    description string
    buttons     []Button
    focusedID   string
}

func New() Model
func (m *Model) SetTitle(value string)
func (m *Model) SetDescription(value string)
func (m *Model) SetButtons(buttons []Button)
func (m *Model) SetFocusedID(id string)
func (m Model) FocusedID() (string, bool)
func (m *Model) Update(msg tea.KeyMsg) (bool, tea.Cmd)
func (m Model) View(width, height int) string
```

Current planned uses:

- confirm worktree deletion
- confirm branch creation

The screen handler should interpret `enter` on the focused button and call the corresponding `app` method.

### `status`

Responsibilities:

- render one or more status items using a shared visual style
- support multiple simultaneous status messages
- support status kinds such as info, success, and error
- not decide lifecycle rules

Suggested API:

```go
type Kind string

const (
    KindInfo    Kind = "info"
    KindSuccess Kind = "success"
    KindError   Kind = "error"
)

type Item struct {
    ID      string
    Kind    Kind
    Message string
}

type Model struct {}

func New() Model
func (m Model) View(items []Item, width int) []string
```

The lifecycle rule is owned by `app`: status items should stay visible until the next keystroke, then all current status items should be cleared together.

### `textinput`

Responsibilities:

- own current text value
- own focus state
- own placeholder
- support one-line ASCII editing
- support normal rune insertion, backspace, and paste-like rune insertion
- not define submit semantics

Suggested API:

```go
type Model struct {
    value       string
    placeholder string
    focused     bool
}

func New(placeholder string) Model
func (m *Model) SetPlaceholder(value string)
func (m *Model) SetValue(value string)
func (m Model) Value() string
func (m *Model) Clear()
func (m *Model) Focus()
func (m *Model) Blur()
func (m Model) Focused() bool
func (m *Model) Update(msg tea.KeyMsg) (bool, tea.Cmd)
func (m Model) View() string
```

`Update` should return whether the key was consumed.

### `selectlist`

Responsibilities:

- own items
- own selected stable item ID
- own selected index and scroll state
- support up/down navigation
- preserve selection by stable ID when items reload
- not define what Enter means

Suggested item shape:

```go
type Item struct {
    ID    string
    Label string
}
```

Suggested API:

```go
type Model struct {
    items      []Item
    selectedID string
    selected   int
    scroll     int
    emptyLabel string
}

func New(emptyLabel string) Model
func (m *Model) SetItems(items []Item)
func (m *Model) SetSelectedID(id string)
func (m Model) SelectedID() (string, bool)
func (m Model) SelectedItem() (Item, bool)
func (m *Model) Update(msg tea.KeyMsg) (bool, tea.Cmd)
func (m Model) View(height int) string
```

Optional extension:

- emit a Bubble Tea message like `SelectionChangedMsg`

This is optional because the parent screen can also read the selected item directly after update.

## Screen Structure

Each screen should be split into:

- app-layer business state and semantics
- ui-layer screen adapter with Bubble Tea components and keymaps

Recommended UI screen files:

- `ui/screens/change.go`
- `ui/screens/add.go`
- `ui/screens/docs.go`

Each screen should be a struct with:

- local reusable UI components
- one or more key handler maps
- methods for key handlers
- methods for syncing from `app` state
- methods for rendering

Suggested shape:

```go
type ChangeScreen struct {
    dialog          dialog.Model
    search          textinput.Model
    list            selectlist.Model
    defaultHandlers KeyHandlers
    confirmHandlers KeyHandlers
}
```

This follows the chosen direction:

- handlers are methods on the screen struct
- key routing uses a shared normalized key layer

The top-level UI model should also own shared presentational components:

- `loading.Model`
- `status.Model`

Those are cross-screen overlays rather than screen-local controls.

## Key Handling Plan

Key handling should be centralized and easy to extend.

### Shared Key Normalization

Create a shared key layer in `ui/keys`.

Suggested shape:

```go
type Key string

const (
    KeyUnknown   Key = ""
    KeyEnter     Key = "enter"
    KeyEsc       Key = "esc"
    KeyCtrlC     Key = "ctrl+c"
    KeyCtrlA     Key = "ctrl+a"
    KeyCtrlD     Key = "ctrl+d"
    KeyQuestion  Key = "?"
    KeyUp        Key = "up"
    KeyDown      Key = "down"
    KeyTab       Key = "tab"
    KeyShiftTab  Key = "shift+tab"
    KeyBackspace Key = "backspace"
    KeyQ         Key = "q"
)

func Normalize(msg tea.KeyMsg) Key
```

This prevents raw `msg.String()` checks from spreading across the UI.

### Per-Screen Key Handler Maps

Each screen should define one or more key handler maps.

Suggested shape:

```go
type Handler func(*ScreenContext, tea.KeyMsg) tea.Cmd
type KeyHandlers map[keys.Key]Handler
```

And then for a screen:

```go
func (s *ChangeScreen) initHandlers() {
    s.defaultHandlers = KeyHandlers{
        keys.KeyEnter:    s.handleEnter,
        keys.KeyCtrlA:    s.handleOpenAdd,
        keys.KeyCtrlD:    s.handleStartRemove,
        keys.KeyQuestion: s.handleOpenDocs,
        keys.KeyEsc:      s.handleQuit,
        keys.KeyCtrlC:    s.handleQuit,
    }

    s.confirmHandlers = KeyHandlers{
        keys.KeyEsc:   s.handleCancelRemove,
        keys.KeyCtrlC: s.handleQuit,
    }
}
```

This keeps the screen-specific event surface explicit and easy to extend.

### Handler Methods On Screen Structs

This is the chosen direction for handler organization.

Example:

```go
func (s *ChangeScreen) handleEnter(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd
func (s *ChangeScreen) handleOpenDocs(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd
func (s *ChangeScreen) handleStartRemove(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd
```

Benefits:

- handlers can access screen-local UI component state directly
- related behavior stays colocated with the screen that owns it
- confirm-substate handlers are easy to organize
- new shortcuts are easy to register without adding giant `switch` blocks

### Routing Order

Each screen should route keys in a consistent order.

Recommended order:

1. Decide which screen substate is active
2. If status entries are visible, clear them on this keystroke through `app.DismissStatuses()`
3. If a dialog is active for the current screen, route keys to the dialog first
4. Give the focused component first chance to consume the key when appropriate
5. Normalize the key using `ui/keys`
6. Look up a handler in the active key handler map
7. If found, run the handler
8. Otherwise, ignore the key

Example for change screen:

- search input gets rune input and backspace first
- list gets up/down and tab movement keys
- screen handlers own Enter, docs, remove, add, quit
- if a confirmation dialog is active, the dialog consumes left/right navigation and focused-button changes first

This keeps component concerns and business concerns separate.

### `ScreenContext`

Screen handlers need a narrow integration surface.

Suggested shape:

```go
type ScreenContext struct {
    App       *app.App
    RunEffect func(app.Effect) tea.Cmd
}
```

Handlers can:

- read current app state
- call semantic app methods
- execute returned effects

This avoids giving handlers direct service access.

## Detailed Transition Plan

The safest approach is incremental.

### Phase 1: Service Renames And Extraction

1. Rename `worktree.Manager` to `worktree.Service`
2. Rename constructors to `NewService` and `NewServiceWithRunner`
3. Update imports, tests, and documentation
4. Create `docs/` package with `docs.Service`
5. Move docs loading and formatting logic from `ui/commands.go` into `docs/`
6. Add tests for docs formatting behavior

Deliverable:

- `worktree` and `docs` are explicit service packages

### Phase 2: Introduce `app`

1. Create `app.Services`
2. Create `app.App`
3. Move selected-path output state from `ui` into `app`
4. Move worktree list state, docs data, and confirmation semantics into `app`
5. Define effect types and result types
6. Add `app` unit tests for business flows

Deliverable:

- business rules no longer live in `ui/update_*.go`

### Phase 3: Extract Reusable Bubble Tea Components

1. Add `ui/components/loading`
2. Add `ui/components/dialog`
3. Add `ui/components/status`
4. Add `ui/components/textinput`
5. Add `ui/components/selectlist`
6. Port change-screen search behavior to `textinput`
7. Port change-screen list behavior to `selectlist`
8. Port add-screen path and branch fields to `textinput`
9. Replace inline confirmation rendering with `dialog`
10. Replace inline status rendering with `status`
11. Add component unit tests

Deliverable:

- repeated input and selection logic is removed from screens

### Phase 4: Screen Adapters

1. Create `ui/screens/change.go`
2. Create `ui/screens/add.go`
3. Create `ui/screens/docs.go`
4. Move mode-specific routing out of `ui/model.go`
5. Give each screen its own component set and handler maps
6. Keep rendering split by screen

Deliverable:

- per-screen UI logic is isolated and easier to extend

### Phase 5: Shared Key Layer

1. Add `ui/keys`
2. Replace raw `msg.String()` checks with normalized keys
3. Define handler maps per screen and substate
4. Convert screen shortcut logic from `switch` blocks to handler maps

Deliverable:

- key handling becomes explicit, extensible, and consistent

### Phase 6: Thin Top-Level UI Model

1. Reduce `ui.Model` to runtime orchestration
2. Let it hold `app.App` plus screen adapters
3. Convert `app.Effect` values to Bubble Tea commands
4. Convert async command results back into `app` results
5. Preserve existing startup and final output behavior

Deliverable:

- top-level UI model becomes a thin adapter around app and screens

## Testing Plan

Tests should be moved and expanded with the architecture.

### `worktree/`

- keep existing tests with renamed types

### `docs/`

- add docs load and formatting tests
- test CRLF normalization
- test backspace-overstrike stripping

### `app/`

- selection submit flow
- add flow with existing branch
- add flow with missing branch and confirmation
- remove flow with confirmation
- docs open and close flow
- reload clearing submitted output

### `ui/components/textinput`

- focus and blur behavior
- ASCII rune insertion
- backspace behavior
- clear behavior
- placeholder rendering

### `ui/components/selectlist`

- selection by stable ID
- move up/down behavior
- item reload preserving selection by ID
- scroll clamping
- empty state rendering

### `ui/components/dialog`

- variable button counts
- focused button movement
- stable focused button ID
- popup rendering

### `ui/components/status`

- multiple simultaneous status items
- kind-specific rendering
- rendering order

### `ui/components/loading`

- message rendering
- layout behavior

### `ui/screens/`

- handler map dispatch
- component routing order
- screen-specific key behavior

## Key Decisions To Preserve During Refactor

- keep one top-level Bubble Tea program
- keep selected-path `stdout` contract unchanged
- keep Git subprocesses as the source of truth
- keep service seams easy to test with injected runners
- keep reusable UI components in Bubble Tea land
- keep shared overlay components in Bubble Tea land
- keep app logic free of Bubble Tea imports
- keep screen handlers as methods on screen structs
- keep shared key normalization in one UI key layer
- keep per-screen key handler maps explicit and easy to extend

## Open Implementation Notes

These should guide implementation choices while refactoring.

### On Enter Handling

`textinput` and `selectlist` should not own Enter semantics.

Examples:

- on change screen, Enter means submit selected worktree
- on add screen, Enter means submit path and branch to `app`
- on docs screen, Enter may do nothing

Enter belongs to the screen handler map, not the reusable component.

### On Selection Ownership

`selectlist` should own current selected ID and selection movement state, but the business layer should be able to read the selected ID at any time.

This fits the requirement that the app can reason about current selection state without owning the list widget itself.

### On Focus Ownership

Focus should stay in `ui` because it is a presentation and interaction concern. The business layer should know which screen is active, but not which Bubble Tea component has focus unless a screen intentionally reflects that as semantic state.

### On Status Lifecycle

Status lifecycle is a semantic rule and should therefore live in `app`.

Required behavior:

- one or more status entries may be visible at once
- status entries remain visible until the next keystroke
- on the next keystroke, the current batch of status entries is cleared before normal routing continues

This allows screens and services to report multiple outcomes without coupling status behavior to any one screen.

### On Loading Ownership

Loading messages should be set by `app` when async work begins and cleared by `app` when result messages are processed.

Examples:

- `creating worktree`
- `checking branch`
- `removing worktree`
- `loading worktrees`
- `loading docs`

The reusable loading component should only render that state.

### On Dialog Ownership

The dialog component should own button focus and popup rendering, but the `app` layer should own why the dialog exists and what choices are available.

Examples:

- remove confirmation dialog buttons: `confirm`, `cancel`
- create branch dialog buttons: `create`, `cancel`

This keeps confirmation semantics out of the rendering layer.

## Definition Of Done

The refactor is complete when:

- `worktree.Service` replaces `worktree.Manager`
- docs are loaded through `docs.Service`
- `app` owns business rules, async effect definitions, and selected-path output state
- reusable `loading`, `dialog`, `status`, `textinput`, and `selectlist` components exist under `ui/components`
- each screen is implemented as its own adapter with handler maps and screen methods
- all raw key-string switching is replaced by normalized keys and per-screen handler maps where appropriate
- the top-level Bubble Tea model is thin
- existing behavior is preserved or intentionally improved
- tests cover services, app flows, and reusable components
