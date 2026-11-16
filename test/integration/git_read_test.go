package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gaspar-suarezz/repora/internal/domain"
)

func seedRepository(t *testing.T, dir string) domain.Repository {
	t.Helper()
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	gitRun(t, dir, "init", "-q")
	// The adapter commits without -c identity flags, and CI runners have no
	// global Git identity, so pin it in the repository configuration.
	gitRun(t, dir, "config", "user.name", "Repora Test")
	gitRun(t, dir, "config", "user.email", "test@example.com")
	writeFile(t, filepath.Join(dir, "README.md"), "hello\n")
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-qm", "initial commit")
	writeFile(t, filepath.Join(dir, "app.go"), "package main\n")
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-qm", "add app")
	return domain.Repository{ID: "r", Name: "R", Path: dir}
}

func TestBranchesAgainstRealRepository(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	repo := seedRepository(t, dir)
	gitRun(t, dir, "branch", "feature/one")

	branches, err := adapter.Branches(context.Background(), repo)
	if err != nil {
		t.Fatalf("Branches: %v", err)
	}
	if len(branches) != 2 {
		t.Fatalf("branches = %d, want 2: %+v", len(branches), branches)
	}
	if !branches[0].IsCurrent || branches[0].Name != "main" {
		t.Errorf("current branch is not first: %+v", branches[0])
	}
	if branches[1].Name != "feature/one" || branches[1].IsRemote {
		t.Errorf("unexpected second branch: %+v", branches[1])
	}
}

func TestLogAgainstRealRepository(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	repo := seedRepository(t, dir)

	commits, err := adapter.Log(context.Background(), repo, 10)
	if err != nil {
		t.Fatalf("Log: %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("commits = %d, want 2", len(commits))
	}
	if commits[0].Subject != "add app" || commits[1].Subject != "initial commit" {
		t.Errorf("order/subjects wrong: %q, %q", commits[0].Subject, commits[1].Subject)
	}
	if len(commits[0].Hash) != 40 {
		t.Errorf("hash = %q", commits[0].Hash)
	}
}

func TestLogOnUnbornBranchIsEmpty(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q")

	commits, err := adapter.Log(context.Background(), domain.Repository{ID: "r", Name: "R", Path: dir}, 10)
	if err != nil {
		t.Fatalf("Log: %v", err)
	}
	if len(commits) != 0 {
		t.Fatalf("commits = %d, want 0", len(commits))
	}
}

func TestRemotesAgainstRealRepository(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	repo := seedRepository(t, dir)
	gitRun(t, dir, "remote", "add", "origin", "git@example.com:acme/r.git")

	remotes, err := adapter.Remotes(context.Background(), repo)
	if err != nil {
		t.Fatalf("Remotes: %v", err)
	}
	if len(remotes) != 1 {
		t.Fatalf("remotes = %d, want 1", len(remotes))
	}
	if remotes[0].Name != "origin" || remotes[0].FetchURL != "git@example.com:acme/r.git" {
		t.Errorf("remote = %+v", remotes[0])
	}
}

func TestDiffAgainstRealRepository(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	repo := seedRepository(t, dir)

	if diff, err := adapter.Diff(context.Background(), repo, domain.DiffUnstaged, ""); err != nil {
		t.Fatalf("Diff: %v", err)
	} else if diff != "" {
		t.Errorf("clean repository produced a diff: %q", diff)
	}

	writeFile(t, filepath.Join(dir, "README.md"), "hello\nworld\n")

	diff, err := adapter.Diff(context.Background(), repo, domain.DiffUnstaged, "README.md")
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if !strings.Contains(diff, "+world") {
		t.Errorf("diff missing added line:\n%s", diff)
	}
	if strings.Contains(diff, "\x1b[") {
		t.Error("diff contains ANSI escapes")
	}

	gitRun(t, dir, "add", "README.md")
	staged, err := adapter.Diff(context.Background(), repo, domain.DiffStaged, "README.md")
	if err != nil {
		t.Fatalf("staged Diff: %v", err)
	}
	if !strings.Contains(staged, "+world") {
		t.Errorf("staged diff missing added line:\n%s", staged)
	}
}

func TestShowCommitAgainstRealRepository(t *testing.T) {
	adapter := newAdapter(t)
	dir := t.TempDir()
	repo := seedRepository(t, dir)

	commits, err := adapter.Log(context.Background(), repo, 1)
	if err != nil || len(commits) != 1 {
		t.Fatalf("Log: %v (%d commits)", err, len(commits))
	}

	out, err := adapter.ShowCommit(context.Background(), repo, commits[0].Hash)
	if err != nil {
		t.Fatalf("ShowCommit: %v", err)
	}
	if !strings.Contains(out, "app.go") {
		t.Errorf("commit detail missing file:\n%s", out)
	}
}
