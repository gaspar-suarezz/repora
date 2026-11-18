// Package components contains stateless renderers shared by Repora screens.
//
// Components receive data and a theme; they never touch application services,
// Git, or model state.
package components

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/gasparsuarez/repora/internal/tui/theme"
)

// Symbol is the Repora mark.
const Symbol = "◈"

// HeaderData configures the application header.
type HeaderData struct {
	Width    int
	Warnings int
	Version  string
}

// Header renders the compact brand header with optional warnings and version.
func Header(t theme.Theme, d HeaderData) string {
	left := t.Accent.Render(Symbol) + "  " + t.Title.Render("Repora") +
		"   " + t.Bar.Render("WORKSPACE MANAGER")

	var right []string
	if d.Warnings > 0 {
		right = append(right, t.Warning.Render("!")+" "+
			t.Dim.Render(plural(d.Warnings, "warning", "warnings")))
	}
	if d.Version != "" {
		right = append(right, t.Bar.Render(truncate(d.Version, 24)))
	}

	if len(right) == 0 {
		return left
	}

	badges := strings.Join(right, "   ")
	if d.Width <= 0 {
		return left + "   " + badges
	}
	return Spread(left, badges, d.Width)
}

// Tags renders repository names as compact chips, dropping overflow into +N.
func Tags(t theme.Theme, names []string, width int) string {
	if len(names) == 0 || width <= 0 {
		return ""
	}

	var parts []string
	used := 0

	for i, name := range names {
		chip := "[" + name + "]"
		needed := lipgloss.Width(chip)
		if len(parts) > 0 {
			needed++ // separating space
		}

		if used+needed > width {
			suffix := "+" + strconv.Itoa(len(names)-i)
			suffixWidth := lipgloss.Width(suffix)
			if len(parts) > 0 {
				suffixWidth++
			}
			if used+suffixWidth <= width {
				parts = append(parts, t.Chip.Render(suffix))
			}
			break
		}

		parts = append(parts, t.Chip.Render(chip))
		used += needed
	}

	return strings.Join(parts, " ")
}

// Count is one aggregated status entry.
type Count struct {
	Glyph  string
	Label  string
	Status string // success | warning | danger | info
}

// StatusLine renders aggregated counts, dropping entries that do not fit.
func StatusLine(t theme.Theme, counts []Count, width int) string {
	if len(counts) == 0 {
		return ""
	}

	var parts []string
	used := 0
	for _, count := range counts {
		part := t.StatusStyle(count.Status).Render(count.Glyph + " " + count.Label)
		needed := lipgloss.Width(part)
		if len(parts) > 0 {
			needed += 2
		}
		if width > 0 && used+needed > width {
			break
		}
		parts = append(parts, part)
		used += needed
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "  ")
}

func plural(n int, singular, plural string) string {
	if n == 1 {
		return strconv.Itoa(n) + " " + singular
	}
	return strconv.Itoa(n) + " " + plural
}

// Breadcrumb renders a context trail with the active segment in the accent
// color. right is optional, right-aligned metadata.
func Breadcrumb(t theme.Theme, segments []string, width int, right string) string {
	separator := t.Bar.Render("  /  ")
	maxSegment := max(8, width/3)
	parts := make([]string, 0, len(segments))
	for i, segment := range segments {
		style := t.Dim
		if i == len(segments)-1 {
			style = t.Accent
		}
		parts = append(parts, style.Render(strings.ToUpper(truncate(segment, maxSegment))))
	}

	crumb := strings.Join(parts, separator)
	if right == "" {
		return crumb
	}
	return Spread(crumb, right, width)
}

// Rule renders a hairline separator that spans the given width.
func Rule(t theme.Theme, width int) string {
	if width <= 0 {
		return ""
	}
	return t.Border.Render(strings.Repeat("─", width))
}
