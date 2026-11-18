package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/gasparsuarez/repora/internal/domain"
)

func aggregateModel(t *testing.T) *Model {
	t.Helper()

	cfg := &domain.Config{
		Version: 1,
		Repositories: []domain.Repository{
			{ID: "api", Name: "API", Path: "/tmp/api"},
			{ID: "web", Name: "Web", Path: "/tmp/web"},
			{ID: "infra", Name: "Infrastructure", Path: "/tmp/infra"},
			{ID: "shared", Name: "Shared", Path: "/tmp/shared"},
		},
		Workspaces: []domain.Workspace{
			{ID: "product", Name: "Product", Description: "Product workspace",
				RepositoryIDs: []domain.RepositoryID{"api", "web", "infra", "shared"}},
		},
	}

	m := testNew(cfg, "/tmp/config.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m.homeLoaded = true
	m.homeStates = map[domain.RepositoryID]domain.RepositoryState{
		"api": {RepositoryID: "api", WorkingTree: domain.WorkingTree{Untracked: []string{"a.md"}}},
		"web": {RepositoryID: "web"},
		"infra": {RepositoryID: "infra", Branch: domain.BranchState{
			Head: "main", Upstream: "origin/main", HasUpstream: true, Behind: 3,
		}},
		"shared": {RepositoryID: "shared", WorkingTree: domain.WorkingTree{
			Conflicts: []domain.FileChange{{Path: "go.mod", Index: domain.CodeUnmerged}},
		}},
	}
	return m
}

func TestHomeAggregatesRealStates(t *testing.T) {
	m := aggregateModel(t)
	out := m.View()

	for _, want := range []string{"1 conflict", "1 modified", "1 sync", "1 clean"} {
		if !strings.Contains(out, want) {
			t.Errorf("home missing aggregate %q:\n%s", want, out)
		}
	}
}

func TestHomeEmptyWorkspaceCard(t *testing.T) {
	cfg := &domain.Config{
		Version: 1,
		Workspaces: []domain.Workspace{
			{ID: "empty", Name: "Empty", Description: "Nothing assigned yet"},
		},
	}
	m := testNew(cfg, "/tmp/config.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m.homeLoaded = true

	out := m.View()
	for _, want := range []string{"Empty", "0 repos", "Nothing assigned yet", "No repositories yet", "open, then a to add"} {
		if !strings.Contains(out, want) {
			t.Errorf("empty card missing %q:\n%s", want, out)
		}
	}
}

func TestHomeShowsLoadingBeforeStatesArrive(t *testing.T) {
	cfg := &domain.Config{
		Version:      1,
		Repositories: []domain.Repository{{ID: "api", Name: "API", Path: "/tmp/api"}},
		Workspaces: []domain.Workspace{
			{ID: "alpha", Name: "Alpha", RepositoryIDs: []domain.RepositoryID{"api"}},
		},
	}
	m := testNew(cfg, "/tmp/config.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m.homeLoaded = true
	m.homeLoading = true

	if out := m.View(); !strings.Contains(out, "loading status") {
		t.Errorf("expected loading state:\n%s", out)
	}
}

func TestHomeWindowKeepsSelectionVisible(t *testing.T) {
	cfg := &domain.Config{Version: 1}
	for i := 0; i < 10; i++ {
		cfg.Workspaces = append(cfg.Workspaces, domain.Workspace{
			ID:   domain.WorkspaceID(string(rune('a' + i))),
			Name: "Workspace " + string(rune('A'+i)),
		})
	}

	m := testNew(cfg, "/tmp/config.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m.homeLoaded = true

	if _, end := m.homeWindow(len(cfg.Workspaces), 18); end == 0 {
		t.Fatal("window is empty")
	}
	if m.homeOffset != 0 {
		t.Fatalf("initial offset = %d, want 0", m.homeOffset)
	}

	m.wsIndex = 9
	_, end := m.homeWindow(len(cfg.Workspaces), 18)
	if m.wsIndex < m.homeOffset || m.wsIndex >= end {
		t.Fatalf("selection %d outside window [%d,%d)", m.wsIndex, m.homeOffset, end)
	}
	if !strings.Contains(m.View(), "Workspace J") {
		t.Errorf("selected workspace not rendered:\n%s", m.View())
	}
}

func TestHomeFitsTerminalWidth(t *testing.T) {
	for _, width := range []int{80, 100, 120} {
		m := aggregateModel(t)
		m.Update(tea.WindowSizeMsg{Width: width, Height: 30})

		for _, line := range strings.Split(m.View(), "\n") {
			if got := lipgloss.Width(line); got > width {
				t.Fatalf("width %d: line overflows (%d): %q", width, got, line)
			}
		}
	}
}

func TestHomeScrollsWithManyWorkspaces(t *testing.T) {
	cfg := &domain.Config{Version: 1}
	for i := 0; i < 8; i++ {
		cfg.Workspaces = append(cfg.Workspaces, domain.Workspace{
			ID:   domain.WorkspaceID(string(rune('a' + i))),
			Name: "Workspace " + string(rune('A'+i)),
		})
	}

	m := testNew(cfg, "/tmp/config.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m.homeLoaded = true

	for i := 0; i < 7; i++ {
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	}

	if m.wsIndex != 7 {
		t.Fatalf("wsIndex = %d, want 7", m.wsIndex)
	}
	if m.homeOffset == 0 {
		t.Fatal("offset did not advance while scrolling")
	}
	if out := m.View(); !strings.Contains(out, "Workspace H") {
		t.Errorf("last workspace not visible:\n%s", out)
	}
}
