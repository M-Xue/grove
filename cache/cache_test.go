package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/M-Xue/grove/worktree"
)

const testKey = "/home/user/project/.git"

func sampleList() []worktree.Info {
	return []worktree.Info{
		{
			Path:                  "/home/user/project",
			Branch:                "main",
			CommitLabel:           "initial commit",
			CommitHash:            "abc123",
			HasUncommittedChanges: true,
			CreatedAt:             time.Date(2026, 8, 18, 10, 0, 0, 0, time.UTC),
		},
		{
			Path:       "/home/user/project-feature",
			Branch:     "feature",
			CommitHash: "def456",
		},
	}
}

func TestSaveThenLoadRoundTrips(t *testing.T) {
	c := NewAt(t.TempDir())
	want := sampleList()

	if err := c.Save(testKey, want); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	got, hit := c.Load(testKey)
	if !hit {
		t.Fatal("Load reported a miss after Save")
	}
	if len(got) != len(want) {
		t.Fatalf("Load returned %d entries, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Path != want[i].Path ||
			got[i].Branch != want[i].Branch ||
			got[i].CommitLabel != want[i].CommitLabel ||
			got[i].CommitHash != want[i].CommitHash ||
			got[i].HasUncommittedChanges != want[i].HasUncommittedChanges ||
			!got[i].CreatedAt.Equal(want[i].CreatedAt) {
			t.Fatalf("entry %d round-tripped incorrectly:\n got %+v\nwant %+v", i, got[i], want[i])
		}
	}
}

func TestLoadMissWhenNoFileExists(t *testing.T) {
	c := NewAt(t.TempDir())
	if _, hit := c.Load(testKey); hit {
		t.Fatal("Load reported a hit with no file present")
	}
}

func TestLoadMissOnMalformedFile(t *testing.T) {
	dir := t.TempDir()
	c := NewAt(dir)
	// Save first so the target path exists, then corrupt it.
	if err := c.Save(testKey, sampleList()); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if err := os.WriteFile(c.pathFor(testKey), []byte("{not json"), 0o644); err != nil {
		t.Fatalf("failed to corrupt cache file: %v", err)
	}

	if _, hit := c.Load(testKey); hit {
		t.Fatal("Load reported a hit on a malformed file")
	}
}

func TestEmptyListIsAHitNotAMiss(t *testing.T) {
	c := NewAt(t.TempDir())
	if err := c.Save(testKey, []worktree.Info{}); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	got, hit := c.Load(testKey)
	if !hit {
		t.Fatal("an empty cached list should be a hit, not a miss")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(got))
	}
}

func TestSaveCreatesMissingDirectory(t *testing.T) {
	// Root at a path that does not exist yet; Save must create it.
	root := filepath.Join(t.TempDir(), "does", "not", "exist")
	c := NewAt(root)
	if err := c.Save(testKey, sampleList()); err != nil {
		t.Fatalf("Save did not create its directory: %v", err)
	}
	if _, hit := c.Load(testKey); !hit {
		t.Fatal("Load missed after Save created the directory")
	}
}

func TestSaveLeavesNoTempFileBehind(t *testing.T) {
	dir := t.TempDir()
	c := NewAt(dir)
	if err := c.Save(testKey, sampleList()); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read cache dir: %v", err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".tmp" {
			t.Fatalf("Save left a temp file behind: %s", entry.Name())
		}
	}
}

func TestDifferentKeysDoNotCollide(t *testing.T) {
	c := NewAt(t.TempDir())
	listA := []worktree.Info{{Path: "/repo-a", Branch: "main"}}
	listB := []worktree.Info{{Path: "/repo-b", Branch: "dev"}}

	if err := c.Save("/home/a/.git", listA); err != nil {
		t.Fatalf("Save A: %v", err)
	}
	if err := c.Save("/home/b/.git", listB); err != nil {
		t.Fatalf("Save B: %v", err)
	}

	gotA, hitA := c.Load("/home/a/.git")
	gotB, hitB := c.Load("/home/b/.git")
	if !hitA || !hitB {
		t.Fatalf("expected hits for both keys, got A=%v B=%v", hitA, hitB)
	}
	if gotA[0].Path != "/repo-a" || gotB[0].Path != "/repo-b" {
		t.Fatalf("keys collided: A=%q B=%q", gotA[0].Path, gotB[0].Path)
	}
}
