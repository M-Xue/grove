package app

import "fmt"

type StatusKind string

const (
	StatusInfo    StatusKind = "info"
	StatusSuccess StatusKind = "success"
	StatusError   StatusKind = "error"
)

type StatusEntry struct {
	ID      string
	Kind    StatusKind
	Message string
}

// ClearStatus removes every status notice.
func (a *App) ClearStatus() {
	a.state.Statuses = nil
}

func (a *App) appendStatus(kind StatusKind, message string) {
	a.state.Statuses = append(a.state.Statuses, StatusEntry{
		ID:      nextStatusID(len(a.state.Statuses)),
		Kind:    kind,
		Message: message,
	})
}

func nextStatusID(existingCount int) string {
	return fmt.Sprintf("status-%d", existingCount+1)
}
