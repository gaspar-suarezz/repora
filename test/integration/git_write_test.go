package integration

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gasparsuarez/repora/internal/domain"
)

func TestStageAndCommit(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	repo := seedRepository(t, dir)

	writeFile(t, filepath.Join(dir, "README.md"), "hello\nworld\n")

	if err := adapter.Stage(context.Background(), repo, []string{"README.md"}); err != nil {
		t.Fatalf("Stage: %v", err)
	}

	state := adapter.Status(context.Background(), repo)
	if state.StagedCount() != 1 {
		t.Fatalf("staged = %d, want 1", state.StagedCount())
	}
	if state.UnstagedCount() != 0 {
		t.Errorf("unstaged = %d, want 0", state.UnstagedCount())
	}

	if err := adapter.Commit(context.Background(), repo, "update readme"); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	commits, err := adapter.Log(context.Background(), repo, 1)
	if err != nil || len(commits) != 1 {
		t.Fatalf("Log: %v (%d)", err, len(commits))
	}
	if commits[0].Subject != "update readme" {
		t.Errorf("subject = %q", commits[0].Subject)
	}

	state = adapter.Status(context.Background(), repo)
	if !state.IsClean() {
		t.Errorf("working tree should be clean, got %+v", state.WorkingTree)
	}
}

func TestStageAllIncludesUntrackedAndDeletions(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	repo := seedRepository(t, dir)

	writeFile(t, filepath.Join(dir, "README.md"), "changed\n")
	writeFile(t, filepath.Join(dir, "new.txt"), "new\n")

	if err := adapter.StageAll(context.Background(), repo); err != nil {
		t.Fatalf("StageAll: %v", err)
	}

	state := adapter.Status(context.Background(), repo)
	if state.StagedCount() != 2 {
		t.Errorf("staged = %d, want 2", state.StagedCount())
	}
	if state.UntrackedCount() != 0 {
		t.Errorf("untracked = %d, want 0", state.UntrackedCount())
	}
}

func TestUnstageAll(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	repo := seedRepository(t, dir)

	writeFile(t, filepath.Join(dir, "README.md"), "changed\n")
	if err := adapter.StageAll(context.Background(), repo); err != nil {
		t.Fatalf("StageAll: %v", err)
	}

	if err := adapter.UnstageAll(context.Background(), repo); err != nil {
		t.Fatalf("UnstageAll: %v", err)
	}

	state := adapter.Status(context.Background(), repo)
	if state.StagedCount() != 0 {
		t.Errorf("staged = %d, want 0", state.StagedCount())
	}
	if state.UnstagedCount() != 1 {
		t.Errorf("unstaged = %d, want 1", state.UnstagedCount())
	}
}

func TestUnstageAllOnUnbornBranch(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q")

	writeFile(t, filepath.Join(dir, "first.txt"), "content\n")
	if err := adapter.StageAll(context.Background(), domain.Repository{ID: "r", Name: "R", Path: dir}); err != nil {
		t.Fatalf("StageAll: %v", err)
	}
	if err := adapter.UnstageAll(context.Background(), domain.Repository{ID: "r", Name: "R", Path: dir}); err != nil {
		t.Fatalf("UnstageAll on unborn branch: %v", err)
	}
}

func TestCreateBranchAndCheckout(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	repo := seedRepository(t, dir)
	ctx := context.Background()

	if err := adapter.CreateBranch(ctx, repo, "feature/retry", true); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
	if state := adapter.Status(ctx, repo); state.Branch.Head != "feature/retry" {
		t.Fatalf("head = %q, want feature/retry", state.Branch.Head)
	}

	if err := adapter.Checkout(ctx, repo, "main"); err != nil {
		t.Fatalf("Checkout: %v", err)
	}
	if state := adapter.Status(ctx, repo); state.Branch.Head != "main" {
		t.Fatalf("head = %q, want main", state.Branch.Head)
	}
}

