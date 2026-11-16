package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gaspar-suarezz/repora/internal/app"
	"github.com/gaspar-suarezz/repora/internal/domain"
)

type repoTab int

const (
	tabOverview repoTab = iota
	tabChanges
	tabBranches
	tabHistory
)

var repoTabNames = []string{"Overview", "Changes", "Branches", "History"}

// changeItem is one selectable row in the Changes tab.
type changeItem struct {
	Path    string
	Code    domain.StatusCode
	Section string
	Target  domain.DiffTarget
}

func buildChanges(state domain.RepositoryState) []changeItem {
	var items []changeItem

	for _, change := range state.WorkingTree.Conflicts {
		items = append(items, changeItem{
			Path:    change.Path,
			Code:    change.Index,
			Section: "conflicts",
			Target:  domain.DiffUnstaged,
		})
	}
	for _, change := range state.WorkingTree.Staged {
		items = append(items, changeItem{
			Path:    change.Path,
			Code:    change.Index,
			Section: "staged",
			Target:  domain.DiffStaged,
		})
	}
	for _, change := range state.WorkingTree.Unstaged {
		items = append(items, changeItem{
			Path:    change.Path,
			Code:    change.Worktree,
			Section: "unstaged",
			Target:  domain.DiffUnstaged,
		})
	}
	for _, path := range state.WorkingTree.Untracked {
		items = append(items, changeItem{
			Path:    path,
			Code:    "?",
			Section: "untracked",
			Target:  domain.DiffUnstaged,
		})
	}

	return items
}

func (m *Model) currentRow() *repoRow {
	if m.rowIndex < 0 || m.rowIndex >= len(m.rows) {
		return nil
	}
	return &m.rows[m.rowIndex]
}

func (m *Model) currentRepo() (domain.Repository, bool) {
	row := m.currentRow()
	if row == nil {
		return domain.Repository{}, false
	}
	return row.repo, true
}

// openRepository enters the repository screen for the selected row.
func (m *Model) openRepository() tea.Cmd {
	row := m.currentRow()
	if row == nil {
		return nil
	}
	if row.unavailable {
		m.notice = "repository is unavailable: " + row.repo.Path
		m.noticeError = true
		return nil
	}

	m.screen = screenRepository
	m.repoTab = tabOverview
	m.detail = app.RepositoryDetail{}
	m.changes = buildChanges(row.state)
	m.changeIndex = 0
	m.branchIndex = 0
	m.historyIndex = 0
	m.diffOpen = false
	m.diffLoading = false
	m.commitOpen = false
	m.notice = ""
	m.noticeError = false

	return m.loadDetail()
}

func (m *Model) closeRepository() {
	m.screen = screenWorkspace
	m.diffOpen = false
	m.commitOpen = false
	m.detail = app.RepositoryDetail{}
	m.detailLoading = false
	m.detailGen++
	m.notice = ""
	m.noticeError = false
}

func (m *Model) loadDetail() tea.Cmd {
	repo, ok := m.currentRepo()
	if !ok {
		return nil
	}

	m.detailGen++
	gen := m.detailGen
	m.detailLoading = true
	svc := m.repos

	return func() tea.Msg {
		if svc == nil {
			return detailMsg{gen: gen, repoID: repo.ID}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return detailMsg{gen: gen, repoID: repo.ID, detail: svc.Load(ctx, repo)}
	}
}

func (m *Model) loadDiff(item changeItem) tea.Cmd {
	repo, ok := m.currentRepo()
	if !ok {
		return nil
	}
	if item.Section == "untracked" {
		m.diffOpen = true
		m.diffTitle = item.Path + " (untracked)"
		m.diffContent = "This file is untracked.\nStage it to see a diff."
		m.diffErr = nil
		m.diffScroll = 0
		return nil
	}

	gen := m.detailGen
	m.diffLoading = true
	svc := m.repos
	target := item.Target
	path := item.Path

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		title := path + " (" + string(target) + ")"
		if svc == nil {
			return diffMsg{gen: gen, title: title, err: fmt.Errorf("repository service unavailable")}
		}
		content, err := svc.Diff(ctx, repo, target, path)
		return diffMsg{gen: gen, title: title, content: content, err: err}
	}
}

func (m *Model) refreshCurrentRepository() tea.Cmd {
	repo, ok := m.currentRepo()
	if !ok {
		return nil
	}
	return m.refreshRepoID(repo.ID)
}

