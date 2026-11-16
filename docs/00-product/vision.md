# Vision

**Status:** Defined
**Scope:** Product strategy. No technical decisions.

---

## What Repora Is

Repora is a terminal-native application that centralizes Git repository management and
developer workspace orchestration.

It lets a developer group independent Git repositories into logical workspaces and
operate them through a single, keyboard-driven interface, while every repository keeps
its own branches, remotes, history, and configuration.

Repora combines two capabilities:

1. **Git management** — branches, staging, commits, diffs, history, remotes, and
   synchronization.
2. **Workspace orchestration** — grouping repositories into projects and, eventually,
   running project-specific development commands against them.

### Workspaces belong to Repora, not to Git

A workspace is a container that Repora owns. It is stored only in Repora's configuration
and has no equivalent in Git and no directory on disk.

The user decides exactly which repositories belong to a workspace — `work`, `personal`,
`client-a`, `everything` — and the same repository can belong to as many workspaces as
needed. Grouping never changes a repository's branches, remotes, hooks, or history; it only
changes what Repora shows together.

Git has no concept of a workspace, and Repora never writes workspace information into a
repository. See [`workspaces.md`](workspaces.md) for the authoritative model.

---

## The Problem

Developers working across multiple repositories — microservices, monorepos split into
repos, frontend/backend/infrastructure splits — spend a disproportionate amount of time
on mechanical context switching:

- Moving between directories to run Git commands.
- Opening several instances of Git tooling at once.
- Losing sight of the overall state of a project.
- Checking branch and synchronization status repository by repository.
- Running development commands in scattered terminals.
- Having no single place that answers: *what is the state of this project right now?*

The tooling market solves two halves of this problem separately: repository-level Git
clients are excellent at depth, and multi-repository viewers are excellent at breadth.
Repora's thesis is that the value is in the combination — breadth for awareness, depth
for action, without leaving the terminal.

---

## Vision

> **One workspace. Every repository. Full control.**

A developer opens Repora, sees every repository in the active project, understands in
seconds which ones are dirty, diverged, or broken, and acts on a specific one with full
confidence about which repository is being targeted.

---

## Principles

These principles are normative. When a design or implementation choice conflicts with
one of them, the principle wins or a new decision is recorded.

1. **Repository isolation.** Every Git operation targets an explicitly selected
   repository. There is no implicit "current repository" and no operation that silently
   fans out across repositories.
2. **Workspace awareness.** A workspace's aggregate state must be visible without
   opening each repository individually.
3. **Keyboard first.** Every important action has a keybinding. Mouse support is not a
   requirement.
4. **Fast interaction.** Minimize keystrokes and screen transitions for common tasks.
5. **Non-destructive by default.** Destructive or irreversible actions require explicit
   confirmation. Drafts and inputs are never lost silently.
6. **Native Git compatibility.** Repora delegates to the installed Git executable. It
   does not reimplement Git semantics.
7. **Minimal interface.** Show the information needed to decide and act. Nothing else.
8. **Extensibility without redesign.** New capabilities must fit the existing layers
   rather than require architectural rewrites.

---

## What Repora Is Not

- **Not a clone of an existing terminal Git client.** Other tools are a reference for
  keyboard interaction at most, not for layout, scope, or feature priority.
- **Not a Git reimplementation.** Repora never parses packfiles, computes diffs itself,
  or manages refs directly.
- **Not a GUI in a terminal.** It should feel like a native TUI, not a web layout
  squeezed into a grid.
- **Not a CI/CD tool or a hosted service.** It runs locally, on the developer's machine,
  against local repositories.
- **Not a replacement for the shell.** Terminal escape hatches may exist, but the
  product's value is a curated workflow, not a general shell.

---

## Non-Goals for v0.1

- Multi-user or shared configuration.
- Hosting-provider integrations (GitHub, GitLab, Bitbucket APIs).
- Remote repository creation.
- Merge conflict resolution tooling.
- A general-purpose process supervisor.

See [`mvp-scope.md`](mvp-scope.md) for the authoritative scope.

---

## Success Definition

Repora v0.1 succeeds if a developer with a workspace of four or more repositories can:

1. Launch Repora once.
2. Understand the state of the entire project at a glance.
3. Perform the day-to-day Git operations they would otherwise do in separate terminals.
4. Do all of it without leaving the application or worrying about which directory a
   command will run in.
