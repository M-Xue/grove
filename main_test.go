package main

import (
	"testing"

	"github.com/M-Xue/grove/ui"
)

func TestSelectedPathOutputReturnsSubmittedPath(t *testing.T) {
	m := ui.New(nil)

	if got := selectedPathOutput(m); got != "" {
		t.Fatalf("expected empty path, got %q", got)
	}
}
