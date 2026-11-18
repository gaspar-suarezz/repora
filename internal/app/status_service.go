package app

import (
	"context"
	"sync"

	"github.com/gasparsuarez/repora/internal/domain"
)

// StatusService refreshes repository state concurrently with a bounded worker
// pool.
type StatusService struct {
	git     domain.GitPort
	workers int
}

// NewStatusService returns a service using the configured worker count.
func NewStatusService(git domain.GitPort, workers int) *StatusService {
	if workers < 1 {
		workers = 1
	}
	if workers > 16 {
		workers = 16
	}
	return &StatusService{git: git, workers: workers}
}

// Refresh inspects every repository and calls emit once per repository, in
// completion order. It returns when all repositories have been processed or the
// context is cancelled.
//
// The caller owns delivery: emit is invoked from worker goroutines, so it must
// be safe for concurrent use.
func (s *StatusService) Refresh(ctx context.Context, repos []domain.Repository, emit func(domain.RepositoryState)) {
	if len(repos) == 0 {
		return
	}

	workers := s.workers
	if workers > len(repos) {
		workers = len(repos)
	}

	jobs := make(chan domain.Repository, len(repos))
	for _, repo := range repos {
		jobs <- repo
	}
	close(jobs)

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for repo := range jobs {
				if ctx.Err() != nil {
					return
				}
				state := s.git.Status(ctx, repo)
				if ctx.Err() != nil {
					return
				}
				emit(state)
			}
		}()
	}
	wg.Wait()
}

// Status returns the state of a single repository.
func (s *StatusService) Status(ctx context.Context, repo domain.Repository) domain.RepositoryState {
	return s.git.Status(ctx, repo)
}

// Workers reports the effective worker count.
func (s *StatusService) Workers() int { return s.workers }
