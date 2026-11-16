# Glossary

**Status:** Defined
**Scope:** Shared vocabulary. All other documents use these terms with these meanings.

---

| Term | Definition |
|---|---|
| **Workspace** | A container owned by Repora that groups user-chosen repositories. It exists only in Repora's configuration: it is not a Git concept and not a directory. See [`workspaces.md`](workspaces.md). |
| **Repository** | A local Git repository registered in the configuration, identified by an ID and a filesystem path. |
| **Repository State** | Dynamic, in-memory Git information for a repository (branch, upstream, ahead/behind, working tree). Never persisted. |
| **Registration** | The act of adding a repository to the global repository list in configuration. |
| **Association** | Linking an existing registered repository to a workspace. |
| **Working tree status** | The set of staged, unstaged, untracked, and conflicting changes in a repository. |
| **Ahead / Behind** | Commit divergence relative to the configured upstream branch. Only meaningful when an upstream exists. |
| **Upstream** | The remote-tracking branch a local branch is configured to compare against (for example `origin/main`). |
| **Detached HEAD** | A repository state where `HEAD` points to a commit rather than a branch. |
| **Unborn branch** | A branch with no commits yet (for example, a freshly initialized repository). |
| **Clean** | A repository with no staged, unstaged, untracked, or conflicting changes and no divergence. |
| **Dirty / Modified** | A repository with at least one staged, unstaged, or untracked change. |
| **Sync pending** | A repository that is ahead of, or behind, its upstream. |
| **Diverged** | A repository that is both ahead and behind its upstream. |
| **Conflict** | An unmerged path recorded by Git that requires user resolution. |
| **Executor** | The infrastructure component that runs the Git binary with an explicit working directory and structured error output. |
| **Adapter** | The infrastructure component that translates Git command output into domain types. |
| **Application service** | A use case coordinating domain entities and infrastructure ports (for example, refreshing a workspace). |
| **Port** | An interface defined by an inner layer and implemented by an outer layer. |
| **Refresh** | A read-only operation that re-reads repository state from Git. |
| **Write operation** | Any Git operation that mutates the repository (stage, commit, checkout, fetch that writes refs, pull, push). |
| **Worker pool** | The bounded set of goroutines used to refresh repository states concurrently. |
| **Generation / epoch** | A monotonic counter used to discard stale asynchronous results after navigation or a newer refresh. |
| **Workspace Command** | A developer-defined shell command associated with a workspace, executed with an explicit working directory and explicit user action (post-MVP). |
