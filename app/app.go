package app

type App struct {
	services      Services
	state         State
	statusCounter int
}

func New(services Services) *App {
	return &App{
		services: services,
		state: State{
			Screen: ScreenChange,
		},
	}
}

func (a *App) Init() Effect {
	a.setLoading("loading worktrees")
	return LoadWorktreesEffect{}
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

func (a *App) DismissStatuses() {
	a.state.Statuses = nil
}
