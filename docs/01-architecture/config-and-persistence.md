# Configuration and Persistence

**Status:** Defined
**Scope:** Configuration file location, schema, loading, validation, and safe writes.

See D-002, D-005, D-014, and D-015 in [`decisions.md`](decisions.md).

---

## Location

Resolution order, first match wins:

1. `--config <path>` flag.
2. `REPORA_CONFIG` environment variable.
3. `$XDG_CONFIG_HOME/repora/config.yaml`.
4. `~/.config/repora/config.yaml` (default on Linux and macOS).

Supporting files live next to the configuration file:

```text
~/.config/repora/
├── config.yaml
└── logs/
    └── repora.log        # optional, only when logging is enabled
```

Rules:

- The directory is created with mode `0700` if missing.
- `config.yaml` is created with mode `0600` if missing.
- The resolved path is shown in the Help screen and in startup errors.
- Repora MUST NOT write anywhere else except the OS temp directory for atomic-write
  staging.

---

## Schema — Version 1

```yaml
version: 1

repositories:
  - id: api
    name: API
    path: /Users/dev/code/product-x/api

  - id: web
    name: Web
    path: /Users/dev/code/product-x/web

  - id: infra
    name: Infrastructure
    path: /Users/dev/code/product-x/infra

workspaces:
  - id: product-x
    name: Product X
    description: Customer-facing product
    repositories:
      - api
      - web
      - infra

  - id: platform
    name: Platform
    repositories:
      - api

preferences:
  theme: dark
  refresh_interval: 30
  status_workers: 4
  default_workspace: product-x
```

Field reference:

| Path | Type | Required | Notes |
|---|---|---|---|
| `version` | int | Yes | Must be `1` |
| `repositories[].id` | string | Yes | Slug, unique |
| `repositories[].name` | string | Yes | Display name |
| `repositories[].path` | string | Yes | Absolute or `~`-prefixed |
| `workspaces[].id` | string | Yes | Slug, unique |
| `workspaces[].name` | string | Yes | Display name |
| `workspaces[].description` | string | No | Shown in Home |
| `workspaces[].repositories` | []string | No | Ordered; must reference existing repository IDs |
| `preferences.theme` | string | No | `dark` default |
| `preferences.refresh_interval` | int | No | Seconds, `0` disables, default `30`, min `5` if non-zero |
| `preferences.status_workers` | int | No | Default `4`, clamped to `1..16` |
| `preferences.default_workspace` | string | No | Must reference an existing workspace |

Additional rules:

- Workspace membership is a Repora-only concept: nothing is written into repositories.
- Unknown top-level keys are preserved on save (D-014).
- Unknown keys inside known objects are also preserved where feasible; if not feasible,
  they are preserved at the object level via raw node merging.
- No credentials, tokens, passwords, or remote URLs are stored.

---

## Loading

```text
1. Resolve path
2. Read file
   - missing file      → create default config, continue
   - permission error  → fail fast with path and OS error
3. Parse YAML into raw node tree
4. Check version
   - missing           → treat as 1 with a warning
   - 1                 → decode
   - > 1               → fail: "config created by a newer version"
5. Decode known fields into Config
6. Validate
7. Expand and canonicalize repository paths
8. Return Config + warnings
```

Warnings are non-fatal and surfaced in the UI (status line or Help):

- Repository path does not exist.
- Repository path exists but is not a Git repository.
- Duplicate repository path (second entry marked invalid).
- `default_workspace` references a missing workspace.
- `refresh_interval` out of range (clamped).

Errors are fatal and prevent entering the TUI:

- YAML syntax error (include line and column).
- Unsupported version.
- Duplicate IDs.
- Dangling workspace → repository reference.
- Unreadable file.

Fatal errors MUST include the config path and a short remediation hint.

---

## Path Handling

At load and at registration:

```text
expand leading "~" or "~/"
→ filepath.Abs
→ filepath.Clean
→ EvalSymlinks if it exists
→ git -C <path> rev-parse --show-toplevel
```

- The stored value is always the result of `rev-parse --show-toplevel`.
- Repository paths are stored absolute for v0.1. Relative paths are rejected.
- Home-directory expansion happens at load; the `~` form is never persisted.
- Paths containing newlines are rejected.
- A path that is a subdirectory of a repository is normalized to its root, and the user is
  informed when that happens.

---

## Saving

Saving MUST be atomic and crash-safe.

```text
1. Serialize Config → YAML node tree
2. Merge into the original raw node tree (preserve unknown fields)
3. Marshal with 2-space indentation
4. Create temp file in the same directory: config.yaml.tmp-<random>
5. Write bytes
6. fsync file
7. close
8. rename temp → config.yaml
9. fsync parent directory
```

Rules:

- Never write in place.
- Never truncate the original before the rename succeeds.
- On any failure, the original file remains untouched and the temp file is removed.
- File mode is `0600`; directory mode is `0700`.
- A failed save MUST surface as an error and MUST NOT leave in-memory state diverged from
  disk. If the write fails, the UI keeps the previous state and reports the failure.
- Confirmation dialogs that lead to a save MUST NOT close before the save result is known.

---

## Preserving Unknown Fields

Approach:

1. Decode the file into a `yaml.Node` document.
2. Decode the same bytes into typed structs for use.
3. On save, walk the node tree and replace only the keys Repora owns.
4. Keys not present in the schema are left untouched.

Consequences:

- A future `commands:` section authored by a newer Repora is not destroyed by an older
  binary.
- Hand-added comments are not guaranteed to survive (documented limitation), but unknown
  keys are.
- Round-trip tests must assert that a config with unknown keys is preserved byte-for-byte
  in those regions.

---

## Versioning and Migration

- `version` gates the schema.
- Migration functions are a registry: `func(cfg *yaml.Node) error` keyed by source
  version, applied in order.
- v0.1 only writes `version: 1`. The migration registry exists but contains no entries.
- If a config declares a version newer than the binary supports, Repora refuses to write
  and offers a read-only mode message. It MUST NOT downgrade silently.

---

## Concurrency and Multiple Instances

- Repora does not support concurrent writers to the same config.
- On startup, acquire an advisory lock file `config.lock` in the config directory.
- If the lock is held, start in read-only mode with a visible banner and disable all
  writes.
- The lock is released on clean exit; stale locks are detected via process liveness or an
  age heuristic and reported, never silently removed.

---

## Default Configuration

If `config.yaml` does not exist, Repora writes:

```yaml
version: 1

repositories: []

workspaces: []

preferences:
  theme: dark
  refresh_interval: 30
  status_workers: 4
```

The Home screen then shows an empty state with a call to action: create a workspace and
register a repository.

---

## Example: Minimal Valid Config

```yaml
version: 1

repositories:
  - id: api
    name: API
    path: ~/code/api

workspaces:
  - id: default
    name: Default
    repositories:
      - api
```

This is the smallest configuration Repora accepts with a non-empty workspace.

---

## Testing Requirements

- Load: missing file, empty file, valid v1, unsupported version, malformed YAML,
  duplicate IDs, dangling references.
- Paths: `~` expansion, trailing slash, symlink, subdirectory of a repository, missing
  path, non-repository path, path with spaces and unicode.
- Save: atomicity (simulate write failure), permission preservation, unknown-field
  preservation, concurrent-read consistency.
- Round trip: load → save → load produces an equivalent `Config`.
