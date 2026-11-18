package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gasparsuarez/repora/internal/domain"
	"github.com/gasparsuarez/repora/internal/tui/components"
)

// aggregate summarizes repository states for one workspace.
type aggregate struct {
	clean     int
	modified  int
	conflicts int
	sync      int
	errors    int

	known int
	total int
}

func (a aggregate) counts() []components.Count {
	counts := make([]components.Count, 0, 5)

	if a.conflicts > 0 {
		counts = append(counts, components.Count{Glyph: "✖", Label: plural(a.conflicts, "conflict", "conflicts"), Status: "danger"})
	}
	if a.modified > 0 {
		counts = append(counts, components.Count{Glyph: "●", Label: plural(a.modified, "modified", "modified"), Status: "warning"})
	}
	if a.sync > 0 {
		counts = append(counts, components.Count{Glyph: "↑", Label: plural(a.sync, "sync", "sync"), Status: "info"})
	}
	if a.errors > 0 {
		counts = append(counts, components.Count{Glyph: "!", Label: plural(a.errors, "error", "errors"), Status: "danger"})
	}
	if a.clean > 0 {
		counts = append(counts, components.Count{Glyph: "✓", Label: plural(a.clean, "clean", "clean"), Status: "success"})
	}
	return counts
}

// aggregateFor summarizes a workspace from the states currently cached for the
// home screen. Repositories without a state yet are not counted.
func (m *Model) aggregateFor(ws domain.Workspace) aggregate {
	agg := aggregate{total: len(ws.RepositoryIDs)}

	for _, id := range ws.RepositoryIDs {
		state, ok := m.homeStates[id]
		if !ok {
			continue
		}
		agg.known++

		switch state.Class() {
		case domain.ClassConflicts:
			agg.conflicts++
		case domain.ClassDiverged, domain.ClassBehind, domain.ClassAhead:
			agg.sync++
		case domain.ClassDirty:
			agg.modified++
		case domain.ClassError:
			agg.errors++
		default:
			agg.clean++
		}
	}

	return agg
}

// allRepositories returns every repository referenced by a workspace, once.
func (m *Model) allRepositories() []domain.Repository {
	seen := make(map[domain.RepositoryID]struct{})
	repos := make([]domain.Repository, 0)

	for _, ws := range m.cfg.Workspaces {
		for _, id := range ws.RepositoryIDs {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			if repo, ok := m.cfg.RepositoryByID(id); ok {
				repos = append(repos, repo)
			}
		}
	}

	return repos
}

// loadHomeSummaries refreshes the status of every repository referenced by a
// workspace so the explorer can aggregate real data. It is read-only and
// bounded by the status service worker pool.
func (m *Model) loadHomeSummaries() tea.Cmd {
	if m.status == nil {
		m.homeLoaded = true
		m.homeLoading = false
		return nil
	}

	repos := m.allRepositories()
	if len(repos) == 0 {
		m.homeLoaded = true
		m.homeLoading = false
		return nil
	}

	if m.homeCancel != nil {
		m.homeCancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.homeCancel = cancel

	m.homeGen++
	gen := m.homeGen
	m.homeLoading = true
	m.homeLoaded = true

	ch := make(chan tea.Msg)
	m.homeCh = ch
	svc := m.status

	start := func() tea.Msg {
		go func() {
			svc.Refresh(ctx, repos, func(state domain.RepositoryState) {
				select {
				case ch <- homeStatusMsg{gen: gen, state: state}:
				case <-ctx.Done():
				}
			})
			select {
			case ch <- homeDoneMsg{gen: gen}:
			case <-ctx.Done():
			}
			close(ch)
		}()
		return nil
	}

	return tea.Batch(start, m.waitHome())
}

func (m *Model) waitHome() tea.Cmd {
	if m.homeCh == nil {
		return nil
	}
	ch := m.homeCh
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

// reloadHome discards cached states and refreshes them.
func (m *Model) reloadHome() tea.Cmd {
	m.homeStates = make(map[domain.RepositoryID]domain.RepositoryState)
	m.homeLoaded = false
	return m.loadHomeSummaries()
}

// mergeRowsIntoHome caches the states already loaded by the workspace
// dashboard so returning home does not trigger a second round of Git calls.
func (m *Model) mergeRowsIntoHome() {
	if m.homeStates == nil {
		m.homeStates = make(map[domain.RepositoryID]domain.RepositoryState)
	}
	for _, row := range m.rows {
		if row.loading || row.unavailable {
			continue
		}
		m.homeStates[row.repo.ID] = row.state
	}
}

// homeWindow returns the slice of workspace cards that fits in avail lines.
func (m *Model) homeWindow(total, avail int) (start, end int) {
	if total == 0 {
		return 0, 0
	}

	perCard := components.CardHeight + 1
	visible := max(1, (avail+1)/perCard)
	if visible > total {
		visible = total
	}

	start = m.homeOffset
	if start > total-visible {
		start = total - visible
	}
	if start < 0 {
		start = 0
	}
	if m.wsIndex < start {
		start = m.wsIndex
	}
	if m.wsIndex >= start+visible {
		start = m.wsIndex - visible + 1
	}
	if start > total-visible {
		start = max(0, total-visible)
	}

	end = min(total, start+visible)
	return start, end
}

// homeAvail is the number of lines available for cards.
func (m *Model) homeAvail() int {
	avail := m.height - 9 // top margin, gap, status, footer, header, rule, section
	if avail < components.CardHeight {
		avail = components.CardHeight
	}
	return avail
}

// syncHomeWindow stores the first visible card so navigation can scroll.
func (m *Model) syncHomeWindow() {
	start, _ := m.homeWindow(len(m.cfg.Workspaces), m.homeAvail())
	m.homeOffset = start
}
