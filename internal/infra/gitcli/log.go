package gitcli

import (
	"context"
	"fmt"
	"strings"

	"github.com/gasparsuarez/repora/internal/domain"
)

// DefaultLogLimit is the default number of commits requested.
const DefaultLogLimit = 50

// MaxLogLimit caps history queries.
const MaxLogLimit = 500

// Log returns the most recent commits, newest first.
func (r *Repository) Log(ctx context.Context, repo domain.Repository, limit int) ([]domain.Commit, error) {
	if limit <= 0 {
		limit = DefaultLogLimit
	}
	if limit > MaxLogLimit {
		limit = MaxLogLimit
	}

	ctx, cancel := context.WithTimeout(ctx, ReadTimeout)
	defer cancel()

	stdout, stderr, err := r.ex.Execute(ctx, repo.Path,
		"log", "-n", fmt.Sprintf("%d", limit), "--date=iso-strict",
		"--pretty=format:%H%x1f%h%x1f%an%x1f%ae%x1f%aI%x1f%s%x1f%D",
	)
	if err != nil {
		// An unborn branch has no commits; that is an empty history, not a failure.
		if strings.Contains(string(stderr), "does not have any commits yet") {
			return nil, nil
		}
		return nil, fmt.Errorf("reading history in %s: %w", repo.Path, exitSummary(err, stderr))
	}

	return ParseLog(stdout), nil
}

// ParseLog parses the output produced by Log's pretty format.
func ParseLog(out []byte) []domain.Commit {
	var commits []domain.Commit

	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\x1f")
		if len(fields) < 6 {
			continue
		}

		commit := domain.Commit{
			Hash:    fields[0],
			Short:   fields[1],
			Author:  fields[2],
			Email:   fields[3],
			Date:    parseISODate(fields[4]),
			Subject: fields[5],
		}
		if len(fields) >= 7 && strings.TrimSpace(fields[6]) != "" {
			for _, ref := range strings.Split(fields[6], ", ") {
				if ref = strings.TrimSpace(ref); ref != "" {
					commit.Refs = append(commit.Refs, ref)
				}
			}
		}
		commits = append(commits, commit)
	}

	return commits
}
