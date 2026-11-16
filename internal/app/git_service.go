package app

import (
	"context"
	"fmt"

	"github.com/gaspar-suarezz/repora/internal/domain"
)

// GitService runs mutating Git operations with per-repository serialization and
// input validation.
type GitService struct {
	writer domain.GitWriter
	locks  *Locks
}

// NewGitService returns a service backed by the Git write adapter.
func NewGitService(writer domain.GitWriter, locks *Locks) *GitService {
	return &GitService{writer: writer, locks: locks}
}

// IsBusy reports whether a write operation holds the repository lock.
func (s *GitService) IsBusy(id domain.RepositoryID) bool {
	return s.locks != nil && s.locks.IsLocked(id)
}

// Stage adds the given paths to the index.
func (s *GitService) Stage(ctx context.Context, repo domain.Repository, paths []string) error {
	if len(paths) == 0 {
		return domain.ErrNothingToStage
	}
	return s.run(ctx, repo.ID, func(ctx context.Context) error {
		return s.writer.Stage(ctx, repo, paths)
	})
}

// StageAll stages every change, including deletions and untracked files.
func (s *GitService) StageAll(ctx context.Context, repo domain.Repository) error {
	return s.run(ctx, repo.ID, func(ctx context.Context) error {
		return s.writer.StageAll(ctx, repo)
	})
}

// Unstage removes specific paths from the index.
func (s *GitService) Unstage(ctx context.Context, repo domain.Repository, paths []string) error {
	if len(paths) == 0 {
		return domain.ErrNothingToStage
	}
	return s.run(ctx, repo.ID, func(ctx context.Context) error {
		return s.writer.Unstage(ctx, repo, paths)
	})
}

// UnstageAll removes every staged change from the index.
func (s *GitService) UnstageAll(ctx context.Context, repo domain.Repository) error {
	return s.run(ctx, repo.ID, func(ctx context.Context) error {
		return s.writer.UnstageAll(ctx, repo)
	})
}

// Commit creates a commit with the given message.
func (s *GitService) Commit(ctx context.Context, repo domain.Repository, message string) error {
	if err := domain.ValidateCommitMessage(message); err != nil {
		return err
	}
	return s.run(ctx, repo.ID, func(ctx context.Context) error {
		return s.writer.Commit(ctx, repo, message)
	})
}

// CreateBranch creates a branch and checks it out.
func (s *GitService) CreateBranch(ctx context.Context, repo domain.Repository, name string) error {
	if err := domain.ValidateBranchName(name); err != nil {
		return err
	}
	return s.run(ctx, repo.ID, func(ctx context.Context) error {
		return s.writer.CreateBranch(ctx, repo, name, true)
	})
}

// Checkout switches to an existing branch.
func (s *GitService) Checkout(ctx context.Context, repo domain.Repository, branch string) error {
	if err := domain.ValidateBranchName(branch); err != nil {
		return err
	}
	return s.run(ctx, repo.ID, func(ctx context.Context) error {
		return s.writer.Checkout(ctx, repo, branch)
	})
}

// DeleteBranch deletes a local branch. The safe form refuses unmerged branches.
func (s *GitService) DeleteBranch(ctx context.Context, repo domain.Repository, name string) error {
	if err := domain.ValidateBranchName(name); err != nil {
		return err
	}
	return s.run(ctx, repo.ID, func(ctx context.Context) error {
		return s.writer.DeleteBranch(ctx, repo, name, false)
	})
}

// Fetch updates remote-tracking refs. It writes refs, so it is serialized.
func (s *GitService) Fetch(ctx context.Context, repo domain.Repository) error {
	return s.run(ctx, repo.ID, func(ctx context.Context) error {
		return s.writer.Fetch(ctx, repo)
	})
}

// Pull fast-forwards the current branch.
func (s *GitService) Pull(ctx context.Context, repo domain.Repository) error {
	return s.run(ctx, repo.ID, func(ctx context.Context) error {
		return s.writer.Pull(ctx, repo)
	})
}

// Push publishes the current branch.
func (s *GitService) Push(ctx context.Context, repo domain.Repository) error {
	return s.run(ctx, repo.ID, func(ctx context.Context) error {
		return s.writer.Push(ctx, repo)
	})
}

func (s *GitService) run(ctx context.Context, id domain.RepositoryID, fn func(context.Context) error) error {
	if s.locks == nil {
		return fn(ctx)
	}
	release, err := s.locks.Acquire(ctx, id)
	if err != nil {
		return fmt.Errorf("repository %q is busy: %w", id, err)
	}
	defer release()
	return fn(ctx)
}
