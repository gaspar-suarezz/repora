package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gaspar-suarezz/repora/internal/app"
	"github.com/gaspar-suarezz/repora/internal/domain"
)

type fakeWriter struct {
	calls []string
	err   error
}

func (w *fakeWriter) record(op string) error {
	w.calls = append(w.calls, op)
	return w.err
}

func (w *fakeWriter) Stage(context.Context, domain.Repository, []string) error {
	return w.record("stage")
}
func (w *fakeWriter) StageAll(context.Context, domain.Repository) error {
	return w.record("stage-all")
}
func (w *fakeWriter) Unstage(context.Context, domain.Repository, []string) error {
	return w.record("unstage")
}
func (w *fakeWriter) UnstageAll(context.Context, domain.Repository) error {
	return w.record("unstage-all")
}
func (w *fakeWriter) Commit(context.Context, domain.Repository, string) error {
	return w.record("commit")
}
func (w *fakeWriter) CreateBranch(context.Context, domain.Repository, string, bool) error {
	return w.record("create-branch")
}
func (w *fakeWriter) Checkout(context.Context, domain.Repository, string) error {
	return w.record("checkout")
}
func (w *fakeWriter) DeleteBranch(context.Context, domain.Repository, string, bool) error {
	return w.record("delete-branch")
}
func (w *fakeWriter) Fetch(context.Context, domain.Repository) error {
	return w.record("fetch")
}
func (w *fakeWriter) Pull(context.Context, domain.Repository) error {
	return w.record("pull")
}
func (w *fakeWriter) Push(context.Context, domain.Repository) error {
	return w.record("push")
}

func writeModel(t *testing.T) (*Model, *fakeWriter) {
	t.Helper()
	m := repositoryModel(t)
	w := &fakeWriter{}
	m.git = app.NewGitService(w, app.NewLocks())
	m.repoTab = tabChanges
	return m, w
}

func typeText(t *testing.T, m *Model, text string) {
	t.Helper()
	for _, r := range text {
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

func TestCommitPromptFlow(t *testing.T) {
	m, writer := writeModel(t)

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if !m.prompt.active() || m.prompt.kind != promptCommit {
		t.Fatalf("commit prompt did not open: %+v", m.prompt.kind)
	}
	if out := m.View(); !strings.Contains(out, "COMMIT MESSAGE") {
		t.Errorf("prompt view missing title:\n%s", out)
	}

	typeText(t, m, "fix retry")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("submitting the prompt produced no command")
	}
	if m.prompt.active() {
		t.Error("prompt should close after submit")
	}

	msg := cmd()
	done, ok := msg.(writeDoneMsg)
	if !ok {
		t.Fatalf("msg = %T", msg)
	}
	if done.err != nil {
		t.Fatalf("write error: %v", done.err)
	}
	if len(writer.calls) != 1 || writer.calls[0] != "commit" {
		t.Errorf("calls = %v", writer.calls)
	}

	_, _ = m.Update(done)
	if !strings.Contains(m.notice, "commit") || m.noticeError {
		t.Errorf("notice = %q (error=%v)", m.notice, m.noticeError)
	}
}

func TestCommitPromptRejectsEmptyMessage(t *testing.T) {
	m, writer := writeModel(t)

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Error("empty message should not start a command")
	}
	if !m.prompt.active() {
		t.Error("prompt closed on invalid input")
	}
	if m.prompt.err == nil {
		t.Error("expected an inline validation error")
	}
	if len(writer.calls) != 0 {
		t.Errorf("writer was called: %v", writer.calls)
	}
}

func TestBranchPromptValidatesName(t *testing.T) {
	m, writer := writeModel(t)
	m.repoTab = tabBranches

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if m.prompt.kind != promptNewBranch {
		t.Fatalf("branch prompt did not open: %v", m.prompt.kind)
	}

	typeText(t, m, "bad..name")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("invalid branch name should not start a command")
	}
	if m.prompt.err == nil {
		t.Error("expected a validation error")
	}
	if len(writer.calls) != 0 {
		t.Errorf("writer was called: %v", writer.calls)
	}

	// Escape clears the prompt.
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.prompt.active() {
		t.Error("prompt did not close on escape")
	}
}

func TestBranchPromptCreatesBranch(t *testing.T) {
	m, writer := writeModel(t)
	m.repoTab = tabBranches

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	typeText(t, m, "feature/retry")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command")
	}

	msg := cmd()
	if done, ok := msg.(writeDoneMsg); !ok || done.err != nil {
		t.Fatalf("unexpected message: %#v", msg)
	}
	if len(writer.calls) != 1 || writer.calls[0] != "create-branch" {
		t.Errorf("calls = %v", writer.calls)
	}
}

