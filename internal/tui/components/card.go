package components

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/gasparsuarez/repora/internal/tui/theme"
)

// CardData configures a workspace card.
type CardData struct {
	Name        string
	Description string
	RepoCount   int
	Tags        []string
	Status      []Count
	Loading     bool
	Loaded      bool
	Selected    bool
	Width       int
}

// CardHeight is the fixed height of a card, in lines, including borders.
const CardHeight = 6

// Card renders a workspace card. All cards share the same height so the list
// does not jitter when moving the selection.
func Card(t theme.Theme, d CardData) string {
	// Lip Gloss Width includes padding but not the border, so the box is two
	// cells narrower than the card, and the usable text area loses one more
	// cell on each side to padding.
	box := d.Width - 2
	if box < 14 {
		box = 14
	}
	content := box - 2
	inner := content

	marker := t.Bar.Render("◇")
	nameStyle := t.Header
	if d.Selected {
		marker = t.Accent.Render("◆")
		nameStyle = t.Selected
	}

	name := truncate(d.Name, max(4, inner-14))
	count := t.Dim.Render(plural(d.RepoCount, "repo", "repos"))
	first := Spread(marker+" "+nameStyle.Render(name), count, inner)

	description := ""
	if d.Description != "" {
		description = t.Dim.Render(truncate(d.Description, inner))
	}

	third := Tags(t, d.Tags, inner)
	if len(d.Tags) == 0 {
		if d.RepoCount == 0 {
			third = t.Bar.Render("No repositories yet")
		} else {
			third = t.Bar.Render("…")
		}
	}

	fourth := cardStatus(t, d, inner)

	rendered := strings.Join([]string{first, description, third, fourth}, "\n")

	style := t.Card
	if d.Selected {
		style = t.CardSelected
	}
	return style.Width(box).Render(rendered)
}

func cardStatus(t theme.Theme, d CardData, width int) string {
	switch {
	case d.Loading:
		return t.Bar.Render("◐ loading status…")
	case !d.Loaded:
		return t.Bar.Render("◦ status not loaded")
	case d.RepoCount == 0:
		return t.Bar.Render("↵ open, then a to add")
	case len(d.Status) == 0:
		return t.Success.Render("✓ all clean")
	default:
		return StatusLine(t, d.Status, width)
	}
}

// Spread places right at the right edge of a line of the given width.
func Spread(left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		return left + " " + right
	}
	return left + strings.Repeat(" ", gap) + right
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	return string(runes[:width-1]) + "…"
}

// FooterItem is one contextual action.
type FooterItem struct {
	Key   string
	Label string
}

// Footer renders the contextual action bar, dropping actions that do not fit.
func Footer(t theme.Theme, width int, items []FooterItem) string {
	var parts []string
	used := 0

	for _, item := range items {
		part := t.Key.Render(item.Key) + " " + t.Label.Render(item.Label)
		needed := lipgloss.Width(part)
		if len(parts) > 0 {
			needed += 3
		}
		if width > 0 && used+needed > width-1 {
			break
		}
		parts = append(parts, part)
		used += needed
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "   ")
}

// ProgressHint describes scroll overflow for a list.
func ProgressHint(shown, total int) string {
	if total <= shown {
		return ""
	}
	return "↓ " + strconv.Itoa(total-shown) + " more"
}
