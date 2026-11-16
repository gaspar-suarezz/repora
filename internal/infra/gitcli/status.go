package gitcli

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/gaspar-suarezz/repora/internal/domain"
)

type entryKind int

const (
	kindStaged entryKind = iota
	kindUnstaged
	kindBoth
	kindConflict
)

type statusEntry struct {
	change domain.FileChange
	kind   entryKind
}

// ParseStatus parses the output of:
//
//	git -c core.quotepath=false status --porcelain=v2 --branch \
//	    --untracked-files=all -z
//
// The parser tolerates both NUL-terminated and LF-terminated records, because
// header termination has varied across Git versions.
func ParseStatus(repoID domain.RepositoryID, out []byte) domain.RepositoryState {
	state := domain.RepositoryState{
		RepositoryID: repoID,
		FetchedAt:    time.Now(),
	}

	var entries []statusEntry
	var untracked []string
	pendingRename := -1

	for _, chunk := range bytes.Split(out, []byte{0}) {
		for _, raw := range bytes.Split(chunk, []byte{'\n'}) {
			line := strings.TrimSuffix(string(raw), "\r")
			if line == "" {
				continue
			}

			if pendingRename >= 0 {
				entries[pendingRename].change.OrigPath = line
				pendingRename = -1
				continue
			}

			switch line[0] {
			case '#':
				parseBranchHeader(&state.Branch, line)
			case '1':
				if fc, ok := parseOrdinaryEntry(line); ok {
					entries = append(entries, statusEntry{change: fc, kind: kindOf(fc)})
				}
			case '2':
				if fc, ok := parseRenameEntry(line); ok {
					entries = append(entries, statusEntry{change: fc, kind: kindOf(fc)})
					pendingRename = len(entries) - 1
				}
			case 'u':
				if fc, ok := parseUnmergedEntry(line); ok {
					entries = append(entries, statusEntry{change: fc, kind: kindConflict})
				}
			case '?':
				if len(line) > 2 {
					untracked = append(untracked, line[2:])
				}
			case '!':
				// Ignored entries are never requested; skip defensively.
			}
		}
	}

	for _, e := range entries {
		switch e.kind {
		case kindStaged:
			state.WorkingTree.Staged = append(state.WorkingTree.Staged, e.change)
		case kindUnstaged:
			state.WorkingTree.Unstaged = append(state.WorkingTree.Unstaged, e.change)
		case kindBoth:
			state.WorkingTree.Staged = append(state.WorkingTree.Staged, e.change)
			state.WorkingTree.Unstaged = append(state.WorkingTree.Unstaged, e.change)
		case kindConflict:
			state.WorkingTree.Conflicts = append(state.WorkingTree.Conflicts, e.change)
		}
	}
	state.WorkingTree.Untracked = untracked

	return state
}

func parseBranchHeader(branch *domain.BranchState, line string) {
	fields := strings.SplitN(line, " ", 3)
	if len(fields) < 3 {
		return
	}
	key, value := fields[1], fields[2]

	switch key {
	case "branch.oid":
		if value == "(initial)" {
			branch.Unborn = true
			return
		}
		branch.OID = value
	case "branch.head":
		if value == "(detached)" {
			branch.Detached = true
			return
		}
		branch.Head = value
	case "branch.upstream":
		branch.Upstream = value
		branch.HasUpstream = true
	case "branch.ab":
		for _, field := range strings.Fields(value) {
			if len(field) < 2 {
				continue
			}
			n, err := strconv.Atoi(field[1:])
			if err != nil {
				continue
			}
			switch field[0] {
			case '+':
				branch.Ahead = n
			case '-':
				branch.Behind = n
			}
		}
	}
}

// parseOrdinaryEntry parses: 1 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <path>
func parseOrdinaryEntry(line string) (domain.FileChange, bool) {
	fields := strings.SplitN(line, " ", 9)
	if len(fields) < 9 {
		return domain.FileChange{}, false
	}
	return domain.FileChange{
		Index:    statusCode(fields[1], 0),
		Worktree: statusCode(fields[1], 1),
		Path:     fields[8],
	}, true
}

// parseRenameEntry parses: 2 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <X><score> <path>
// The original path follows as the next NUL-terminated field.
func parseRenameEntry(line string) (domain.FileChange, bool) {
	fields := strings.SplitN(line, " ", 10)
	if len(fields) < 10 {
		return domain.FileChange{}, false
	}
	return domain.FileChange{
		Index:    statusCode(fields[1], 0),
		Worktree: statusCode(fields[1], 1),
		Path:     fields[9],
	}, true
}

// parseUnmergedEntry parses:
// u <XY> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
func parseUnmergedEntry(line string) (domain.FileChange, bool) {
	fields := strings.SplitN(line, " ", 11)
	if len(fields) < 11 {
		return domain.FileChange{}, false
	}
	return domain.FileChange{
		Index:    statusCode(fields[1], 0),
		Worktree: statusCode(fields[1], 1),
		Path:     fields[10],
	}, true
}

func statusCode(xy string, position int) domain.StatusCode {
	if len(xy) <= position {
		return domain.CodeUnmodified
	}
	return domain.StatusCode(string(xy[position]))
}

func kindOf(change domain.FileChange) entryKind {
	hasIndex := change.Index != domain.CodeUnmodified && change.Index != ""
	hasWorktree := change.Worktree != domain.CodeUnmodified && change.Worktree != ""

	switch {
	case hasIndex && hasWorktree:
		return kindBoth
	case hasIndex:
		return kindStaged
	default:
		return kindUnstaged
	}
}