func TestPushSetsUpstream(t *testing.T) {
	adapter := newAdapter(t)
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	work := filepath.Join(root, "work")

	if err := gitInitBare(t, remote); err != nil {
		t.Fatalf("init bare: %v", err)
	}
	repo := seedRepository(t, work)
	gitRun(t, work, "remote", "add", "origin", remote)

	if err := adapter.Push(context.Background(), repo); err != nil {
		t.Fatalf("Push: %v", err)
	}

	upstream := strings.TrimSpace(gitRun(t, work, "rev-parse", "--abbrev-ref", "@{upstream}"))
	if upstream != "origin/main" {
		t.Fatalf("upstream = %q, want origin/main", upstream)
	}

	head := strings.TrimSpace(gitRun(t, work, "rev-parse", "HEAD"))
	remoteHead := strings.TrimSpace(gitRun(t, remote, "rev-parse", "refs/heads/main"))
	if head != remoteHead {
		t.Errorf("remote head = %q, want %q", remoteHead, head)
	}
}

func TestPullFastForward(t *testing.T) {
	adapter := newAdapter(t)
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	workA := filepath.Join(root, "a")
	workB := filepath.Join(root, "b")

	if err := gitInitBare(t, remote); err != nil {
		t.Fatalf("init bare: %v", err)
	}

	repoA := seedRepository(t, workA)
	gitRun(t, workA, "remote", "add", "origin", remote)
	if err := adapter.Push(context.Background(), repoA); err != nil {
		t.Fatalf("Push: %v", err)
	}

	gitRun(t, root, "clone", "-q", remote, workB)
	gitRun(t, workB, "config", "user.name", "Repora Test")
	gitRun(t, workB, "config", "user.email", "test@example.com")

	writeFile(t, filepath.Join(workA, "second.txt"), "two\n")
	gitRun(t, workA, "add", ".")
	gitRun(t, workA, "commit", "-qm", "second")
	if err := adapter.Push(context.Background(), repoA); err != nil {
		t.Fatalf("second Push: %v", err)
	}

	repoB := domain.Repository{ID: "b", Name: "B", Path: workB}
	if err := adapter.Pull(context.Background(), repoB); err != nil {
		t.Fatalf("Pull: %v", err)
	}

	if _, err := filepath.Glob(filepath.Join(workB, "second.txt")); err != nil {
		t.Fatalf("glob: %v", err)
	}
	state := adapter.Status(context.Background(), repoB)
	if state.Branch.Behind != 0 {
		t.Errorf("behind = %d, want 0", state.Branch.Behind)
	}
}

func TestPullDivergedFails(t *testing.T) {
	adapter := newAdapter(t)
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	workA := filepath.Join(root, "a")
	workB := filepath.Join(root, "b")

	if err := gitInitBare(t, remote); err != nil {
		t.Fatalf("init bare: %v", err)
	}

	repoA := seedRepository(t, workA)
	gitRun(t, workA, "remote", "add", "origin", remote)
	if err := adapter.Push(context.Background(), repoA); err != nil {
		t.Fatalf("Push: %v", err)
	}

	gitRun(t, root, "clone", "-q", remote, workB)
	gitRun(t, workB, "config", "user.name", "Repora Test")
	gitRun(t, workB, "config", "user.email", "test@example.com")
	writeFile(t, filepath.Join(workB, "local.txt"), "local\n")
	gitRun(t, workB, "add", ".")
	gitRun(t, workB, "commit", "-qm", "local only")

	writeFile(t, filepath.Join(workA, "remote.txt"), "remote\n")
	gitRun(t, workA, "add", ".")
	gitRun(t, workA, "commit", "-qm", "remote only")
	gitRun(t, workA, "push", "-q")

	repoB := domain.Repository{ID: "b", Name: "B", Path: workB}
	err := adapter.Pull(context.Background(), repoB)
	if err == nil {
		t.Fatal("expected pull to fail on diverged branch")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "fast-forward") {
		t.Errorf("error does not mention fast-forward: %v", err)
	}
}

