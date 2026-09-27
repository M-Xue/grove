package pr

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

type stubRunner struct {
	output []byte
	err    error
	calls  [][]string
}

func (s *stubRunner) Output(name string, args ...string) ([]byte, error) {
	s.calls = append(s.calls, append([]string{name}, args...))
	return s.output, s.err
}

func TestForBranchRunsGHAndSelectsPR(t *testing.T) {
	runner := &stubRunner{output: []byte(openPRFixture)}
	service := NewService(runner)

	info, found, err := service.ForBranch("fix-remote-annotation-url-policy")
	if err != nil {
		t.Fatalf("ForBranch returned error: %v", err)
	}
	if !found || info.Number != 2745 || info.Status != StatusOpen {
		t.Fatalf("unexpected result: found=%v info=%#v", found, info)
	}

	want := []string{
		"gh", "pr", "list",
		"--head", "fix-remote-annotation-url-policy",
		"--state", "all",
		"--limit", "10",
		"--json", listFields,
	}
	if len(runner.calls) != 1 || strings.Join(runner.calls[0], " ") != strings.Join(want, " ") {
		t.Fatalf("unexpected gh invocation: %v", runner.calls)
	}
}

func TestForBranchNoPRIsNotAnError(t *testing.T) {
	runner := &stubRunner{output: []byte("[]\n")}
	service := NewService(runner)

	_, found, err := service.ForBranch("feature")
	if err != nil {
		t.Fatalf("ForBranch returned error: %v", err)
	}
	if found {
		t.Fatal("expected no PR to be found")
	}
}

func TestForBranchRequiresBranchName(t *testing.T) {
	service := NewService(&stubRunner{})
	if _, _, err := service.ForBranch("  "); err == nil {
		t.Fatal("expected an error for a blank branch name")
	}
}

func TestForBranchClassifiesUnavailableEnvironments(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"gh not installed", &exec.Error{Name: "gh", Err: exec.ErrNotFound}},
		{"not authenticated", errors.New("exit status 4: To get started with GitHub CLI, please run:  gh auth login")},
		{"no github remote", errors.New("exit status 1: could not determine base repo: no default remote repository")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewService(&stubRunner{err: tc.err})
			_, _, err := service.ForBranch("feature")
			if !errors.Is(err, ErrUnavailable) {
				t.Fatalf("expected ErrUnavailable, got %v", err)
			}
		})
	}
}

func TestForBranchKeepsTransientErrorsTransient(t *testing.T) {
	service := NewService(&stubRunner{err: fmt.Errorf("exit status 1: connect: network is unreachable")})
	_, _, err := service.ForBranch("feature")
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, ErrUnavailable) {
		t.Fatal("network errors must stay transient, not unavailable")
	}
}
