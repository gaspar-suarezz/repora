# Security and Reliability

**Status:** Defined
**Scope:** Safety rules for operating on real repositories and configurations.

Repora runs on a developer's machine with their credentials and their most valuable working
state. Reliability is a product requirement, not a quality nice-to-have.

---

## Threat Model

### Assets

- The user's Git repositories and uncommitted work.
- The user's configuration file.
- The user's credentials, held by Git, SSH, and credential helpers (never by Repora).
- The user's terminal session.

### Adversaries and Hazards

| Hazard | Example |
|---|---|
| Command injection | A branch name or path crafted to alter the invoked command |
| Path traversal | A registered path escaping its intended root |
| Destructive operation | Accidental branch deletion or history rewrite |
| Data loss | Overwriting uncommitted work or truncating config |
| Credential leakage | Storing tokens in config or logs |
| Process runaway | A hook or command that never terminates |
| Concurrent mutation | Two Git writes racing on the same index |
| Malicious repository | A repository path or config authored elsewhere |

Repora is not a sandbox. It trusts the user's environment but must not amplify mistakes.

---

## Rules

### Command execution

1. **No shell.** Never invoke `sh -c`, `bash -c`, or any shell. Use argument vectors.
2. **Explicit working directory.** Every Git command sets `cmd.Dir` to the resolved
   repository root.
3. **Argument separation.** Use `--` before pathspecs to prevent option injection.
4. **Validated identifiers.** Branch names are validated against Git's own rules before
   use.
5. **No user-controlled flags.** User input is never interpreted as an option. If a value
   could start with `-`, it is rejected or separated.
6. **Non-interactive.** `GIT_TERMINAL_PROMPT=0`, no stdin, no editor, no pager.
7. **Bounded output.** Diff and log output are capped before rendering.

### Repository safety

8. **No destructive operations in v0.1.** No force push, no hard reset, no clean, no
   branch deletion from the UI.
9. **No silent overwrite.** Repora never discards uncommitted changes to resolve a
   conflict or complete an operation.
10. **No implicit writes.** Refresh, startup, and navigation never mutate repositories.
11. **Serialized writes.** One mutating operation per repository at a time.
12. **Post-operation re-read.** State is always re-read from Git after a write; Repora does
    not assume success left the repository in an expected shape.
13. **Lock files are never deleted.** `index.lock` is surfaced, not removed.
14. **Hooks run.** Repora does not bypass hooks in v0.1; a hook failure is reported
    verbatim.

### Configuration safety

15. **Atomic writes.** Temp file, fsync, rename, fsync directory.
16. **No silent downgrade.** A newer config version is refused, not rewritten.
17. **Unknown fields preserved.** Older binaries do not destroy newer configuration.
18. **Restrictive permissions.** `0600` file, `0700` directory.
19. **No credentials.** Configuration never stores tokens, passwords, or secrets.
20. **Single writer.** A lock file prevents concurrent Repora instances from writing.

### Data handling

21. **No telemetry.** No analytics, no crash reporting, no network calls other than Git.
22. **No content exfiltration.** Repository contents are not logged or transmitted.
23. **Logs are opt-in.** If debug logging is enabled, paths and operation names are logged;
    file contents and credentials are not.
24. **Error output is local.** Git's stderr is displayed to the user and not persisted
    beyond the session unless logging is enabled.

### Process management (v0.2)

25. **Explicit action only.** Commands run only when the user invokes them.
26. **argv only.** No shell interpretation.
27. **Working directory containment.** The resolved working directory must stay inside the
    repository or workspace root.
28. **Bounded logs.** Ring buffers with hard caps.
29. **Group termination.** Stop signals the process group; escalate to `SIGKILL` after a
    timeout.
30. **No orphans.** Exit policy is explicit and confirmed.

---

## Failure Modes and Expected Behavior

| Failure | Expected behavior |
|---|---|
| Config unreadable | Fatal error before TUI; names path and OS error |
| Config malformed | Fatal error with line/column; original untouched |
| Config write fails | In-memory state reverted; error shown; original intact |
| Repository path missing | Row marked `unavailable`; app usable |
| Not a Git repository | Marked unavailable with explanation |
| Git not installed | Fatal error with install guidance |
| Git too old | Fatal error with detected and required versions |
| Auth required | Fast failure; guidance to authenticate in a terminal |
| Network timeout | Cancellable; no partial-state assumption; refresh after |
| Hook rejects commit | Commit fails; hook output shown |
| `index.lock` present | Error explaining a concurrent Git process |
| Context cancelled | Operation abandoned; no user-facing error dialog |
| Terminal resized mid-operation | Layout recalculated; operation continues |
| Config lock held | Read-only mode with a visible banner |

---

## Cancellation Semantics

- Cancelling a read operation leaves no side effects.
- Cancelling a write delegates to process termination; the repository may be left in an
  intermediate state, so Repora refreshes and reports the observed state rather than
  claiming success or failure.
- Cancelling `fetch` may leave partial refs; this is safe because refs are atomic at the
  Git level.
- Cancelling `pull` or `push` is reported as "interrupted" with a recommendation to verify
  state.

---

## Concurrency Safety

- Per-repository locks for all writes, including `fetch`.
- Generation counters for all asynchronous reads.
- No shared mutable state between goroutines other than guarded maps.
- `go test -race ./...` is a release gate.

---

## Privacy

Repora stores:

- Repository display names and absolute paths.
- Workspace names, descriptions, and associations.
- Preferences.

Repora does not store:

- File contents.
- Commit messages beyond the in-memory session.
- Remotes' credentials.
- Environment variables.
- Terminal history.

---

## Review Checklist

Before merging a change that touches Git execution, config writes, or concurrency:

- [ ] No shell invocation added.
- [ ] Every subprocess has an explicit `cmd.Dir`.
- [ ] Pathspecs are separated with `--`.
- [ ] User input is validated before use as an argument or identifier.
- [ ] A write holds the repository lock.
- [ ] Results are re-read after a write.
- [ ] Errors preserve raw Git output.
- [ ] No new `panic` reachable from user input.
- [ ] Tests cover the failure path, not only the success path.
- [ ] No credentials or file contents are logged.