func TestStageSelectedFile(t *testing.T) {
	m, writer := writeModel(t)

	// Index 1 is the unstaged file.
	m.changeIndex = 1
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	if cmd == nil {
		t.Fatal("stage produced no command")
	}
	msg := cmd()
	if done, ok := msg.(writeDoneMsg); !ok || done.err != nil {
		t.Fatalf("unexpected message: %#v", msg)
	}
	if len(writer.calls) != 1 || writer.calls[0] != "stage" {
		t.Errorf("calls = %v", writer.calls)
	}
	if !strings.Contains(m.notice, "stage") {
		t.Errorf("notice = %q", m.notice)
	}
}

func TestSpaceOnStagedFileUnstages(t *testing.T) {
	m, writer := writeModel(t)

	// Index 0 is already staged, so Space unstages it.
	m.changeIndex = 0
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	if cmd == nil {
		t.Fatal("unstage produced no command")
	}
	if done, ok := cmd().(writeDoneMsg); !ok || done.err != nil {
		t.Fatalf("unexpected result")
	}
	if len(writer.calls) != 1 || writer.calls[0] != "unstage" {
		t.Errorf("calls = %v", writer.calls)
	}
}

func TestWriteFailureShowsErrorDetail(t *testing.T) {
	m, _ := writeModel(t)

	_, _ = m.Update(writeDoneMsg{
		label:  "pull",
		repoID: "api",
		err:    errors.New("fatal: Not possible to fast-forward, aborting."),
	})

	if !m.noticeError {
		t.Error("notice should be flagged as an error")
	}
	if !strings.Contains(m.errDetail, "fast-forward") {
		t.Errorf("errDetail = %q", m.errDetail)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if !m.errOpen {
		t.Fatal("error overlay did not open")
	}
	if out := m.View(); !strings.Contains(out, "fast-forward") {
		t.Errorf("overlay missing error text:\n%s", out)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.errOpen {
		t.Error("error overlay did not close")
	}
}

func TestFetchFromWorkspaceDashboard(t *testing.T) {
	m, writer := writeModel(t)
	m.screen = screenWorkspace
	m.repoTab = tabOverview

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	if cmd == nil {
		t.Fatal("fetch produced no command")
	}
	if done, ok := cmd().(writeDoneMsg); !ok || done.err != nil {
		t.Fatalf("unexpected result")
	}
	if len(writer.calls) != 1 || writer.calls[0] != "fetch" {
		t.Errorf("calls = %v", writer.calls)
	}
}

func TestWriteKeysRequireService(t *testing.T) {
	m := repositoryModel(t)
	m.git = nil
	m.repoTab = tabChanges

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	if cmd != nil {
		t.Error("stage all should be a no-op without a Git service")
	}
	if !m.noticeError {
		t.Errorf("expected an error notice, got %q", m.notice)
	}
}

func TestDeleteBranchRequiresConfirmation(t *testing.T) {
	m, writer := writeModel(t)
	m.repoTab = tabBranches
	m.detail.Branches = []domain.Branch{
		{Name: "feature/x", HasUpstream: true},
		{Name: "main", IsCurrent: true},
	}
	m.branchIndex = 0

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if !m.confirm.active || m.confirm.action != confirmDeleteBranch {
		t.Fatalf("confirm = %+v", m.confirm)
	}
	if !strings.Contains(m.View(), "Delete branch") {
		t.Errorf("confirmation title missing:\n%s", m.View())
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("confirming produced no command")
	}
	if done, ok := cmd().(writeDoneMsg); !ok || done.err != nil {
		t.Fatalf("unexpected result")
	}
	if len(writer.calls) != 1 || writer.calls[0] != "delete-branch" {
		t.Errorf("calls = %v", writer.calls)
	}
}

func TestDeleteCurrentBranchBlocked(t *testing.T) {
	m, writer := writeModel(t)
	m.repoTab = tabBranches
	m.detail.Branches = []domain.Branch{{Name: "main", IsCurrent: true}}
	m.branchIndex = 0

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	if m.confirm.active {
		t.Fatal("confirmation should not open for the current branch")
	}
	if !m.noticeError {
		t.Errorf("expected an error notice, got %q", m.notice)
	}
	if len(writer.calls) != 0 {
		t.Errorf("writer was called: %v", writer.calls)
	}
}

func TestDeleteRemoteBranchBlocked(t *testing.T) {
	m, writer := writeModel(t)
	m.repoTab = tabBranches
	m.detail.Branches = []domain.Branch{{Name: "origin/main", IsRemote: true}}
	m.branchIndex = 0

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	if m.confirm.active {
		t.Fatal("remote branches must not be deletable")
	}
	if len(writer.calls) != 0 {
		t.Errorf("writer was called: %v", writer.calls)
	}
}

func TestQuitKeyNavigatesBack(t *testing.T) {
	m := repositoryModel(t)

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if m.screen != screenWorkspace {
		t.Fatalf("screen = %v, want workspace", m.screen)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if m.screen != screenHome {
		t.Fatalf("screen = %v, want home", m.screen)
	}
}
