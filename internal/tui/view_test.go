package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/gaspar-suarezz/repora/internal/app"
	"github.com/gaspar-suarezz/repora/internal/domain"
)

func testConfig() *domain.Config {
	return &domain.Config{
		Version: 1,
		Repositories: []domain.Repository{
			{ID: "api", Name: "API", Path: "/tmp/api"},
		},
		Workspaces: []domain.Workspace{
			{ID: "product", Name: "Product X", RepositoryIDs: []domain.RepositoryID{"api"}},
		},
	}
}

func testNew(cfg *domain.Config, path string, reload Reloader) *Model {
	m := New(Deps{Config: cfg, ConfigPath: path, NoColor: true, Reload: reload})
	m.tick = nil // timers are exercised separately
	return m
}

func sizeModel(t *testing.T, width, height int) *Model {
	t.Helper()
	m := testNew(testConfig(), "/tmp/config.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return m
}

func TestHomeRendersWorkspaces(t *testing.T) {
	m := sizeModel(t, 80, 24)
	out := m.View()

	for _, want := range []string{"Repora", "WORKSPACE MANAGER", "YOUR WORKSPACES", "Product X", "1 repo"} {
		if !strings.Contains(out, want) {
			t.Errorf("home view missing %q:\n%s", want, out)
		}
	}
}

func TestHomeFitsTerminalHeight(t *testing.T) {
	m := sizeModel(t, 80, 24)
	if got := strings.Count(m.View(), "\n") + 1; got != 24 {
		t.Fatalf("rendered %d lines, want 24", got)
	}
}

func TestHomeEmptyState(t *testing.T) {
	cfg := &domain.Config{Version: 1}
	m := testNew(cfg, "/tmp/config.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	out := m.View()
	if !strings.Contains(out, "No workspaces yet.") {
		t.Errorf("missing empty state:\n%s", out)
	}
	if !strings.Contains(out, "/tmp/config.yaml") {
		t.Errorf("missing config path:\n%s", out)
	}
}

func TestWorkspaceRendersRepositoryRows(t *testing.T) {
	m := sizeModel(t, 120, 30)
	m.screen = screenWorkspace
	m.rows = []repoRow{
		{
			repo: m.cfg.Repositories[0],
			state: domain.RepositoryState{
				RepositoryID: "api",
				Branch:       domain.BranchState{Head: "main"},
			},
		},
	}

	out := m.View()
	for _, want := range []string{"WORKSPACES", "PRODUCT X", "API", "main", "✓ clean"} {
		if !strings.Contains(out, want) {
			t.Errorf("workspace view missing %q:\n%s", want, out)
		}
	}
}

func TestWorkspaceStatesRenderDistinctly(t *testing.T) {
	m := sizeModel(t, 120, 30)
	m.screen = screenWorkspace

	cases := []struct {
		name string
		row  repoRow
		want string
	}{
		{"unavailable", repoRow{repo: m.cfg.Repositories[0], unavailable: true}, "! unavailable"},
		{"loading", repoRow{repo: m.cfg.Repositories[0], loading: true}, "◐ refreshing"},
		{"dirty", repoRow{repo: m.cfg.Repositories[0], state: domain.RepositoryState{
			WorkingTree: domain.WorkingTree{Untracked: []string{"a.txt"}},
			Branch:      domain.BranchState{Head: "main"},
		}}, "● modified"},
		{"behind", repoRow{repo: m.cfg.Repositories[0], state: domain.RepositoryState{
			Branch: domain.BranchState{Head: "main", Behind: 2},
		}}, "↓ behind 2"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m.rows = []repoRow{tc.row}
			if out := m.View(); !strings.Contains(out, tc.want) {
				t.Errorf("missing %q:\n%s", tc.want, out)
			}
		})
	}
}

func TestHelpToggles(t *testing.T) {
	m := sizeModel(t, 100, 30)

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !m.showHelp {
		t.Fatal("help did not open")
	}
	if out := m.View(); !strings.Contains(out, "Global") {
		t.Errorf("help view missing content:\n%s", out)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if m.showHelp {
		t.Fatal("help did not close")
	}
}

func TestNavigationChangesSelection(t *testing.T) {
	cfg := testConfig()
	cfg.Workspaces = append(cfg.Workspaces, domain.Workspace{ID: "platform", Name: "Platform"})
	m := testNew(cfg, "/tmp/config.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.wsIndex != 1 {
		t.Fatalf("wsIndex = %d, want 1", m.wsIndex)
	}
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.wsIndex != 0 {
		t.Fatalf("wsIndex = %d, want 0", m.wsIndex)
	}
}

func TestAllScreensShowBrandHeader(t *testing.T) {
	m := sizeModel(t, 100, 30)

	check := func(name string) {
		t.Helper()
		out := m.View()
		for _, want := range []string{"◈  Repora", "WORKSPACE MANAGER"} {
			if !strings.Contains(out, want) {
				t.Errorf("%s is missing %q:\n%s", name, want, out)
			}
		}
	}

	check("home")

	m.screen = screenWorkspace
	m.rows = []repoRow{{repo: m.cfg.Repositories[0], state: domain.RepositoryState{RepositoryID: "api"}}}
	check("workspace")

	m.screen = screenRepository
	m.rowIndex = 0
	m.detail = app.RepositoryDetail{}
	check("repository")
}

func TestBreadcrumbHighlightsActiveSegment(t *testing.T) {
	m := sizeModel(t, 100, 30)
	m.screen = screenWorkspace
	m.detail = app.RepositoryDetail{}

	out := m.View()
	if !strings.Contains(out, "WORKSPACES  /") {
		t.Errorf("breadcrumb missing ancestors:\n%s", out)
	}
	if !strings.Contains(out, "PRODUCT X") {
		t.Errorf("breadcrumb missing active segment:\n%s", out)
	}
}

func TestWorkspaceListKeepsSelectionVisible(t *testing.T) {
	cfg := &domain.Config{Version: 1}
	ws := domain.Workspace{ID: "many", Name: "Many"}
	for i := 0; i < 30; i++ {
		name := fmt.Sprintf("repo-%02d", i)
		id := domain.RepositoryID(name)
		cfg.Repositories = append(cfg.Repositories, domain.Repository{ID: id, Name: name, Path: "/tmp/" + name})
		ws.RepositoryIDs = append(ws.RepositoryIDs, id)
	}
	cfg.Workspaces = []domain.Workspace{ws}

	m := testNew(cfg, "/tmp/config.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m.screen = screenWorkspace
	m.rows = make([]repoRow, 0, 30)
	for _, repo := range cfg.Repositories {
		m.rows = append(m.rows, repoRow{repo: repo, state: domain.RepositoryState{RepositoryID: repo.ID}})
	}

	// Scrolled to the bottom: the last row is visible, no downward hint.
	m.rowIndex = 29
	out := m.View()
	if !strings.Contains(out, "repo-29") {
		t.Errorf("selected row is not visible:\n%s", out)
	}

	// At the top with 30 rows, the overflow hint must appear.
	m.rowIndex = 0
	if out := m.View(); !strings.Contains(out, "↓") {
		t.Errorf("missing overflow hint:\n%s", out)
	}
}

func TestPromptFitsTerminalWidth(t *testing.T) {
	for _, width := range []int{80, 100} {
		m := sizeModel(t, width, 24)
		m.prompt = newPrompt()
		m.prompt.kind = promptWorkspaceDescription
		m.prompt.pendingName = "A very long workspace name that should be truncated"
		m.prompt.input.SetValue("A long description that should stay inside the panel")
		m.prompt.input.CursorEnd()

		for _, line := range strings.Split(m.View(), "\n") {
			if got := lipgloss.Width(line); got > width {
				t.Fatalf("width %d: prompt line overflows (%d): %q", width, got, line)
			}
		}
		if !strings.Contains(m.View(), "Repora") {
			t.Errorf("prompt is missing the brand header")
		}
	}
}

func TestContentHasMargins(t *testing.T) {
	m := sizeModel(t, 80, 24)
	lines := strings.Split(m.View(), "\n")

	if strings.TrimSpace(lines[0]) != "" {
		t.Errorf("missing top margin: %q", lines[0])
	}
	for i, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, "  ") {
			t.Errorf("line %d lacks the left margin: %q", i+1, line)
		}
		if lipgloss.Width(line) > 80 {
			t.Errorf("line %d overflows: %q", i+1, line)
		}
	}
}
