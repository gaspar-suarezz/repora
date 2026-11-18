package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gasparsuarez/repora/internal/domain"
	"github.com/gasparsuarez/repora/internal/tui/components"
)

type confirmAction int

const (
	confirmNone confirmAction = iota
	confirmDeleteWorkspace
	confirmRemoveRepository
	confirmDeleteBranch
)

type confirmState struct {
	active  bool
	title   string
	message string
	action  confirmAction

	wsID   domain.WorkspaceID
	repoID domain.RepositoryID
	branch string
}

func (m *Model) openDeleteWorkspaceConfirm() {
	if m.wsIndex < 0 || m.wsIndex >= len(m.cfg.Workspaces) {
		return
	}
	ws := m.cfg.Workspaces[m.wsIndex]

	m.confirm = confirmState{
		active:  true,
		title:   `Delete workspace "` + ws.Name + `"?`,
		message: "Removes the workspace from Repora.\nRepositories are not deleted from disk.",
		action:  confirmDeleteWorkspace,
		wsID:    ws.ID,
	}
}

func (m *Model) openRemoveRepositoryConfirm() {
	row := m.currentRow()
	if row == nil {
		return
	}
	wsID := m.selectedWorkspaceID()
	m.confirm = confirmState{
		active:  true,
		title:   `Remove "` + row.repo.Name + `" from this workspace?`,
		message: "The repository stays registered and on disk.\nOther workspaces are unaffected.",
		action:  confirmRemoveRepository,
		wsID:    wsID,
		repoID:  row.repo.ID,
	}
}

func (m *Model) openDeleteBranchConfirm() {
	row := m.currentRow()
	if row == nil || m.branchIndex < 0 || m.branchIndex >= len(m.detail.Branches) {
		return
	}

	branch := m.detail.Branches[m.branchIndex]
	switch {
	case branch.IsRemote:
		m.notice = "remote branches are read-only"
		m.noticeError = false
		return
	case branch.IsCurrent:
		m.notice = "cannot delete the checked-out branch"
		m.noticeError = true
		return
	}

	m.confirm = confirmState{
		active:  true,
		title:   `Delete branch "` + branch.Name + `"?`,
		message: "Git refuses to delete a branch that is not fully merged.\nThis cannot be undone.",
		action:  confirmDeleteBranch,
		repoID:  row.repo.ID,
		branch:  branch.Name,
	}
}

func (m *Model) handleConfirmKey(key string) tea.Cmd {
	switch key {
	case "y":
		state := m.confirm
		m.confirm = confirmState{}
		switch state.action {
		case confirmDeleteWorkspace:
			return m.deleteWorkspaceCmd(state.wsID)
		case confirmRemoveRepository:
			return m.removeRepositoryCmd(state.wsID, state.repoID)
		case confirmDeleteBranch:
			return m.deleteBranchCmd(state.repoID, state.branch)
		}
		return nil
	default:
		// Enter and Esc cancel: the destructive choice is never the default.
		m.confirm = confirmState{}
		return nil
	}
}

func (m *Model) renderConfirm() string {
	var b strings.Builder
	b.WriteString(m.brandHeader())
	b.WriteString("\n" + components.Rule(m.st, m.contentWidth()) + "\n")
	b.WriteString(components.Breadcrumb(m.st, []string{"Workspaces", "Confirm"}, m.contentWidth(), ""))
	b.WriteString("\n\n")

	body := []string{
		m.st.Header.Bold(true).Render(truncate(m.confirm.title, max(10, m.contentWidth()-8))),
		"",
	}
	for _, line := range strings.Split(m.confirm.message, "\n") {
		body = append(body, m.st.Bar.Render(truncate(line, max(10, m.contentWidth()-8))))
	}

	b.WriteString(m.renderPanelWith(strings.Join(body, "\n"), len(body)+2, m.st.CardSelected))

	return m.frame(b.String(), [][2]string{
		{"y", "Confirm"},
		{"Esc", "Cancel"},
	})
}
