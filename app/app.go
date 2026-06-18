package app

type App struct {
	services        Services
	state           State
	loadingCounter  int
	branchCommitSeq int
}

type Option func(*App)

func WithInitialScreen(screen ScreenID) Option {
	return func(a *App) {
		a.state.Screen = screen
	}
}

func New(services Services, options ...Option) *App {
	a := &App{
		services: services,
		state: State{
			Screen: ScreenChange,
		},
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		option(a)
	}
	return a
}

func (a *App) Init() Command {
	switch a.state.Screen {
	case ScreenAdd:
		return nil
	default:
		return a.loadWorktrees()
	}
}

func (a *App) State() State {
	return a.state
}

// Quit signals that grove should exit without selecting a path. The terminal
// is controlled by ui, so this surfaces as a QuitRequested message.
func (a *App) Quit() Command {
	return func() Message { return QuitRequested{} }
}

func (a *App) SubmittedPath() string {
	return a.state.SubmittedPath
}

func (a *App) DismissCompletedLoading() {
	if len(a.state.Loading) == 0 {
		return
	}
	filtered := a.state.Loading[:0]
	for _, entry := range a.state.Loading {
		if entry.Active && entry.Completed {
			continue
		}
		filtered = append(filtered, entry)
	}
	a.state.Loading = filtered
}
