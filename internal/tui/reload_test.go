package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gasparsuarez/repora/internal/app"
	"github.com/gasparsuarez/repora/internal/domain"
)

func configWith(ids ...string) *domain.Config {
	cfg := &domain.Config{Version: 1}
	for _, id := range ids {
		cfg.Workspaces = append(cfg.Workspaces, domain.Workspace{
			ID:   domain.WorkspaceID(id),
			Name: strings.ToUpper(id),
		})
	}
	return cfg
}

func TestReloadPreservesSelectedWorkspaceByID(t *testing.T) {
	m := testNew(configWith("alpha", "beta"), "/tmp/c.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.wsIndex = 1 // beta

	_, _ = m.Update(configReloadedMsg{cfg: configWith("beta", "alpha", "gamma")})

	if got := m.cfg.Workspaces[m.wsIndex].ID; got != "beta" {
		t.Fatalf("selection = %q, want beta", got)
	}
	if m.notice != "configuration reloaded" {
		t.Errorf("notice = %q", m.notice)
	}
}

func TestReloadClampsSelectionWhenWorkspaceDisappears(t *testing.T) {
	m := testNew(configWith("alpha", "beta", "gamma"), "/tmp/c.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.wsIndex = 2

	_, _ = m.Update(configReloadedMsg{cfg: configWith("alpha")})

	if m.wsIndex != 0 {
		t.Fatalf("wsIndex = %d, want 0", m.wsIndex)
	}
}

func TestReloadFromWorkspaceFallsBackToHomeWhenRemoved(t *testing.T) {
	m := testNew(configWith("alpha"), "/tmp/c.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.screen = screenWorkspace
	m.rows = []repoRow{{repo: domain.Repository{ID: "r", Name: "R", Path: "/tmp/r"}}}

	_, _ = m.Update(configReloadedMsg{cfg: configWith("other")})

	if m.screen != screenHome {
		t.Fatalf("screen = %v, want home", m.screen)
	}
	if !strings.Contains(m.notice, "no longer exists") {
		t.Errorf("notice = %q", m.notice)
	}
}

func TestReloadErrorKeepsPreviousState(t *testing.T) {
	cfg := configWith("alpha")
	m := testNew(cfg, "/tmp/c.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	_, _ = m.Update(configReloadedMsg{err: errors.New("yaml: line 2")})

	if m.cfg != cfg {
		t.Fatal("config was replaced after a failed reload")
	}
	if !m.noticeError || !strings.Contains(m.notice, "reload failed") {
		t.Errorf("notice = %q (error=%v)", m.notice, m.noticeError)
	}
}

func TestReloadKeyInvokesReloader(t *testing.T) {
	reloaded := configWith("gamma")
	called := false

	m := testNew(configWith("alpha"), "/tmp/c.yaml",
		func() (*domain.Config, []string, *app.StatusService, error) {
			called = true
			return reloaded, []string{"warn"}, nil, nil
		})
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	if cmd == nil {
		t.Fatal("Ctrl+R produced no command")
	}

	msg := cmd()
	if !called {
		t.Fatal("reloader was not called")
	}
	_, _ = m.Update(msg)

	if got := m.cfg.Workspaces[0].ID; got != "gamma" {
		t.Fatalf("workspace = %q, want gamma", got)
	}
	if len(m.warnings) != 1 {
		t.Errorf("warnings = %v", m.warnings)
	}
}

func TestConfigKeyWithoutEditorShowsNotice(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	previous := lookPath
	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	defer func() { lookPath = previous }()

	m := testNew(configWith("alpha"), "/tmp/c.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{','}})
	if cmd != nil {
		t.Fatal("expected no command without an editor")
	}
	if !strings.Contains(m.notice, "EDITOR") {
		t.Errorf("notice = %q", m.notice)
	}
	if !strings.Contains(m.notice, "/tmp/c.yaml") {
		t.Errorf("notice should name the config path: %q", m.notice)
	}
}

func TestConfigKeyFallsBackToInstalledEditor(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	previous := lookPath
	lookPath = func(name string) (string, error) {
		if name == "vim" {
			return "/usr/bin/vim", nil
		}
		return "", errors.New("not found")
	}
	defer func() { lookPath = previous }()

	m := testNew(configWith("alpha"), "/tmp/c.yaml", nil)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{','}})
	if cmd == nil {
		t.Fatal("expected the fallback editor to be launched")
	}
	if m.notice != "" {
		t.Errorf("notice = %q, want none", m.notice)
	}
}

func TestResolveEditorPrecedence(t *testing.T) {
	t.Setenv("VISUAL", "visual-editor --wait")
	t.Setenv("EDITOR", "editor")

	editor, ok := resolveEditor()
	if !ok || editor != "visual-editor --wait" {
		t.Fatalf("editor = %q (ok=%v), want $VISUAL", editor, ok)
	}

	t.Setenv("VISUAL", "")
	editor, ok = resolveEditor()
	if !ok || editor != "editor" {
		t.Fatalf("editor = %q (ok=%v), want $EDITOR", editor, ok)
	}
}

func TestReloadWithoutReloaderReportsError(t *testing.T) {
	m := testNew(configWith("alpha"), "/tmp/c.yaml", nil)
	msg := m.reloadCmd()()
	reloaded, ok := msg.(configReloadedMsg)
	if !ok {
		t.Fatalf("msg = %T", msg)
	}
	if reloaded.err == nil {
		t.Fatal("expected an error when no reloader is configured")
	}
}
