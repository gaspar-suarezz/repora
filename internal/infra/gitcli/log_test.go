package gitcli

import "testing"

func TestParseLog(t *testing.T) {
	out := "abcdef1234567890\x1fabcdef1\x1fAna Ruiz\x1fana@example.com\x1f2026-01-02T10:00:00+01:00\x1fAdd payment retry\x1fHEAD -> main, origin/main\n" +
		"1234567890abcdef\x1f1234567\x1fLuis M.\x1fluis@example.com\x1f2026-01-01T09:30:00+01:00\x1fRefactor café handler\x1f\n"

	commits := ParseLog([]byte(out))

	if len(commits) != 2 {
		t.Fatalf("commits = %d, want 2", len(commits))
	}

	first := commits[0]
	if first.Hash != "abcdef1234567890" || first.Short != "abcdef1" {
		t.Errorf("hash = %q / %q", first.Hash, first.Short)
	}
	if first.Subject != "Add payment retry" {
		t.Errorf("subject = %q", first.Subject)
	}
	if first.Author != "Ana Ruiz" || first.Email != "ana@example.com" {
		t.Errorf("author = %q <%q>", first.Author, first.Email)
	}
	if first.Date.IsZero() {
		t.Error("date was not parsed")
	}
	if len(first.Refs) != 2 || first.Refs[0] != "HEAD -> main" || first.Refs[1] != "origin/main" {
		t.Errorf("refs = %v", first.Refs)
	}

	if commits[1].Subject != "Refactor café handler" {
		t.Errorf("unicode subject = %q", commits[1].Subject)
	}
	if len(commits[1].Refs) != 0 {
		t.Errorf("refs = %v, want none", commits[1].Refs)
	}
}

func TestParseLogEmpty(t *testing.T) {
	if got := ParseLog(nil); len(got) != 0 {
		t.Errorf("got %d commits", len(got))
	}
	if got := ParseLog([]byte("\n\n")); len(got) != 0 {
		t.Errorf("got %d commits", len(got))
	}
}

func TestParseLogIgnoresMalformedLines(t *testing.T) {
	out := "only-one-field\n" +
		"abcdef\x1fabcdef\x1fAna\x1fana@example.com\x1f2026-01-02T10:00:00Z\x1fok\n"
	if got := ParseLog([]byte(out)); len(got) != 1 {
		t.Fatalf("commits = %d, want 1", len(got))
	}
}
