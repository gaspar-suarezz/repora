# Testing Strategy

**Status:** Defined
**Scope:** Test levels, required coverage, fixtures, and tooling.

Testing is part of implementation, not a phase. Every deliverable ships with tests.

---

## Principles

1. Tests for business logic and infrastructure are mandatory. Tests for rendering are
   required where behavior is not trivially readable from the code.
2. Tests never touch the user's real repositories, real config, or global Git
   configuration.
3. Tests are deterministic: no reliance on wall-clock time, network access, or ordering.
4. A failing test is never skipped to make a build pass; it is fixed or explicitly
   quarantined with an issue.
5. The race detector is part of the default gate.

---

## Levels

| Level | Location | Scope | Speed |
|---|---|---|---|
| Unit | alongside packages | Pure functions, validation, parsing | ms |
| Component | `internal/tui/**` | Model transitions, rendering | ms |
| Integration | `test/integration` | Real Git, real filesystem, temp dirs | 100 ms – 2 s |
| End-to-end TUI | `internal/tui/app` | Message-driven full flows | 100 ms – 1 s |
| Fuzz | `internal/infra/gitcli` | Parser robustness | seconds |

---

## What to Test, by Layer

### Domain — `internal/domain`

- ID normalization and validation (valid, invalid, edge cases).
- Name validation (empty, whitespace-only, too long, unicode).
- `Classify()` precedence: unavailable > error > conflicts > diverged > behind > ahead >
  dirty > clean.
- Working tree helpers: `IsClean`, `HasConflicts`, `IsDiverged`, `NeedsSync`.
- Invariants: duplicate associations, dangling references.
- Import boundary test: `domain` imports no non-stdlib packages.

Table-driven, no I/O.

### Config — `internal/infra/config`

Load:

- Missing file → default created.
- Empty file, whitespace-only file.
- Valid version 1.
- Version 2 → refused with a clear error.
- Malformed YAML → error with line/column.
- Duplicate repository IDs, duplicate paths, dangling workspace references.
- Preferences out of range → clamped or rejected per spec.

Paths:

- `~` expansion, trailing slashes, relative path rejection.
- Symlinked repository, subdirectory of a repository, missing path, non-repository path.
- Paths with spaces, unicode, and trailing whitespace.

Save:

- Atomic write success.
- Write failure (injected) leaves the original intact.
- Permissions `0600` file / `0700` directory.
- Unknown top-level and nested keys preserved.
- Round trip: load → save → load equals the original `Config`.

### Git executor — `internal/infra/gitcli/executor.go`

- Non-zero exit → `*ExitError` with args, code, dir, stderr.
- Cancellation → context error, not `ExitError`.
- Standard input is nil; a stdin-reading command does not hang.
- Environment includes `GIT_TERMINAL_PROMPT=0`, `LC_ALL=C`.
- Timeout produces the documented error.

### Status parser — `internal/infra/gitcli/status_parse.go`

Table-driven cases, one per record shape:

- Header-only, no entries.
- `1` entries for each X/Y combination.
- `2` rename with score.
- `u` unmerged entries, multiple.
- `?` untracked, including spaces and unicode in paths.
- Branch headers: normal, detached, unborn, no upstream, ahead/behind, diverged.
- Mixed record streams.
- CRLF noise and trailing NUL, to confirm tolerance.
- Empty input.

Fuzz:

```go
func FuzzStatusParser(f *testing.F) {
    // seed with captured outputs
    f.Fuzz(func(t *testing.T, data []byte) {
        _, _ = ParseStatus(data) // must never panic
    })
}
```

### Git adapter — integration

Against temporary repositories and a local bare remote:

- Fresh `git init`: unborn branch, no commits.
- Staged, unstaged, untracked, and mixed states.
- Deleted files staged and unstaged.
- Rename with and without staged content.
- Merge conflict with 1 and 2+ conflicted files.
- Detached HEAD.
- Ahead, behind, and diverged after pushes to a bare remote.
- Upstream deleted (`[gone]`).
- Branch create, checkout, invalid names, dirty-tree refusal.
- Stage, stage all, unstage all, commit.
- Commit rejected by a hook.
- Pull fast-forward, pull divergence.
- Push with and without upstream.
- Log parsing with unicode and `%` characters in subjects.
- Diff for text, binary, and very large files.

### Application — `internal/app`

With fakes for ports:

- Workspace CRUD and association rules.
- Repository registration validation and duplicate detection.
- Refresh: all results delivered, worker count respected.
- Generation: stale results discarded.
- Locking: same repository serialized, different repositories concurrent.
- Lock cancellation: waiting returns promptly with no side effect.
- Post-write refresh scheduling (success and failure).
- Error wrapping preserves the original.

### TUI — `internal/tui`

Model tests (no rendering):

- Navigation stack push/pop and selection restoration.
- Key routing precedence: `Ctrl+C` → modal → input → screen → global.
- Enabled predicates per binding.
- Row state transitions: queued → refreshing → ready / error.
- Stale message discard.
- Notification lifecycle.
- Modal input preservation after a failed operation.

Rendering:

- Golden files per screen at 80x24, 100x30, 120x40 with `--no-color`.
- Column dropping order at each breakpoint.
- Truncation of long names, paths, and branch names.
- Empty states.

End-to-end (teatest):

- Home → workspace → repository → back, asserting screen content.
- Stage → commit → push loop with fake services.
- Cancel a long operation and verify the UI remains responsive.

---

## Fixtures

`test/fixtures` builds repositories deterministically.

Requirements:

- Every repository is created in `t.TempDir()`.
- `GIT_CONFIG_GLOBAL` and `GIT_CONFIG_SYSTEM` are redirected to a temp file so the
  developer's identity and aliases do not leak in.
- Each repository sets local `user.name` and `user.email`.
- Initial commit message is fixed; dates are fixed for reproducible history.
- Helpers compose: `NewRepo`, `WriteFile`, `Commit`, `Branch`, `Checkout`, `Merge`,
  `NewBareRemote`, `Push`.
- No helper leaves a process running.

---

## Gate Commands

```bash
gofmt -l .                  # must print nothing
go vet ./...                # must be clean
go test ./...               # must pass
go test -race ./...         # must pass
go test -coverprofile=cover.out ./...
go tool cover -func=cover.out
```

Additional, run before release:

```bash
go test -run Fuzz -fuzz FuzzStatusParser -fuzztime 60s ./internal/infra/gitcli
```

Coverage targets:

| Package | Minimum |
|---|---|
| `internal/domain` | 90% |
| `internal/infra/config` | 85% |
| `internal/infra/gitcli` | 80% |
| `internal/app` | 80% |
| `internal/tui` | 60% (behavior-focused) |

Coverage is a signal, not a goal. Untested critical paths are unacceptable regardless of
the number.

---

## Handling Flaky Tests

- No `time.Sleep` to synchronize; use channels or explicit polling with a deadline.
- No dependence on filesystem timestamps with 1-second resolution.
- No dependence on network; remotes are local bare repositories.
- A flaky test is deleted or fixed the same day; it is never left skipped.

---

## Continuous Integration

Minimum pipeline on every push:

```text
1. gofmt -l .        (fail if output)
2. go vet ./...
3. go build ./...
4. go test ./...
5. go test -race ./...
```

Matrix: `darwin/arm64`, `linux/amd64`.

The pipeline MUST NOT require network access beyond module download at setup.
