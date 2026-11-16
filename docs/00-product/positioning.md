# Positioning

**Status:** Defined
**Scope:** Audience, use cases, and differentiation.

---

## Target Audience

### Primary

- Full-stack, backend, and frontend developers working in multi-repository projects.
- Developers working with microservices or split frontend/backend/infrastructure repos.
- Developers maintaining multiple independent client or product codebases.
- Terminal-first developers who already live in the shell.

### Secondary

- Platform and DevOps engineers who coordinate several infrastructure repositories.
- Consultants and contractors who switch between client projects frequently.
- Developers using `git worktree`-centric workflows (future support).

### Explicitly Not Targeted

- Users who prefer graphical Git clients and are not comfortable in a terminal.
- Teams needing a shared, server-side repository dashboard.
- Users looking for a full terminal multiplexer or IDE replacement.

---

## Core Use Cases

### 1. Project status at a glance

A developer starts the day, opens the workspace for a client project containing six
repositories, and immediately sees:

- which repositories are clean
- which have uncommitted work
- which are behind or ahead of their upstream
- which have merge conflicts blocking progress

No `cd`, no repeated `git status`, no lost terminal tabs.

### 2. Focused repository operation

The developer selects one repository and performs operations with an explicit target:
stage files, commit, switch branch, fetch, pull, push. The interface always states which
repository the action will affect.

### 3. Context switching between projects

The developer moves from Client A to Client B and gets the same consolidated view
without reconfiguring anything, because workspaces are logical groupings independent of
directory structure.

### 4. Shared repository, multiple contexts

The same repository can belong to several workspaces (for example `api` belongs to both
`product-x` and `platform-core`) without duplication of configuration or state.

---

## Differentiation

| Capability | Single-repository Git TUIs | Multi-repository status tools | Shell scripts | **Repora** |
|---|---|---|---|---|
| Deep single-repo Git workflow | Yes | No | Partial | Yes |
| Multi-repo consolidated view | No | Yes | Partial | Yes |
| Logical workspaces (not tied to a directory) | No | Rarely | No | Yes |
| Explicit per-repository operation targets | N/A | No | Manual | Yes |
| Concurrent status refresh with per-repo states | N/A | Sometimes | No | Yes |
| Workspace-specific development commands | No | No | Yes | Planned (v0.2) |
| Keyboard-first, minimal TUI | Yes | Varies | N/A | Yes |

### Against single-repository Git TUIs

Depth-first single-repository clients are strong at what they do. Repora does not compete
on the depth of Git features in v0.1. It competes on **breadth plus safe action**: seeing
many repositories and operating any one of them without ambiguity.

### Against multi-repository status tools

Most such tools are read-only dashboards or scripts that print a table. Repora closes
the loop: after seeing the state, the developer acts on it in the same interface.

### Against shell scripts and aliases

Scripts are programmable but brittle, non-interactive, and hard to extend with
confirmations, cancellation, and per-repository error reporting. They also scale poorly
when repositories have different remotes, branches, and states.

---

## Value Proposition

**For developers who coordinate multiple Git repositories, Repora is a terminal
workspace that shows the whole project's Git state and lets you act on any repository
individually — because switching directories and juggling terminals is the tax you pay
for a multi-repository architecture, and it should not be.**

---

## Positioning Constraints

- Repora MUST stay useful with a single repository, but it MUST NOT optimize for that
  case at the expense of multi-repository workflows.
- Repora MUST NOT depend on a hosted service, account, or network connection to be
  useful.
- Repora MUST NOT require repository restructuring; it adapts to how the developer
  already organizes code.
