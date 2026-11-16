# UX Principles

**Status:** Defined
**Scope:** Interaction and layout principles. Visual tokens are in
[`visual-system.md`](visual-system.md); screens in [`screens.md`](screens.md).

---

## The Core Feeling

Repora should feel like a precision instrument: quiet, fast, and unambiguous. The user
should never wonder:

- which repository an action will affect
- whether a command is still running
- whether an operation succeeded
- why a row looks different from another

If any of those questions arise, the design has failed, regardless of how it looks.

---

## Principles

### 1. The target is always explicit

Every screen that performs a repository-scoped action shows the target repository name in
a fixed position (header or action bar). There is no implicit "current repository" carried
across screens.

- Confirmation dialogs repeat the target: **"Commit to `api`?"**, not **"Commit?"**.
- The command palette, when introduced, shows the resolved target next to each action.

### 2. Breadth before depth

The workspace dashboard is the most important screen. It must answer "what is the state of
this project?" in a single glance, above the fold.

Repository detail screens are secondary and must be reachable in one keystroke from the row
they belong to.

### 3. Keyboard first, not keyboard only

- Every action has a keybinding.
- Every keybinding is discoverable: the action bar and Help screen always list what is
  available in the current context.
- Destructive actions are never bound to a single unmodified key.

### 4. Fast paths beat completeness

The common loop is: open workspace → scan rows → act on one repo → return. Any design that
adds a screen, a menu, or a confirmation to that loop must justify itself.

### 5. Non-destructive by default

- Destructive or irreversible actions require confirmation.
- Confirmation defaults to "No"; `Esc` cancels.
- Inputs (commit message, branch name) are preserved if an operation fails, so the user
  can retry without retyping.

### 6. Every operation has visible state

| Phase | Feedback |
|---|---|
| Requested | Row enters loading state immediately |
| Running | Spinner plus the operation name in the status line |
| Succeeded | One-line confirmation, auto-dismissed |
| Failed | Persistent error banner with details, dismissible |
| Cancelled | Subtle status line notice |

No operation is silent. No operation blocks the interface.

### 7. Errors are actionable

Every error message answers three questions:

1. What failed?
2. Which repository?
3. What can the user do next?

Git's raw output is always available, one keystroke away, never stuffed into a modal by
default.

### 8. Density with air

Terminal space is precious and users scan quickly.

- One line per repository in the dashboard, plus a metadata line only when it carries
  information (dirty, diverged, error).
- No nested borders around every panel.
- Section headers are labels, not boxes.
- Padding is applied to the content area, not to each cell.

### 9. No color-only meaning

Every status encoded by color is also encoded by a glyph or word. This keeps the interface
usable with colorblindness, limited palettes, and `--no-color`.

### 10. Stable layout

- The header, action bar, and status line keep fixed positions.
- Content areas do not jump when a spinner appears; reserve the space.
- Truncation is consistent: ellipsis at the end for names, middle truncation for paths.

### 11. Motion is feedback, not decoration

- Spinner only while work is in flight.
- No animated transitions between screens.
- Flashing is reserved for errors that require attention.

### 12. Interruptible

- `Esc` cancels the current input or modal.
- `Ctrl+C` is a global escape hatch: first press cancels running operations, second press
  quits.
- Quitting with operations in flight requires confirmation.

---

## Layout Grammar

Every primary screen follows the same vertical structure:

```text
┌─────────────────────────────────────────────┐
│ Header: breadcrumb + target + counts        │  1 line (2 when wrapped)
├─────────────────────────────────────────────┤
│                                             │
│ Content                                     │  flexible
│                                             │
├─────────────────────────────────────────────┤
│ Status line: operation / transient notice   │  1 line
├─────────────────────────────────────────────┤
│ Action bar: contextual keys                 │  1 line
└─────────────────────────────────────────────┘
```

Rules:

- Header is always present, even when empty, to keep content from shifting.
- Status line is always present; when idle it shows a hint or nothing.
- Action bar lists at most 6 actions, most useful first; `?` reveals the rest.
- Modals overlay the content area and keep the header visible so the target remains
  visible.

---

## Interaction Patterns

### Selection

- A single row is selected at all times in list screens.
- Selection is indicated by a left gutter marker plus a subtle background change, never by
  color alone.
- `↑`/`↓` move one row; `k`/`j` also work; `g`/`G` jump to first/last.
- Selection is preserved across refresh when the row still exists.

### Focus

- Focus cycles only between interactive regions that exist on the current screen.
- `Tab`/`Shift+Tab` move focus where more than one region exists.
- The focused region has a visible marker in its header.

### Modals

- Modals are a stack, not a mode switch; `Esc` pops one level.
- While a modal is open, global shortcuts that would navigate away are disabled, except
  `Ctrl+C`.
- A modal that triggers a save or an operation stays open until the result is known, then
  closes on success or shows the error inline on failure.

### Prompts

- Single-line text prompts for branch names and commit subjects.
- Commit message supports a single line in v0.1; multiline is post-MVP.
- Validation feedback appears inline, under the input, before submission.
- Enter submits; `Esc` cancels and preserves nothing (explicit cancel is intentional).

### Confirmations

```text
┌ Delete workspace "Product X"? ─────────────┐
│ This removes the workspace from Repora.    │
│ Repositories are not deleted from disk.    │
│                                            │
│            [Esc] Cancel   [Enter] Delete   │
└────────────────────────────────────────────┘
```

- The dialog states the exact consequence, including what is *not* affected.
- The destructive choice is never the default focus.

---

## Empty States

Every list screen defines one:

| Screen | Empty state |
|---|---|
| Home | "No workspaces yet. Press `n` to create one." |
| Workspace dashboard | "This workspace has no repositories. Press `a` to add one." |
| Changes | "Working tree clean." |
| Branches | Never empty (at least one branch exists). |
| History | "No commits yet." |

Empty states are informative, not decorative, and always name the next action.

---

## Loading and Latency

- Anything under 100 ms shows nothing.
- 100 ms – 1 s: spinner only.
- Over 1 s: spinner plus the operation name.
- Over 5 s: add a hint that `Esc` cancels.

Never show a progress percentage for Git operations; it would be fabricated.

---

## Accessibility

- `--no-color` renders the full interface with glyphs and words only.
- No information depends on hue alone.
- Contrast targets follow the palette in [`visual-system.md`](visual-system.md).
- The interface must be usable at 80x24 and comfortable at 120x40.
- Below 80 columns, secondary columns are dropped in a defined order rather than wrapped
  unpredictably.

---

## Anti-Patterns

Reject these in review:

- A dashboard row that requires horizontal scrolling.
- A confirmation that names the action but not the repository.
- An error that only says "failed".
- Automatic refresh that moves the selection.
- Silent success.
- Blocking the UI on a network operation.
- Modals nested more than two levels deep.
- Shortcuts that differ between screens without a visible reason.
