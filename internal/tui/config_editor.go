package tui

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gasparsuarez/repora/internal/app"
	"github.com/gasparsuarez/repora/internal/domain"
)

// lookPath is replaceable in tests.
var lookPath = exec.LookPath

// editorCandidates are tried in order when neither $VISUAL nor $EDITOR is set.
var editorCandidates = []string{"nano", "vim", "vi"}

// resolveEditor returns the editor command to use. It honours $VISUAL, then
// $EDITOR, and finally falls back to the first common editor found on PATH.
func resolveEditor() (string, bool) {
	for _, name := range []string{"VISUAL", "EDITOR"} {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value, true
		}
	}
	for _, candidate := range editorCandidates {
		if _, err := lookPath(candidate); err == nil {
			return candidate, true
		}
	}
	return "", false
}

// openConfigEditor suspends the TUI and opens the configuration file in the
// user's editor. Repora reloads the configuration when the editor exits.
func (m *Model) openConfigEditor() tea.Cmd {
	editor, ok := resolveEditor()
	if !ok {
		m.notice = "set $EDITOR to edit the config, or open " + m.configPath
		m.noticeError = false
		return nil
	}

	fields := strings.Fields(editor)
	if len(fields) == 0 {
		m.notice = "empty $EDITOR"
		m.noticeError = true
		return nil
	}

	args := append(append([]string{}, fields[1:]...), m.configPath)
	cmd := exec.Command(fields[0], args...)
	cmd.Env = os.Environ()

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorClosedMsg{err: err}
	})
}

// reloadCmd reads the configuration from disk again and rebuilds derived
// services. It never writes to disk.
func (m *Model) reloadCmd() tea.Cmd {
	reload := m.reload
	return func() tea.Msg {
		if reload == nil {
			return configReloadedMsg{err: errors.New("reload is not available")}
		}
		cfg, warnings, status, err := reload()
		return configReloadedMsg{cfg: cfg, warnings: warnings, status: status, err: err}
	}
}

// applyReload replaces the in-memory configuration, preserving the selected
// workspace by ID when possible.
func (m *Model) applyReload(cfg *domain.Config, warnings []string, status *app.StatusService) tea.Cmd {
	if cfg == nil {
		return nil
	}

	previousID := m.selectedWorkspaceID()

	m.cfg = cfg
	m.warnings = warnings
	if status != nil {
		m.status = status
	}

	if idx := indexOfWorkspace(cfg, previousID); idx >= 0 {
		m.wsIndex = idx
	} else if m.wsIndex >= len(cfg.Workspaces) {
		m.wsIndex = max(0, len(cfg.Workspaces)-1)
	}

	if m.screen == screenWorkspace {
		if previousID != "" && indexOfWorkspace(cfg, previousID) >= 0 {
			m.notice = "configuration reloaded"
			m.noticeError = false
			return m.rebuildRowsAndRefresh()
		}
		m.closeWorkspace()
		m.notice = "workspace no longer exists — back to home"
		m.noticeError = false
		return nil
	}

	m.notice = "configuration reloaded"
	m.noticeError = false
	return nil
}

func (m *Model) selectedWorkspaceID() domain.WorkspaceID {
	if m.wsIndex < 0 || m.wsIndex >= len(m.cfg.Workspaces) {
		return ""
	}
	return m.cfg.Workspaces[m.wsIndex].ID
}

func indexOfWorkspace(cfg *domain.Config, id domain.WorkspaceID) int {
	if id == "" {
		return -1
	}
	for i, ws := range cfg.Workspaces {
		if ws.ID == id {
			return i
		}
	}
	return -1
}
