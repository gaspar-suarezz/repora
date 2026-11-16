# Concurrency and State Management

**Status:** Defined
**Scope:** Worker pool, per-repository locking, cancellation, and message semantics between
application services and the TUI.

See D-007, D-008, and D-013 in [`decisions.md`](decisions.md).

---

## Model

Repora is single-process and single-threaded at the UI level. All concurrency lives in the
application layer and is expressed to the TUI as messages.

```text
TUI (Update loop, sequential)
   │ tea.Cmd → runs in its own goroutine
   ▼
Application service
   │ worker pool / per-repo locks
   ▼
Git CLI subprocesses
   │
   └── results → tea.Msg → Update loop
```

Rules:

- `Update` MUST NOT block. No I/O, no `time.Sleep`, no locks held across rendering.
- Goroutines MUST NOT mutate TUI state directly. They return messages.
- Application services MUST be safe for concurrent use.
- Domain values passed to the UI are immutable snapshots (value types, fresh slices).

---

## Read Path: Workspace Refresh

### Pipeline

```text
RefreshWorkspace(ctx, workspaceID)
  1. gen := atomic.AddUint64(&generation, 1)
  2. repos := resolveAssociations(workspaceID)
  3. for each repo: send job{repoID, gen} to jobs channel
  4. N workers (default 4) loop:
        state := git.Status(ctx, repo.Path)
        emit RepositoryStatusMsg{RepoID, Gen: gen, State: state}
  5. close(jobs) after all sends; close(results) after all workers exit
```

### Delivery to the TUI

Two viable delivery mechanisms:

**A. One message per repository (recommended).**

- Each worker sends a `RepositoryStatusMsg` through a `chan tea.Msg` consumed by a
  `tea.Cmd` that returns the first message and re-registers itself.
- Pros: progressive rendering, per-row loading states, no giant message.
- Cons: requires a pump.

Pump pattern:

```go
func waitForStatus(ch <-chan tea.Msg) tea.Cmd {
    return func() tea.Msg { return <-ch }
}
```

After handling a message, `Update` returns `waitForStatus(ch)` again.

**B. One aggregated message when all repositories finish.**

- Simpler, but loses progressive feedback and makes a single slow repository block the
  whole dashboard.
- Rejected for the workspace dashboard.
- Acceptable for a single repository refresh.

### Generation counter

- A single counter per TUI session, incremented on every refresh request.
- Every in-flight result carries the generation it was produced for.
- `Update` discards messages with `Gen < currentGen`.
- Discarding is silent; no UI change, no notification.

This solves:

- Navigating away while a refresh is in flight.
- Two refreshes triggered in quick succession.
- Switching workspaces during a long fetch-free status batch.

### Deduplication

- A refresh request for a repository that already has an in-flight refresh with the same
  generation MUST NOT enqueue a second job.
- Track in-flight repository IDs in a set keyed by generation.
- A newer generation supersedes and cancels older contexts.

### Cancellation

- Each refresh has a derived `context.Context`.
- Leaving the workspace cancels it.
- Triggering a newer refresh cancels the previous one.
- Workers MUST check `ctx.Err()` before starting a job and pass `ctx` to the executor.
- Cancelled results are dropped, not surfaced as errors.

---

## Write Path: Mutating Operations

### Per-repository locks

```go
type Locker struct {
    mu    sync.Mutex
    locks map[domain.RepositoryID]*sync.Mutex
}

func (l *Locker) Lock(id domain.RepositoryID) func() {
    l.mu.Lock()
    m, ok := l.locks[id]
    if !ok {
        m = &sync.Mutex{}
        l.locks[id] = m
    }
    l.mu.Unlock()
    m.Lock()
    return m.Unlock
}
```

Rules:

- Every mutating operation acquires the lock for its repository.
- Locks are held for the duration of the Git subprocess, not longer.
- Different repositories mutate concurrently.
- Lock acquisition MUST be cancellable: if the context is cancelled while waiting, return
  `ErrCancelled` without running the command. Use a channel-based semaphore per
  repository instead of `sync.Mutex` if cancellation while waiting is required.

Recommended implementation: a buffered channel of size 1 per repository.

```go
type RepoLock struct{ ch chan struct{} }

func (r *RepoLock) Acquire(ctx context.Context) error {
    select {
    case r.ch <- struct{}{}:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
func (r *RepoLock) Release() { <-r.ch }
```

### Serialization scope

- Serialized: `stage`, `unstage`, `commit`, `checkout`, `create branch`, `delete branch`,
  `pull`, `push`, `fetch`.
- `fetch` writes remote-tracking refs and therefore counts as a write for locking
  purposes, even though it does not touch the working tree.
- Read operations (`status`, `log`, `diff`, `branches`, `remotes`) do not take the write
  lock. They may observe intermediate states during a write; the UI refreshes after the
  write completes.

### Post-operation behavior

```text
write succeeds
  → invalidate cached state for that repository
  → schedule a refresh for that repository (new generation)
  → show success notification with a one-line summary

write fails
  → show error with Git output
  → schedule a refresh anyway (state may have partially changed)
```

Never assume the operation left the repository in the expected state; always re-read.

### Index lock protection

- Repora's own per-repository lock prevents self-inflicted `index.lock` contention.
- External Git processes can still hold `index.lock`. Repora MUST NOT delete it.
- On `ErrIndexLocked`, the UI explains the likely cause and suggests checking for running
  Git processes.

---

## Startup and Validation Concurrency

- Repository validation at load may run concurrently with the same worker pool.
- Validation results are advisory: a failing repository is marked unavailable rather than
  aborting startup.
- Git version detection runs once, before the pool starts.

---

## UI-visible States

Every repository row has one of these states:

| State | Meaning |
|---|---|
| `Pending` | Queued, not started |
| `Loading` | Git command running |
| `Ready` | State available |
| `Error` | Refresh failed; error retained |
| `Unavailable` | Path missing or not a repository |

Transitions are explicit and tested. A row MUST NOT render stale data while `Loading`
unless the previous successful snapshot is deliberately retained with a "refreshing"
marker. Recommended: retain the snapshot and show a subtle spinner so the row does not
flicker.

---

## Global Constraints

1. Maximum concurrent Git subprocesses = pool size (default 4). This includes validation
   and refresh, but writes are not part of the pool and may add up to one subprocess per
   locked repository.
2. No global mutable repository state. State is a map owned by the application service,
   guarded by an `RWMutex`, from which snapshots are copied.
3. No goroutine leaks. Every worker exits when its context is done; tests assert this with
   `goleak`.
4. No unbounded channels. Jobs and results channels are sized to the workspace repository
   count.
5. `go test -race ./...` MUST pass.

---

## Testing

- Deterministic worker pool tests using a fake executor with controllable latency.
- Generation counter: stale messages are discarded.
- Cancellation: navigate away mid-refresh; no messages applied, no goroutine leak.
- Locking: two concurrent commits on the same repository run sequentially; on different
  repositories they overlap.
- Lock acquisition cancellation: waiting for a lock returns promptly on cancel.
- Failure isolation: one repository erroring does not block others.