func TestFetchUpdatesRemoteTracking(t *testing.T) {
	adapter := newAdapter(t)
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	workA := filepath.Join(root, "a")
	workB := filepath.Join(root, "b")

	if err := gitInitBare(t, remote); err != nil {
		t.Fatalf("init bare: %v", err)
	}

	repoA := seedRepository(t, workA)
	gitRun(t, workA, "remote", "add", "origin", remote)
	if err := adapter.Push(context.Background(), repoA); err != nil {
		t.Fatalf("Push: %v", err)
	}

	gitRun(t, root, "clone", "-q", remote, workB)
	gitRun(t, workB, "config", "user.name", "Repora Test")
	gitRun(t, workB, "config", "user.email", "test@example.com")

	writeFile(t, filepath.Join(workA, "third.txt"), "three\n")
	gitRun(t, workA, "add", ".")
	gitRun(t, workA, "commit", "-qm", "third")
	if err := adapter.Push(context.Background(), repoA); err != nil {
		t.Fatalf("second Push: %v", err)
	}

	repoB := domain.Repository{ID: "b", Name: "B", Path: workB}
	if err := adapter.Fetch(context.Background(), repoB); err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	state := adapter.Status(context.Background(), repoB)
	if state.Branch.Behind != 1 {
		t.Errorf("behind = %d, want 1", state.Branch.Behind)
	}
}

func gitInitBare(t *testing.T, dir string) error {
	t.Helper()
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	gitRun(t, dir, "init", "-q", "--bare")
	return nil
}

func TestUnstageSinglePath(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	repo := seedRepository(t, dir)

	writeFile(t, filepath.Join(dir, "a.txt"), "a\n")
	writeFile(t, filepath.Join(dir, "b.txt"), "b\n")
	if err := adapter.StageAll(context.Background(), repo); err != nil {
		t.Fatalf("StageAll: %v", err)
	}

	if err := adapter.Unstage(context.Background(), repo, []string{"a.txt"}); err != nil {
		t.Fatalf("Unstage: %v", err)
	}

	state := adapter.Status(context.Background(), repo)
	if state.StagedCount() != 1 {
		t.Fatalf("staged = %d, want 1", state.StagedCount())
	}
	if state.WorkingTree.Staged[0].Path != "b.txt" {
		t.Errorf("staged path = %q, want b.txt", state.WorkingTree.Staged[0].Path)
	}
}

func TestDeleteMergedBranch(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	repo := seedRepository(t, dir)
	ctx := context.Background()

	if err := adapter.CreateBranch(ctx, repo, "feature/merged", true); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
	if err := adapter.Checkout(ctx, repo, "main"); err != nil {
		t.Fatalf("Checkout: %v", err)
	}

	if err := adapter.DeleteBranch(ctx, repo, "feature/merged", false); err != nil {
		t.Fatalf("DeleteBranch: %v", err)
	}

	branches, err := adapter.Branches(ctx, repo)
	if err != nil {
		t.Fatalf("Branches: %v", err)
	}
	for _, branch := range branches {
		if branch.Name == "feature/merged" {
			t.Fatal("branch still exists after deletion")
		}
	}
}

func TestDeleteUnmergedBranchRefused(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	repo := seedRepository(t, dir)
	ctx := context.Background()

	if err := adapter.CreateBranch(ctx, repo, "feature/unmerged", true); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
	writeFile(t, filepath.Join(dir, "only-here.txt"), "x\n")
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-qm", "unmerged work")
	if err := adapter.Checkout(ctx, repo, "main"); err != nil {
		t.Fatalf("Checkout: %v", err)
	}

	err := adapter.DeleteBranch(ctx, repo, "feature/unmerged", false)
	if err == nil {
		t.Fatal("expected git to refuse deleting an unmerged branch")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not fully merged") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteCurrentBranchRefused(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	repo := seedRepository(t, dir)

	err := adapter.DeleteBranch(context.Background(), repo, "main", false)
	if !errors.Is(err, domain.ErrCurrentBranch) {
		t.Fatalf("err = %v, want ErrCurrentBranch", err)
	}
}
