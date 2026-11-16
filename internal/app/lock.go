package app

import (
	"context"
	"sync"

	"github.com/gaspar-suarezz/repora/internal/domain"
)

// Locks serializes mutating operations per repository.
type Locks struct {
	mu    sync.Mutex
	locks map[domain.RepositoryID]chan struct{}
}

// NewLocks returns an empty lock registry.
func NewLocks() *Locks {
	return &Locks{locks: make(map[domain.RepositoryID]chan struct{})}
}

// Acquire blocks until the repository lock is held or ctx is done. It returns
// the release function on success.
func (l *Locks) Acquire(ctx context.Context, id domain.RepositoryID) (func(), error) {
	l.mu.Lock()
	ch, ok := l.locks[id]
	if !ok {
		ch = make(chan struct{}, 1)
		l.locks[id] = ch
	}
	l.mu.Unlock()

	select {
	case ch <- struct{}{}:
		var once sync.Once
		return func() { once.Do(func() { <-ch }) }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// IsLocked reports whether an operation currently holds the repository lock.
func (l *Locks) IsLocked(id domain.RepositoryID) bool {
	l.mu.Lock()
	ch, ok := l.locks[id]
	l.mu.Unlock()
	if !ok {
		return false
	}
	return len(ch) == 1
}
