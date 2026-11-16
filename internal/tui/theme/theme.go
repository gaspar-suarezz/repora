// Package theme centralizes Repora's design tokens and reusable styles.
//
// It has no dependency on the TUI or on application packages, so screens and
// components can share one visual language without import cycles.
package theme

import "github.com/charmbracelet/lipgloss"

// Palette holds the raw color tokens.
type Palette struct {
	Background      lipgloss.TerminalColor
	Surface         lipgloss.TerminalColor
	SurfaceElevated lipgloss.TerminalColor
	Border          lipgloss.TerminalColor
	Accent          lipgloss.TerminalColor
	TextPrimary     lipgloss.TerminalColor
	TextSecondary   lipgloss.TerminalColor
	TextMuted       lipgloss.TerminalColor
	Success         lipgloss.TerminalColor
	Warning         lipgloss.TerminalColor
	Danger          lipgloss.TerminalColor
	Info            lipgloss.TerminalColor
}

// DefaultPalette is the Repora dark theme.
var DefaultPalette = Palette{
	Background:      lipgloss.Color("#17191E"),
	Surface:         lipgloss.Color("#22252C"),
	SurfaceElevated: lipgloss.Color("#252932"),
	Border:          lipgloss.Color("#30343D"),
	Accent:          lipgloss.Color("#A3E635"),
	TextPrimary:     lipgloss.Color("#F4F4F5"),
	TextSecondary:   lipgloss.Color("#9CA3AF"),
	TextMuted:       lipgloss.Color("#6B7280"),
	Success:         lipgloss.Color("#A3E635"),
	Warning:         lipgloss.Color("#F5B84B"),
	Danger:          lipgloss.Color("#F16B6B"),
	Info:            lipgloss.Color("#60A5FA"),
}

// Theme is the resolved style set used by every screen and component.
type Theme struct {
	Palette Palette
	NoColor bool

	Title    lipgloss.Style
	Header   lipgloss.Style
	Dim      lipgloss.Style
	Accent   lipgloss.Style
	Success  lipgloss.Style
	Warning  lipgloss.Style
	Danger   lipgloss.Style
	Info     lipgloss.Style
	Selected lipgloss.Style
	Key      lipgloss.Style
	Label    lipgloss.Style
	Border   lipgloss.Style

	// Card styles. The border color and the title marker distinguish the
	// selected card without changing its size.
	Card         lipgloss.Style
	CardSelected lipgloss.Style
	Panel        lipgloss.Style
	Bar          lipgloss.Style
	Chip         lipgloss.Style
	ChipAccent   lipgloss.Style
}

// New builds the theme. When noColor is true every style is plain, so the
// interface remains readable through glyphs and words alone.
func New(noColor bool) Theme {
	palette := DefaultPalette

	if noColor {
		plain := lipgloss.NewStyle()
		bold := lipgloss.NewStyle().Bold(true)
		card := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)

		return Theme{
			Palette:      palette,
			NoColor:      true,
			Title:        bold,
			Header:       plain,
			Dim:          plain,
			Accent:       bold,
			Success:      plain,
			Warning:      plain,
			Danger:       plain,
			Info:         plain,
			Selected:     bold,
			Key:          bold,
			Label:        plain,
			Border:       plain,
			Card:         card,
			CardSelected: card,
			Panel:        card,
			Bar:          plain,
			Chip:         plain,
			ChipAccent:   bold,
		}
	}

	return Theme{
		Palette:      palette,
		Title:        lipgloss.NewStyle().Foreground(palette.TextPrimary).Bold(true),
		Header:       lipgloss.NewStyle().Foreground(palette.TextPrimary),
		Dim:          lipgloss.NewStyle().Foreground(palette.TextSecondary),
		Accent:       lipgloss.NewStyle().Foreground(palette.Accent).Bold(true),
		Success:      lipgloss.NewStyle().Foreground(palette.Success),
		Warning:      lipgloss.NewStyle().Foreground(palette.Warning),
		Danger:       lipgloss.NewStyle().Foreground(palette.Danger),
		Info:         lipgloss.NewStyle().Foreground(palette.Info),
		Selected:     lipgloss.NewStyle().Foreground(palette.TextPrimary).Bold(true),
		Key:          lipgloss.NewStyle().Foreground(palette.Accent).Bold(true),
		Label:        lipgloss.NewStyle().Foreground(palette.TextSecondary),
		Border:       lipgloss.NewStyle().Foreground(palette.Border),
		Card:         lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(palette.Border).Padding(0, 1),
		CardSelected: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(palette.Accent).Padding(0, 1),
		Panel:        lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(palette.Border).Padding(0, 1),
		Bar:          lipgloss.NewStyle().Foreground(palette.TextMuted),
		Chip:         lipgloss.NewStyle().Foreground(palette.TextSecondary),
		ChipAccent:   lipgloss.NewStyle().Foreground(palette.Accent).Bold(true),
	}
}

// StatusStyle maps a semantic status name to its style.
func (t Theme) StatusStyle(status string) lipgloss.Style {
	switch status {
	case "success":
		return t.Success
	case "warning":
		return t.Warning
	case "danger":
		return t.Danger
	case "info":
		return t.Info
	default:
		return t.Dim
	}
}
