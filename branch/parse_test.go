package branch

import "testing"

func TestHeadCheckoutBranchName(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{name: "checkout to branch", line: "checkout: moving from main to feature/a", want: "feature/a"},
		{name: "leading whitespace", line: "  checkout: moving from feature/a to main", want: "main"},
		{name: "not a checkout line", line: "commit: add thing", want: ""},
		{name: "destination is a ref", line: "checkout: moving from main to refs/heads/x", want: ""},
		{name: "destination is HEAD", line: "checkout: moving from main to HEAD", want: ""},
		{name: "missing to separator", line: "checkout: moving from main", want: ""},
		{name: "empty", line: "", want: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := headCheckoutBranchName(test.line); got != test.want {
				t.Fatalf("headCheckoutBranchName(%q) = %q, want %q", test.line, got, test.want)
			}
		})
	}
}

func TestReflogBranchName(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{name: "local ref with selector", line: "refs/heads/feature/a@{0}", want: "feature/a"},
		{name: "remote ref with selector", line: "refs/remotes/origin/main@{2}", want: "origin/main"},
		{name: "local ref without selector", line: "refs/heads/main", want: "main"},
		{name: "trailing brace trimmed", line: "refs/heads/main@{1}", want: "main"},
		{name: "no recognized prefix", line: "HEAD@{0}", want: ""},
		{name: "empty", line: "", want: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := reflogBranchName(test.line); got != test.want {
				t.Fatalf("reflogBranchName(%q) = %q, want %q", test.line, got, test.want)
			}
		})
	}
}

func TestNormalizeBranch(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "strips refs/heads prefix", value: "refs/heads/feature/a", want: "feature/a"},
		{name: "trims whitespace", value: "  main  ", want: "main"},
		{name: "plain name unchanged", value: "main", want: "main"},
		{name: "empty", value: "", want: ""},
		{name: "whitespace only", value: "   ", want: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizeBranch(test.value); got != test.want {
				t.Fatalf("normalizeBranch(%q) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}
