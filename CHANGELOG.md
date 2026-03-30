# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- `ExitCodeSuccess`, `ExitCodeError`, `ExitCodeUsage`, `ExitCodeNotFound` constants for named exit codes following POSIX conventions.
- `ExitError` type that bundles a human-readable message and an exit code, with `NewExitError` constructor and `ExitCodeOf` helper.
- `CommandDeprecated` interface — implement `DeprecationMessage() string` to have the CLI automatically print a warning before dispatching the command.
- `NoColorFlag` field on `CLI` (default `"no-color"`) — exposes a `--no-color` / `-no-color` global flag that disables ANSI color output via `fatih/color`. The `NO_COLOR` env var (https://no-color.org) is respected automatically.
- `EnvDefault(key, fallback string) string` — returns an env var value, falling back to the default when unset or empty.
- `EnvDefaultBool(key string, fallback bool) bool` — same for boolean env vars.
- `EnvDefaultInt(key string, fallback int) int` — same for integer env vars.
- `UiWriterLevel` type and constants (`LevelInfo`, `LevelOutput`, `LevelWarn`, `LevelError`) — `UiWriter.Level` controls which `Ui` method each write is routed to (default `LevelInfo` preserves backward compatibility).

### Changed
- Auto-generated parent commands for nested subcommand trees now use an internal `noopParentCommand` type instead of the public `MockCommand` test helper, removing a coupling between production and test code.
- README field reference table now documents `VersionFunc` and `AutocompleteNoDefaultFlags`.
- README now includes a comprehensive "Extending gocli" section with concrete examples for every extension interface.

---

## [0.1.0] — Initial Release

### Added
- Flat and nested subcommands with O(log n) radix-tree routing.
- `CommandV2` interface for context-aware execution with cancellation and deadline propagation.
- `CommandAliases` for hidden command aliases.
- `HiddenCommands` to exclude commands from help and autocomplete.
- Fuzzy "did you mean" suggestions (Levenshtein distance ≤ 2) for mistyped commands.
- `BeforeRun` / `AfterRun` middleware hooks.
- Shell autocompletion for bash, zsh, and fish via `posener/complete`.
- `CommandAutocomplete` interface for fine-grained flag and argument completion.
- `CommandHelpTemplate` interface for Sprig-powered per-command help templates.
- `BasicUi`, `ConcurrentUi`, `ColoredUi`, `PrefixedUi`, and `MockUi` UI implementations.
- `UiWriter` adapter for routing `io.Writer` output through the `Ui` system.

[Unreleased]: https://github.com/timkrebs/gocli/compare/HEAD...HEAD
[0.1.0]: https://github.com/timkrebs/gocli/releases/tag/v0.1.0
