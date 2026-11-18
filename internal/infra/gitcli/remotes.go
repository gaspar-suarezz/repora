package gitcli

import (
	"context"
	"fmt"
	"strings"

	"github.com/gasparsuarez/repora/internal/domain"
)

// Remotes lists configured remotes.
func (r *Repository) Remotes(ctx context.Context, repo domain.Repository) ([]domain.Remote, error) {
	ctx, cancel := context.WithTimeout(ctx, ReadTimeout)
	defer cancel()

	stdout, stderr, err := r.ex.Execute(ctx, repo.Path, "remote", "-v")
	if err != nil {
		return nil, fmt.Errorf("listing remotes in %s: %w", repo.Path, exitSummary(err, stderr))
	}

	return ParseRemotes(stdout), nil
}

// ParseRemotes parses `git remote -v` output.
func ParseRemotes(out []byte) []domain.Remote {
	var remotes []domain.Remote
	index := make(map[string]int)

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		name, url, kind := fields[0], fields[1], fields[2]

		pos, ok := index[name]
		if !ok {
			remotes = append(remotes, domain.Remote{Name: name})
			pos = len(remotes) - 1
			index[name] = pos
		}
		switch kind {
		case "(fetch)":
			remotes[pos].FetchURL = url
		case "(push)":
			remotes[pos].PushURL = url
		}
	}

	return remotes
}
