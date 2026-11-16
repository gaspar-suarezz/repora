package app

import (
	"errors"
	"strings"
	"testing"
)

func TestExplainCommonFailures(t *testing.T) {
	tests := []struct {
		message string
		want    string
	}{
		{"git switch develop: exit code 128: fatal: cannot switch branch while cherry-picking", "cherry-pick --continue"},
		{"fatal: cannot switch branch while merging", "merge --continue"},
		{"fatal: cannot switch branch while rebasing", "rebase --continue"},
		{"error: the branch 'x' is not fully merged", "branch -D"},
		{"fatal: The current branch x has no upstream branch", "upstream"},
		{"fatal: Not possible to fast-forward, aborting.", "diverged"},
		{"fatal: cannot lock ref 'HEAD': index.lock exists", "lock file"},
		{"fatal: Authentication failed for 'https://example.com'", "authenticate"},
	}

	for _, tc := range tests {
		hint := Explain(errors.New(tc.message))
		if hint == "" {
			t.Errorf("Explain(%q) returned no hint", tc.message)
			continue
		}
		if !strings.Contains(strings.ToLower(hint), strings.ToLower(tc.want)) {
			t.Errorf("Explain(%q) = %q, want it to mention %q", tc.message, hint, tc.want)
		}
	}
}

func TestExplainUnknownFailure(t *testing.T) {
	if hint := Explain(errors.New("something unexpected")); hint != "" {
		t.Errorf("hint = %q, want empty", hint)
	}
	if hint := Explain(nil); hint != "" {
		t.Errorf("hint = %q, want empty", hint)
	}
}
