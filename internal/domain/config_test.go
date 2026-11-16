package domain

import (
	"errors"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	base := func() *Config {
		return &Config{
			Version: 1,
			Repositories: []Repository{
				{ID: "api", Name: "API", Path: "/tmp/api"},
				{ID: "web", Name: "Web", Path: "/tmp/web"},
			},
			Workspaces: []Workspace{
				{ID: "product", Name: "Product", RepositoryIDs: []RepositoryID{"api", "web"}},
			},
		}
	}

	t.Run("valid", func(t *testing.T) {
		if err := base().Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("defaults version", func(t *testing.T) {
		cfg := base()
		cfg.Version = 0
		if err := cfg.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Version != CurrentConfigVersion {
			t.Fatalf("version = %d, want %d", cfg.Version, CurrentConfigVersion)
		}
	})

	t.Run("unsupported version", func(t *testing.T) {
		cfg := base()
		cfg.Version = 99
		if err := cfg.Validate(); !errors.Is(err, ErrUnsupportedVersion) {
			t.Fatalf("err = %v, want ErrUnsupportedVersion", err)
		}
	})

	t.Run("duplicate repository id", func(t *testing.T) {
		cfg := base()
		cfg.Repositories[1].ID = "api"
		if err := cfg.Validate(); !errors.Is(err, ErrDuplicateID) {
			t.Fatalf("err = %v, want ErrDuplicateID", err)
		}
	})

	t.Run("duplicate path", func(t *testing.T) {
		cfg := base()
		cfg.Repositories[1].Path = "/tmp/api"
		if err := cfg.Validate(); !errors.Is(err, ErrDuplicatePath) {
			t.Fatalf("err = %v, want ErrDuplicatePath", err)
		}
	})

	t.Run("dangling reference", func(t *testing.T) {
		cfg := base()
		cfg.Workspaces[0].RepositoryIDs = append(cfg.Workspaces[0].RepositoryIDs, "ghost")
		if err := cfg.Validate(); !errors.Is(err, ErrDanglingReference) {
			t.Fatalf("err = %v, want ErrDanglingReference", err)
		}
	})

	t.Run("duplicate association", func(t *testing.T) {
		cfg := base()
		cfg.Workspaces[0].RepositoryIDs = []RepositoryID{"api", "api"}
		if err := cfg.Validate(); !errors.Is(err, ErrDuplicateID) {
			t.Fatalf("err = %v, want ErrDuplicateID", err)
		}
	})

	t.Run("default workspace dangling", func(t *testing.T) {
		cfg := base()
		cfg.Preferences.DefaultWorkspace = "ghost"
		if err := cfg.Validate(); !errors.Is(err, ErrDanglingReference) {
			t.Fatalf("err = %v, want ErrDanglingReference", err)
		}
	})

	t.Run("repository shared across workspaces", func(t *testing.T) {
		cfg := base()
		cfg.Workspaces = append(cfg.Workspaces, Workspace{
			ID:            "personal",
			Name:          "Personal",
			RepositoryIDs: []RepositoryID{"api"},
		})
		if err := cfg.Validate(); err != nil {
			t.Fatalf("a repository must be allowed in several workspaces: %v", err)
		}
	})

	t.Run("empty workspace allowed", func(t *testing.T) {
		cfg := base()
		cfg.Workspaces = append(cfg.Workspaces, Workspace{ID: "scratch", Name: "Scratch"})
		if err := cfg.Validate(); err != nil {
			t.Fatalf("an empty workspace must be allowed: %v", err)
		}
	})
}

func TestConfigRepositoriesOf(t *testing.T) {
	cfg := &Config{
		Version: 1,
		Repositories: []Repository{
			{ID: "api", Name: "API", Path: "/tmp/api"},
			{ID: "web", Name: "Web", Path: "/tmp/web"},
			{ID: "infra", Name: "Infra", Path: "/tmp/infra"},
		},
		Workspaces: []Workspace{
			{ID: "w", Name: "W", RepositoryIDs: []RepositoryID{"web", "api"}},
		},
	}

	repos := cfg.RepositoriesOf(cfg.Workspaces[0])
	if len(repos) != 2 {
		t.Fatalf("len = %d, want 2", len(repos))
	}
	if repos[0].ID != "web" || repos[1].ID != "api" {
		t.Fatalf("order not preserved: %v", []RepositoryID{repos[0].ID, repos[1].ID})
	}
}

func TestPreferencesNormalized(t *testing.T) {
	got := Preferences{}.Normalized()
	if got.Theme != "dark" {
		t.Errorf("theme = %q, want dark", got.Theme)
	}
	if got.StatusWorkers != DefaultStatusWorkers {
		t.Errorf("workers = %d, want %d", got.StatusWorkers, DefaultStatusWorkers)
	}

	got = Preferences{StatusWorkers: 100, RefreshInterval: 1}.Normalized()
	if got.StatusWorkers != 16 {
		t.Errorf("workers = %d, want clamp to 16", got.StatusWorkers)
	}
	if got.RefreshInterval != 5 {
		t.Errorf("interval = %d, want clamp to 5", got.RefreshInterval)
	}
}

func TestRepoStateClass(t *testing.T) {
	tests := []struct {
		name  string
		state RepositoryState
		want  StateClass
	}{
		{"clean", RepositoryState{}, ClassClean},
		{"error wins", RepositoryState{Err: errors.New("boom"), Branch: BranchState{Ahead: 1}}, ClassError},
		{
			"conflicts win over divergence",
			RepositoryState{
				WorkingTree: WorkingTree{Conflicts: []FileChange{{Path: "a.go"}}},
				Branch:      BranchState{Ahead: 1, Behind: 1},
			},
			ClassConflicts,
		},
		{"diverged", RepositoryState{Branch: BranchState{Ahead: 1, Behind: 2}}, ClassDiverged},
		{"behind", RepositoryState{Branch: BranchState{Behind: 2}}, ClassBehind},
		{"ahead", RepositoryState{Branch: BranchState{Ahead: 3}}, ClassAhead},
		{
			"dirty",
			RepositoryState{WorkingTree: WorkingTree{Untracked: []string{"a.txt"}}},
			ClassDirty,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.state.Class(); got != tc.want {
				t.Fatalf("Class() = %q, want %q", got, tc.want)
			}
		})
	}
}
