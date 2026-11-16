# Definition of Done — v0.1

**Status:** Active
**Scope:** The release gate for v0.1. Every item must be true before tagging `v0.1.0`.

---

## Primary Acceptance Scenario

Using a workspace with at least four independent Git repositories:

1. Run `repora` from any directory.
2. Repora loads the configured workspaces.
3. Create a new workspace with `n`.
4. Register an existing local repository by path.
5. Associate the same repository with a second workspace.
6. Open the workspace dashboard.
7. See the state of every repository, including branch, changes, ahead/behind, and
   conflicts.
8. Identify modified, clean, conflicting, and unsynchronized repositories from the
   dashboard alone.
9. Enter one repository.
10. Create a branch and switch to it.
11. Stage files and create a commit.
12. Fetch, pull, and push.
13. View recent commit history.
14. Return to the dashboard.
15. Select a different repository and perform an independent operation.
16. Repeat steps 9–15 for the remaining repositories.
17. Quit and relaunch; configuration persists and state is re-read from Git.

All of this without leaving Repora and without ambiguity about which repository each
operation affected.

---

## Functional Checklist

### Configuration

- [ ] `~/.config/repora/config.yaml` is created on first run with a valid default.
- [ ] `--config` and `REPORA_CONFIG` override the default path.
- [ ] `~` is expanded in paths and never persisted unexpanded.
- [ ] Repository paths are canonicalized to the Git root.
- [ ] Invalid or missing paths produce a warning, not a crash.
- [ ] Duplicate repository paths are rejected.
- [ ] Dangling workspace references are a load error with a clear message.
- [ ] Saves are atomic; a failed save leaves the original intact.
- [ ] Unknown configuration fields survive a save.
- [ ] A config from a newer version is refused, not rewritten.

### Workspaces

- [ ] Create workspace.
- [ ] Rename workspace (display name only; ID stable).
- [ ] Delete workspace (removes workspace, not repositories).
- [ ] Add an existing repository to a workspace.
- [ ] Remove a repository from a workspace.
- [ ] A repository can belong to multiple workspaces.
- [ ] IDs are unique and slug-validated.

### Discovery

- [ ] Register a repository by manual path.
- [ ] Detect whether a path is a Git working tree.
- [ ] Reject bare repositories with a clear message.
- [ ] Resolve a subdirectory to its repository root.

### Dashboard

- [ ] Every repository in the workspace is listed.
- [ ] Rows show name, branch, staged/unstaged/untracked counts, conflicts, ahead/behind.
- [ ] Refresh is concurrent and never freezes the UI.
- [ ] Each row has independent loading and error states.
- [ ] Selection is preserved across refresh.
- [ ] Filtering works and clears with `Esc`.
- [ ] Unavailable repositories are listed and explained.

### Repository View

- [ ] Overview shows branch, upstream, divergence, counts, remotes, recent commits.
- [ ] Changes shows conflicts, staged, unstaged, and untracked sections.
- [ ] Diff matches `git diff` for the selected target and file.
- [ ] Branches shows local and remote branches with tracking info.
- [ ] History lists recent commits with hash, subject, author, and date.
- [ ] Tabs and back navigation behave per the navigation spec.

### Git operations

- [ ] Stage selected file.
- [ ] Stage all.
- [ ] Unstage all.
- [ ] Commit with a message; hooks run.
- [ ] Commit with nothing staged is rejected with a notice.
- [ ] Create branch with validation.
- [ ] Checkout branch; conflicting changes produce a clear error.
- [ ] Fetch updates remote-tracking refs.
- [ ] Pull fast-forwards; divergence fails clearly.
- [ ] Push works with and without an upstream.
- [ ] Push failure (rejected) shows Git's message.
- [ ] Every operation targets only the selected repository.

### Reliability

- [ ] Every operation names its target repository.
- [ ] Every operation has visible running, success, and error feedback.
- [ ] Long network operations are cancellable.
- [ ] Writes to the same repository are serialized.
- [ ] `index.lock` from an external process is reported, never deleted.
- [ ] Raw Git output is available for every failure.
- [ ] No operation silently discards uncommitted changes.

### Interface

- [ ] Usable at 80x24.
- [ ] Comfortable at 120x40.
- [ ] `--no-color` produces a complete interface.
- [ ] Every state is encoded by glyph and word, not only color.
- [ ] Action bar reflects the current context and hides disabled actions.
- [ ] Help is generated from the key registry.
- [ ] Empty states exist for every list that can be empty.
- [ ] Confirmations state the target and the consequence.
- [ ] Destructive actions are never the default choice.

### Quality gates

- [ ] `gofmt -l .` prints nothing.
- [ ] `go vet ./...` is clean.
- [ ] `go build ./...` succeeds.
- [ ] `go test ./...` passes.
- [ ] `go test -race ./...` passes.
- [ ] No goroutine leaks in TUI lifecycle tests.
- [ ] Coverage targets met per [`testing-strategy.md`](testing-strategy.md).
- [ ] Fuzz target runs clean for 60 seconds.
- [ ] No `TODO`, `FIXME`, or placeholder implementation in shipped packages.

### Packaging

- [ ] Binaries built for `darwin/arm64`, `darwin/amd64`, `linux/amd64`, `linux/arm64`.
- [ ] `repora --version` prints version, commit, and build date.
- [ ] `SHA256SUMS` published.
- [ ] README install instructions verified on a clean machine.
- [ ] CHANGELOG entry written.
- [ ] Known limitations documented.

---

## Explicit Non-Blockers

These are known and accepted for v0.1:

- No command palette.
- No branch deletion from the UI.
- No per-file unstage.
- No amend.
- No stash, rebase, cherry-pick, worktrees.
- No multi-line commit messages.
- No light theme.
- No Windows support.
- No automatic updates.

If any of these are needed to satisfy the primary acceptance scenario, the scope decision
must be revisited and recorded in
[`../01-architecture/decisions.md`](../01-architecture/decisions.md).

---

## Sign-off

| Role | Responsibility |
|---|---|
| Implementation | Functional checklist and quality gates |
| Review | Security checklist in [`security-and-reliability.md`](security-and-reliability.md) |
| Manual acceptance | Executed on a clean machine against four real repositories |
