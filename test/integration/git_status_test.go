package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gasparsuarez/repora/internal/domain"
	"github.com/gasparsuarez/repora/internal/infra/gitcli"
)

// gitRun runs git inside dir with an isolated configuration so the developer's
// identity and aliases cannot leak into the fixtures.
func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()

	full := append([]string{
		"-c", "user.name=Repora Test",
		"-c", "user.email=test@example.com",
		"-c", "init.defaultBranch=main",
		"-c", "commit.gpgsign=false",
	}, args...)

	cmd := exec.Command("git", full...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_TERMINAL_PROMPT=0",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func newAdapter(t *testing.T) *gitcli.Repository {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	return gitcli.NewRepository(gitcli.NewExecutor())
}

func TestRootResolvesRepositoryRoot(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q")

	nested := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(nested, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	root, err := adapter.Root(context.Background(), nested)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}

	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		want = dir
	}
	if root != want {
		t.Fatalf("root = %q, want %q", root, want)
	}
}

func TestRootRejectsNonRepository(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()

	_, err := adapter.Root(context.Background(), dir)
	if err == nil {
		t.Fatal("expected error for non-repository path")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStatusUnbornRepository(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q")

	state := adapter.Status(context.Background(), domain.Repository{ID: "r", Name: "R", Path: dir})
	if state.Err != nil {
		t.Fatalf("Status: %v", state.Err)
	}
	if !state.Branch.Unborn {
		t.Errorf("expected unborn branch, got %+v", state.Branch)
	}
}

func TestStatusCleanThenDirty(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q")

	readme := filepath.Join(dir, "README.md")
	writeFile(t, readme, "hello\n")
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-qm", "initial")

	repo := domain.Repository{ID: "r", Name: "R", Path: dir}

	state := adapter.Status(context.Background(), repo)
	if state.Err != nil {
		t.Fatalf("Status: %v", state.Err)
	}
	if !state.IsClean() {
		t.Fatalf("expected clean, got %+v", state.WorkingTree)
	}
	if state.Branch.Head != "main" {
		t.Errorf("head = %q, want main", state.Branch.Head)
	}

	writeFile(t, readme, "hello\nworld\n")
	writeFile(t, filepath.Join(dir, "notes", "todo.md"), "- item\n")

	state = adapter.Status(context.Background(), repo)
	if state.Err != nil {
		t.Fatalf("Status: %v", state.Err)
	}
	if got := state.UnstagedCount(); got != 1 {
		t.Errorf("unstaged = %d, want 1", got)
	}
	if got := state.UntrackedCount(); got != 1 {
		t.Errorf("untracked = %d, want 1", got)
	}
	if state.Class() != domain.ClassDirty {
		t.Errorf("class = %q, want modified", state.Class())
	}
}

func TestStatusStagedRenameAndPathWithSpaces(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q")

	writeFile(t, filepath.Join(dir, "old name.md"), "content\n")
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-qm", "initial")
	gitRun(t, dir, "mv", "old name.md", "new name.md")

	state := adapter.Status(context.Background(), domain.Repository{ID: "r", Name: "R", Path: dir})
	if state.Err != nil {
		t.Fatalf("Status: %v", state.Err)
	}
	if got := state.StagedCount(); got != 1 {
		t.Fatalf("staged = %d, want 1", got)
	}

	change := state.WorkingTree.Staged[0]
	if change.Path != "new name.md" {
		t.Errorf("path = %q, want %q", change.Path, "new name.md")
	}
	if change.OrigPath != "old name.md" {
		t.Errorf("origPath = %q, want %q", change.OrigPath, "old name.md")
	}
}

func TestStatusUnicodePath(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q")

	writeFile(t, filepath.Join(dir, "notas", "acción.md"), "contenido\n")

	state := adapter.Status(context.Background(), domain.Repository{ID: "r", Name: "R", Path: dir})
	if state.Err != nil {
		t.Fatalf("Status: %v", state.Err)
	}
	if got := state.UntrackedCount(); got != 1 {
		t.Fatalf("untracked = %d, want 1", got)
	}
	if got := state.WorkingTree.Untracked[0]; got != "notas/acción.md" {
		t.Errorf("untracked = %q", got)
	}
}

func TestStatusAheadBehindWithLocalRemote(t *testing.T) {
	adapter := newAdapter(t)
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	work := filepath.Join(root, "work")

	if err := os.MkdirAll(remote, 0o750); err != nil {
		t.Fatalf("mkdir remote: %v", err)
	}
	gitRun(t, remote, "init", "-q", "--bare")
	if err := os.MkdirAll(work, 0o750); err != nil {
		t.Fatalf("mkdir work: %v", err)
	}
	gitRun(t, work, "init", "-q")
	writeFile(t, filepath.Join(work, "README.md"), "hello\n")
	gitRun(t, work, "add", ".")
	gitRun(t, work, "commit", "-qm", "initial")
	gitRun(t, work, "remote", "add", "origin", remote)
	gitRun(t, work, "push", "-q", "-u", "origin", "main")

	repo := domain.Repository{ID: "r", Name: "R", Path: work}

	state := adapter.Status(context.Background(), repo)
	if state.Err != nil {
		t.Fatalf("Status: %v", state.Err)
	}
	if !state.Branch.HasUpstream || state.Branch.Upstream != "origin/main" {
		t.Fatalf("upstream = %q (has=%v)", state.Branch.Upstream, state.Branch.HasUpstream)
	}
	if state.Branch.Ahead != 0 || state.Branch.Behind != 0 {
		t.Fatalf("ahead/behind = %d/%d, want 0/0", state.Branch.Ahead, state.Branch.Behind)
	}

	writeFile(t, filepath.Join(work, "second.txt"), "two\n")
	gitRun(t, work, "add", ".")
	gitRun(t, work, "commit", "-qm", "second")

	state = adapter.Status(context.Background(), repo)
	if state.Branch.Ahead != 1 {
		t.Fatalf("ahead = %d, want 1", state.Branch.Ahead)
	}
	if state.Class() != domain.ClassAhead {
		t.Errorf("class = %q, want ahead", state.Class())
	}
}

func TestStatusConflict(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q")

	file := filepath.Join(dir, "conflict.txt")
	writeFile(t, file, "base\n")
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-qm", "base")

	gitRun(t, dir, "checkout", "-qb", "other")
	writeFile(t, file, "other\n")
	gitRun(t, dir, "commit", "-qam", "other")

	gitRun(t, dir, "checkout", "-q", "main")
	writeFile(t, file, "main\n")
	gitRun(t, dir, "commit", "-qam", "main")

	// The merge is expected to fail; run it directly.
	cmd := exec.Command("git", "-c", "user.name=Repora Test", "-c", "user.email=test@example.com", "merge", "other")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	_ = cmd.Run()

	state := adapter.Status(context.Background(), domain.Repository{ID: "r", Name: "R", Path: dir})
	if state.Err != nil {
		t.Fatalf("Status: %v", state.Err)
	}
	if state.ConflictCount() != 1 {
		t.Fatalf("conflicts = %d, want 1", state.ConflictCount())
	}
	if state.Class() != domain.ClassConflicts {
		t.Errorf("class = %q, want conflicts", state.Class())
	}
}
