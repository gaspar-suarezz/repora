package app

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gaspar-suarezz/repora/internal/domain"
)

func TestLocksSerializePerRepository(t *testing.T) {
	locks := NewLocks()
	ctx := context.Background()

	release, err := locks.Acquire(ctx, "api")
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	if !locks.IsLocked("api") {
		t.Error("api should be locked")
	}
	if locks.IsLocked("web") {
		t.Error("web should not be locked")
	}

	acquired := make(chan struct{})
	go func() {
		release2, err := locks.Acquire(ctx, "api")
		if err != nil {
			t.Errorf("second Acquire: %v", err)
			return
		}
		close(acquired)
		release2()
	}()

	select {
	case <-acquired:
		t.Fatal("second lock acquired while the first was held")
	case <-time.After(50 * time.Millisecond):
	}

	release()

	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("second lock was never acquired after release")
	}

	if locks.IsLocked("api") {
		t.Error("api should be unlocked after release")
	}
}

func TestLocksDifferentRepositoriesDoNotBlock(t *testing.T) {
	locks := NewLocks()
	ctx := context.Background()

	releaseAPI, err := locks.Acquire(ctx, "api")
	if err != nil {
		t.Fatalf("Acquire api: %v", err)
	}
	defer releaseAPI()

	done := make(chan error, 1)
	go func() {
		releaseWeb, err := locks.Acquire(ctx, "web")
		if err == nil {
			releaseWeb()
		}
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Acquire web: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("independent repository lock blocked")
	}
}

func TestLocksAcquireHonorsCancellation(t *testing.T) {
	locks := NewLocks()
	release, err := locks.Acquire(context.Background(), "api")
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	defer release()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	_, err = locks.Acquire(ctx, "api")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("cancelled acquisition did not return promptly")
	}
}

func TestGitServiceSerializesSameRepository(t *testing.T) {
	writer := &recordingWriter{block: make(chan struct{}), started: make(chan struct{}, 2)}
	svc := NewGitService(writer, NewLocks())
	repo := domain.Repository{ID: "api", Name: "API", Path: "/tmp/api"}

	first := make(chan error, 1)
	go func() {
		first <- svc.Stage(context.Background(), repo, []string{"a.go"})
	}()
	<-writer.started

	second := make(chan error, 1)
	go func() {
		second <- svc.Stage(context.Background(), repo, []string{"b.go"})
	}()

	select {
	case <-writer.started:
		t.Fatal("second operation started before the first released the lock")
	case <-time.After(50 * time.Millisecond):
	}

	close(writer.block)

	if err := <-first; err != nil {
		t.Fatalf("first: %v", err)
	}
	if err := <-second; err != nil {
		t.Fatalf("second: %v", err)
	}

	if got := writer.count(); got != 2 {
		t.Errorf("calls = %d, want 2", got)
	}
}

func TestGitServiceValidatesInputsBeforeCallingGit(t *testing.T) {
	writer := &recordingWriter{}
	svc := NewGitService(writer, NewLocks())
	repo := domain.Repository{ID: "api", Name: "API", Path: "/tmp/api"}

	if err := svc.Commit(context.Background(), repo, "   "); !errors.Is(err, domain.ErrEmptyMessage) {
		t.Errorf("Commit error = %v, want ErrEmptyMessage", err)
	}
	if err := svc.CreateBranch(context.Background(), repo, "bad..name"); !errors.Is(err, domain.ErrInvalidBranch) {
		t.Errorf("CreateBranch error = %v, want ErrInvalidBranch", err)
	}
	if err := svc.Stage(context.Background(), repo, nil); !errors.Is(err, domain.ErrNothingToStage) {
		t.Errorf("Stage error = %v, want ErrNothingToStage", err)
	}
	if err := svc.Unstage(context.Background(), repo, nil); !errors.Is(err, domain.ErrNothingToStage) {
		t.Errorf("Unstage error = %v, want ErrNothingToStage", err)
	}
	if err := svc.DeleteBranch(context.Background(), repo, "bad..name"); !errors.Is(err, domain.ErrInvalidBranch) {
		t.Errorf("DeleteBranch error = %v, want ErrInvalidBranch", err)
	}
	if got := writer.count(); got != 0 {
		t.Errorf("writer was called %d times for invalid input", got)
	}
}

type recordingWriter struct {
	mu      sync.Mutex
	calls   int
	block   chan struct{}
	started chan struct{}
}

func (w *recordingWriter) enter() {
	if w.started != nil {
		w.started <- struct{}{}
	}
	if w.block != nil {
		<-w.block
	}
}

func (w *recordingWriter) count() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.calls
}

func (w *recordingWriter) record() {
	w.mu.Lock()
	w.calls++
	w.mu.Unlock()
}

func (w *recordingWriter) Stage(context.Context, domain.Repository, []string) error {
	w.record()
	w.enter()
	return nil
}
func (w *recordingWriter) StageAll(context.Context, domain.Repository) error {
	w.record()
	w.enter()
	return nil
}
func (w *recordingWriter) Unstage(context.Context, domain.Repository, []string) error {
	w.record()
	w.enter()
	return nil
}
func (w *recordingWriter) UnstageAll(context.Context, domain.Repository) error {
	w.record()
	w.enter()
	return nil
}
func (w *recordingWriter) Commit(context.Context, domain.Repository, string) error {
	w.record()
	w.enter()
	return nil
}
func (w *recordingWriter) CreateBranch(context.Context, domain.Repository, string, bool) error {
	w.record()
	w.enter()
	return nil
}
func (w *recordingWriter) Checkout(context.Context, domain.Repository, string) error {
	w.record()
	w.enter()
	return nil
}
func (w *recordingWriter) DeleteBranch(context.Context, domain.Repository, string, bool) error {
	w.record()
	w.enter()
	return nil
}
func (w *recordingWriter) Fetch(context.Context, domain.Repository) error {
	w.record()
	w.enter()
	return nil
}
func (w *recordingWriter) Pull(context.Context, domain.Repository) error {
	w.record()
	w.enter()
	return nil
}
func (w *recordingWriter) Push(context.Context, domain.Repository) error {
	w.record()
	w.enter()
	return nil
}
