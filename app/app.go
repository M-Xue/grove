package app

type App struct {
	services       Services
	state          State
	loadingCounter int
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

func (a *App) Init() Effect {
	switch a.state.Screen {
	case ScreenAdd:
		return nil
	case ScreenBranch:
		a.setLoading("loading worktrees")
		return LoadWorktreesEffect{}
	default:
		a.setLoading("loading worktrees")
		return LoadWorktreesEffect{}
	}
}

func (a *App) State() State {
	return a.state
}

func (a *App) Services() Services {
	return a.services
}

func (a *App) SubmittedPath() string {
	return a.state.SubmittedPath
}

func (a *App) DismissDialog() {
	a.clearDialog()
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
