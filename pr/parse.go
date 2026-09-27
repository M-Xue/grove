package pr

import (
	"encoding/json"
	"fmt"
)

// listedPR mirrors one element of `gh pr list --json` output. Only the fields
// grove consumes are declared; unknown fields are ignored by encoding/json.
type listedPR struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	IsDraft bool   `json:"isDraft"`
	URL     string `json:"url"`
	Author  struct {
		Login string `json:"login"`
	} `json:"author"`
	StatusCheckRollup []rollupNode `json:"statusCheckRollup"`
}

// rollupNode is one entry of statusCheckRollup. gh emits two shapes keyed by
// __typename: a CheckRun (GitHub Actions et al.) carries name/status/conclusion,
// while a StatusContext (commit-status API) carries context/state. The struct
// is the union of both; normalization picks the fields for the actual shape.
type rollupNode struct {
	TypeName   string `json:"__typename"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Context    string `json:"context"`
	State      string `json:"state"`
}

// parsePRList converts `gh pr list --json` output into Infos, in gh's order.
func parsePRList(data []byte) ([]Info, error) {
	var listed []listedPR
	if err := json.Unmarshal(data, &listed); err != nil {
		return nil, fmt.Errorf("parsing gh pr list output: %w", err)
	}
	infos := make([]Info, 0, len(listed))
	for _, entry := range listed {
		infos = append(infos, Info{
			Number: entry.Number,
			Title:  entry.Title,
			Status: normalizeStatus(entry.State, entry.IsDraft),
			Author: entry.Author.Login,
			URL:    entry.URL,
			Checks: normalizeChecks(entry.StatusCheckRollup),
		})
	}
	return infos, nil
}

// selectPR picks the PR to surface for a branch when gh returns several with
// the same head: the most recent open (or draft) one wins, else the most
// recent of any state. Recency follows PR number.
func selectPR(prs []Info) (Info, bool) {
	best, found := Info{}, false
	bestOpen, foundOpen := Info{}, false
	for _, pr := range prs {
		if !found || pr.Number > best.Number {
			best, found = pr, true
		}
		if pr.Status == StatusOpen || pr.Status == StatusDraft {
			if !foundOpen || pr.Number > bestOpen.Number {
				bestOpen, foundOpen = pr, true
			}
		}
	}
	if foundOpen {
		return bestOpen, true
	}
	return best, found
}

// normalizeStatus folds gh's state/isDraft pair into one Status. gh reports a
// draft PR as state OPEN with isDraft set.
func normalizeStatus(state string, isDraft bool) Status {
	switch state {
	case "OPEN":
		if isDraft {
			return StatusDraft
		}
		return StatusOpen
	case "MERGED":
		return StatusMerged
	case "CLOSED":
		return StatusClosed
	default:
		return StatusClosed
	}
}

func normalizeChecks(nodes []rollupNode) []Check {
	if len(nodes) == 0 {
		return nil
	}
	checks := make([]Check, 0, len(nodes))
	for _, node := range nodes {
		if node.TypeName == "StatusContext" {
			checks = append(checks, Check{Name: node.Context, State: statusContextState(node.State)})
			continue
		}
		checks = append(checks, Check{Name: node.Name, State: checkRunState(node.Status, node.Conclusion)})
	}
	return checks
}

// checkRunState maps a CheckRun's status/conclusion pair: anything not yet
// completed is pending; a completed run resolves by its conclusion.
func checkRunState(status, conclusion string) CheckState {
	if status != "COMPLETED" {
		return CheckPending
	}
	switch conclusion {
	case "SUCCESS":
		return CheckPassed
	case "SKIPPED":
		return CheckSkipped
	case "CANCELLED":
		return CheckCancelled
	case "NEUTRAL":
		return CheckNeutral
	case "FAILURE", "TIMED_OUT", "ACTION_REQUIRED", "STARTUP_FAILURE":
		return CheckFailed
	default:
		return CheckNeutral
	}
}

// SummarizeChecks folds a PR's checks into one rollup state for a summary
// line: any failure or cancellation dominates, then anything still pending,
// then passed. A list with no consequential checks (empty, or all
// skipped/neutral) is neutral.
func SummarizeChecks(checks []Check) CheckState {
	pending, passed := false, false
	for _, check := range checks {
		switch check.State {
		case CheckFailed, CheckCancelled:
			return CheckFailed
		case CheckPending:
			pending = true
		case CheckPassed:
			passed = true
		}
	}
	if pending {
		return CheckPending
	}
	if passed {
		return CheckPassed
	}
	return CheckNeutral
}

func statusContextState(state string) CheckState {
	switch state {
	case "SUCCESS":
		return CheckPassed
	case "PENDING", "EXPECTED":
		return CheckPending
	case "FAILURE", "ERROR":
		return CheckFailed
	default:
		return CheckNeutral
	}
}
