package domain

import "time"

// StatusCode is a single character from Git's porcelain v2 alphabet.
type StatusCode string

// Git status codes.
const (
	CodeUnmodified StatusCode = "."
	CodeModified   StatusCode = "M"
	CodeTypeChange StatusCode = "T"
	CodeAdded      StatusCode = "A"
	CodeDeleted    StatusCode = "D"
	CodeRenamed    StatusCode = "R"
	CodeCopied     StatusCode = "C"
	CodeUnmerged   StatusCode = "U"
)

// FileChange is a single changed path.
//
// A file can be changed on both the index and worktree sides; in that case it
// appears in both Staged and Unstaged with the same Path.
type FileChange struct {
	Path     string
	OrigPath string
	Index    StatusCode
	Worktree StatusCode
}

// WorkingTree is the set of changes in a repository.
type WorkingTree struct {
	Staged    []FileChange
	Unstaged  []FileChange
	Untracked []string
	Conflicts []FileChange
}

// Dirty reports whether the working tree has any change.
func (w WorkingTree) Dirty() bool {
	return len(w.Staged) > 0 || len(w.Unstaged) > 0 || len(w.Untracked) > 0 || len(w.Conflicts) > 0
}

// BranchState is the state of HEAD and its upstream.
type BranchState struct {
	Head        string
	Detached    bool
	Unborn      bool
	OID         string
	Upstream    string
	HasUpstream bool
	Ahead       int
	Behind      int
}

// Label returns the branch name to display, falling back to synthetic labels.
func (b BranchState) Label() string {
	switch {
	case b.Unborn:
		return "(no commits)"
	case b.Detached:
		return "HEAD (detached)"
	case b.Head == "":
		return "(unknown)"
	default:
		return b.Head
	}
}

// RepositoryState is a snapshot of Git-derived information. It is never
// persisted.
type RepositoryState struct {
	RepositoryID RepositoryID
	FetchedAt    time.Time
	Err          error

	Branch      BranchState
	WorkingTree WorkingTree
}

// StateClass is the dashboard classification of a repository.
type StateClass string

// Repository classifications, ordered by precedence.
const (
	ClassError       StateClass = "error"
	ClassConflicts   StateClass = "conflicts"
	ClassDiverged    StateClass = "diverged"
	ClassBehind      StateClass = "behind"
	ClassAhead       StateClass = "ahead"
	ClassDirty       StateClass = "modified"
	ClassClean       StateClass = "clean"
	ClassUnavailable StateClass = "unavailable"
)

// Class returns the classification with the highest precedence.
//
// Unavailable is not derived here: it is a property of the registration, not of
// Git output, so callers set it explicitly.
func (s RepositoryState) Class() StateClass {
	switch {
	case s.Err != nil:
		return ClassError
	case len(s.WorkingTree.Conflicts) > 0:
		return ClassConflicts
	case s.Branch.Ahead > 0 && s.Branch.Behind > 0:
		return ClassDiverged
	case s.Branch.Behind > 0:
		return ClassBehind
	case s.Branch.Ahead > 0:
		return ClassAhead
	case s.WorkingTree.Dirty():
		return ClassDirty
	default:
		return ClassClean
	}
}

// IsClean reports whether the repository has no changes and no divergence.
func (s RepositoryState) IsClean() bool { return s.Class() == ClassClean }

// StagedCount returns the number of staged entries.
func (s RepositoryState) StagedCount() int { return len(s.WorkingTree.Staged) }

// UnstagedCount returns the number of unstaged entries.
func (s RepositoryState) UnstagedCount() int { return len(s.WorkingTree.Unstaged) }

// UntrackedCount returns the number of untracked files.
func (s RepositoryState) UntrackedCount() int { return len(s.WorkingTree.Untracked) }

// ConflictCount returns the number of unmerged entries.
func (s RepositoryState) ConflictCount() int { return len(s.WorkingTree.Conflicts) }
