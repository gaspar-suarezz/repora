# Visual System

**Status:** Defined
**Scope:** Color tokens, semantic status encoding, layout metrics, typography, and
rendering rules. Implementation lives in `internal/tui/styles`.

---

## Design Direction

Minimal, premium, terminal-native. Charcoal surfaces, a single lime accent, restrained
color, generous vertical rhythm, and no unnecessary chrome.

Rules:

- Surfaces are differentiated by lightness, not by heavy borders.
- Borders are hairlines or absent; boxed panels are the exception, not the default.
- The accent appears only on: the brand mark, the selected row gutter, focused inputs, and
  primary confirmations.
- Semantic colors are reserved for repository state. They never decorate.

---

## Color Tokens

```go
// internal/tui/theme/theme.go
type Palette struct {
    Background      lipgloss.TerminalColor // #17191E
    Surface         lipgloss.TerminalColor // #22252C
    SurfaceElevated lipgloss.TerminalColor // #252932
    Border          lipgloss.TerminalColor // #30343D
    Accent          lipgloss.TerminalColor // #A3E635
    TextPrimary     lipgloss.TerminalColor // #F4F4F5
    TextSecondary   lipgloss.TerminalColor // #9CA3AF
    TextMuted       lipgloss.TerminalColor // #6B7280
    Success         lipgloss.TerminalColor // #A3E635
    Warning         lipgloss.TerminalColor // #F5B84B
    Danger          lipgloss.TerminalColor // #F16B6B
    Info            lipgloss.TerminalColor // #60A5FA
}
```

| Token | Hex | Usage |
|---|---|---|
| Background | `#17191E` | Application background |
| Surface | `#22252C` | Cards, elevated panels |
| SurfaceElevated | `#252932` | Highlighted elements inside a surface |
| Border | `#30343D` | Card borders and separators |
| Accent | `#A3E635` | Brand mark, selection, primary actions |
| TextPrimary | `#F4F4F5` | Names and headings |
| TextSecondary | `#9CA3AF` | Metadata, descriptions |
| TextMuted | `#6B7280` | Hints, scroll indicators, disabled text |
| Success | `#A3E635` | Clean repository state |
| Warning | `#F5B84B` | Modified or unsynchronized state |
| Danger | `#F16B6B` | Conflicts and failures |
| Info | `#60A5FA` | Informational state and sync direction |

Rules:

- Accent and Success share a hue by design. They are used in different contexts
  (interaction versus repository state), and meaning never depends on hue alone.
- Text on `Background` must meet WCAG AA contrast (4.5:1) for body text.
- `TextMuted` is for non-essential metadata only; never for actionable content.
- If the terminal reports fewer than 256 colors, degrade to the ANSI-16 approximation
  below.

### ANSI-16 fallback

| Token | ANSI |
|---|---|
| Background | default |
| Surface | black / bright black |
| SurfaceElevated | bright black |
| TextPrimary | default foreground |
| TextSecondary | bright black |
| Accent | bright green |
| Success | green |
| Warning | yellow |
| Error | red |
| Info | blue |

`--no-color` disables all color; glyphs and words carry the meaning.

---

## Repository Status Encoding

Color is never the only signal. Every state has a glyph and a word.

| State | Glyph | Word | Color | Precedence |
|---|---|---|---|---|
| Unavailable | `!` | `unavailable` | Error | 1 |
| Error | `✗` | `error` | Error | 2 |
| Refreshing | `◐` | `refreshing` | TextMuted | 3 |
| Pending | `·` | `queued` | TextMuted | 4 |
| Conflicts | `✖` | `conflicts` | Error | 5 |
| Diverged | `⇅` | `diverged` | Error | 6 |
| Behind | `↓` | `behind N` | Info | 7 |
| Ahead | `↑` | `ahead N` | Warning | 8 |
| Dirty | `●` | `modified` | Warning | 9 |
| Clean | `✓` | `clean` | Success | 10 |
| Detached | `⌥` | `HEAD` | Info | marker |

Notes:

- Unavailable and Error rank above conflicts because no reliable state exists.
- Diverged outranks ahead/behind because both actions are needed.
- Detached is a marker, not a class: a detached HEAD can also be dirty.
- Untracked-only repositories count as Dirty with the label `modified`.
- The label is preferred over the glyph in narrow layouts; both appear when width allows.

### Row composition

```text
> api            ● modified   main                 2 staged  ↑1
  web            ✓ clean      develop              up to date
  infra          ↓ behind 2   main
  legacy-api     ! unavailable
```

