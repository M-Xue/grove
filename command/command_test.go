package command

import (
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCombinedOutputReturnsErrorForMissingBinary(t *testing.T) {
	r := New()
	if _, err := r.CombinedOutput("grove-nonexistent-binary-xyz"); err == nil {
		t.Fatal("expected error for nonexistent command")
	}
}

func TestNewAppliesDefaultTimeout(t *testing.T) {
	if got := New().timeout; got != DefaultTimeout {
		t.Fatalf("expected default timeout %v, got %v", DefaultTimeout, got)
	}
}

func TestCombinedOutputEnforcesTimeout(t *testing.T) {
	// A command that runs longer than the runner's timeout must be cancelled
	// rather than blocking indefinitely.
	r := Runner{timeout: 10 * time.Millisecond}
	name, args := sleepCommand()
	start := time.Now()
	if _, err := r.CombinedOutput(name, args...); err == nil {
		t.Fatal("expected error when command exceeds timeout")
	}
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Fatalf("expected timeout to cancel quickly, took %v", elapsed)
	}
}

// sleepCommand returns a long-running command appropriate for the host OS.
func sleepCommand() (string, []string) {
	if runtime.GOOS == "windows" {
		// ping -n 61 takes roughly 60 seconds and needs no console.
		return "ping", []string{"-n", "61", "127.0.0.1"}
	}
	return "sleep", []string{"60"}
}

func TestOutputSeparatesStdoutFromStderr(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires sh")
	}
	r := New()
	output, err := r.Output("sh", "-c", "echo notice >&2; echo '{\"data\":1}'")
	if err != nil {
		t.Fatalf("Output returned error: %v", err)
	}
	if got := string(output); got != "{\"data\":1}\n" {
		t.Fatalf("expected stdout only, got %q", got)
	}
}

func TestOutputAttachesStderrToError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires sh")
	}
	r := New()
	_, err := r.Output("sh", "-c", "echo boom >&2; exit 1")
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected stderr in error text, got %v", err)
	}
}
