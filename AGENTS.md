# AGENTS.md

Operating rules for anyone — human or automated agent — changing this repository.

Repora is a terminal-native, multi-repository Git workspace manager written in Go.

---

## 1. Start here

Read these before making changes, in order:

1. `README.md`
2. `docs/00-product/vision.md`
3. `docs/00-product/workspaces.md`
4. `docs/00-product/mvp-scope.md`
5. `docs/01-architecture/architecture-overview.md`
6. `docs/01-architecture/decisions.md`
7. `docs/01-architecture/project-structure.md`

If a request conflicts with those documents, either update the documents deliberately in
the same change, or stop and ask. Do not silently diverge.

---

## 2. Source of truth

When two documents disagree, the earlier entry wins:

1. `docs/01-architecture/decisions.md`
2. `docs/00-product/*`
3. `docs/01-architecture/*`
4. `docs/02-tui/*`
5. `docs/04-quality/*`
6. `README.md`

---

## 2.1 Internal documents

The repository keeps a few working documents locally that are intentionally **not**
published:

- `repora.md` — original product specification.
- `docs/PROGRESS.md` — implementation status and session log.
- `docs/03-phases/` — delivery plans.

They are listed in `.gitignore`. Do not add them to the public repository and do not link
to them from published documentation.

## 3. Quick facts

| Item | Value |
|---|---|
| Language | Go 1.24+ |
| Binary | `make build` → `bin/repora` |
| Git | 2.30+ required |
| Platforms | macOS and Linux |
| Module | `github.com/gasparsuarez/repora` |
| Config | `~/.config/repora/config.yaml` (override with `--config` or `REPORA_CONFIG`) |
| Gate | `make gate` (format, vet, test, race) |
| Distribution | `make dist` (four binaries + `SHA256SUMS`) |

---

## 4. Product invariants

These must not be broken:

- Workspaces are owned by Repora, not by Git.
- Workspace membership exists only in Repora configuration.
- A repository may belong to multiple workspaces.
- Repositories remain operationally independent at all times.
- Every Git write operation targets exactly one explicitly selected repository.
- Repora must not fan out writes across repositories automatically.
- The product is workspace-first, not single-repository-first.
- Grouping repositories never changes their branches, remotes, hooks, or history.

If a change weakens any of these, stop and document the reason first.

---

## 5. Architecture invariants

- Use the native Git CLI through `os/exec`.
- Never use `sh -c` or shell interpolation.
- Keep Git execution out of the TUI layer.
- Keep domain types free of Bubble Tea, Lip Gloss, and infrastructure imports.
- Preserve the layer direction:

  ```text
  cmd → tui → app → domain
  cmd → infra → domain
  ```

- `app` must not import `infra`, `tui`, or `cmd`.
- `infra` must not import `app`, `tui`, or `cmd`.
- `tui` must not import `infra`.
- Prefer small concrete services over premature abstraction.
- Do not introduce a database, event bus, or plugin system without updating the docs and
  recording a decision.

---

## 6. Git and safety rules

- Always run Git with an explicit working directory.
- Always pass arguments as a vector; never build a command string.
- Separate pathspecs with `--`.
- Validate branch names and commit messages before invoking Git.
- Keep `pull` fast-forward-only unless a decision changes it.
- Preserve Git's stderr, exit code, and argument vector in errors.
- Never delete `index.lock`; report it.
- Never auto-run destructive or network operations from refresh, navigation, or startup.
- Never expose force push, hard reset, clean, stash drop, or history rewriting without:
  1. a recorded decision,
  2. a confirmation flow,
  3. tests.
- Serialize writes per repository; never hold a lock across rendering.

---

## 7. TUI and UX rules

- Keyboard-first always. Do not add mouse-only interactions.
- Keep every screen usable at 80x24 and comfortable at 120x40.
- Never encode meaning with color alone; pair it with a glyph or a word.
- Keep the target repository visible on every repository-scoped screen.
- Never block the update loop with I/O; use `tea.Cmd`.
- Discard stale asynchronous results by generation counter.
- Preserve selection across refresh unless the selected item disappeared.
- Success notices auto-dismiss; error notices persist until replaced.
- Do not silently discard an in-progress input after a failed operation.
- Prefer overlays over nested panes unless the screen specification says otherwise.

---

## 8. Configuration rules

- Follow `docs/01-architecture/config-and-persistence.md`.
- Keep configuration YAML human-editable.
- Write atomically: temp file, fsync, rename, directory fsync.
- Preserve unknown fields on save (D-014, D-020).
- Validate the whole configuration before writing; never persist an invalid state.
- Deleting a workspace must not delete repositories from disk or unregister them.
- Removing a repository from a workspace must not unregister it.
- Never store credentials, tokens, or secrets in configuration.

---

## 9. Change workflow

For any non-trivial change:

1. Read `docs/PROGRESS.md`.
2. Decide whether it belongs to v0.1.x polish, v0.2, or a later version.
3. Implement the smallest complete vertical slice.
4. Add or update tests in the same change.
5. Update documentation if behavior or constraints changed.
6. Add a factual entry to the `Session log` in `docs/PROGRESS.md`.

Never present a placeholder as finished functionality. If something is partial, say so
in *In progress*.

### Repository hygiene

- Do not commit unless the user asks.
- Do not rewrite history, force push, or change the committed identity.
- Do not add files outside the project without being asked.
- Keep `bin/`, `dist/`, and generated artifacts out of version control.

---

## 10. Testing and release gates

Before considering work complete:

```bash
make gate
```

`make gate` runs format checking, vet, the full test suite, and the race detector.

Also required when relevant:

| Change | Additional check |
|---|---|
| TUI rendering | Golden tests pass; update them intentionally with `-update` |
| Packaging or release | `make dist` |
| Parsing or concurrency | `go test -race ./...` plus targeted tests |
| Documentation links | All relative links resolve |

Tests must never touch the user's real repositories, real configuration, or global Git
configuration. Use temporary directories and isolated Git configuration.

---

## 11. What not to do

- Do not turn Repora into a clone of another terminal Git client.
- Do not optimize for single-repository workflows at the expense of multi-repository
  visibility.
- Do not replace explicit user configuration with implicit filesystem inference.
- Do not add frameworks or dependencies without a clear, documented benefit.
- Do not bypass decisions to move faster.
- Do not mark work as done when tests or documentation were skipped.
- Do not expand scope silently; propose it first.

---

When in doubt, preserve repository safety and workspace semantics over feature breadth.
