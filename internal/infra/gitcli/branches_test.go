package gitcli

import "testing"

func TestParseBranches(t *testing.T) {
	line := func(ref, short, head, upstream, track, sha, author, date, subject string) string {
		return ref + "\x00" + short + "\x00" + head + "\x00" + upstream + "\x00" + track + "\x00" +
			sha + "\x00" + author + "\x00" + date + "\x00" + subject + "\n"
	}

	out := line("refs/heads/main", "main", "*", "origin/main", "[ahead 1, behind 2]", "aaa111", "Ana", "2026-01-02T10:00:00+01:00", "Add retry") +
		line("refs/heads/feature/x", "feature/x", " ", "", "", "bbb222", "Luis", "2026-01-01T09:00:00+01:00", "Feature work") +
		line("refs/remotes/origin/main", "origin/main", " ", "", "", "ccc333", "Ana", "2026-01-02T10:00:00+01:00", "Add retry") +
		line("refs/remotes/origin/HEAD", "origin/HEAD", " ", "", "", "ccc333", "Ana", "2026-01-02T10:00:00+01:00", "Add retry")

	branches := ParseBranches([]byte(out))

	if len(branches) != 3 {
		t.Fatalf("branches = %d, want 3 (remote HEAD must be filtered)", len(branches))
	}

	main := branches[0]
	if main.Name != "main" || !main.IsCurrent || main.IsRemote {
		t.Fatalf("unexpected first branch: %+v", main)
	}
	if main.Ahead != 1 || main.Behind != 2 {
		t.Errorf("ahead/behind = %d/%d, want 1/2", main.Ahead, main.Behind)
	}
	if !main.HasUpstream || main.Upstream != "origin/main" {
		t.Errorf("upstream = %q", main.Upstream)
	}
	if main.LastCommit.Short != "aaa111" || main.LastCommit.Subject != "Add retry" {
		t.Errorf("last commit = %+v", main.LastCommit)
	}
	if main.LastCommit.Date.IsZero() {
		t.Error("date was not parsed")
	}

	feature := branches[1]
	if feature.Name != "feature/x" || feature.IsCurrent || feature.HasUpstream {
		t.Fatalf("unexpected second branch: %+v", feature)
	}

	if !branches[2].IsRemote {
		t.Errorf("third branch should be remote: %+v", branches[2])
	}
}

func TestParseBranchesGoneUpstream(t *testing.T) {
	out := "refs/heads/main\x00main\x00*\x00origin/main\x00[gone]\x00aaa\x00Ana\x002026-01-02T10:00:00Z\x00msg\n"
	branches := ParseBranches([]byte(out))
	if len(branches) != 1 {
		t.Fatalf("branches = %d", len(branches))
	}
	if !branches[0].Gone {
		t.Error("expected gone upstream")
	}
}

func TestParseBranchesEmpty(t *testing.T) {
	if got := ParseBranches(nil); len(got) != 0 {
		t.Errorf("got %d branches", len(got))
	}
}
