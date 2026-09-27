// Package pr provides read-only pull-request information for branches, backed
// by the GitHub CLI (gh). gh is used instead of the GitHub HTTP API because it
// owns the two hard problems — authentication and remote/host resolution —
// and follows grove's service pattern of a thin wrapper over an injected
// command runner.
package pr

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Status is the folded state of a pull request.
type Status string

const (
	StatusOpen   Status = "open"
	StatusDraft  Status = "draft"
	StatusMerged Status = "merged"
	StatusClosed Status = "closed"
)

// CheckState is the normalized outcome of one CI check.
type CheckState string

const (
	CheckPending   CheckState = "pending"
	CheckPassed    CheckState = "passed"
	CheckFailed    CheckState = "failed"
	CheckSkipped   CheckState = "skipped"
	CheckCancelled CheckState = "cancelled"
	CheckNeutral   CheckState = "neutral"
)

// Check is one CI check on a pull request.
type Check struct {
	Name  string
	State CheckState
}

// Info is the pull request grove surfaces for a branch.
type Info struct {
	Number int
	Title  string
	Status Status
	Author string
	URL    string
	Checks []Check
}

// ErrUnavailable marks a lookup that cannot work in this environment — gh not
// installed, not authenticated, or the repository has no GitHub remote — as
// opposed to a transient failure. Callers should stop querying once seen.
var ErrUnavailable = errors.New("pull request lookup unavailable")

// Runner executes external commands, returning stdout alone so gh's JSON is
// never corrupted by advisory notices it writes to stderr.
type Runner interface {
	Output(name string, args ...string) ([]byte, error)
}

type Service interface {
	// ForBranch returns the pull request whose head is the named branch.
	// found=false with a nil error is the ordinary "no PR" case. An error
	// wrapping ErrUnavailable means lookups cannot work at all in this
	// environment.
	ForBranch(branch string) (Info, bool, error)
}

type service struct {
	runner Runner
}

// NewService returns a PR Service backed by the injected command runner.
func NewService(runner Runner) Service {
	return service{runner: runner}
}

// listLimit bounds how many same-head PRs are fetched for selection; more than
// a handful sharing one head branch is pathological.
const listLimit = 10

const listFields = "number,title,state,isDraft,url,author,statusCheckRollup"

func (s service) ForBranch(branch string) (Info, bool, error) {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return Info{}, false, errors.New("branch name is required")
	}

	output, err := s.runner.Output(
		"gh",
		"pr", "list",
		"--head", branch,
		"--state", "all",
		"--limit", fmt.Sprintf("%d", listLimit),
		"--json", listFields,
	)
	if err != nil {
		if isUnavailable(err) {
			return Info{}, false, fmt.Errorf("%w: %v", ErrUnavailable, err)
		}
		return Info{}, false, err
	}

	prs, err := parsePRList(output)
	if err != nil {
		return Info{}, false, err
	}
	info, found := selectPR(prs)
	return info, found, nil
}

// isUnavailable classifies an error as environmental rather than transient:
// the gh binary is missing, gh wants a login, or it cannot resolve a GitHub
// repository from the remotes. The stderr checks are heuristic string matches
// against gh's stable user guidance; anything unrecognized stays transient so
// the next lookup retries.
func isUnavailable(err error) bool {
	if errors.Is(err, exec.ErrNotFound) {
		return true
	}
	message := strings.ToLower(err.Error())
	for _, marker := range []string{
		"gh auth login",
		"could not determine base repo",
		"no default remote repository",
		"none of the git remotes",
		"not a git repository",
	} {
		if strings.Contains(message, marker) {
			return true
		}
	}
	return false
}
