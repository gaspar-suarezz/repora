package app

import (
	"errors"
	"testing"

	"github.com/gaspar-suarezz/repora/internal/domain"
)

type memStore struct {
	cfg   *domain.Config
	err   error
	saves int
}

func (m *memStore) Load() (*domain.Config, []string, error) { return m.cfg, nil, nil }

func (m *memStore) Save(cfg *domain.Config) error {
	if m.err != nil {
		return m.err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	m.cfg = cfg.Clone()
	m.saves++
	return nil
}

func emptyConfig() *domain.Config {
	return &domain.Config{Version: 1}
}

func TestWorkspaceCreateDerivesSlug(t *testing.T) {
	store := &memStore{}
	svc := NewWorkspaceService(store)
	cfg := emptyConfig()

	id, err := svc.Create(cfg, "  Product X  ", "Customer-facing product")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != "product-x" {
		t.Errorf("id = %q, want product-x", id)
	}
	if len(store.cfg.Workspaces) != 1 || store.cfg.Workspaces[0].Name != "Product X" {
		t.Fatalf("stored = %+v", store.cfg.Workspaces)
	}
	if got := store.cfg.Workspaces[0].Description; got != "Customer-facing product" {
		t.Errorf("description = %q", got)
	}
}

func TestWorkspaceCreateAvoidsCollision(t *testing.T) {
	store := &memStore{}
	svc := NewWorkspaceService(store)
	cfg := &domain.Config{Version: 1, Workspaces: []domain.Workspace{{ID: "product-x", Name: "Product X"}}}

	id, err := svc.Create(cfg, "Product X", "Customer-facing product")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != "product-x-2" {
		t.Errorf("id = %q, want product-x-2", id)
	}
}

func TestWorkspaceCreateRejectsEmptyName(t *testing.T) {
	svc := NewWorkspaceService(&memStore{})
	if _, err := svc.Create(emptyConfig(), "   ", ""); !errors.Is(err, domain.ErrInvalidName) {
		t.Fatalf("err = %v, want ErrInvalidName", err)
	}
}

func TestWorkspaceCreatePropagatesSaveError(t *testing.T) {
	store := &memStore{err: errors.New("disk full")}
	svc := NewWorkspaceService(store)

	if _, err := svc.Create(emptyConfig(), "Work", ""); err == nil {
		t.Fatal("expected save error")
	}
}

func TestWorkspaceRenameKeepsID(t *testing.T) {
	store := &memStore{}
	svc := NewWorkspaceService(store)
	cfg := &domain.Config{Version: 1, Workspaces: []domain.Workspace{{ID: "work", Name: "Work"}}}

	if err := svc.Rename(cfg, "work", "Day Job", "Updated description"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if store.cfg.Workspaces[0].ID != "work" || store.cfg.Workspaces[0].Name != "Day Job" {
		t.Fatalf("stored = %+v", store.cfg.Workspaces[0])
	}
	if got := store.cfg.Workspaces[0].Description; got != "Updated description" {
		t.Errorf("description = %q", got)
	}
}

func TestWorkspaceRenameMissing(t *testing.T) {
	svc := NewWorkspaceService(&memStore{})
	err := svc.Rename(emptyConfig(), "ghost", "X", "")
	if !errors.Is(err, domain.ErrWorkspaceNotFound) {
		t.Fatalf("err = %v, want ErrWorkspaceNotFound", err)
	}
}

func TestWorkspaceDeleteClearsDefault(t *testing.T) {
	store := &memStore{}
	svc := NewWorkspaceService(store)
	cfg := &domain.Config{
		Version:     1,
		Workspaces:  []domain.Workspace{{ID: "work", Name: "Work"}},
		Preferences: domain.Preferences{DefaultWorkspace: "work"},
	}

	if err := svc.Delete(cfg, "work"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(store.cfg.Workspaces) != 0 {
		t.Errorf("workspaces = %+v", store.cfg.Workspaces)
	}
	if store.cfg.Preferences.DefaultWorkspace != "" {
		t.Errorf("default workspace = %q, want empty", store.cfg.Preferences.DefaultWorkspace)
	}
}

func TestWorkspaceAssociateRules(t *testing.T) {
	store := &memStore{}
	svc := NewWorkspaceService(store)
	cfg := &domain.Config{
		Version:      1,
		Repositories: []domain.Repository{{ID: "api", Name: "API", Path: "/tmp/api"}},
		Workspaces:   []domain.Workspace{{ID: "work", Name: "Work"}},
	}

	if err := svc.Associate(cfg, "work", "api"); err != nil {
		t.Fatalf("Associate: %v", err)
	}
	if got := store.cfg.Workspaces[0].RepositoryIDs; len(got) != 1 || got[0] != "api" {
		t.Fatalf("ids = %v", got)
	}

	if err := svc.Associate(cfg, "work", "api"); !errors.Is(err, domain.ErrDuplicateID) {
		t.Errorf("duplicate association err = %v", err)
	}
	if err := svc.Associate(cfg, "work", "ghost"); !errors.Is(err, domain.ErrRepositoryNotFound) {
		t.Errorf("unknown repository err = %v", err)
	}
	if err := svc.Associate(cfg, "ghost", "api"); !errors.Is(err, domain.ErrWorkspaceNotFound) {
		t.Errorf("unknown workspace err = %v", err)
	}

	if err := svc.Disassociate(cfg, "work", "api"); err != nil {
		t.Fatalf("Disassociate: %v", err)
	}
	if got := store.cfg.Workspaces[0].RepositoryIDs; len(got) != 0 {
		t.Fatalf("ids after removal = %v", got)
	}
}

func TestWorkspaceSharedRepositoryAcrossWorkspaces(t *testing.T) {
	store := &memStore{}
	svc := NewWorkspaceService(store)
	cfg := &domain.Config{
		Version:      1,
		Repositories: []domain.Repository{{ID: "api", Name: "API", Path: "/tmp/api"}},
		Workspaces: []domain.Workspace{
			{ID: "work", Name: "Work"},
			{ID: "personal", Name: "Personal"},
		},
	}

	if err := svc.Associate(cfg, "work", "api"); err != nil {
		t.Fatalf("Associate work: %v", err)
	}
	if err := svc.Associate(cfg, "personal", "api"); err != nil {
		t.Fatalf("Associate personal: %v", err)
	}

	if len(store.cfg.Workspaces[0].RepositoryIDs) != 1 || len(store.cfg.Workspaces[1].RepositoryIDs) != 1 {
		t.Fatalf("shared repository not associated with both: %+v", store.cfg.Workspaces)
	}
}
