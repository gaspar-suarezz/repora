# Workspaces

**Status:** Defined — normative
**Scope:** The workspace concept. This document is authoritative for what a workspace is
and is not.

---

## Definition

A **workspace** is a container owned by Repora that groups the repositories the user
chooses.

It exists **only inside Repora's own configuration**. It is not a Git concept, it is not a
filesystem concept, and it has no representation on disk other than its entry in
`~/.config/repora/config.yaml`.

> Git does not know that workspaces exist. Repora does not ask Git to create, group, or
> relate anything.

---

## The Model

```text
Repora configuration
├── repositories[]              registered globally, one entry per local repository
│      id, name, path
│
└── workspaces[]                user-defined containers
       id, name, description
       repositories: [repository ids]   ← explicit, ordered membership
```

Two independent layers:

1. **Registration** — a local Git repository is registered once, globally.
2. **Association** — that registration is linked to any number of workspaces.

A workspace never owns a repository. It only references it.

---

## Normative Rules

1. **Repora owns workspaces.** No workspace data is written to a repository and no Git
   command is aware of workspaces.
2. **Membership is explicit.** A repository belongs to a workspace only because the user
   added it. Nothing is discovered, inferred, or inherited.
3. **A repository can belong to many workspaces.** The same repository may appear in
   `work` and `personal` at the same time, with no duplication and no second clone.
4. **Workspaces never change repository behavior.** They do not affect branches, remotes,
   hooks, history, configuration, or synchronization. They are a view, not a modifier.
5. **Repositories in a workspace may live anywhere.** There is no required common parent
   directory, no shared drive, and no naming convention.
6. **A workspace may be empty.** Creating a workspace before choosing repositories is
   valid.
7. **Workspace order is meaningful.** The order of repository IDs defines the display
   order in the dashboard.
8. **Deleting a workspace is non-destructive.** It removes the workspace entry only.
   Repositories remain registered and remain on disk.
9. **Unregistering a repository is destructive to configuration only.** It removes the
   repository from the global list and from every workspace, after confirmation. It never
   touches the repository on disk.
10. **IDs are immutable; names are editable.** A workspace keeps its ID when renamed.
11. **Repository state is per repository, not per workspace.** If a repository belongs to
    two workspaces, both dashboards show the same state, because it is the same working
    tree.

---

## Example: Work and Personal

The motivating case. One developer, two contexts, overlapping repositories.

```yaml
version: 1

repositories:
  - id: api
    name: API
    path: /Users/you/code/company/api

  - id: web
    name: Web
    path: /Users/you/code/company/web

  - id: dotfiles
    name: Dotfiles
    path: /Users/you/dotfiles

  - id: side-project
    name: Side Project
    path: /Users/you/code/side-project

workspaces:
  - id: work
    name: Work
    description: Day job
    repositories: [api, web]

  - id: personal
    name: Personal
    description: My own projects
    repositories: [side-project, dotfiles]

  - id: everything
    name: Everything
    repositories: [api, web, side-project, dotfiles]
```

Observations:

- `api` is not "inside" the Work workspace physically; it is only referenced.
- `personal` contains no company code, even though both workspaces live in the same
  configuration file.
- `everything` shares repositories with both workspaces without duplicating them.
- Removing `web` from `work` leaves it available in `everything` and untouched on disk.
- Repositories form two different directory trees; Repora does not care.

---

## What a Workspace Is Not

| Concept | Repora workspace |
|---|---|
| **Git repository** | A repository is code under version control. A workspace is a list of references to such repositories. |
| **Monorepo** | A monorepo puts many projects in one repository. A workspace spans many separate repositories. |
| **Submodule** | A submodule is a Git relationship stored in a repository. A workspace is a Repora-only relationship. |
| **`git worktree`** | A worktree is an additional checkout of one repository. Workspaces group several repositories. Future Repora worktree support will sit *inside* a repository, not replace workspaces. |
| **Project directory** | A directory is a filesystem location. Workspace repositories can be scattered across the disk. |
| **Editor workspace file** | An editor workspace (for example a `.code-workspace`) is a file inside a project. A Repora workspace is an entry in Repora's configuration and can include unrelated projects. |
| **Branch or tag** | Branches are Git state inside one repository. Workspaces are independent of branches, and repositories in a workspace may be on entirely different branches. |

---

## Lifecycle

```text
create workspace
  → optionally add a description
  → register a repository (global, once)
  → associate the repository with the workspace
  → open the workspace, inspect state, operate per repository
  → remove an association when the repository is no longer relevant
  → delete the workspace when the context is gone
```

Rules at each step:

- Creating a workspace requires only a name. The ID is derived as a slug.
- Registering a repository validates the path and resolves the Git root.
- Associating an already-registered repository never duplicates it.
- Removing an association never deletes the registration.
- Deleting a workspace never deletes repositories or their registrations.

---

## Consequences for the Rest of the Product

These follow directly from the model and are reflected in the architecture:

1. **Configuration is the only source of workspace truth.** There is no scanning of the
   filesystem to rebuild workspaces, and no Git query that returns membership.
2. **Repository identity is global.** State, locks, and operations are keyed by repository
   ID, so a repository behaves identically in every workspace it belongs to.
3. **Per-repository isolation is unchanged by grouping.** Being in a workspace does not
   grant an operation a broader scope; every write still targets exactly one repository.
4. **Aggregate status is derived.** A workspace summary is computed from the member
   repositories' snapshots; it is never stored.
5. **Exporting is trivial.** A workspace is just a list of IDs plus display metadata, which
   is why templates (v1.0) and sharing can be considered later without a schema redesign.

---

## Non-Goals

- Nested workspaces or workspace hierarchies.
- Workspace-level branch, remote, or hook configuration.
- Automatic membership from directory structure or Git remotes.
- Workspaces as a Git concept (no `.repora` file inside repositories in v0.1).
- Workspace templates or import/export (v1.0 and later).

---

## Acceptance Criteria

- [ ] A user can create a workspace without any repository existing yet.
- [ ] A user can create `work` and `personal` and place different repositories in each.
- [ ] The same repository can be associated with both workspaces.
- [ ] Removing a repository from `work` leaves it in `personal` and on disk.
- [ ] Deleting `personal` leaves every repository registered and untouched.
- [ ] Renaming a workspace changes only its display name.
- [ ] Opening either workspace shows the same state for a shared repository.
