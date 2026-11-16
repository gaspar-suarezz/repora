package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNoticeExpiryClearsSuccess(t *testing.T) {
	m := repositoryModel(t)
	m.notice = "fetch · done"
	m.noticeError = false
	m.noticeSeq = 7

	_, _ = m.Update(noticeExpiredMsg{seq: 7})

	if m.notice != "" {
		t.Errorf("notice = %q, want empty", m.notice)
	}
}

func TestNoticeExpiryKeepsErrors(t *testing.T) {
	m := repositoryModel(t)
	m.notice = "pull failed"
	m.noticeError = true
	m.noticeSeq = 7

	_, _ = m.Update(noticeExpiredMsg{seq: 7})

	if m.notice == "" {
		t.Error("error notices must persist")
	}
}

func TestNoticeExpiryIgnoresStaleSequence(t *testing.T) {
	m := repositoryModel(t)
	m.notice = "newer notice"
	m.noticeError = false
	m.noticeSeq = 9

	_, _ = m.Update(noticeExpiredMsg{seq: 8})

	if m.notice != "newer notice" {
		t.Errorf("notice = %q, want it preserved", m.notice)
	}
}

func TestUpdateSchedulesNoticeTimer(t *testing.T) {
	cfg := configWith("alpha")
	m := testNew(cfg, "/tmp/config.yaml", nil)

	var scheduled time.Duration
	m.tick = func(d time.Duration, fn func(time.Time) tea.Msg) tea.Cmd {
		scheduled = d
		return func() tea.Msg { return fn(time.Now()) }
	}

	_, cmd := m.Update(configChangedMsg{action: "create workspace", cfg: cfg})
	if cmd == nil {
		t.Fatal("expected a timer command")
	}
	if scheduled != noticeDuration {
		t.Errorf("duration = %v, want %v", scheduled, noticeDuration)
	}

	msg := cmd()
	expired, ok := msg.(noticeExpiredMsg)
	if !ok {
		t.Fatalf("msg = %T", msg)
	}
	_, _ = m.Update(expired)

	if m.notice != "" {
		t.Errorf("notice = %q, want it cleared", m.notice)
	}
}

func TestUpdateDoesNotScheduleTimerForErrors(t *testing.T) {
	m := testNew(configWith("alpha"), "/tmp/config.yaml", nil)

	called := false
	m.tick = func(time.Duration, func(time.Time) tea.Msg) tea.Cmd {
		called = true
		return nil
	}

	_, _ = m.Update(writeDoneMsg{label: "push", repoID: "api", err: errBoom{}})

	if called {
		t.Error("error notices must not auto-dismiss")
	}
}

func TestWriteOperationMarksRowBusy(t *testing.T) {
	m, _ := writeModel(t)

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

	if !m.rows[0].busy {
		t.Fatal("row should be busy during the operation")
	}
	if m.rows[0].busyLabel == "" {
		t.Error("busy label is empty")
	}
	if out := m.View(); !containsAny(out, "◐") {
		t.Errorf("busy state not rendered:\n%s", out)
	}

	_, _ = m.Update(writeDoneMsg{label: m.rows[0].busyLabel, repoID: "api"})

	if m.rows[0].busy {
		t.Error("row still busy after the operation completed")
	}
}

type errBoom struct{}

func (errBoom) Error() string { return "boom" }

func containsAny(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
