package gitcli

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gaspar-suarezz/repora/internal/domain"
)

const branchFormat = "%(refname)%00%(refname:short)%00%(HEAD)%00%(upstream:short)%00%(upstream:track)%00%(objectname:short)%00%(authorname)%00%(authordate:iso-strict)%00%(contents:subject)"

// Branches lists local and remote branches.
func (r *Repository) Branches(ctx context.Context, repo domain.Repository) ([]domain.Branch, error) {
	ctx, cancel := context.WithTimeout(ctx, ReadTimeout)
	defer cancel()

	stdout, stderr, err := r.ex.Execute(ctx, repo.Path,
		"for-each-ref", "--format="+branchFormat, "refs/heads", "refs/remotes",
	)
	if err != nil {
		return nil, fmt.Errorf("listing branches in %s: %w", repo.Path, exitSummary(err, stderr))
	}

	branches := ParseBranches(stdout)
	sortBranches(branches)
	return branches, nil
}

// ParseBranches parses for-each-ref output produced with branchFormat.
func ParseBranches(out []byte) []domain.Branch {
	var branches []domain.Branch

	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\x00")
		if len(fields) < 9 {
			continue
		}

		refname := fields[0]
		short := fields[1]
		if strings.HasSuffix(refname, "/HEAD") {
			continue
		}

		branch := domain.Branch{
			Name:      short,
			IsCurrent: strings.TrimSpace(fields[2]) == "*",
			IsRemote:  strings.HasPrefix(refname, "refs/remotes/"),
			Upstream:  fields[3],
		}
		branch.HasUpstream = branch.Upstream != ""
		branch.Gone = strings.Contains(fields[4], "gone")
		parseTrack(fields[4], &branch)

		branch.LastCommit = domain.Commit{
			Short:   fields[5],
			Author:  fields[6],
			Date:    parseISODate(fields[7]),
			Subject: fields[8],
		}
		branches = append(branches, branch)
	}

	return branches
}

func parseTrack(track string, branch *domain.Branch) {
	track = strings.TrimSpace(track)
	if track == "" {
		return
	}
	track = strings.TrimPrefix(track, "[")
	track = strings.TrimSuffix(track, "]")

	for _, part := range strings.Split(track, ",") {
		part = strings.TrimSpace(part)
		switch {
		case strings.HasPrefix(part, "ahead "):
			branch.Ahead = atoiSafe(strings.TrimPrefix(part, "ahead "))
		case strings.HasPrefix(part, "behind "):
			branch.Behind = atoiSafe(strings.TrimPrefix(part, "behind "))
		case part == "gone":
			branch.Gone = true
		}
	}
}

func atoiSafe(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return n
}

func sortBranches(branches []domain.Branch) {
	// Stable partition: current first, then local, then remote, each alphabetical.
	rank := func(b domain.Branch) int {
		switch {
		case b.IsCurrent:
			return 0
		case !b.IsRemote:
			return 1
		default:
			return 2
		}
	}

	for i := 1; i < len(branches); i++ {
		for j := i; j > 0; j-- {
			if lessBranch(branches[j], branches[j-1], rank) {
				branches[j], branches[j-1] = branches[j-1], branches[j]
				continue
			}
			break
		}
	}
}

func lessBranch(a, b domain.Branch, rank func(domain.Branch) int) bool {
	ra, rb := rank(a), rank(b)
	if ra != rb {
		return ra < rb
	}
	return a.Name < b.Name
}

func parseISODate(value string) time.Time {
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}
	}
	return t
}
