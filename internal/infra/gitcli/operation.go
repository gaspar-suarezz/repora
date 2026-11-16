package gitcli

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/gaspar-suarezz/repora/internal/domain"
)

// Operation reports a multi-step Git operation that is in progress. Git blocks
// several commands while one is pending, so the UI must surface it.
func (r *Repository) Operation(ctx context.Context, path string) domain.Operation {
	ctx, cancel := context.WithTimeout(ctx, ReadTimeout)
	defer cancel()

	out, _, err := r.ex.Execute(ctx, path, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return domain.OperationNone
	}

	gitDir := strings.TrimSpace(string(out))
	if gitDir == "" {
		return domain.OperationNone
	}

	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(gitDir, name))
		return err == nil
	}

	switch {
	case exists("CHERRY_PICK_HEAD"):
		return domain.OperationCherryPick
	case exists("REVERT_HEAD"):
		return domain.OperationRevert
	case exists("MERGE_HEAD"):
		return domain.OperationMerge
	case exists("rebase-merge"), exists("rebase-apply"):
		return domain.OperationRebase
	case exists(filepath.Join("sequencer", "todo")):
		// A leftover sequence blocks switching branches, and Git reports it as a
		// cherry-pick even when no CHERRY_PICK_HEAD remains.
		return domain.OperationCherryPick
	default:
		return domain.OperationNone
	}
}
