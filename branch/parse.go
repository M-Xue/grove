package branch

import "strings"

// headCheckoutBranchName extracts the destination branch from a HEAD reflog
// "checkout: moving from X to Y" line, returning "" for lines that do not name
// a concrete branch (refs, detached HEAD, malformed input).
func headCheckoutBranchName(line string) string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "checkout: moving from ") {
		return ""
	}
	parts := strings.Split(line, " to ")
	if len(parts) != 2 {
		return ""
	}
	name := strings.TrimSpace(parts[1])
	if name == "" || strings.HasPrefix(name, "refs/") || strings.EqualFold(name, "HEAD") {
		return ""
	}
	return name
}

// reflogBranchName extracts the branch name from a reflog selector line such as
// "refs/heads/feature@{0}", stripping the refs/heads/ or refs/remotes/ prefix
// and the trailing @{…} selector.
func reflogBranchName(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	line = strings.TrimSuffix(line, "}")
	prefixes := []string{"refs/heads/", "refs/remotes/"}
	for _, prefix := range prefixes {
		start := strings.Index(line, prefix)
		if start == -1 {
			continue
		}
		name := line[start+len(prefix):]
		if at := strings.Index(name, "@"); at != -1 {
			name = name[:at]
		}
		return strings.TrimSpace(name)
	}
	return ""
}

// normalizeBranch trims surrounding whitespace and strips a leading refs/heads/
// prefix, yielding the short branch name.
func normalizeBranch(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.TrimPrefix(value, "refs/heads/")
}
