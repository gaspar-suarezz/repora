# Contributing

Thanks for taking the time to improve Repora.

## Before you start

1. Start with [`AGENTS.md`](AGENTS.md) and [`docs/README.md`](docs/README.md).
2. Check the issue list before opening a new one.

## Requirements

- Go 1.24 or newer
- Git 2.30 or newer
- macOS or Linux

## Development

```bash
make build     # bin/repora
make run       # go run ./cmd/repora
make test      # go test ./...
make race      # go test -race ./...
```

Before opening a pull request, run the full gate:

```bash
make gate
```

`make gate` checks formatting, runs `go vet`, the test suite, and the race detector. CI
runs the same command on Linux and macOS.

## Project rules

- Keep Git execution out of the TUI layer; the layering is described in
  [`AGENTS.md`](AGENTS.md).
- Never introduce shell interpolation. Commands are argument vectors.
- Every write operation targets exactly one repository.
- Add tests with the change: unit tests for logic, integration tests against temporary
  repositories for anything touching Git.
- Update documentation when behavior or constraints change.

## Tests and fixtures

Integration tests create temporary repositories under `t.TempDir()` with an isolated Git
configuration. They never touch your real repositories. Never point a test at a path
outside a temporary directory.

## Using Repora on your own repositories

Write operations are safe by design, but during development use a disposable sandbox:

```bash
./scripts/dev-sandbox.sh
./bin/repora --config /tmp/repora-sandbox/config.yaml
```

## Pull requests

- Keep the change focused and describe the motivation.
- Include screenshots for visual changes.
- Mention any user-facing behavior change and update the docs.

## License

By contributing, you agree that your contributions are licensed under the MIT License.
