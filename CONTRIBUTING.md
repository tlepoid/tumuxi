# Contributing to tumuxi

## Development

```bash
git clone https://github.com/tlepoid/tumuxi.git
cd tumuxi
./scripts/install-hooks.sh
make run
```

Run the local checks that mirror CI:

```bash
make devcheck
```

`make devcheck` is the required pre-PR gate: it runs vet, tests, and lint (including file-length checks).

`golangci-lint` is required locally. Install instructions: https://golangci-lint.run/welcome/install/

For style-only cleanup, run:

```bash
make fmt
```

Before opening larger PRs, also run strict ratcheted lint on changed code:

```bash
make lint-strict-new
```

Pull requests are CI-gated (automated). For local confidence before opening a PR:

- always: `make devcheck`, `make lint-strict-new`
- if touching `internal/ui/`, `internal/vterm/`, or `cmd/tumuxi-harness/`: `make harness-presets`
- if touching `internal/tmux/`, `internal/e2e/`, or `internal/pty/`: `go test ./internal/tmux ./internal/e2e`

Architecture references:

- `internal/app/ARCHITECTURE.md`
- `internal/app/MESSAGE_FLOW.md`

## Platform Support

tumuxi is developed and tested on **Linux and macOS** only. CI runs exclusively on Ubuntu.

Windows stub files (`*_windows.go`) exist to allow the package to compile on Windows, but Windows support is **experimental and untested**. PRs that improve Windows compatibility are welcome, but Windows-only regressions will not block Linux/macOS releases.

## Release

Versioning follows SemVer and tags are `vX.Y.Z`. Pushing a tag triggers the GitHub Actions release job.

Fast path:

```bash
git pull --ff-only
make release VERSION=v0.0.5
```

Manual steps:

```bash
make release-check
git tag -a v0.0.5 -m "v0.0.5"
git push origin v0.0.5
```

Notes:

- `make release` runs `release-check`, creates an annotated tag, and pushes it. The worktree must be clean.
- Release builds use the commit timestamp for `main.date`, which keeps the timestamp deterministic for a given commit. If you need strict bit-for-bit reproducibility, consider adding `-trimpath` and a stable build ID to the build flags.

### Homebrew tap

The Homebrew tap lives in `tlepoid/homebrew-tumuxi` and auto-bumps the formula after a release.

- After `make release VERSION=vX.Y.Z`, the tap workflow updates `Formula/tumuxi.rb` (daily at 06:00 UTC).
- To update immediately, run the **Bump tumuxi formula** workflow in the tap repo.
- Users upgrade with `brew upgrade tumuxi`.
