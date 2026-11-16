package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gaspar-suarezz/repora/internal/app"
	"github.com/gaspar-suarezz/repora/internal/domain"
)

func repositoryModel(t *testing.T) *Model {
	t.Helper()

	cfg := testConfig()
	cfg.Repositories[0].Path = "/tmp/api"

	m := testNew(cfg, "/tmp/config.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 32})

	m.screen = screenWorkspace
	m.rows = []repoRow{{
		repo: cfg.Repositories[0],
		state: domain.RepositoryState{
			RepositoryID: "api",
			Branch:       domain.BranchState{Head: "main", Upstream: "origin/main", HasUpstream: true, Ahead: 1},
			WorkingTree: domain.WorkingTree{
				Staged:    []domain.FileChange{{Path: "handler.go", Index: domain.CodeModified}},
				Unstaged:  []domain.FileChange{{Path: "README.md", Worktree: domain.CodeModified}},
				Untracked: []string{"notes/todo.md"},
			},
		},
	}}

	m.screen = screenRepository
	m.changes = buildChanges(m.rows[0].state)
	m.detail = app.RepositoryDetail{
		Branches: []domain.Branch{
			{Name: "main", IsCurrent: true, HasUpstream: true, Upstream: "origin/main", Ahead: 1},
			{Name: "feature/x"},
			{Name: "origin/main", IsRemote: true},
		},
		Commits: []domain.Commit{
			{Hash: "abcdef1234567890", Short: "abcdef1", Subject: "Add retry logic", Author: "Ana", Date: time.Now().Add(-2 * time.Hour)},
		},
		Remotes: []domain.Remote{{Name: "origin", FetchURL: "git@example.com:acme/api.git"}},
	}

	return m
}

func TestRepositoryOverviewRenders(t *testing.T) {
	m := repositoryModel(t)
	out := m.View()

	for _, want := range []string{"WORKSPACES", "API", "Overview", "Changes (3)", "Branches", "History", "main → origin/main", "origin"} {
		if !strings.Contains(out, want) {
			t.Errorf("repository view missing %q:\n%s", want, out)
		}
	}
}

func TestRepositoryTabSwitching(t *testing.T) {
	m := repositoryModel(t)

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if m.repoTab != tabChanges {
		t.Fatalf("tab = %v, want changes", m.repoTab)
	}
	out := m.View()
	for _, want := range []string{"STAGED (1)", "UNSTAGED (1)", "UNTRACKED (1)", "handler.go"} {
		if !strings.Contains(out, want) {
			t.Errorf("changes tab missing %q:\n%s", want, out)
		}
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	if out := m.View(); !strings.Contains(out, "LOCAL") || !strings.Contains(out, "REMOTE") {
		t.Errorf("branches tab missing sections:\n%s", out)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	if out := m.View(); !strings.Contains(out, "Add retry logic") {
		t.Errorf("history tab missing commit:\n%s", out)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if m.repoTab != tabOverview {
		t.Errorf("tab = %v, want overview", m.repoTab)
	}
}

func TestRepositoryChangesNavigation(t *testing.T) {
	m := repositoryModel(t)
	m.repoTab = tabChanges

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.changeIndex != 1 {
		t.Fatalf("changeIndex = %d, want 1", m.changeIndex)
	}
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.changeIndex != 2 {
		t.Fatalf("changeIndex should clamp at 2, got %d", m.changeIndex)
	}
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.changeIndex != 1 {
		t.Fatalf("changeIndex = %d, want 1", m.changeIndex)
	}
}

func TestRepositoryEnterOpensDetail(t *testing.T) {
	m := repositoryModel(t)

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.commitOpen {
		t.Fatal("commit overlay did not open")
	}
	if cmd == nil {
		t.Fatal("expected the commit patch to load")
	}
	_, _ = m.Update(cmd())
	if out := m.View(); !strings.Contains(out, "abcdef1234567890") {
		t.Errorf("commit overlay missing hash:\n%s", out)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.commitOpen {
		t.Fatal("commit overlay did not close")
	}
}

func TestRepositoryEscReturnsToWorkspace(t *testing.T) {
	m := repositoryModel(t)

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.screen != screenWorkspace {
		t.Fatalf("screen = %v, want workspace", m.screen)
	}
	if m.rowIndex != 0 {
		t.Errorf("row selection lost: %d", m.rowIndex)
	}
}

func TestWorkspaceEnterOpensRepository(t *testing.T) {
	m := repositoryModel(t)
	m.screen = screenWorkspace

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.screen != screenRepository {
		t.Fatalf("screen = %v, want repository", m.screen)
	}
	if cmd == nil {
		t.Fatal("expected a detail loading command")
	}
	if m.repoTab != tabOverview {
		t.Errorf("tab = %v, want overview", m.repoTab)
	}
}

func TestDiffOverlayRendersAndScrolls(t *testing.T) {
	m := repositoryModel(t)
	m.diffOpen = true
	m.diffTitle = "README.md (unstaged)"
	m.diffContent = "@@ -1 +1 @@\n-hello\n+hello world\n"

	out := m.View()
	if !strings.Contains(out, "Diff · README.md") || !strings.Contains(out, "+hello world") {
		t.Errorf("diff overlay missing content:\n%s", out)
	}

	m.diffScroll = 99
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.diffScroll < 0 {
		t.Errorf("scroll went negative: %d", m.diffScroll)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.diffOpen {
		t.Error("diff overlay did not close")
	}
}

func TestUntrackedDiffExplainsItself(t *testing.T) {
	m := repositoryModel(t)
	m.repoTab = tabChanges
	// Index 2 is the untracked entry.
	m.changeIndex = 2

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("untracked diff should not run a command")
	}
	if !m.diffOpen {
		t.Fatal("diff overlay did not open")
	}
	if !strings.Contains(m.diffContent, "untracked") {
		t.Errorf("content = %q", m.diffContent)
	}
}