// refreshRepoID re-reads a single repository and updates its row.
func (m *Model) refreshRepoID(id domain.RepositoryID) tea.Cmd {
	repo, ok := m.cfg.RepositoryByID(id)
	if !ok || m.status == nil {
		return nil
	}

	for i := range m.rows {
		if m.rows[i].repo.ID == id {
			m.rows[i].loading = true
		}
	}

	svc := m.status
	gen := m.gen

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return statusMsg{gen: gen, state: svc.Status(ctx, repo)}
	}
}

func (m *Model) handleRepositoryKey(key string) tea.Cmd {
	if m.diffOpen {
		return m.handleDiffKey(key)
	}
	if m.commitOpen {
		page := max(1, m.height-14)
		switch key {
		case "esc", "q", "enter":
			m.commitOpen = false
		case "down", "j":
			m.commitScroll++
		case "up", "k":
			m.commitScroll--
		case "pgdown", " ":
			m.commitScroll += page
		case "pgup":
			m.commitScroll -= page
		case "g":
			m.commitScroll = 0
		case "G":
			m.commitScroll = len(strings.Split(m.commitContent, "\n"))
		}
		m.clampCommitScroll(page)
		return nil
	}
	if m.errOpen {
		switch key {
		case "esc", "enter", "q":
			m.errOpen = false
		}
		return nil
	}

	repo, hasRepo := m.currentRepo()

	switch key {
	case "f", "p", "P", "c", "n", "A", "u", " ":
		if !m.gitReady() {
			return nil
		}
	}

	switch key {
	case "esc", "q":
		m.closeRepository()

	case "1", "2", "3", "4":
		m.repoTab = repoTab(int(key[0] - '1'))

	case "left", "h":
		if m.repoTab > tabOverview {
			m.repoTab--
		}
	case "right", "l":
		if int(m.repoTab) < len(repoTabNames)-1 {
			m.repoTab++
		}

	case "up", "k":
		m.moveInTab(-1)
	case "down", "j":
		m.moveInTab(1)
	case "g":
		m.jumpInTab(0)
	case "G":
		m.jumpInTab(-1)

	case "b":
		m.repoTab = tabBranches

	case "d":
		m.openDeleteBranchConfirm()
		return nil

	case "enter":
		return m.enterInTab()

	case "r":
		return tea.Batch(m.refreshCurrentRepository(), m.loadDetail())

	case "c":
		return m.openCommitPrompt()
	case "n":
		return m.openBranchPrompt()
	case " ":
		return m.stageSelected()
	case "A":
		if hasRepo {
			return m.runWrite("stage all", repo, func(ctx context.Context) error {
				return m.git.StageAll(ctx, repo)
			})
		}
	case "u":
		if hasRepo {
			return m.runWrite("unstage all", repo, func(ctx context.Context) error {
				return m.git.UnstageAll(ctx, repo)
			})
		}

	case "f":
		if !hasRepo {
			return nil
		}
		return m.runWrite("fetch "+repo.Name, repo, func(ctx context.Context) error {
			return m.git.Fetch(ctx, repo)
		})
	case "p":
		if !hasRepo {
			return nil
		}
		return m.runWrite("pull "+repo.Name, repo, func(ctx context.Context) error {
			return m.git.Pull(ctx, repo)
		})
	case "P":
		if !hasRepo {
			return nil
		}
		return m.runWrite("push "+repo.Name, repo, func(ctx context.Context) error {
			return m.git.Push(ctx, repo)
		})

	case "e":
		m.openErrorDetail()
	}

	return nil
}

func (m *Model) handleDiffKey(key string) tea.Cmd {
	page := max(1, m.height-8)

	switch key {
	case "esc", "q", "enter":
		m.diffOpen = false
	case "down", "j":
		m.diffScroll++
	case "up", "k":
		m.diffScroll--
	case "pgdown", " ":
		m.diffScroll += page
	case "pgup":
		m.diffScroll -= page
	case "g":
		m.diffScroll = 0
	case "G":
		m.diffScroll = len(strings.Split(m.diffContent, "\n"))
	}

	m.clampDiffScroll(page)
	return nil
}

func (m *Model) clampDiffScroll(page int) {
	total := len(strings.Split(m.diffContent, "\n"))
	maxScroll := max(0, total-page)
	if m.diffScroll > maxScroll {
		m.diffScroll = maxScroll
	}
	if m.diffScroll < 0 {
		m.diffScroll = 0
	}
}

