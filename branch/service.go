package branch

import (
	"fmt"
	"sort"
	"strings"
)

type Scope string

const (
	ScopeLocal          Scope = "local"
	ScopeRemoteTracking Scope = "remote-tracking"
)

type Info struct {
	Name                string
	CheckedOutHere      bool
	CheckedOutElsewhere bool
}

type CommitInfo struct {
	Hash    string
	Author  string
	Subject string
}

type DeleteAllSummary struct {
	Deleted []string
	Skipped []string
}

type Runner interface {
	CombinedOutput(name string, args ...string) ([]byte, error)
}

type Service interface {
	List() ([]Info, Scope, error)
	RecentCommits(name string, limit int) ([]CommitInfo, error)
	ToggleScope() Scope
	Checkout(name string) error
	Delete(name string) error
	DeleteAllLocal() (DeleteAllSummary, error)
	Fetch() error
}

type service struct {
	runner Runner
	scope  Scope
}

// NewService returns a branch Service backed by the injected command runner.
func NewService(runner Runner) Service {
	return &service{runner: runner, scope: ScopeLocal}
}

func (s *service) List() ([]Info, Scope, error) {
	switch s.scope {
	case ScopeRemoteTracking:
		branches, err := s.remoteTrackingBranches()
		return branches, s.scope, err
	default:
		branches, err := s.localBranches()
		return branches, s.scope, err
	}
}

func (s *service) ToggleScope() Scope {
	if s.scope == ScopeRemoteTracking {
		s.scope = ScopeLocal
		return s.scope
	}
	s.scope = ScopeRemoteTracking
	return s.scope
}

func (s *service) Checkout(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("branch name is required")
	}

	if s.scope == ScopeRemoteTracking {
		_, err := s.runner.CombinedOutput("git", "switch", "--track", name)
		return err
	}

	_, err := s.runner.CombinedOutput("git", "switch", name)
	return err
}

func (s *service) Delete(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("branch name is required")
	}

	if s.scope == ScopeRemoteTracking {
		return fmt.Errorf("remote-tracking branches cannot be deleted")
	}

	_, err := s.runner.CombinedOutput("git", "branch", "-d", name)
	return err
}

func (s *service) RecentCommits(name string, limit int) ([]CommitInfo, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("branch name is required")
	}
	if limit <= 0 {
		limit = 10
	}

	output, err := s.runner.CombinedOutput(
		"git",
		"log",
		fmt.Sprintf("-n%d", limit),
		"--format=%h%x1f%an%x1f%s",
		name,
	)
	if err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return nil, nil
	}

	lines := strings.Split(trimmed, "\n")
	commits := make([]CommitInfo, 0, len(lines))
	for _, line := range lines {
		parts := strings.SplitN(line, "\x1f", 3)
		if len(parts) != 3 {
			continue
		}
		commits = append(commits, CommitInfo{
			Hash:    strings.TrimSpace(parts[0]),
			Author:  strings.TrimSpace(parts[1]),
			Subject: strings.TrimSpace(parts[2]),
		})
	}

	return commits, nil
}

func (s *service) DeleteAllLocal() (DeleteAllSummary, error) {
	branches, err := s.localBranches()
	if err != nil {
		return DeleteAllSummary{}, err
	}

	summary := DeleteAllSummary{
		Deleted: make([]string, 0, len(branches)),
		Skipped: make([]string, 0, len(branches)),
	}
	for _, branch := range branches {
		if branch.CheckedOutHere || branch.CheckedOutElsewhere {
			summary.Skipped = append(summary.Skipped, branch.Name)
			continue
		}
		if _, err := s.runner.CombinedOutput("git", "branch", "-D", branch.Name); err != nil {
			return summary, err
		}
		summary.Deleted = append(summary.Deleted, branch.Name)
	}

	return summary, nil
}

func (s *service) Fetch() error {
	_, err := s.runner.CombinedOutput("git", "fetch")
	return err
}

