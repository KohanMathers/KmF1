# Contributing to KmF1

Thanks for looking at contributing. This document covers how to get set up locally and what's enforced on pull requests.

## Getting set up

```sh
git clone https://github.com/KohanMathers/KmF1.git
cd KmF1
python3 seed_test_data.py   # downloads real fixture captures the test suite needs
cd go
go build ./...
go test ./...
```

`seed_test_data.py` downloads real, historical F1 live-timing/Jolpica/MultiViewer captures into the various `testdata/` directories. These aren't committed (see `.gitignore`) to keep the repo lean, so run it once after cloning and again any time you see a test fail with a missing-file error.

## Before opening a PR

Run these from `go/` — they're exactly what CI checks, so a clean run locally means CI will pass too:

```sh
cd go
go build ./...
go vet ./...
gofmt -l .      # must print nothing
go test ./...
```

Also run `go mod tidy` inside `go/` if you added, removed, or upgraded a dependency, and commit the resulting `go.mod`/`go.sum` changes.

## What CI enforces

Every push to `main` and every pull request runs `.github/workflows/ci.yml`:

- `go build ./...` — the module builds.
- `go vet ./...` — no vet warnings.
- `gofmt -l .` — every file is gofmt'ed; fix with `gofmt -w .`.
- `go test ./...` — the full test suite passes, fixtures seeded automatically in CI.

Pull requests additionally run `.github/workflows/pr.yml`:

- `go.mod`/`go.sum` are tidy: CI runs `go mod tidy` and fails if that produces a diff. Run it yourself and commit the result if your change touches dependencies.
- Exported API has doc comments: every exported type, func, method, const, and var declared directly in the `go/` package root (the `kmf1` public API) must have a doc comment immediately above it, in normal [Go doc comment style](https://go.dev/doc/comment) (starts with the identifier's name, one line is fine, see the existing files for examples). This is checked by `.github/scripts/docguard.go`, which only looks at `go/*.go` (not `internal/`), so it won't block work that's purely internal.

  Run it yourself before pushing:

  ```sh
  go run .github/scripts/docguard.go go
  ```

## Style notes

- No comments unless they explain a non-obvious why (a subtle invariant, a workaround, something that would surprise a reader); the doc-comment rule above is the one required exception, since it's about documenting the public API's contract, not explaining implementation.
- Keep doc comments to one or two sentences, on a single unwrapped line (`gofmt`/`go vet` don't enforce wrapping, and long lines document just as well as wrapped ones, don't hand-wrap them).
- Prefer editing existing files over adding new ones; don't introduce abstractions a change doesn't need.
