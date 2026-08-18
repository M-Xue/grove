package repo

import (
	"errors"
	"strings"
)

// ErrNoCommonDir is returned when git reports no common directory for the
// current repository, so a stable cache key cannot be derived.
var ErrNoCommonDir = errors.New("could not determine repository common git dir")

// WorktreeCacheKey returns a stable identifier for the repository, used to key
// its persisted worktree cache. It is the absolute common git directory, which
// every worktree of a repository shares — so all of them resolve to the same
// cache entry. The path-format flag forces an absolute path regardless of the
// directory grove was launched from.
func WorktreeCacheKey(runner Runner) (string, error) {
	output, err := runner.CombinedOutput(
		"git", "rev-parse", "--path-format=absolute", "--git-common-dir",
	)
	if err != nil {
		return "", err
	}
	key := strings.TrimSpace(string(output))
	if key == "" {
		return "", ErrNoCommonDir
	}
	return key, nil
}
