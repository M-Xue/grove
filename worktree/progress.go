package worktree

import (
	"regexp"
	"strconv"
)

// Progress reports how far a worktree checkout has advanced, as the number of
// files written so far out of the total to write.
type Progress struct {
	Done  int
	Total int
}

// updatingFilesRe matches git's checkout progress line, e.g.
// "Updating files:  47% (27/57)" or the terminal "Updating files: 100%
// (57/57), done." Only the file counts are captured; the percent text is
// recomputed from them so the reported percentage stays exact.
var updatingFilesRe = regexp.MustCompile(`Updating files:\s+\d+% \((\d+)/(\d+)\)`)

// parseUpdatingFiles extracts checkout progress from a single git stderr line.
// It reports ok=false for any line that is not a checkout-progress line (e.g.
// "Preparing worktree …") or whose total is not a positive number.
func parseUpdatingFiles(line string) (Progress, bool) {
	m := updatingFilesRe.FindStringSubmatch(line)
	if m == nil {
		return Progress{}, false
	}
	done, err1 := strconv.Atoi(m[1])
	total, err2 := strconv.Atoi(m[2])
	if err1 != nil || err2 != nil || total <= 0 {
		return Progress{}, false
	}
	return Progress{Done: done, Total: total}, true
}
