package docs

import (
	"errors"
	"testing"
)

type stubRunner struct {
	output []byte
	err    error
}

func (s stubRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	return s.output, s.err
}

func TestWorktreeHelpFormatsOutput(t *testing.T) {
	service := NewServiceWithRunner(stubRunner{output: []byte("A\r\nB\n")})
	lines, err := service.WorktreeHelp()
	if err != nil {
		t.Fatalf("WorktreeHelp returned error: %v", err)
	}
	if len(lines) != 2 || lines[0] != "A" || lines[1] != "B" {
		t.Fatalf("unexpected lines: %#v", lines)
	}
}

func TestWorktreeHelpReturnsRunnerError(t *testing.T) {
	service := NewServiceWithRunner(stubRunner{err: errors.New("git failed")})
	if _, err := service.WorktreeHelp(); err == nil {
		t.Fatal("expected error")
	}
}
