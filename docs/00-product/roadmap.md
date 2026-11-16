# Roadmap

**Status:** Directional
**Scope:** Version evolution. Not a commitment to deliver every item.

---

## Version Overview

| Version | Theme | Status |
|---|---|---|
| v0.1 | Multi-repository Git management | Defined in [`mvp-scope.md`](mvp-scope.md) |
| v0.2 | Workspace orchestration | Planned |
| v0.3 | Advanced Git management | Planned |
| v1.0 | Developer workspace platform | Directional |

---

## v0.1 — Multi-Repository Git Management

**Goal:** Prove the core thesis — consolidated multi-repository awareness plus safe,
explicitly targeted Git actions.

Includes:

- YAML configuration and workspace management.
- Home / Workspace Explorer.
- Workspace Dashboard with concurrent status refresh.
- Repository View (overview, changes, branches, history).
- Fetch, pull, push, create branch, checkout, stage, unstage all, commit.
- Read-only diff viewer.
- Keyboard-driven navigation with contextual action bar.
- macOS and Linux binaries.

Excludes everything listed in the out-of-scope table of
[`mvp-scope.md`](mvp-scope.md).

**Exit criteria:** the acceptance checklist in
[`../04-quality/definition-of-done.md`](../04-quality/definition-of-done.md) passes.

---

## v0.2 — Workspace Orchestration

**Goal:** Move from "manage repositories" to "manage the development environment".

Adds:

- **Command palette** as the primary action surface, context-aware by screen and
  selection.
- **Workspace commands:** developer-defined commands stored in configuration, bound to a
  workspace and optionally to a specific repository, with an explicit working directory.
- **Process lifecycle:** start, stop, restart, and status for each command.
- **Process logs:** ring-buffered output per process, viewable in a dedicated screen.
- **Running-process indicators** on the workspace dashboard and repository rows.
- **Command palette extensions:** run commands, jump to processes, filter by context.

Design constraints carried forward:

- Commands execute only after explicit user action.
- Commands never run as part of workspace refresh.
- Each command has an explicit working directory.
- Process output is bounded in memory.

**Exit criteria:** a developer can start the backend and frontend of a project from
Repora, watch their logs, restart them, and stop them without leaving the TUI.

---

## v0.3 — Advanced Git Management

**Goal:** Close the remaining daily-workflow gaps for developers who rarely leave the
TUI.

Adds:

- Worktree management (list, create, remove, open).
- Stash management (push, list, pop, apply, drop).
- Cherry-pick.
- Interactive rebase (at minimum: reword, squash, drop, reorder).
- Improved commit history browsing (graph, filtering, commit detail view).
- Advanced branch operations (rename, delete with unmerged detection, set upstream).
- Better diff navigation (file tree, hunk jumping, word-level diff).

**Exit criteria:** a developer can complete a full feature branch lifecycle — including
history cleanup — without leaving Repora.

---

## v1.0 — Developer Workspace Platform

**Goal:** Turn Repora into the default entry point for a developer's working day.

Candidate additions:

- Workspace templates (scaffold a workspace from a manifest).
- Improved repository discovery (bounded scanning with preview and bulk import).
- Custom workspace configurations and per-workspace preferences.
- Enhanced theming and layout customization.
- Additional integrations where there is demonstrated demand.
- Optional session persistence (restore last workspace and selection).

Not planned at any version:

- Hosted service or account system.
- Team collaboration features.
- Telemetry by default.

---

## Sequencing Rationale

1. **v0.1 first, depth second.** Orchestration without a reliable repository layer would
   be built on sand. Status parsing, executor safety, and concurrency are prerequisites
   for everything else.
2. **Command palette before commands.** Once commands exist, the palette becomes the
   natural surface for them. Shipping the palette earlier would mean rewriting it.
3. **Advanced Git last.** It has the highest complexity and the lowest marginal value for
   the multi-repository thesis.
4. **Release hardening is continuous.** Phase 05 is a stabilization gate, not the first
   time quality is considered.

---

## What Would Change This Roadmap

- Evidence that users mainly want orchestration, not Git depth: promote v0.2 earlier.
- Evidence that users mainly operate one repository at a time: reconsider the
  multi-repository emphasis entirely.
- Requests for Windows support: requires path handling and terminal capability review;
  would become its own phase.
