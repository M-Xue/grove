// Package cache persists grove's last-known worktree list to disk so the
// change screen can paint instantly on launch while the authoritative list is
// refreshed in the background (stale-while-revalidate).
//
// The cache is never load-bearing: any read failure — missing, unreadable, or
// malformed file — is reported as a miss so callers fall back to a live git
// listing. Writes are atomic (temp file + rename) so a crash mid-write cannot
// leave a half-written file that breaks the next launch.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/M-Xue/grove/worktree"
)

// subdir namespaces grove's worktree caches beneath the OS cache root.
const subdir = "grove/worktrees"

// Cache reads and writes per-repo worktree lists rooted at a base directory.
type Cache struct {
	dir string
}

// New returns a Cache rooted at the OS user cache directory. The second return
// is false when that directory cannot be resolved (e.g. a headless environment
// with no HOME), in which case caching should be disabled entirely.
func New() (Cache, bool) {
	root, err := os.UserCacheDir()
	if err != nil {
		return Cache{}, false
	}
	return Cache{dir: filepath.Join(root, subdir)}, true
}

// NewAt returns a Cache rooted at dir. It is used by tests to isolate the cache
// from the real user cache directory.
func NewAt(dir string) Cache {
	return Cache{dir: dir}
}

// Load returns the cached worktree list for key and whether it was a hit. Any
// failure — no file, unreadable, malformed JSON — is a miss (nil, false); Load
// never surfaces an error, keeping the cache non-load-bearing. An empty cached
// list is a legitimate hit.
func (c Cache) Load(key string) ([]worktree.Info, bool) {
	data, err := os.ReadFile(c.pathFor(key))
	if err != nil {
		return nil, false
	}
	var list []worktree.Info
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, false
	}
	if list == nil {
		list = []worktree.Info{}
	}
	return list, true
}

// Save writes list as the cached worktree list for key. It creates the cache
// directory if needed and writes atomically: the JSON is written to a temp file
// in the same directory and renamed over the target, which replaces the
// destination wholesale so a corrupt or truncated file is fully overwritten.
func (c Cache) Save(key string, list []worktree.Info) error {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}

	target := c.pathFor(key)
	tmp, err := os.CreateTemp(c.dir, filepath.Base(target)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Best-effort cleanup if we fail before the rename succeeds.
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, target)
}

// pathFor returns the cache file path for a repo key. The key (a repo's
// absolute common git dir) is hashed so the filename is fixed-length and free
// of path separators and illegal characters across platforms.
func (c Cache) pathFor(key string) string {
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(c.dir, hex.EncodeToString(sum[:])+".json")
}
