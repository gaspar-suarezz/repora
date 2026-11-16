package gitcli

import "testing"

func TestParseRemotes(t *testing.T) {
	out := `origin	git@github.com:acme/api.git (fetch)
origin	git@github.com:acme/api.git (push)
upstream	https://example.com/upstream.git (fetch)
upstream	https://example.com/upstream.git (push)
`

	remotes := ParseRemotes([]byte(out))

	if len(remotes) != 2 {
		t.Fatalf("remotes = %d, want 2", len(remotes))
	}
	if remotes[0].Name != "origin" || remotes[0].FetchURL != "git@github.com:acme/api.git" {
		t.Errorf("remote[0] = %+v", remotes[0])
	}
	if remotes[1].Name != "upstream" || remotes[1].PushURL != "https://example.com/upstream.git" {
		t.Errorf("remote[1] = %+v", remotes[1])
	}
}

func TestParseRemotesEmpty(t *testing.T) {
	if got := ParseRemotes(nil); len(got) != 0 {
		t.Errorf("got %d remotes", len(got))
	}
}
