package domain

import "context"

// GitWriter is the mutating Git contract. Implementations must not run two
// operations against the same repository concurrently; the application layer
// guarantees that with per-repository locks.
type GitWriter interface {
	Stage(ctx context.Context, repo Repository, paths []string) error
	StageAll(ctx context.Context, repo Repository) error
	Unstage(ctx context.Context, repo Repository, paths []string) error
	UnstageAll(ctx context.Context, repo Repository) error
	Commit(ctx context.Context, repo Repository, message string) error

	CreateBranch(ctx context.Context, repo Repository, name string, checkout bool) error
	Checkout(ctx context.Context, repo Repository, branch string) error
	DeleteBranch(ctx context.Context, repo Repository, name string, force bool) error

	Fetch(ctx context.Context, repo Repository) error
	Pull(ctx context.Context, repo Repository) error
	Push(ctx context.Context, repo Repository) error
}