func (m *Model) moveInTab(delta int) {
	switch m.repoTab {
	case tabChanges:
		m.changeIndex = clampIndex(m.changeIndex+delta, len(m.changes))
	case tabBranches:
		m.branchIndex = clampIndex(m.branchIndex+delta, len(m.detail.Branches))
	case tabHistory:
		m.historyIndex = clampIndex(m.historyIndex+delta, len(m.detail.Commits))
	}
}

func (m *Model) jumpInTab(which int) {
	count := 0
	switch m.repoTab {
	case tabChanges:
		count = len(m.changes)
	case tabBranches:
		count = len(m.detail.Branches)
	case tabHistory:
		count = len(m.detail.Commits)
	}
	if count == 0 {
		return
	}
	if which < 0 {
		which = count - 1
	}
	target := clampIndex(which, count)
	switch m.repoTab {
	case tabChanges:
		m.changeIndex = target
	case tabBranches:
		m.branchIndex = target
	case tabHistory:
		m.historyIndex = target
	}
}

func (m *Model) enterInTab() tea.Cmd {
	switch m.repoTab {
	case tabChanges:
		if m.changeIndex >= 0 && m.changeIndex < len(m.changes) {
			return m.loadDiff(m.changes[m.changeIndex])
		}
	case tabHistory:
		if m.historyIndex >= 0 && m.historyIndex < len(m.detail.Commits) {
			m.commitOpen = true
			return m.loadCommitDetail()
		}
	case tabBranches:
		if m.branchIndex < 0 || m.branchIndex >= len(m.detail.Branches) {
			return nil
		}
		branch := m.detail.Branches[m.branchIndex]
		switch {
		case branch.IsRemote:
			m.notice = "remote branches are read-only"
			m.noticeError = false
		case branch.IsCurrent:
			m.notice = "already on " + branch.Name
			m.noticeError = false
		case m.git == nil:
			m.notice = "write operations are unavailable"
			m.noticeError = true
		default:
			repo, _ := m.currentRepo()
			return m.runWrite("checkout "+branch.Name, repo, func(ctx context.Context) error {
				return m.git.Checkout(ctx, repo, branch.Name)
			})
		}
	}
	return nil
}

func clampIndex(i, n int) int {
	if n == 0 {
		return 0
	}
	if i < 0 {
		return 0
	}
	if i >= n {
		return n - 1
	}
	return i
}

// gitReady reports whether write operations are available, showing a notice
// when they are not.
func (m *Model) gitReady() bool {
	if m.git == nil {
		m.notice = "write operations are unavailable"
		m.noticeError = true
		return false
	}
	return true
}

// setRowBusy marks a repository row as running an operation.
func (m *Model) setRowBusy(id domain.RepositoryID, label string) {
	for i := range m.rows {
		if m.rows[i].repo.ID != id {
			continue
		}
		m.rows[i].busy = label != ""
		m.rows[i].busyLabel = label
	}
}

// deleteBranchCmd deletes a local branch after confirmation.
func (m *Model) deleteBranchCmd(repoID domain.RepositoryID, branch string) tea.Cmd {
	repo, ok := m.cfg.RepositoryByID(repoID)
	if !ok || m.git == nil {
		return nil
	}
	return m.runWrite("delete branch "+branch, repo, func(ctx context.Context) error {
		return m.git.DeleteBranch(ctx, repo, branch)
	})
}

// loadCommitDetail fetches the patch for the selected history entry.
func (m *Model) loadCommitDetail() tea.Cmd {
	repo, ok := m.currentRepo()
	if !ok || m.historyIndex < 0 || m.historyIndex >= len(m.detail.Commits) {
		return nil
	}

	commit := m.detail.Commits[m.historyIndex]
	gen := m.detailGen
	svc := m.repos
	m.commitLoading = true
	m.commitContent = ""
	m.commitErr = nil
	m.commitScroll = 0

	return func() tea.Msg {
		if svc == nil {
			return commitDetailMsg{gen: gen}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		content, err := svc.Commit(ctx, repo, commit.Hash)
		return commitDetailMsg{gen: gen, content: content, err: err}
	}
}

func (m *Model) clampCommitScroll(page int) {
	total := len(strings.Split(m.commitContent, "\n"))
	maxScroll := max(0, total-page)
	if m.commitScroll > maxScroll {
		m.commitScroll = maxScroll
	}
	if m.commitScroll < 0 {
		m.commitScroll = 0
	}
}
