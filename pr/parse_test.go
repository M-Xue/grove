package pr

import (
	"reflect"
	"testing"
)

// openPRFixture is a trimmed capture of `gh pr list --json` output for an open
// PR with one of each rollup shape.
const openPRFixture = `[
  {
    "number": 2745,
    "title": "feat(remote-tasks): gate remote annotation URLs",
    "state": "OPEN",
    "isDraft": false,
    "url": "https://github.com/tektoncd/pipelines-as-code/pull/2745",
    "author": {"login": "chmouel"},
    "statusCheckRollup": [
      {"__typename": "CheckRun", "name": "e2e tests (gitea_1)", "status": "COMPLETED", "conclusion": "SUCCESS"},
      {"__typename": "CheckRun", "name": "linters", "status": "IN_PROGRESS", "conclusion": ""},
      {"__typename": "CheckRun", "name": "optional", "status": "COMPLETED", "conclusion": "SKIPPED"},
      {"__typename": "StatusContext", "context": "EasyCLA", "state": "SUCCESS"}
    ]
  }
]`

func TestParsePRListNormalizesFieldsAndChecks(t *testing.T) {
	prs, err := parsePRList([]byte(openPRFixture))
	if err != nil {
		t.Fatalf("parsePRList returned error: %v", err)
	}
	want := []Info{{
		Number: 2745,
		Title:  "feat(remote-tasks): gate remote annotation URLs",
		Status: StatusOpen,
		Author: "chmouel",
		URL:    "https://github.com/tektoncd/pipelines-as-code/pull/2745",
		Checks: []Check{
			{Name: "e2e tests (gitea_1)", State: CheckPassed},
			{Name: "linters", State: CheckPending},
			{Name: "optional", State: CheckSkipped},
			{Name: "EasyCLA", State: CheckPassed},
		},
	}}
	if !reflect.DeepEqual(prs, want) {
		t.Fatalf("unexpected parse result:\ngot  %#v\nwant %#v", prs, want)
	}
}

func TestParsePRListEmptyArrayMeansNoPRs(t *testing.T) {
	prs, err := parsePRList([]byte("[]\n"))
	if err != nil {
		t.Fatalf("parsePRList returned error: %v", err)
	}
	if len(prs) != 0 {
		t.Fatalf("expected no PRs, got %#v", prs)
	}
}

func TestParsePRListRejectsNonJSON(t *testing.T) {
	if _, err := parsePRList([]byte("gh: not logged in")); err == nil {
		t.Fatal("expected an error for non-JSON input")
	}
}

func TestNormalizeStatusFoldsDraftAndTerminalStates(t *testing.T) {
	cases := []struct {
		state   string
		isDraft bool
		want    Status
	}{
		{"OPEN", false, StatusOpen},
		{"OPEN", true, StatusDraft},
		{"MERGED", false, StatusMerged},
		{"MERGED", true, StatusMerged},
		{"CLOSED", false, StatusClosed},
		{"SOMETHING_NEW", false, StatusClosed},
	}
	for _, tc := range cases {
		if got := normalizeStatus(tc.state, tc.isDraft); got != tc.want {
			t.Fatalf("normalizeStatus(%q, %v) = %q, want %q", tc.state, tc.isDraft, got, tc.want)
		}
	}
}

func TestCheckRunStateMapsConclusions(t *testing.T) {
	cases := []struct {
		status, conclusion string
		want               CheckState
	}{
		{"QUEUED", "", CheckPending},
		{"IN_PROGRESS", "", CheckPending},
		{"COMPLETED", "SUCCESS", CheckPassed},
		{"COMPLETED", "FAILURE", CheckFailed},
		{"COMPLETED", "TIMED_OUT", CheckFailed},
		{"COMPLETED", "ACTION_REQUIRED", CheckFailed},
		{"COMPLETED", "CANCELLED", CheckCancelled},
		{"COMPLETED", "SKIPPED", CheckSkipped},
		{"COMPLETED", "NEUTRAL", CheckNeutral},
		{"COMPLETED", "SOMETHING_NEW", CheckNeutral},
	}
	for _, tc := range cases {
		if got := checkRunState(tc.status, tc.conclusion); got != tc.want {
			t.Fatalf("checkRunState(%q, %q) = %q, want %q", tc.status, tc.conclusion, got, tc.want)
		}
	}
}

func TestSelectPRPrefersOpenThenMostRecent(t *testing.T) {
	merged := Info{Number: 10, Status: StatusMerged}
	closed := Info{Number: 30, Status: StatusClosed}
	open := Info{Number: 20, Status: StatusOpen}
	draft := Info{Number: 25, Status: StatusDraft}

	if got, found := selectPR([]Info{merged, open, closed, draft}); !found || got.Number != 25 {
		t.Fatalf("expected the most recent open/draft PR (25), got %#v found=%v", got, found)
	}
	if got, found := selectPR([]Info{merged, closed}); !found || got.Number != 30 {
		t.Fatalf("expected the most recent PR (30), got %#v found=%v", got, found)
	}
	if _, found := selectPR(nil); found {
		t.Fatal("expected no selection from an empty list")
	}
}

func TestSummarizeChecks(t *testing.T) {
	cases := []struct {
		name   string
		checks []Check
		want   CheckState
	}{
		{"empty", nil, CheckNeutral},
		{"all passed", []Check{{State: CheckPassed}, {State: CheckPassed}}, CheckPassed},
		{"failure dominates", []Check{{State: CheckPassed}, {State: CheckFailed}, {State: CheckPending}}, CheckFailed},
		{"cancelled counts as failed", []Check{{State: CheckPassed}, {State: CheckCancelled}}, CheckFailed},
		{"pending beats passed", []Check{{State: CheckPassed}, {State: CheckPending}}, CheckPending},
		{"skips are inconsequential", []Check{{State: CheckSkipped}, {State: CheckNeutral}}, CheckNeutral},
		{"passed among skips", []Check{{State: CheckSkipped}, {State: CheckPassed}}, CheckPassed},
	}
	for _, tc := range cases {
		if got := SummarizeChecks(tc.checks); got != tc.want {
			t.Fatalf("%s: SummarizeChecks = %q, want %q", tc.name, got, tc.want)
		}
	}
}
