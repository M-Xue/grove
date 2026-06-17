package command

import (
	"runtime"
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