Columns, in order of dropping priority when width shrinks:

1. Selection gutter (always)
2. Repository name (always, truncated with ellipsis)
3. State glyph and word (always)
4. Branch (always; muted, truncated to the remaining width)
5. Change summary (dropped below 104 columns)
6. Divergence appears in the state column

---

## Typography and Glyphs

- Monospace only; no font configuration.
- Glyphs MUST be single-width in UTF-8. Avoid emoji.
- Reserved glyphs:

| Purpose | Glyph |
|---|---|
| Selection marker | `>` |
| Separator dot | `·` |
| Ahead | `↑` |
| Behind | `↓` |
| Diverged | `⇅` |
| Clean | `✓` |
| Error | `✗` |
| Conflict | `✖` |
| Unavailable | `!` |
| Breadcrumb separator | `/` |
| Loading frames | `◐ ◓ ◑ ◒` |
| Ellipsis | `…` |

- Table headers are uppercase and `TextSecondary`.
- Metadata separators are ` · ` (space, dot, space).
- Paths use middle truncation: `~/code/pro…/api`.

---

## Spacing

Base unit: 1 cell. All spacing is a multiple of it.

Global margins:

- Two cells on the left and right of every screen.
- One blank line above the content.
- One blank line between content and the status line.

The header, the rule, panels, cards, and the footer all live inside those margins, so
nothing touches the terminal edge.

| Context | Horizontal | Vertical |
|---|---|---|
| Screen padding | 1 cell left, 1 right | 1 line top, 1 bottom |
| Between sections | — | 1 blank line |
| List row | 1 cell gutter, 1 cell gap after gutter | 0 (single-line rows) |
| Modal padding | 2 cells | 1 line |
| Action bar | 2 cells between items | fixed at bottom |

Rules:

- Never exceed 2 cells of horizontal padding inside a panel.
- Section headers are followed by exactly one blank line.
- Do not add blank lines between rows.

---

## Borders and Separators

- Prefer a single hairline row of `─` or no separator at all.
- Rounded borders are used only for modals and focused input boxes.
- Focused panel: `BorderFocus` on all sides.
- Unfocused panel: no border, or `Border` hairlines only.

---

## Layout Metrics

| Breakpoint | Width | Behavior |
|---|---|---|
| Comfortable | >= 120 | All columns visible |
| Standard | 100 – 119 | Branch moves to a second line for dirty rows |
| Compact | 80 – 99 | Change summary dropped; state word shown |
| Minimum | < 80 | Single-column layout; show a notice that the terminal is narrow |

Minimum supported size: **80x24**. Below 80 columns, Repora renders a single-column
fallback rather than producing broken tables.

Vertical budget:

- Header: 1 line (2 when the breadcrumb wraps)
- Action bar: 1 line
- Status line: 1 line
- Content: remaining lines, minimum 10

---

## Component Style Rules

### Rule

A hairline of `Border` color separates the brand header from the screen context. It spans
the full terminal width and never appears inside panels.

### Panel

The single working surface of an internal screen.

- Rounded border in `Border`, padding of one cell on each side.
- One panel per screen; never nest panels.
- Content is truncated with `…` rather than clipping the bottom border.
- The Workspace panel hugs its content; the Repository panel fills the available height so
  tabs do not jump between sections.

### Screen header (breadcrumb)

```text
WORKSPACES / PRODUCT X                              4 repositories
```

- Segments are uppercase.
- Ancestor segments use `TextMuted`; the active segment uses `Accent` bold.
- Separators `/` use `TextMuted`.
- Right-aligned metadata (count, repository state) uses `TextSecondary`.
- The breadcrumb always sits below the brand header, never replaces it.

### Application header

```text
◈  Repora   WORKSPACE MANAGER
```

- Symbol `◈` in `Accent`, wordmark in `TextPrimary` bold.
- Descriptor `WORKSPACE MANAGER` in `TextMuted`, uppercase.
- Warnings appear right-aligned as `! N warnings` in `Warning`.
- The build version appears at the far right in `TextMuted`, for example `v0.1.0` or `dev`.
- Warnings and version share the right side, separated by three spaces.
- One line; never taller.

### Workspace card

```text
╭──────────────────────────────────────────────╮
│ ◆ Product X                          4 repos │
│ Customer-facing product                      │
│ [API] [Web] [Infrastructure] [Shared]        │
│ ✖ 1 conflict  ● 1 modified  ↑ 1 sync  ✓ 1 clean │
╰──────────────────────────────────────────────╯
```

