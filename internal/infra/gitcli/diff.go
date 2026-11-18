package gitcli

import (
	"context"
	"fmt"
	"strings"

	"github.com/gasparsuarez/repora/internal/domain"
)

// MaxDiffLines bounds rendered diff size.
const MaxDiffLines = 5000

// Diff returns a text diff for a single path, or for the whole worktree when
// path is empty.
func (r *Repository) Diff(ctx context.Context, repo domain.Repository, target domain.DiffTarget, path string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, ReadTimeout)
	defer cancel()

	args := []string{"--no-pager", "diff", "--no-color", "--no-ext-diff"}
	switch target {
	case domain.DiffStaged:
		args = append(args, "--cached")
	case domain.DiffUnstaged, domain.DiffCommit:
		// working tree diff
	default:
		return "", fmt.Errorf("unsupported diff target %q", target)
	}

	if path != "" {
		args = append(args, "--", path)
	}

	stdout, stderr, err := r.ex.Execute(ctx, repo.Path, args...)
	if err != nil {
		return "", fmt.Errorf("diff in %s: %w", repo.Path, exitSummary(err, stderr))
	}

	return truncateDiff(string(stdout)), nil
}

// ShowCommit returns the diff introduced by a single commit.
func (r *Repository) ShowCommit(ctx context.Context, repo domain.Repository, hash string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, ReadTimeout)
	defer cancel()

	stdout, stderr, err := r.ex.Execute(ctx, repo.Path,
		"--no-pager", "show", "--no-color", "--no-ext-diff", "--stat", "--patch", hash,
	)
	if err != nil {
		return "", fmt.Errorf("showing %s in %s: %w", hash, repo.Path, exitSummary(err, stderr))
	}
	return truncateDiff(string(stdout)), nil
}

func truncateDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	if len(lines) <= MaxDiffLines {
		return diff
	}
	return strings.Join(lines[:MaxDiffLines], "\n") +
		fmt.Sprintf("\n… diff truncated at %d lines\n", MaxDiffLines)
}
