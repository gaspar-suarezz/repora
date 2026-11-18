package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/gasparsuarez/repora/internal/tui/theme"
)

func TestTagsFitWithinWidth(t *testing.T) {
	th := theme.New(true)
	names := []string{"API", "Web", "Infrastructure", "Shared", "Dotfiles"}

	out := Tags(th, names, 40)
	if lipgloss.Width(out) > 40 {
		t.Fatalf("tags overflow: %q (%d)", out, lipgloss.Width(out))
	}
	if !strings.Contains(out, "+") {
		t.Errorf("overflow was not indicated: %q", out)
	}
}

func TestTagsRenderAllWhenSpaceAllows(t *testing.T) {
	th := theme.New(true)
	out := Tags(th, []string{"API", "Web"}, 80)
	if !strings.Contains(out, "[API]") || !strings.Contains(out, "[Web]") {
		t.Errorf("tags = %q", out)
	}
	if strings.Contains(out, "+") {
		t.Errorf("unexpected overflow marker: %q", out)
	}
}

func TestFooterDropsItemsOnNarrowTerminals(t *testing.T) {
	th := theme.New(true)
	items := []FooterItem{
		{Key: "↑↓", Label: "Navigate"},
		{Key: "↵", Label: "Open"},
		{Key: "n", Label: "New"},
		{Key: "e", Label: "Rename"},
		{Key: "d", Label: "Delete"},
		{Key: "r", Label: "Refresh"},
		{Key: "?", Label: "Help"},
	}

	wide := Footer(th, 120, items)
	if !strings.Contains(wide, "Help") {
		t.Errorf("wide footer dropped items: %q", wide)
	}

	narrow := Footer(th, 40, items)
	if lipgloss.Width(narrow) > 40 {
		t.Errorf("narrow footer overflows: %q (%d)", narrow, lipgloss.Width(narrow))
	}
	if strings.Contains(narrow, "Help") {
		t.Errorf("narrow footer should drop trailing items: %q", narrow)
	}
}

func TestCardHasFixedHeight(t *testing.T) {
	th := theme.New(true)

	full := Card(th, CardData{
		Name: "Product", Description: "Description", RepoCount: 2,
		Tags: []string{"API", "Web"}, Width: 60, Selected: true,
		Status: []Count{{Glyph: "✓", Label: "2 clean", Status: "success"}},
	})
	empty := Card(th, CardData{Name: "Empty", RepoCount: 0, Width: 60})

	if got := strings.Count(full, "\n") + 1; got != CardHeight {
		t.Errorf("full card height = %d, want %d", got, CardHeight)
	}
	if got := strings.Count(empty, "\n") + 1; got != CardHeight {
		t.Errorf("empty card height = %d, want %d", got, CardHeight)
	}
}

func TestCardWidthMatchesRequest(t *testing.T) {
	th := theme.New(true)
	for _, width := range []int{60, 80, 120} {
		card := Card(th, CardData{Name: "Product", RepoCount: 1, Tags: []string{"API"}, Width: width})
		for _, line := range strings.Split(card, "\n") {
			if got := lipgloss.Width(line); got != width {
				t.Errorf("width %d: line width = %d (%q)", width, got, line)
			}
		}
	}
}

func TestCardSelectedUsesDistinctMarker(t *testing.T) {
	th := theme.New(true)

	selected := Card(th, CardData{Name: "Product", Width: 60, Selected: true})
	unselected := Card(th, CardData{Name: "Product", Width: 60})

	if !strings.Contains(selected, "◆") {
		t.Error("selected card is missing the ◆ marker")
	}
	if !strings.Contains(unselected, "◇") {
		t.Error("unselected card is missing the ◇ marker")
	}
}

func TestStatusLineDropsOverflow(t *testing.T) {
	th := theme.New(true)
	counts := []Count{
		{Glyph: "✖", Label: "2 conflicts", Status: "danger"},
		{Glyph: "●", Label: "3 modified", Status: "warning"},
		{Glyph: "✓", Label: "5 clean", Status: "success"},
	}

	wide := StatusLine(th, counts, 120)
	for _, want := range []string{"2 conflicts", "3 modified", "5 clean"} {
		if !strings.Contains(wide, want) {
			t.Errorf("wide status line missing %q: %q", want, wide)
		}
	}

	narrow := StatusLine(th, counts, 18)
	if lipgloss.Width(narrow) > 18 {
		t.Errorf("narrow status line overflows: %q", narrow)
	}
	if !strings.Contains(narrow, "conflicts") {
		t.Errorf("narrow status line should keep the highest-priority entry: %q", narrow)
	}
}

func TestHeaderShowsWarnings(t *testing.T) {
	th := theme.New(true)
	out := Header(th, HeaderData{Width: 80, Warnings: 2})

	for _, want := range []string{"Repora", "WORKSPACE MANAGER", "2 warnings"} {
		if !strings.Contains(out, want) {
			t.Errorf("header missing %q: %q", want, out)
		}
	}
	if got := lipgloss.Width(out); got > 80 {
		t.Errorf("header overflows: %d", got)
	}
}

func TestBreadcrumbTruncatesLongSegments(t *testing.T) {
	th := theme.New(true)
	out := Breadcrumb(th, []string{"Workspaces", "A very long workspace name that would overflow the terminal"}, 60, "")

	if lipgloss.Width(out) > 60 {
		t.Fatalf("breadcrumb overflows: %q (%d)", out, lipgloss.Width(out))
	}
	if !strings.Contains(out, "…") {
		t.Errorf("long segment was not truncated: %q", out)
	}
	if !strings.Contains(out, "WORKSPACES") {
		t.Errorf("missing ancestors: %q", out)
	}
}

func TestHeaderShowsVersion(t *testing.T) {
	th := theme.New(true)
	out := Header(th, HeaderData{Width: 80, Version: "v0.1.0"})

	if !strings.Contains(out, "v0.1.0") {
		t.Errorf("header missing version: %q", out)
	}
	if got := lipgloss.Width(out); got > 80 {
		t.Errorf("header overflows: %d", got)
	}
}

func TestHeaderShowsWarningsAndVersion(t *testing.T) {
	th := theme.New(true)
	out := Header(th, HeaderData{Width: 100, Warnings: 2, Version: "v0.1.0"})

	for _, want := range []string{"2 warnings", "v0.1.0"} {
		if !strings.Contains(out, want) {
			t.Errorf("header missing %q: %q", want, out)
		}
	}
	if got := lipgloss.Width(out); got > 100 {
		t.Errorf("header overflows: %d", got)
	}
}
