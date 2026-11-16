package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gaspar-suarezz/repora/internal/app"
	"github.com/gaspar-suarezz/repora/internal/domain"
	"github.com/gaspar-suarezz/repora/internal/infra/config"
)

// TestAcceptanceScenario walks the v0.1 acceptance scenario end to end through
// the application layer: a workspace of four repositories, changed states,
// commit, branch, synchronization, and workspace deletion.
func TestAcceptanceScenario(t *testing.T) {
	adapter := newAdapter(t)

	root := t.TempDir()
	configPath := filepath.Join(root, "config", "config.yaml")
	store := config.NewFileStore(configPath)

	cfg, _, err := store.Load()
	if err != nil {
		t.Fatalf("initial config: %v", err)
	}
	if _, statErr := os.Stat(configPath); statErr != nil {
		t.Fatalf("default config was not created: %v", statErr)
	}

	workspaces := app.NewWorkspaceService(store)
	repositories := app.NewRepositoryService(adapter, store)
	gitService := app.NewGitService(adapter, app.NewLocks())
	statusService := app.NewStatusService(adapter, 4)

	// --- four independent repositories -------------------------------------

	names := []string{"api", "web", "infra", "shared"}
	paths := make(map[string]string, len(names))
	for _, name := range names {
		path := filepath.Join(root, "code", name)
		seedRepository(t, path)
		paths[name] = path
	}

	wsID, err := workspaces.Create(cfg, "Acceptance", "End-to-end acceptance workspace")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	repoIDs := make(map[string]domain.RepositoryID, len(names))
	for _, name := range names {
		id, err := repositories.Register(context.Background(), cfg, paths[name])
		if err != nil {
			t.Fatalf("register %s: %v", name, err)
		}
		if err := workspaces.Associate(cfg, wsID, id); err != nil {
			t.Fatalf("associate %s: %v", name, err)
		}
		repoIDs[name] = id
	}

	// --- configuration is durable ------------------------------------------

	cfg, _, err = store.Load()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(cfg.Repositories) != 4 {
		t.Fatalf("repositories = %d, want 4", len(cfg.Repositories))
	}
	ws, ok := cfg.WorkspaceByID(wsID)
	if !ok || len(ws.RepositoryIDs) != 4 {
		t.Fatalf("workspace not persisted: %+v", ws)
	}

	// --- distinct repository states ----------------------------------------

	// web: untracked file.
	writeFile(t, filepath.Join(paths["web"], "notes.md"), "notes\n")

	// infra: upstream configured, then one local commit (ahead).
	infraRemote := filepath.Join(root, "remote", "infra.git")
	if err := gitInitBare(t, infraRemote); err != nil {
		t.Fatalf("init bare remote: %v", err)
	}
	gitRun(t, paths["infra"], "remote", "add", "origin", infraRemote)
	if err := gitService.Push(context.Background(), mustRepo(t, cfg, repoIDs["infra"])); err != nil {
		t.Fatalf("initial push: %v", err)
	}
	writeFile(t, filepath.Join(paths["infra"], "change.txt"), "change\n")
	gitRun(t, paths["infra"], "add", ".")
	gitRun(t, paths["infra"], "commit", "-qm", "local change")

	// shared: merge conflict.
	conflict := filepath.Join(paths["shared"], "conflict.txt")
	writeFile(t, conflict, "base\n")
	gitRun(t, paths["shared"], "add", ".")
	gitRun(t, paths["shared"], "commit", "-qm", "base")
	gitRun(t, paths["shared"], "checkout", "-qb", "other")
	writeFile(t, conflict, "other\n")
	gitRun(t, paths["shared"], "commit", "-qam", "other")
	gitRun(t, paths["shared"], "checkout", "-q", "main")
	writeFile(t, conflict, "main\n")
	gitRun(t, paths["shared"], "commit", "-qam", "main")
	mergeExpectFailure(t, paths["shared"], "other")

	// --- refresh every repository ------------------------------------------

	ordered := make([]domain.Repository, 0, 4)
	for _, name := range names {
		repo, ok := cfg.RepositoryByID(repoIDs[name])
		if !ok {
			t.Fatalf("repository %s missing", name)
		}
		ordered = append(ordered, repo)
	}

	states := make(map[domain.RepositoryID]domain.RepositoryState, 4)
	var mu sync.Mutex
	statusService.Refresh(context.Background(), ordered, func(state domain.RepositoryState) {
		mu.Lock()
		states[state.RepositoryID] = state
		mu.Unlock()
	})
	if len(states) != 4 {
		t.Fatalf("received %d states, want 4", len(states))
	}

	expectClass := map[string]domain.StateClass{
		"api":    domain.ClassClean,
		"web":    domain.ClassDirty,
		"infra":  domain.ClassAhead,
		"shared": domain.ClassConflicts,
	}
	for _, name := range names {
		state := states[repoIDs[name]]
		if state.Err != nil {
			t.Fatalf("%s: %v", name, state.Err)
		}
		if got := state.Class(); got != expectClass[name] {
			t.Errorf("%s class = %q, want %q", name, got, expectClass[name])
		}
	}

	// --- stage and commit in one repository --------------------------------

	api := mustRepo(t, cfg, repoIDs["api"])
	writeFile(t, filepath.Join(paths["api"], "feature.txt"), "feature\n")
	if err := gitService.StageAll(context.Background(), api); err != nil {
		t.Fatalf("stage all: %v", err)
	}
	if err := gitService.Commit(context.Background(), api, "add acceptance feature"); err != nil {
		t.Fatalf("commit: %v", err)
	}
	commits, err := adapter.Log(context.Background(), api, 1)
	if err != nil || len(commits) != 1 {
		t.Fatalf("log: %v", err)
	}
	if commits[0].Subject != "add acceptance feature" {
		t.Errorf("subject = %q", commits[0].Subject)
	}

	// --- branch lifecycle ---------------------------------------------------

	if err := gitService.CreateBranch(context.Background(), api, "feature/acceptance"); err != nil {
		t.Fatalf("create branch: %v", err)
	}
	if state := adapter.Status(context.Background(), api); state.Branch.Head != "feature/acceptance" {
		t.Fatalf("head = %q", state.Branch.Head)
	}
	if err := gitService.Checkout(context.Background(), api, "main"); err != nil {
		t.Fatalf("checkout main: %v", err)
	}

	// --- synchronization ---------------------------------------------------

	infra := mustRepo(t, cfg, repoIDs["infra"])
	if err := gitService.Push(context.Background(), infra); err != nil {
		t.Fatalf("push: %v", err)
	}
	if state := adapter.Status(context.Background(), infra); state.Branch.Ahead != 0 {
		t.Errorf("ahead after push = %d, want 0", state.Branch.Ahead)
	}
	if err := gitService.Fetch(context.Background(), infra); err != nil {
		t.Fatalf("fetch: %v", err)
	}

	// --- deleting the workspace keeps repositories -------------------------

	if err := workspaces.Delete(cfg, wsID); err != nil {
		t.Fatalf("delete workspace: %v", err)
	}

	cfg, _, err = store.Load()
	if err != nil {
		t.Fatalf("final reload: %v", err)
	}
	if len(cfg.Workspaces) != 0 {
		t.Errorf("workspaces = %d, want 0", len(cfg.Workspaces))
	}
	if len(cfg.Repositories) != 4 {
		t.Errorf("repositories = %d, want 4", len(cfg.Repositories))
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("config mode = %o, want 600", got)
	}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if strings.Contains(string(raw), "name: Acceptance") {
		t.Errorf("deleted workspace still present in the file:\n%s", raw)
	}
}

func mustRepo(t *testing.T, cfg *domain.Config, id domain.RepositoryID) domain.Repository {
	t.Helper()
	repo, ok := cfg.RepositoryByID(id)
	if !ok {
		t.Fatalf("repository %q not found", id)
	}
	return repo
}

func mergeExpectFailure(t *testing.T, dir, branch string) {
	t.Helper()

	cmd := exec.Command("git",
		"-c", "user.name=Repora Test",
		"-c", "user.email=test@example.com",
		"merge", branch,
	)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
	)
	if err := cmd.Run(); err == nil {
		t.Fatalf("merge %s unexpectedly succeeded", branch)
	}
}
