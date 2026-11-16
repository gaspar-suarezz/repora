package app

import (
	"fmt"
	"strings"

	"github.com/gaspar-suarezz/repora/internal/domain"
)

// WorkspaceService manages workspaces and their repository associations.
// Every mutation operates on a caller-owned Config copy and persists it.
type WorkspaceService struct {
	store domain.ConfigStore
}

// NewWorkspaceService returns a service backed by store.
func NewWorkspaceService(store domain.ConfigStore) *WorkspaceService {
	return &WorkspaceService{store: store}
}

// Create adds a workspace derived from name and returns its new identifier.
// description is optional.
func (s *WorkspaceService) Create(cfg *domain.Config, name, description string) (domain.WorkspaceID, error) {
	name = strings.TrimSpace(name)
	if err := domain.ValidateName(name); err != nil {
		return "", err
	}

	id := domain.WorkspaceID(uniqueID(domain.NormalizeID(name), func(candidate string) bool {
		_, exists := cfg.WorkspaceByID(domain.WorkspaceID(candidate))
		return exists
	}))
	if id == "" {
		return "", fmt.Errorf("%w: name produces an empty id", domain.ErrInvalidName)
	}

	cfg.Workspaces = append(cfg.Workspaces, domain.Workspace{
		ID:          id,
		Name:        name,
		Description: strings.TrimSpace(description),
	})

	if err := s.store.Save(cfg); err != nil {
		return "", err
	}
	return id, nil
}

// Rename changes the display name and description of a workspace, keeping its
// identifier. description may be empty.
func (s *WorkspaceService) Rename(cfg *domain.Config, id domain.WorkspaceID, name, description string) error {
	name = strings.TrimSpace(name)
	if err := domain.ValidateName(name); err != nil {
		return err
	}

	for i := range cfg.Workspaces {
		if cfg.Workspaces[i].ID != id {
			continue
		}
		cfg.Workspaces[i].Name = name
		cfg.Workspaces[i].Description = strings.TrimSpace(description)
		return s.store.Save(cfg)
	}
	return fmt.Errorf("%w: %q", domain.ErrWorkspaceNotFound, id)
}

// Delete removes a workspace. Repositories are never deleted.
func (s *WorkspaceService) Delete(cfg *domain.Config, id domain.WorkspaceID) error {
	index := -1
	for i := range cfg.Workspaces {
		if cfg.Workspaces[i].ID == id {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("%w: %q", domain.ErrWorkspaceNotFound, id)
	}

	cfg.Workspaces = append(cfg.Workspaces[:index], cfg.Workspaces[index+1:]...)
	if cfg.Preferences.DefaultWorkspace == id {
		cfg.Preferences.DefaultWorkspace = ""
	}
	return s.store.Save(cfg)
}

// Associate adds a registered repository to a workspace.
func (s *WorkspaceService) Associate(cfg *domain.Config, wsID domain.WorkspaceID, repoID domain.RepositoryID) error {
	if _, ok := cfg.RepositoryByID(repoID); !ok {
		return fmt.Errorf("%w: %q", domain.ErrRepositoryNotFound, repoID)
	}

	for i := range cfg.Workspaces {
		if cfg.Workspaces[i].ID != wsID {
			continue
		}
		for _, existing := range cfg.Workspaces[i].RepositoryIDs {
			if existing == repoID {
				return fmt.Errorf("%w: repository %q already in workspace %q", domain.ErrDuplicateID, repoID, wsID)
			}
		}
		cfg.Workspaces[i].RepositoryIDs = append(cfg.Workspaces[i].RepositoryIDs, repoID)
		return s.store.Save(cfg)
	}
	return fmt.Errorf("%w: %q", domain.ErrWorkspaceNotFound, wsID)
}

// Disassociate removes a repository from a workspace, keeping its registration.
func (s *WorkspaceService) Disassociate(cfg *domain.Config, wsID domain.WorkspaceID, repoID domain.RepositoryID) error {
	for i := range cfg.Workspaces {
		if cfg.Workspaces[i].ID != wsID {
			continue
		}
		ids := cfg.Workspaces[i].RepositoryIDs
		for j, existing := range ids {
			if existing != repoID {
				continue
			}
			cfg.Workspaces[i].RepositoryIDs = append(ids[:j], ids[j+1:]...)
			return s.store.Save(cfg)
		}
		return fmt.Errorf("%w: repository %q is not in workspace %q", domain.ErrDanglingReference, repoID, wsID)
	}
	return fmt.Errorf("%w: %q", domain.ErrWorkspaceNotFound, wsID)
}

// uniqueID returns base, or base-2, base-3, ... until taken is false.
func uniqueID(base string, taken func(string) bool) string {
	if base == "" {
		return ""
	}
	if !taken(base) {
		return base
	}
	for i := 2; i < 100; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !taken(candidate) {
			return candidate
		}
	}
	return base
}
