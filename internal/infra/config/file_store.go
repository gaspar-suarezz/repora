package config

import "github.com/gasparsuarez/repora/internal/domain"

// FileStore implements domain.ConfigStore against a single YAML file.
type FileStore struct {
	Path string
}

// NewFileStore returns a store bound to path.
func NewFileStore(path string) *FileStore { return &FileStore{Path: path} }

// Load implements domain.ConfigStore.
func (s *FileStore) Load() (*domain.Config, []string, error) { return Load(s.Path) }

// Save implements domain.ConfigStore.
func (s *FileStore) Save(cfg *domain.Config) error { return Save(s.Path, cfg) }
