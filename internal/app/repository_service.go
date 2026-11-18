package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gasparsuarez/repora/internal/domain"
)

// RepositoryDetail is the bundle of read-only information shown by the
// repository screen. Section errors are reported individually so one failing
// Git command does not blank the whole screen.
type RepositoryDetail struct {
	Branches []domain.Branch
	Commits  []domain.Commit
	Remotes  []domain.Remote

	BranchErr error
	LogErr    error
	RemoteErr error
}

// RepositoryService reads the detail view of a single repository and manages
// repository registration.
type RepositoryService struct {
	git   domain.GitPort
	store domain.ConfigStore
}

// NewRepositoryService returns a service backed by the Git adapter and the
// configuration store.
func NewRepositoryService(git domain.GitPort, store domain.ConfigStore) *RepositoryService {
	return &RepositoryService{git: git, store: store}
}

// Load reads branches, history, and remotes concurrently.
func (s *RepositoryService) Load(ctx context.Context, repo domain.Repository) RepositoryDetail {
	var detail RepositoryDetail
	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		defer wg.Done()
		detail.Branches, detail.BranchErr = s.git.Branches(ctx, repo)
	}()
	go func() {
		defer wg.Done()
		detail.Commits, detail.LogErr = s.git.Log(ctx, repo, 50)
	}()
	go func() {
		defer wg.Done()
		detail.Remotes, detail.RemoteErr = s.git.Remotes(ctx, repo)
	}()
	wg.Wait()

	return detail
}

// Diff returns a diff for a path (or the whole worktree when path is empty).
func (s *RepositoryService) Diff(ctx context.Context, repo domain.Repository, target domain.DiffTarget, path string) (string, error) {
	return s.git.Diff(ctx, repo, target, path)
}

// Commit returns the patch and stat for a single commit.
func (s *RepositoryService) Commit(ctx context.Context, repo domain.Repository, hash string) (string, error) {
	return s.git.ShowCommit(ctx, repo, hash)
}

// Status returns a fresh snapshot of a single repository.
func (s *RepositoryService) Status(ctx context.Context, repo domain.Repository) domain.RepositoryState {
	return s.git.Status(ctx, repo)
}

// Register validates path as a Git working tree and adds it to the config.
// The canonical repository root is stored, not the path as typed.
func (s *RepositoryService) Register(ctx context.Context, cfg *domain.Config, path string) (domain.RepositoryID, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("%w: path is empty", domain.ErrInvalidPath)
	}

	root, err := s.git.Root(ctx, path)
	if err != nil {
		return "", err
	}

	for _, repo := range cfg.Repositories {
		if repo.Path == root {
			return "", fmt.Errorf("%w: %q is already registered as %q", domain.ErrDuplicatePath, root, repo.ID)
		}
	}

	name := filepath.Base(root)
	id := domain.RepositoryID(uniqueID(domain.NormalizeID(name), func(candidate string) bool {
		_, exists := cfg.RepositoryByID(domain.RepositoryID(candidate))
		return exists
	}))
	if id == "" {
		return "", fmt.Errorf("%w: %q produces an empty id", domain.ErrInvalidName, root)
	}

	cfg.Repositories = append(cfg.Repositories, domain.Repository{
		ID:   id,
		Name: name,
		Path: root,
	})

	if err := s.store.Save(cfg); err != nil {
		return "", err
	}
	return id, nil
}

// Unregister removes a repository from the config and from every workspace.
// Nothing is deleted from disk.
func (s *RepositoryService) Unregister(cfg *domain.Config, id domain.RepositoryID) error {
	index := -1
	for i := range cfg.Repositories {
		if cfg.Repositories[i].ID == id {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("%w: %q", domain.ErrRepositoryNotFound, id)
	}

	cfg.Repositories = append(cfg.Repositories[:index], cfg.Repositories[index+1:]...)
	for i := range cfg.Workspaces {
		ids := cfg.Workspaces[i].RepositoryIDs
		kept := ids[:0]
		for _, existing := range ids {
			if existing != id {
				kept = append(kept, existing)
			}
		}
		cfg.Workspaces[i].RepositoryIDs = kept
	}

	return s.store.Save(cfg)
}
