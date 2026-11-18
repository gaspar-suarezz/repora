package gitcli

import (
	"testing"

	"github.com/gasparsuarez/repora/internal/domain"
)

const nul = "\x00"

func TestParseStatusClean(t *testing.T) {
	out := "# branch.oid 1111111111111111111111111111111111111111\n" +
		"# branch.head main\n" +
		"# branch.upstream origin/main\n" +
		"# branch.ab +0 -0\n"

	state := ParseStatus("api", []byte(out))

	if state.Branch.Head != "main" {
		t.Errorf("head = %q, want main", state.Branch.Head)
	}
	if !state.Branch.HasUpstream || state.Branch.Upstream != "origin/main" {
		t.Errorf("upstream = %q (has=%v), want origin/main", state.Branch.Upstream, state.Branch.HasUpstream)
	}
	if state.Branch.Ahead != 0 || state.Branch.Behind != 0 {
		t.Errorf("ahead/behind = %d/%d, want 0/0", state.Branch.Ahead, state.Branch.Behind)
	}
	if state.Class() != domain.ClassClean {
		t.Errorf("class = %q, want clean", state.Class())
	}
}

func TestParseStatusDirtyMixed(t *testing.T) {
	out := "# branch.head feature/x\n" +
		"1 MM N... 100644 100644 100644 aaa bbb handler.go" + nul +
		"1 .M N... 100644 100644 100644 aaa bbb README.md" + nul +
		"1 A. N... 000000 100644 100644 000 aaa new.go" + nul +
		"? notes/my file.md" + nul +
		"? scripts/debug.sh" + nul

	state := ParseStatus("api", []byte(out))

	if got := state.StagedCount(); got != 2 {
		t.Errorf("staged = %d, want 2", got)
	}
	if got := state.UnstagedCount(); got != 2 {
		t.Errorf("unstaged = %d, want 2", got)
	}
	if got := state.UntrackedCount(); got != 2 {
		t.Errorf("untracked = %d, want 2", got)
	}
	if state.WorkingTree.Untracked[0] != "notes/my file.md" {
		t.Errorf("untracked[0] = %q", state.WorkingTree.Untracked[0])
	}

	found := false
	for _, change := range state.WorkingTree.Staged {
		if change.Path == "handler.go" {
			found = true
			if change.Index != domain.CodeModified || change.Worktree != domain.CodeModified {
				t.Errorf("handler.go codes = %q/%q", change.Index, change.Worktree)
			}
		}
	}
	if !found {
		t.Error("handler.go missing from staged entries")
	}
}

func TestParseStatusRenameKeepsOriginalPath(t *testing.T) {
	out := "2 R. N... 100644 100644 100644 aaa bbb R100 internal/new/path.go" + nul + "internal/old/path.go" + nul

	state := ParseStatus("api", []byte(out))

	if len(state.WorkingTree.Staged) != 1 {
		t.Fatalf("staged = %d, want 1", len(state.WorkingTree.Staged))
	}
	change := state.WorkingTree.Staged[0]
	if change.Path != "internal/new/path.go" {
		t.Errorf("path = %q", change.Path)
	}
	if change.OrigPath != "internal/old/path.go" {
		t.Errorf("origPath = %q, want internal/old/path.go", change.OrigPath)
	}
}

func TestParseStatusConflict(t *testing.T) {
	out := "u UU N... 100644 100644 100644 100644 aaa bbb ccc src/conflict.go" + nul

	state := ParseStatus("api", []byte(out))

	if state.ConflictCount() != 1 {
		t.Fatalf("conflicts = %d, want 1", state.ConflictCount())
	}
	if state.Class() != domain.ClassConflicts {
		t.Errorf("class = %q, want conflicts", state.Class())
	}
	if got := state.StagedCount() + state.UnstagedCount(); got != 0 {
		t.Errorf("conflicted file counted in staged/unstaged: %d", got)
	}
}

func TestParseStatusUnbornAndDetached(t *testing.T) {
	unborn := ParseStatus("empty", []byte("# branch.oid (initial)\n# branch.head main\n"))
	if !unborn.Branch.Unborn {
		t.Error("expected unborn branch")
	}
	if unborn.Branch.Label() != "(no commits)" {
		t.Errorf("label = %q", unborn.Branch.Label())
	}

	detached := ParseStatus("api", []byte("# branch.oid abc123\n# branch.head (detached)\n"))
	if !detached.Branch.Detached {
		t.Error("expected detached HEAD")
	}
	if detached.Branch.Label() != "HEAD (detached)" {
		t.Errorf("label = %q", detached.Branch.Label())
	}
}

func TestParseStatusDiverged(t *testing.T) {
	out := "# branch.head main\n# branch.upstream origin/main\n# branch.ab +2 -3\n"
	state := ParseStatus("api", []byte(out))

	if state.Branch.Ahead != 2 || state.Branch.Behind != 3 {
		t.Fatalf("ahead/behind = %d/%d, want 2/3", state.Branch.Ahead, state.Branch.Behind)
	}
	if state.Class() != domain.ClassDiverged {
		t.Errorf("class = %q, want diverged", state.Class())
	}
}

func TestParseStatusEmptyInput(t *testing.T) {
	state := ParseStatus("api", nil)
	if state.Class() != domain.ClassClean {
		t.Errorf("class = %q, want clean", state.Class())
	}
	if state.RepositoryID != "api" {
		t.Errorf("repoID = %q", state.RepositoryID)
	}
}

// FuzzParseStatus asserts the parser never panics on arbitrary input.
func FuzzParseStatus(f *testing.F) {
	f.Add([]byte("# branch.head main\n"))
	f.Add([]byte("1 .M N... 100644 100644 100644 aaa bbb file.go\x00"))
	f.Add([]byte("2 R. N... 100644 100644 100644 aaa bbb R100 new\x00old\x00"))
	f.Add([]byte("u UU N... 100644 100644 100644 100644 a b c x\x00"))
	f.Add([]byte("? weird\x00"))

	f.Fuzz(func(t *testing.T, data []byte) {
		_ = ParseStatus("api", data)
	})
}
