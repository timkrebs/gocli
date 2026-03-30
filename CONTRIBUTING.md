# Contributing to gocli

Thank you for your interest in contributing. This document covers how to
report issues, propose features, and submit pull requests.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Reporting Bugs](#reporting-bugs)
- [Requesting Features](#requesting-features)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Commit Message Format](#commit-message-format)
- [Pull Request Process](#pull-request-process)
- [Testing](#testing)
- [Coding Conventions](#coding-conventions)

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).
By participating you agree to uphold it. Report unacceptable behaviour to the
maintainers via a private GitHub message.

---

## Reporting Bugs

1. Check that the bug has not already been reported in [Issues](https://github.com/timkrebs/gocli/issues).
2. Open a new issue using the **Bug Report** template.
3. Include a minimal, self-contained reproducer — ideally a failing test or a short `main.go`.
4. State the Go version (`go version`), OS, and gocli version.

---

## Requesting Features

1. Check [Issues](https://github.com/timkrebs/gocli/issues) and open PRs — it may already be in progress.
2. Open an issue using the **Feature Request** template before writing code for non-trivial changes.
3. Describe the problem you are solving, not just the solution you have in mind.

---

## Development Setup

**Requirements:** Go 1.23+ and `make`.

```bash
# Clone
git clone https://github.com/timkrebs/gocli.git
cd gocli

# Download dependencies
make deps

# Run tests (standard + race detector)
make test
make testrace

# Run linter (requires golangci-lint)
make lint

# Generate coverage report
make coverage
```

Install `golangci-lint` if you do not have it:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

---

## Making Changes

1. **Fork** the repository and create a branch from `main`:

   ```bash
   git checkout -b feat/my-feature
   ```

2. Make your changes. Keep each PR focused on a single concern.

3. Add or update tests so that coverage does not decrease.

4. Run the full suite before pushing:

   ```bash
   make test testrace lint
   ```

5. Push and open a pull request against `main`.

---

## Commit Message Format

gocli uses **Conventional Commits** so that GoReleaser can generate the
changelog automatically. Every commit on `main` (including squashed PR
commits) must follow this format:

```
<type>(<scope>): <short summary>

[optional body]

[optional footer(s)]
```

### Types

| Type | When to use |
|------|-------------|
| `feat` | New feature or public API addition |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `refactor` | Code change that is neither a fix nor a feature |
| `perf` | Performance improvement |
| `test` | Adding or correcting tests |
| `chore` | Build, tooling, dependency updates |
| `ci` | CI/CD configuration changes |

Append `!` after the type/scope to signal a **breaking change**:
`feat!: remove deprecated Foo method`.

### Examples

```
feat(ui): add UiWriterLevel for per-level output routing
fix(cli): detect --no-color before first output write
docs: add Extending gocli section to README
chore: bump golangci-lint to v2
feat!: remove NoopCommand — replaced by noopParentCommand
```

---

## Pull Request Process

1. Fill in the PR template.
2. Ensure all CI checks pass (test, lint, security).
3. At least one approval from a maintainer is required to merge.
4. PRs are merged by **squash and merge** — the squash commit message
   becomes the `CHANGELOG` entry, so write it in Conventional Commits format.

---

## Testing

- All new behaviour must be covered by tests in the same PR.
- Use the race detector: `go test -race ./...`.
- Tests must pass on all supported Go versions (see the matrix in `test.yml`).
- Use `NewMockUi()` (not `new(MockUi)` or `&MockUi{}`) to avoid nil panics.
- Use `t.Setenv` for environment variable manipulation — it resets automatically.

---

## Coding Conventions

- Follow standard Go formatting (`gofmt` / `goimports`).
- Public API changes require updated `godoc` comments.
- Keep `Synopsis()` text under 50 characters.
- Do not use `interface{}` — use `any` (Go 1.18+).
- Avoid `init()` functions in non-test code; prefer explicit initialization.
- Return named errors or wrap with `fmt.Errorf("context: %w", err)`.
- Do not introduce external dependencies without discussion in an issue first.
  gocli aims to keep its dependency footprint small.
