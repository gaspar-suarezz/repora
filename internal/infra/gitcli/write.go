package gitcli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gasparsuarez/repora/internal/domain"
)

// WriteTimeout bounds local mutating operations.
const WriteTimeout = 30 * time.Second

// Stage adds paths to the index.
func (r *Repository) Stage(ctx context.Context, repo domain.Repository, paths []string) error {
	if len(paths) == 0 {
		return domain.ErrNothingToStage
	}
	args := append([]string{"add", "--"}, paths...)
	return r.write(ctx, repo, "stage", args...)
}

// StageAll stages every change, including deletions and untracked files.
func (r *Repository) StageAll(ctx context.Context, repo domain.Repository) error {
	return r.write(ctx, repo, "stage all", "add", "-A", "--")
}

// Unstage removes specific paths from the index.
func (r *Repository) Unstage(ctx context.Context, repo domain.Repository, paths []string) error {
	if len(paths) == 0 {
		return domain.ErrNothingToStage
	}

	args := append([]string{"restore", "--staged", "--"}, paths...)
	if err := r.write(ctx, repo, "unstage", args...); err != nil {
		// Unborn branch: there is no HEAD to restore from.
		rmArgs := append([]string{"rm", "--cached", "--"}, paths...)
		if fallbackErr := r.write(ctx, repo, "unstage", rmArgs...); fallbackErr != nil {
			return err
		}
	}
	return nil
}

// UnstageAll removes every staged change from the index.
func (r *Repository) UnstageAll(ctx context.Context, repo domain.Repository) error {
	ctx, cancel := context.WithTimeout(ctx, WriteTimeout)
	defer cancel()

	// Works for repositories with at least one commit.
	if _, _, err := r.ex.Execute(ctx, repo.Path, "restore", "--staged", "--", "."); err == nil {
		return nil
	}

	// Unborn branch: there is no HEAD to restore from.
	_, stderr, err := r.ex.Execute(ctx, repo.Path, "rm", "--cached", "-r", "--", ".")
	if err != nil {
		return fmt.Errorf("unstage all in %s: %w", repo.Path, exitSummary(err, stderr))
	}
	return nil
}

// Commit creates a commit with the given message.
func (r *Repository) Commit(ctx context.Context, repo domain.Repository, message string) error {
	if err := domain.ValidateCommitMessage(message); err != nil {
		return err
	}
	return r.write(ctx, repo, "commit", "commit", "-m", message)
}

// CreateBranch creates a branch, optionally checking it out.
func (r *Repository) CreateBranch(ctx context.Context, repo domain.Repository, name string, checkout bool) error {
	if err := domain.ValidateBranchName(name); err != nil {
		return err
	}
	if checkout {
		return r.write(ctx, repo, "create branch", "switch", "-c", name)
	}
	return r.write(ctx, repo, "create branch", "branch", name)
}

// DeleteBranch deletes a local branch. Without force, Git refuses to delete a
// branch that is not fully merged.
func (r *Repository) DeleteBranch(ctx context.Context, repo domain.Repository, name string, force bool) error {
	if err := domain.ValidateBranchName(name); err != nil {
		return err
	}

	current, err := r.currentBranch(ctx, repo)
	if err == nil && current == name {
		return fmt.Errorf("%w: %q", domain.ErrCurrentBranch, name)
	}

	if force {
		return r.write(ctx, repo, "delete branch", "branch", "-D", name)
	}
	return r.write(ctx, repo, "delete branch", "branch", "-d", name)
}

// Checkout switches to an existing branch.
func (r *Repository) Checkout(ctx context.Context, repo domain.Repository, branch string) error {
	if err := domain.ValidateBranchName(branch); err != nil {
		return err
	}
	return r.write(ctx, repo, "checkout", "switch", branch)
}

// Fetch updates remote-tracking refs.
func (r *Repository) Fetch(ctx context.Context, repo domain.Repository) error {
	return r.writeNoTimeout(ctx, repo, "fetch", "fetch", "--prune", "--no-recurse-submodules")
}

// Pull fast-forwards the current branch.
func (r *Repository) Pull(ctx context.Context, repo domain.Repository) error {
	return r.writeNoTimeout(ctx, repo, "pull", "pull", "--ff-only", "--no-recurse-submodules")
}

// Push publishes the current branch, setting the upstream when needed.
func (r *Repository) Push(ctx context.Context, repo domain.Repository) error {
	hasUpstream := r.hasUpstream(ctx, repo)
	if hasUpstream {
		return r.writeNoTimeout(ctx, repo, "push", "push", "--no-recurse-submodules")
	}

	branch, err := r.currentBranch(ctx, repo)
	if err != nil {
		return err
	}
	remote, err := r.defaultRemote(ctx, repo)
	if err != nil {
		return err
	}
	return r.writeNoTimeout(ctx, repo, "push", "push", "--no-recurse-submodules", "-u", remote, branch)
}

func (r *Repository) hasUpstream(ctx context.Context, repo domain.Repository) bool {
	ctx, cancel := context.WithTimeout(ctx, ReadTimeout)
	defer cancel()
	_, _, err := r.ex.Execute(ctx, repo.Path, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	return err == nil
}

func (r *Repository) currentBranch(ctx context.Context, repo domain.Repository) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, ReadTimeout)
	defer cancel()

	out, _, err := r.ex.Execute(ctx, repo.Path, "symbolic-ref", "--short", "-q", "HEAD")
	if err != nil {
		return "", fmt.Errorf("%w: %s", domain.ErrDetachedHead, repo.Path)
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" {
		return "", fmt.Errorf("%w: %s", domain.ErrDetachedHead, repo.Path)
	}
	return branch, nil
}

func (r *Repository) defaultRemote(ctx context.Context, repo domain.Repository) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, ReadTimeout)
	defer cancel()

	out, _, err := r.ex.Execute(ctx, repo.Path, "remote")
	if err != nil {
		return "", fmt.Errorf("listing remotes in %s: %w", repo.Path, err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		if name == "origin" {
			return name, nil
		}
	}
	// No origin: use the first configured remote.
	for _, line := range strings.Split(string(out), "\n") {
		if name := strings.TrimSpace(line); name != "" {
			return name, nil
		}
	}
	return "", fmt.Errorf("%w: %s", domain.ErrNoRemote, repo.Path)
}

func (r *Repository) write(ctx context.Context, repo domain.Repository, op string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, WriteTimeout)
	defer cancel()
	return r.execWrite(ctx, repo, op, args...)
}

func (r *Repository) writeNoTimeout(ctx context.Context, repo domain.Repository, op string, args ...string) error {
	return r.execWrite(ctx, repo, op, args...)
}

func (r *Repository) execWrite(ctx context.Context, repo domain.Repository, op string, args ...string) error {
	_, stderr, err := r.ex.Execute(ctx, repo.Path, args...)
	if err != nil {
		return fmt.Errorf("%s in %s: %w", op, repo.Path, exitSummary(err, stderr))
	}
	return nil
}
