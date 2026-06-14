package docs

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type Runner interface {
	CombinedOutput(name string, args ...string) ([]byte, error)
}

type Service interface {
	WorktreeHelp() ([]string, error)
}

type service struct {
	runner Runner
}

func NewService() Service {
	return service{runner: commandRunner{}}
}

func NewServiceWithRunner(runner Runner) Service {
	if runner == nil {
		return NewService()
	}
	return service{runner: runner}
}

func (s service) WorktreeHelp() ([]string, error) {
	output, err := s.runner.CombinedOutput("git", "--no-pager", "help", "worktree")
	if err != nil {
		return nil, err
	}
	return formatWorktreeDocs(string(output)), nil
}

type commandRunner struct{}

func (commandRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %s", err, message)
	}
	return output, nil
}

var backspaceOverstrikePattern = regexp.MustCompile(`.\x08`)

func formatWorktreeDocs(output string) []string {
	cleaned := strings.ReplaceAll(output, "\r\n", "\n")
	cleaned = strings.ReplaceAll(cleaned, "\r", "\n")
	for backspaceOverstrikePattern.MatchString(cleaned) {
		cleaned = backspaceOverstrikePattern.ReplaceAllString(cleaned, "")
	}
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return []string{"No documentation available."}
	}
	return strings.Split(cleaned, "\n")
}