- Fixed height: four content lines plus border. Cards never change height, so the list
  does not jitter when the selection moves.
- Unselected: rounded border in `Border`, marker `◇` in `TextMuted`.
- Selected: rounded border in `Accent`, marker `◆` in `Accent`, name in `TextPrimary`
  bold. Only the border color and marker change; the size is identical.
- Name is left-aligned; repository count (`N repos`) is right-aligned.
- Description is optional; an empty description keeps the line blank for rhythm.
- Repository tags use `[NAME]` in `TextSecondary`; overflow collapses into `+N`.
- Empty workspace: tags line reads `No repositories yet` and the status line reads
  `↵ open to add repositories`.

### Repository tags

- Format `[NAME]`, separated by one space.
- Never wrap; drop whole tags and append `+N` when the line is full.
- Names longer than the available width are truncated with an ellipsis.

### Status indicators

| Indicator | Glyph | Status color | Meaning |
|---|---|---|---|
| Conflicts | `✖` | Danger | At least one unmerged path |
| Modified | `●` | Warning | Staged, unstaged, or untracked changes |
| Sync | `↑` | Info | Ahead, behind, or diverged |
| Errors | `!` | Danger | Status could not be read |
| Clean | `✓` | Success | No changes and no divergence |

Ordering is by severity: conflicts, modified, sync, errors, clean. Counts equal to zero
are omitted. Every indicator carries a word, so color is never the only signal.

### Footer

```text
↑↓ Navigate   ↵ Open   n New   e Rename   d Delete   r Refresh   ? Help
```

- Keys in `Accent` bold; labels in `TextSecondary`.
- No brackets: the key token is emphasized by weight and color.
- Entries are dropped from the end when the terminal is narrow. The first entry is always
  kept.
- Every entry performs a real action in the current context.

### Action bar

```text
[Enter] Open   [f] Fetch   [p] Pull   [P] Push   [r] Refresh   [?] Help
```

- Key token: `Accent`, bold.
- Label: `TextSecondary`.
- Disabled action (not applicable in context): `TextMuted` and omitted from the bar.

### Status line

- Idle: last success notice or a faint hint.
- Running: spinner plus `fetch api…`.
- Error: `Error` color, prefixed with `✗`.
- Success: `Success` color, prefixed with `✓`, auto-clears after 4 seconds.

### Selection

- Gutter marker `>` in `Accent`.
- Selected row background `Surface`.
- Selected row text `TextPrimary` even if the row is otherwise muted.

### Form panel

Prompts and confirmations reuse the screen shell: brand header, rule, breadcrumb, and a
single focused panel.

- The panel uses the accent border (`CardSelected`) because it holds the focused control.
- The title is uppercase and muted; the question or field label follows inside the panel.
- Hints and validation errors sit on the last line inside the panel.
- Long breadcrumb segments are truncated so the header can never overflow.

### Modal

- Background `Surface`.
- Rounded border, `BorderFocus` when it contains a focused input.
- Title in `TextPrimary`, bold.
- Consequence text in `TextSecondary`.
- Destructive action in `Error` color.

### Diff viewer

| Line type | Prefix | Color |
|---|---|---|
| Addition | `+` | DiffAdd |
| Deletion | `-` | DiffDel |
| Hunk header | `@@` | DiffHunk |
| Context | ` ` | TextSecondary |
| File header | `diff ` | TextPrimary, bold |

Line numbers are shown when width allows; they are not colored.

---

## Truncation and Wrapping

- Names: trailing ellipsis.
- Paths: middle ellipsis, preserving the last two segments.
- Branch names: middle ellipsis, preserving the prefix and the last segment.
- Commit subjects: trailing ellipsis; the full subject appears in the commit detail.
- Diff lines: never wrapped; horizontal scroll or truncation with an indicator.
- No word-aware wrapping in tables; tables truncate.

---

## Accessibility Checklist

- [ ] Every color-coded state has a glyph and a word.
- [ ] `--no-color` produces a complete, legible interface.
- [ ] No text relies on `TextMuted` for meaning.
- [ ] The interface is usable at 80x24.
- [ ] Focus is visible without color (border weight or marker).
- [ ] Spinners stop; no permanent animation on static screens.
- [ ] Selected row is distinguishable from unselected rows in grayscale.

---

## Implementation Notes

- Tokens live in `styles/tokens.go`; themes in `styles/theme.go`.
- Screens never hard-code hex values or ANSI codes.
- Conditionals on width are centralized in `styles/render.go`, not scattered in screens.
- Glyph constants live in `styles/tokens.go` so tests can assert against them.
