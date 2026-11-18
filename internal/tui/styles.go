package tui

import (
	"github.com/gasparsuarez/repora/internal/tui/theme"
)

// styles is the resolved theme used by every screen.
type styles = theme.Theme

func newStyles(noColor bool) styles { return theme.New(noColor) }
