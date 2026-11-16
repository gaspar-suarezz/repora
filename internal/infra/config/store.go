package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gaspar-suarezz/repora/internal/domain"
	"go.yaml.in/yaml/v3"
)

// DefaultYAML is written when no configuration exists.
const DefaultYAML = `# Repora configuration
# Documentation: docs/01-architecture/config-and-persistence.md
version: 1

repositories: []

workspaces: []

preferences:
  theme: dark
  refresh_interval: 30
  status_workers: 4
`

// Load reads and validates the configuration at path. If the file does not
// exist, a default configuration is created first.
//
// Non-fatal problems (for example a repository path that no longer exists) are
// returned as warnings and do not prevent the application from starting.
func Load(path string) (*domain.Config, []string, error) {
	cfg, warnings, err := read(path)
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, domain.ErrConfigNotFound) {
		if createErr := CreateDefault(path); createErr != nil {
			return nil, nil, createErr
		}
		return read(path)
	}
	return cfg, warnings, err
}

// CreateDefault writes the default configuration, creating parent directories.
func CreateDefault(path string) error {
	return WriteFileAtomic(path, []byte(DefaultYAML), 0o600)
}

func read(path string) (*domain.Config, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, fmt.Errorf("%w: %s", domain.ErrConfigNotFound, path)
		}
		return nil, nil, fmt.Errorf("reading configuration %s: %w", path, err)
	}

	var cfg domain.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, nil, fmt.Errorf("parsing configuration %s: %w", path, err)
	}

	if err := normalizeRepositoryPaths(&cfg); err != nil {
		return nil, nil, fmt.Errorf("configuration %s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, nil, fmt.Errorf("configuration %s: %w", path, err)
	}

	return &cfg, pathWarnings(&cfg), nil
}

func normalizeRepositoryPaths(cfg *domain.Config) error {
	for i := range cfg.Repositories {
		repo := &cfg.Repositories[i]
		expanded, err := expandTilde(repo.Path)
		if err != nil {
			return fmt.Errorf("repository %q: %w", repo.ID, err)
		}
		if !filepath.IsAbs(expanded) {
			return fmt.Errorf("repository %q: %w: %q must be absolute", repo.ID, domain.ErrInvalidPath, repo.Path)
		}
		repo.Path = filepath.Clean(expanded)
	}
	return nil
}

func pathWarnings(cfg *domain.Config) []string {
	var warnings []string
	for _, repo := range cfg.Repositories {
		info, err := os.Stat(repo.Path)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			warnings = append(warnings, fmt.Sprintf("repository %q: path does not exist: %s", repo.ID, repo.Path))
		case err != nil:
			warnings = append(warnings, fmt.Sprintf("repository %q: cannot inspect path: %v", repo.ID, err))
		case !info.IsDir():
			warnings = append(warnings, fmt.Sprintf("repository %q: path is not a directory: %s", repo.ID, repo.Path))
		}
	}
	return warnings
}
