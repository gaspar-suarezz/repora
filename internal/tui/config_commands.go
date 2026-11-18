package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gasparsuarez/repora/internal/domain"
)

// createWorkspaceCmd creates a workspace and persists the configuration.
func (m *Model) createWorkspaceCmd(name, description string) tea.Cmd {
	if !m.configReady() {
		return nil
	}
	base := m.cfg.Clone()
	svc := m.workspaces

	return func() tea.Msg {
		id, err := svc.Create(base, name, description)
		if err != nil {
			return configChangedMsg{action: "create workspace", err: err}
		}
		return configChangedMsg{action: "create workspace", cfg: base, selectWS: id}
	}
}

// renameWorkspaceCmd changes a workspace display name.
func (m *Model) renameWorkspaceCmd(id domain.WorkspaceID, name, description string) tea.Cmd {
	if !m.configReady() {
		return nil
	}
	base := m.cfg.Clone()
	svc := m.workspaces

	return func() tea.Msg {
		if err := svc.Rename(base, id, name, description); err != nil {
			return configChangedMsg{action: "rename workspace", err: err}
		}
		return configChangedMsg{action: "rename workspace", cfg: base, selectWS: id}
	}
}

// deleteWorkspaceCmd removes a workspace without touching repositories.
func (m *Model) deleteWorkspaceCmd(id domain.WorkspaceID) tea.Cmd {
	if !m.configReady() {
		return nil
	}
	base := m.cfg.Clone()
	svc := m.workspaces

	return func() tea.Msg {
		if err := svc.Delete(base, id); err != nil {
			return configChangedMsg{action: "delete workspace", err: err}
		}
		return configChangedMsg{action: "delete workspace", cfg: base}
	}
}

// addRepositoryCmd validates a path, registers it, and associates it with the
// selected workspace.
func (m *Model) addRepositoryCmd(path string) tea.Cmd {
	if !m.configReady() {
		return nil
	}
	if m.repos == nil {
		m.notice = "repository registration is unavailable"
		m.noticeError = true
		return nil
	}

	base := m.cfg.Clone()
	wsID := m.selectedWorkspaceID()
	repos := m.repos
	workspaces := m.workspaces

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		repoID, err := repos.Register(ctx, base, path)
		if err != nil {
			return configChangedMsg{action: "add repository", err: err}
		}
		if err := workspaces.Associate(base, wsID, repoID); err != nil {
			// The registration was persisted; report the partial result.
			return configChangedMsg{action: "add repository", cfg: base, err: err}
		}
		return configChangedMsg{action: "add repository", cfg: base}
	}
}

// removeRepositoryCmd disassociates a repository from the selected workspace.
func (m *Model) removeRepositoryCmd(wsID domain.WorkspaceID, repoID domain.RepositoryID) tea.Cmd {
	if !m.configReady() {
		return nil
	}
	base := m.cfg.Clone()
	svc := m.workspaces

	return func() tea.Msg {
		if err := svc.Disassociate(base, wsID, repoID); err != nil {
			return configChangedMsg{action: "remove repository", err: err}
		}
		return configChangedMsg{action: "remove repository", cfg: base}
	}
}

// applyConfig replaces the in-memory configuration and fixes the selection.
func (m *Model) applyConfig(cfg *domain.Config, selectWS domain.WorkspaceID) {
	if cfg == nil {
		return
	}

	previousID := m.selectedWorkspaceID()
	m.cfg = cfg

	switch {
	case selectWS != "" && indexOfWorkspace(cfg, selectWS) >= 0:
		m.wsIndex = indexOfWorkspace(cfg, selectWS)
	case indexOfWorkspace(cfg, previousID) >= 0:
		m.wsIndex = indexOfWorkspace(cfg, previousID)
	case m.wsIndex >= len(cfg.Workspaces):
		m.wsIndex = max(0, len(cfg.Workspaces)-1)
	}
}

func (m *Model) configReady() bool {
	if m.workspaces == nil {
		m.notice = "workspace editing is unavailable"
		m.noticeError = true
		return false
	}
	return true
}
