package app

import (
	"context"
	"errors"
	"testing"

	"github.com/gaspar-suarezz/repora/internal/domain"
)

// rootGit resolves every path to a fixed repository root.
type rootGit struct {
	fakeGit
	root string
	err  error
}

func (g *rootGit) Root(context.Context, string) (string, error) {
	if g.err != nil {
		return "", g.err
	}
	return g.root, nil
}

func TestRegisterAddsCanonicalRepository(t *testing.T) {
	store := &memStore{}
	svc := NewRepositoryService(&rootGit{root: "/tmp/work/api"}, store)
	cfg := emptyConfig()

	id, err := svc.Register(context.Background(), cfg, "/tmp/work/api/subdir")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if id != "api" {
		t.Errorf("id = %q, want api", id)
	}
	if len(store.cfg.Repositories) != 1 {
		t.Fatalf("repositories = %+v", store.cfg.Repositories)
	}
	repo := store.cfg.Repositories[0]
	if repo.Path != "/tmp/work/api" {
		t.Errorf("path = %q, want canonical root", repo.Path)
	}
	if repo.Name != "api" {
		t.Errorf("name = %q", repo.Name)
	}
}

func TestRegisterRejectsDuplicatePath(t *testing.T) {
	store := &memStore{}
	svc := NewRepositoryService(&rootGit{root: "/tmp/work/api"}, store)
	cfg := &domain.Config{
		Version:      1,
		Repositories: []domain.Repository{{ID: "existing", Name: "API", Path: "/tmp/work/api"}},
	}

	_, err := svc.Register(context.Background(), cfg, "/tmp/work/api")
	if !errors.Is(err, domain.ErrDuplicatePath) {
		t.Fatalf("err = %v, want ErrDuplicatePath", err)
	}
}

func TestRegisterPropagatesGitErrors(t *testing.T) {
	store := &memStore{}
	svc := NewRepositoryService(&rootGit{err: domain.ErrNotARepository}, store)

	_, err := svc.Register(context.Background(), emptyConfig(), "/tmp/not-a-repo")
	if !errors.Is(err, domain.ErrNotARepository) {
		t.Fatalf("err = %v, want ErrNotARepository", err)
	}
	if store.saves != 0 {
		t.Errorf("store saved %d times after a failed registration", store.saves)
	}
}

func TestRegisterRejectsEmptyPath(t *testing.T) {
	svc := NewRepositoryService(&rootGit{root: "/tmp/api"}, &memStore{})
	if _, err := svc.Register(context.Background(), emptyConfig(), "   "); !errors.Is(err, domain.ErrInvalidPath) {
		t.Fatalf("err = %v, want ErrInvalidPath", err)
	}
}

func TestUnregisterRemovesFromWorkspaces(t *testing.T) {
	store := &memStore{}
	svc := NewRepositoryService(&rootGit{}, store)
	cfg := &domain.Config{
		Version:      1,
		Repositories: []domain.Repository{{ID: "api", Name: "API", Path: "/tmp/api"}, {ID: "web", Name: "Web", Path: "/tmp/web"}},
		Workspaces: []domain.Workspace{
			{ID: "work", Name: "Work", RepositoryIDs: []domain.RepositoryID{"api", "web"}},
			{ID: "personal", Name: "Personal", RepositoryIDs: []domain.RepositoryID{"api"}},
		},
	}

	if err := svc.Unregister(cfg, "api"); err != nil {
		t.Fatalf("Unregister: %v", err)
	}

	if len(store.cfg.Repositories) != 1 || store.cfg.Repositories[0].ID != "web" {
		t.Fatalf("repositories = %+v", store.cfg.Repositories)
	}
	for _, ws := range store.cfg.Workspaces {
		for _, id := range ws.RepositoryIDs {
			if id == "api" {
				t.Fatalf("workspace %q still references api", ws.ID)
			}
		}
	}
}

func TestUnregisterMissing(t *testing.T) {
	svc := NewRepositoryService(&rootGit{}, &memStore{})
	if err := svc.Unregister(emptyConfig(), "ghost"); !errors.Is(err, domain.ErrRepositoryNotFound) {
		t.Fatalf("err = %v, want ErrRepositoryNotFound", err)
	}
}