func (s *service) localBranches() ([]Info, error) {
	branchNames, err := s.branchNames("refs/heads")
	if err != nil {
		return nil, err
	}

	currentBranch, err := s.currentBranchName()
	if err != nil {
		return nil, err
	}

	checkedOutBranches, err := s.checkedOutWorktreeBranches()
	if err != nil {
		return nil, err
	}

	recencyOrder, err := s.localBranchRecencyOrder()
	if err != nil {
		return nil, err
	}

	branches := make([]Info, 0, len(branchNames))
	for _, name := range branchNames {
		checkedOutHere := currentBranch == name
		branches = append(branches, Info{
			Name:                name,
			CheckedOutHere:      checkedOutHere,
			CheckedOutElsewhere: checkedOutBranches[name] && !checkedOutHere,
		})
	}

	sort.SliceStable(branches, func(i, j int) bool {
		leftIndex, leftSeen := recencyOrder[branches[i].Name]
		rightIndex, rightSeen := recencyOrder[branches[j].Name]
		if leftSeen && rightSeen {
			return leftIndex < rightIndex
		}
		if leftSeen != rightSeen {
			return leftSeen
		}
		return branches[i].Name < branches[j].Name
	})

	return branches, nil
}

func (s *service) remoteTrackingBranches() ([]Info, error) {
	branchNames, err := s.branchNames("refs/remotes")
	if err != nil {
		return nil, err
	}

	branches := make([]Info, 0, len(branchNames))
	for _, name := range branchNames {
		if strings.HasSuffix(name, "/HEAD") {
			continue
		}
		branches = append(branches, Info{Name: name})
	}

	return branches, nil
}

func (s *service) branchNames(ref string) ([]string, error) {
	output, err := s.runner.CombinedOutput(
		"git",
		"for-each-ref",
		"--sort=-committerdate",
		"--format=%(refname:short)",
		ref,
	)
	if err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return nil, nil
	}

	lines := strings.Split(trimmed, "\n")
	branches := make([]string, 0, len(lines))
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		branches = append(branches, name)
	}

	return branches, nil
}

func (s *service) currentBranchName() (string, error) {
	output, err := s.runner.CombinedOutput("git", "branch", "--show-current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (s *service) checkedOutWorktreeBranches() (map[string]bool, error) {
	output, err := s.runner.CombinedOutput("git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	branches := map[string]bool{}
	for _, line := range strings.Split(string(output), "\n") {
		if !strings.HasPrefix(line, "branch ") {
			continue
		}
		name := normalizeBranch(strings.TrimSpace(strings.TrimPrefix(line, "branch ")))
		if name == "" {
			continue
		}
		branches[name] = true
	}

	return branches, nil
}

func (s *service) localBranchRecencyOrder() (map[string]int, error) {
	order, err := s.headCheckoutRecencyOrder()
	if err != nil {
		return nil, err
	}
	if len(order) > 0 {
		return order, nil
	}

	return s.branchReflogRecencyOrder()
}

func (s *service) headCheckoutRecencyOrder() (map[string]int, error) {
	output, err := s.runner.CombinedOutput(
		"git",
		"reflog",
		"show",
		"--format=%gs",
		"HEAD",
	)
	if err != nil {
		return nil, err
	}

	order := map[string]int{}
	index := 0
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		name := headCheckoutBranchName(line)
		if name == "" {
			continue
		}
		if _, ok := order[name]; ok {
			continue
		}
		order[name] = index
		index++
	}

	return order, nil
}

func (s *service) branchReflogRecencyOrder() (map[string]int, error) {
	output, err := s.runner.CombinedOutput(
		"git",
		"reflog",
		"show",
		"--format=%gD",
		"--all",
	)
	if err != nil {
		return nil, err
	}

	order := map[string]int{}
	index := 0
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		name := reflogBranchName(line)
		if name == "" {
			continue
		}
		if _, ok := order[name]; ok {
			continue
		}
		order[name] = index
		index++
	}

	return order, nil
}
