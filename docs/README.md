# Repora Documentation

This directory documents Repora's product, architecture, interface, and quality
standards. It is meant for contributors and users who want to understand how the tool
works and why it makes the choices it makes.

---

## Start Here

- [`../README.md`](../README.md) — install, usage, and current limitations.
- [`../AGENTS.md`](../AGENTS.md) — operating rules for anyone changing this repository.
- [`01-architecture/decisions.md`](01-architecture/decisions.md) — why the code looks the
  way it does.

## Documentation Map

### Product

| Document | Purpose |
|---|---|
| [`00-product/vision.md`](00-product/vision.md) | What Repora is, the problem it solves, and its principles |
| [`00-product/positioning.md`](00-product/positioning.md) | Audience, use cases, and differentiation |
| [`00-product/workspaces.md`](00-product/workspaces.md) | The workspace model: Repora-owned containers and membership rules |
| [`00-product/glossary.md`](00-product/glossary.md) | Shared vocabulary used across all documents |
| [`00-product/mvp-scope.md`](00-product/mvp-scope.md) | What is in and out of the v0.1 MVP |
| [`00-product/roadmap.md`](00-product/roadmap.md) | Version plan from v0.1 to v1.0 |

### Architecture

| Document | Purpose |
|---|---|
| [`01-architecture/architecture-overview.md`](01-architecture/architecture-overview.md) | Layers, dependency rules, and runtime flows |
| [`01-architecture/decisions.md`](01-architecture/decisions.md) | Recorded technical decisions and their rationale |
| [`01-architecture/domain-model.md`](01-architecture/domain-model.md) | Entities, invariants, and state vs. configuration |
| [`01-architecture/config-and-persistence.md`](01-architecture/config-and-persistence.md) | Config location, schema, atomic writes, versioning |
| [`01-architecture/git-integration.md`](01-architecture/git-integration.md) | Git executor contract, commands, parsing, safety rules |
| [`01-architecture/concurrency.md`](01-architecture/concurrency.md) | Worker pool, per-repository locking, cancellation |
| [`01-architecture/project-structure.md`](01-architecture/project-structure.md) | Package layout and import rules |

### TUI

| Document | Purpose |
|---|---|
| [`02-tui/ux-principles.md`](02-tui/ux-principles.md) | Interaction and layout principles |
| [`02-tui/visual-system.md`](02-tui/visual-system.md) | Color tokens, status encoding, spacing, typography |
| [`02-tui/navigation.md`](02-tui/navigation.md) | Screen stack, focus model, modals |
| [`02-tui/screens.md`](02-tui/screens.md) | Per-screen specification |
| [`02-tui/shortcuts.md`](02-tui/shortcuts.md) | Global and contextual keybindings |

### Quality

| Document | Purpose |
|---|---|
| [`04-quality/testing-strategy.md`](04-quality/testing-strategy.md) | Unit, integration, and TUI testing approach |
| [`04-quality/security-and-reliability.md`](04-quality/security-and-reliability.md) | Safety rules for operating on real repositories |
| [`04-quality/definition-of-done.md`](04-quality/definition-of-done.md) | MVP acceptance checklist |

---

## Conventions

- **Language:** documentation is written in English. Code identifiers, commands, and
  screen names are also English.
- **Normative language:** `MUST`, `SHOULD`, and `MAY` follow RFC 2119 semantics.
- **Scope markers:** every document states what it covers and what it explicitly does
  not cover.
- **Phase ownership:** if a capability is not listed in a phase deliverable, it is not
  part of that phase, regardless of previous drafts.
- **Single source of truth:** decisions are recorded in
  [`01-architecture/decisions.md`](01-architecture/decisions.md). Do not re-litigate
  them in phase documents without adding a new decision entry.

---

## Reading Order

For a new contributor or agent:

1. `00-product/vision.md`
2. `00-product/mvp-scope.md`
3. `01-architecture/architecture-overview.md`
4. `01-architecture/domain-model.md`
5. `01-architecture/project-structure.md`
6. `03-phases/phase-01-foundation.md` → onward

For a design pass, read `02-tui/*` after the architecture documents.

---

## Document Status

| Area | Status |
|---|---|
| Product | Defined |
| Architecture | Defined |
| TUI | Defined at specification level |
| Phases | Defined |
| Implementation | v0.1 complete; local release gate passing; publishing pending |

### What exists today

| Area | State |
|---|---|
| `cmd/repora` | `--version`, `--help`, `--config`, `--no-color`, `--debug-status` |
| `internal/domain` | Entities, identifiers, config validation, state classification |
| `internal/infra/config` | Path resolution, load, default creation, atomic writes |
| `internal/infra/gitcli` | Executor, version check, root detection, status parsing, branches, log, remotes, diff |
| `internal/app` | Bounded concurrent refresh, repository detail loading, workspace CRUD, per-repository write locks |
| `internal/tui` | Home, Workspace Dashboard, Repository View (4 tabs), diff, prompts |
| Write operations | Stage, stage all, unstage all, commit, create branch, checkout, fetch, pull, push |
| Configuration access | `,` opens `$EDITOR` and reloads; `Ctrl+R` reloads from disk |
| Workspace editing | Create, rename, delete workspaces and associations from the UI |
| Release gate | `make gate`, `make dist`, CI workflow, golden tests, acceptance test |
| `test/integration` | Real-repository tests using temporary directories |

### Next

Publishing v0.1.0 (remote, CI, release assets) is blocked on account recovery.
See [`PROGRESS.md`](PROGRESS.md).
