# gocli

A powerful, extensible CLI framework for Go that makes building multi-command
command-line applications easy. Inspired by the CLI patterns used in HashiCorp tools.

[![Test](https://github.com/timkrebs/gocli/actions/workflows/test.yml/badge.svg)](https://github.com/timkrebs/gocli/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/timkrebs/gocli.svg)](https://pkg.go.dev/github.com/timkrebs/gocli)
[![Go 1.23+](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org/dl/)

## Features

- **Flat and nested subcommands** with O(log n) radix-tree routing
- **Context-aware execution** via `CommandV2` with cancellation and deadline propagation
- **Command aliases** hidden from help and autocomplete
- **Fuzzy "did you mean" suggestions** for mistyped commands
- **BeforeRun / AfterRun middleware hooks** around every dispatch
- **Shell autocompletion** for bash, zsh, and fish via posener/complete
- **Composable UI layer** — colored, concurrent, prefixed, and mock implementations
- **Sprig-powered help templates** for rich per-command help pages
- **Zero panics** — all error paths handled gracefully

## Installation

```bash
go get github.com/timkrebs/gocli
```

Requires Go 1.23 or later.

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "os"

    cli "github.com/timkrebs/gocli"
)

func main() {
    c := cli.NewCLI("myapp", "1.0.0")
    c.Args        = os.Args[1:]
    c.HelpWriter  = os.Stdout
    c.Commands = map[string]cli.CommandFactory{
        "greet": func() (cli.Command, error) {
            return &GreetCommand{}, nil
        },
    }

    exitStatus, err := c.Run()
    if err != nil {
        log.Println(err)
    }
    os.Exit(exitStatus)
}

type GreetCommand struct{}

func (c *GreetCommand) Help() string     { return "Prints a greeting to stdout." }
func (c *GreetCommand) Synopsis() string { return "Print a greeting" }
func (c *GreetCommand) Run(args []string) int {
    fmt.Println("Hello, world!")
    return 0
}
```

```
$ myapp greet
Hello, world!

$ myapp -h
Usage: myapp [--version] [--help] <command> [<args>]

Available commands are:
    greet    Print a greeting

$ myapp greet -h
Prints a greeting to stdout.
```

## CLI Configuration

### Constructor

```go
c := cli.NewCLI("myapp", "1.2.3")
```

Sets sensible defaults: `BasicHelpFunc`, `Autocomplete: true`, `HelpWriter: os.Stderr`.

### Full field reference

| Field | Type | Description |
|-------|------|-------------|
| `Args` | `[]string` | Command-line args, typically `os.Args[1:]` |
| `Commands` | `map[string]CommandFactory` | Registered subcommands |
| `HiddenCommands` | `[]string` | Commands excluded from help and autocomplete |
| `CommandAliases` | `map[string]string` | Alias → canonical name mapping |
| `Name` | `string` | Binary name (required for autocomplete) |
| `Version` | `string` | Version string printed with `--version` |
| `VersionFunc` | `func() string` | Called for version when `Version` is empty; ignored when `Version` is set |
| `HelpFunc` | `HelpFunc` | Top-level help generator |
| `HelpWriter` | `io.Writer` | Help output destination (default: `os.Stderr`; recommend `os.Stdout`) |
| `ErrorWriter` | `io.Writer` | Error output destination (default: same as `HelpWriter`; recommend `os.Stderr`) |
| `BeforeRun` | `func(name string, args []string) int` | Pre-dispatch hook; non-zero return aborts |
| `AfterRun` | `func(name string, args []string, exitCode int)` | Post-dispatch hook |
| `Autocomplete` | `bool` | Enable shell autocomplete (default `true` via `NewCLI`) |
| `AutocompleteInstall` | `string` | Flag to install autocomplete (default: `autocomplete-install`) |
| `AutocompleteUninstall` | `string` | Flag to uninstall autocomplete (default: `autocomplete-uninstall`) |
| `AutocompleteNoDefaultFlags` | `bool` | Suppress default `-help` / `-version` flags from autocomplete output |
| `AutocompleteGlobalFlags` | `complete.Flags` | Global flags exposed to autocomplete |

### Entrypoints

```go
// Basic run
exitCode, err := c.Run()

// With context — cancellation and deadlines flow into CommandV2 commands
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
exitCode, err := c.RunContext(ctx)
```

## Commands

### Command interface

Every subcommand implements three methods:

```go
type Command interface {
    Help() string
    Run(args []string) int
    Synopsis() string
}
```

| Method | Description |
|--------|-------------|
| `Help()` | Long-form help text: usage line, description, flags |
| `Synopsis()` | One-line description ≤ 50 chars shown in command listings |
| `Run(args []string) int` | Execute and return exit code |

Return the sentinel value `cli.RunResultHelp` from `Run` to display the command's
help text and exit with code 1.

### CommandV2 — context-aware execution

Implement `CommandV2` when a command needs to respect cancellation or deadlines.
The CLI automatically calls `RunContext` for commands that implement this interface,
falling back to `Run` for plain `Command` implementations.

```go
type CommandV2 interface {
    Command
    RunContext(ctx context.Context, args []string) int
}
```

```go
type ServeCommand struct{}

func (c *ServeCommand) Help() string     { return "Starts the server." }
func (c *ServeCommand) Synopsis() string { return "Start the server" }
func (c *ServeCommand) Run(args []string) int {
    return c.RunContext(context.Background(), args)
}

func (c *ServeCommand) RunContext(ctx context.Context, args []string) int {
    srv := startServer()
    <-ctx.Done() // blocks until SIGINT, timeout, etc.
    srv.Shutdown(context.Background())
    return 0
}
```

Pass a context with signal handling for graceful shutdown:

```go
ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer cancel()
exitCode, err := c.RunContext(ctx)
```

### Nested subcommands

Register commands with space-separated keys. Missing parent commands are
auto-created and display help listing their children automatically.

```go
c.Commands = map[string]cli.CommandFactory{
    "server":       serverCmdFactory,
    "server start": serverStartFactory,
    "server stop":  serverStopFactory,
    "db migrate":   dbMigrateFactory,  // "db" parent is auto-created
}
```

```
$ myapp server start    # → ServerStartCommand.Run
$ myapp server stop     # → ServerStopCommand.Run
$ myapp server          # auto-generated: lists "start" and "stop"
$ myapp db migrate      # → DbMigrateCommand.Run
$ myapp db              # auto-generated: lists "migrate"
```

Longest-prefix matching is used, so `myapp server` dispatches to the most
specific registered key that matches the provided arguments.

### Command aliases

Map alias names to canonical command names. Aliases function identically to the
canonical command but are hidden from help listings and autocomplete output.

```go
c.CommandAliases = map[string]string{
    "rm":  "delete",
    "ls":  "list",
}
```

```
$ myapp rm        # identical to: myapp delete
$ myapp ls        # identical to: myapp list
$ myapp -h        # only "delete" and "list" appear — aliases are hidden
```

### Hidden commands

Exclude commands from help and autocomplete while keeping them fully functional:

```go
c.HiddenCommands = []string{"debug-internal", "legacy-cmd"}
```

### Fuzzy "did you mean" suggestions

When a user mistyps a command name, gocli suggests close matches automatically
using Levenshtein distance (threshold ≤ 2 edits):

```
$ myapp deleet
...
Did you mean one of these?
    delete
```

No configuration needed — suggestions fire automatically on any unknown command.

### BeforeRun and AfterRun hooks

Add middleware logic that runs around every command dispatch:

```go
c.BeforeRun = func(name string, args []string) int {
    if !isAuthenticated() {
        fmt.Fprintln(os.Stderr, "error: not authenticated — run 'myapp login' first")
        return 1   // non-zero aborts execution; AfterRun is NOT called
    }
    log.Printf("dispatch: %s %v", name, args)
    return 0
}

c.AfterRun = func(name string, args []string, exitCode int) {
    metrics.Record(name, exitCode)
}
```

- `BeforeRun` fires before dispatch. A non-zero return becomes the process exit code.
- `AfterRun` always fires after the command returns, including on non-zero exit codes.
- Neither hook is called for built-in CLI handling (help, version, autocomplete).

## Help Customisation

### Custom top-level HelpFunc

```go
c.HelpFunc = func(commands map[string]cli.CommandFactory) string {
    var b strings.Builder
    fmt.Fprintf(&b, "myapp v%s\n\nCommands:\n", version)
    for name, factory := range commands {
        cmd, _ := factory()
        fmt.Fprintf(&b, "  %-14s %s\n", name, cmd.Synopsis())
    }
    return b.String()
}
```

`FilteredHelpFunc` wraps any `HelpFunc` to show only a specific subset of commands:

```go
c.HelpFunc = cli.FilteredHelpFunc(
    []string{"deploy", "rollback"},
    cli.BasicHelpFunc("myapp"),
)
```

### Per-command help templates

Implement `CommandHelpTemplate` to use a `text/template` template for the
command's `--help` output. All [Sprig](https://masterminds.github.io/sprig/)
template functions are available.

```go
func (c *DeployCommand) HelpTemplate() string {
    return `{{ .Help }}
{{- if gt (len .Subcommands) 0 }}

Subcommands:
{{- range .Subcommands }}
    {{ .NameAligned }}  {{ .Synopsis }}
{{- end }}
{{- end }}
`
}
```

**Available template data:**

| Key | Type | Description |
|-----|------|-------------|
| `.Name` | `string` | CLI binary name |
| `.SubcommandName` | `string` | The matched subcommand key |
| `.Help` | `string` | Output of `command.Help()` |
| `.Subcommands` | `[]map` | Child subcommands (nested CLIs only) |

Each `.Subcommands` entry exposes `.Name`, `.NameAligned`, `.Help`, `.Synopsis`.

## Shell Autocompletion

Autocompletion supports bash, zsh, and fish via
[posener/complete](https://github.com/posener/complete). Subcommand completion
is automatic. For argument and flag completion, implement `CommandAutocomplete`:

```go
type CommandAutocomplete interface {
    AutocompleteArgs() complete.Predictor
    AutocompleteFlags() complete.Flags
}
```

```go
func (c *DeployCommand) AutocompleteArgs() complete.Predictor {
    return complete.PredictDirs("*")
}

func (c *DeployCommand) AutocompleteFlags() complete.Flags {
    return complete.Flags{
        "--env":     complete.PredictSet("staging", "production"),
        "--dry-run": complete.PredictNothing,
    }
}
```

**Install / uninstall autocomplete** (user runs once):

```
$ myapp -autocomplete-install
$ myapp -autocomplete-uninstall
```

The flag names are configurable via `AutocompleteInstall` and `AutocompleteUninstall`.

## UI Layer

gocli provides a composable, interface-based system for all terminal interaction
that makes testing straightforward.

### Ui interface

```go
type Ui interface {
    Ask(string) (string, error)       // prompt for input
    AskSecret(string) (string, error) // prompt without echo (passwords)
    Output(string)                    // normal stdout
    Info(string)                      // informational (same writer as Output)
    Error(string)                     // error messages
    Warn(string)                      // warnings
}
```

### BasicUi

Direct output to `io.Writer` instances:

```go
ui := &cli.BasicUi{
    Reader:      os.Stdin,
    Writer:      os.Stdout,
    ErrorWriter: os.Stderr,
}

ui.Output("Deploying...")
ui.Error("Connection refused")

name, _ := ui.Ask("Username:")
pass, _ := ui.AskSecret("Password:")
```

> `BasicUi` is not concurrency-safe on its own. Wrap it with `ConcurrentUi`
> when output comes from multiple goroutines.

### ConcurrentUi

Wraps any `Ui` with a mutex for goroutine-safe output:

```go
ui := &cli.ConcurrentUi{
    Ui: &cli.BasicUi{
        Writer:      os.Stdout,
        ErrorWriter: os.Stderr,
    },
}
```

### ColoredUi

Applies ANSI colors per output level. Colors are automatically disabled when
stdout is not a TTY.

```go
ui := &cli.ColoredUi{
    OutputColor: cli.UiColorNone,
    InfoColor:   cli.UiColorGreen,
    ErrorColor:  cli.UiColorRed,
    WarnColor:   cli.UiColorYellow,
    Ui:          baseUi,
}
ui.Error("Build failed!") // printed in red
ui.Info("Done.")          // printed in green
```

**Available colors:** `UiColorNone`, `UiColorRed`, `UiColorGreen`, `UiColorYellow`,
`UiColorBlue`, `UiColorMagenta`, `UiColorCyan`.

Set `.Bold = true` for bold output:

```go
boldRed := cli.UiColor{Code: int(color.FgHiRed), Bold: true}
```

### PrefixedUi

Prepends a fixed string to each output level:

```go
ui := &cli.PrefixedUi{
    InfoPrefix:  "INFO:  ",
    ErrorPrefix: "ERROR: ",
    WarnPrefix:  "WARN:  ",
    Ui:          baseUi,
}
ui.Error("disk full") // prints: ERROR: disk full
```

### UiWriter

Adapts a `Ui` to an `io.Writer`, forwarding each written line as an `Info` call.
Useful for redirecting standard loggers into the UI system:

```go
ui := cli.NewMockUi()
log.SetOutput(&cli.UiWriter{Ui: ui})
log.Println("server started") // routed through ui.Info(...)
```

### Composing layers

Layers can be stacked in any order:

```go
ui := &cli.ConcurrentUi{
    Ui: &cli.ColoredUi{
        ErrorColor: cli.UiColorRed,
        WarnColor:  cli.UiColorYellow,
        Ui: &cli.PrefixedUi{
            ErrorPrefix: "[ERROR] ",
            WarnPrefix:  "[WARN]  ",
            Ui: &cli.BasicUi{
                Writer:      os.Stdout,
                ErrorWriter: os.Stderr,
            },
        },
    },
}
```

## Testing

### MockUi

Captures all UI output in-memory for assertions. Always use the `NewMockUi()`
constructor — direct struct initialisation will cause a nil panic:

```go
func TestMyCommand(t *testing.T) {
    ui := cli.NewMockUi()
    cmd := &MyCommand{Ui: ui}

    code := cmd.Run([]string{"--flag", "value"})

    if code != 0 {
        t.Fatalf("exit %d\nstderr:\n%s", code, ui.ErrorWriter)
    }
    if !strings.Contains(ui.OutputWriter.String(), "expected text") {
        t.Errorf("unexpected output:\n%s", ui.OutputWriter)
    }
}
```

### MockCommand

A minimal `Command` for testing CLI routing and dispatch:

```go
mock := &cli.MockCommand{
    RunResult:    0,
    HelpText:     "long help text",
    SynopsisText: "short synopsis",
}

c := &cli.CLI{
    Args: []string{"serve"},
    Commands: map[string]cli.CommandFactory{
        "serve": func() (cli.Command, error) { return mock, nil },
    },
}

code, err := c.Run()
// mock.RunCalled == true
// mock.RunArgs  == []string{}
```

### MockCommandV2

For testing context-aware command dispatch:

```go
mock := &cli.MockCommandV2{RunContextResult: 0}
ctx  := context.Background()

c := &cli.CLI{
    Args: []string{"serve"},
    Commands: map[string]cli.CommandFactory{
        "serve": func() (cli.Command, error) { return mock, nil },
    },
}
c.RunContext(ctx)

// mock.RunContextCalled == true
// mock.RunContextCtx    == ctx   (exact context forwarded)
// mock.RunContextArgs   == []string{}
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"

    cli "github.com/timkrebs/gocli"
)

func main() {
    // Propagate SIGINT / SIGTERM into context so CommandV2 commands can
    // shut down cleanly.
    ctx, cancel := signal.NotifyContext(
        context.Background(),
        syscall.SIGINT, syscall.SIGTERM,
    )
    defer cancel()

    ui := &cli.ConcurrentUi{
        Ui: &cli.ColoredUi{
            InfoColor:  cli.UiColorGreen,
            ErrorColor: cli.UiColorRed,
            WarnColor:  cli.UiColorYellow,
            Ui: &cli.BasicUi{
                Reader:      os.Stdin,
                Writer:      os.Stdout,
                ErrorWriter: os.Stderr,
            },
        },
    }

    c := cli.NewCLI("myapp", "1.0.0")
    c.Args        = os.Args[1:]
    c.HelpWriter  = os.Stdout
    c.ErrorWriter = os.Stderr
    c.Commands = map[string]cli.CommandFactory{
        "serve":    func() (cli.Command, error) { return &ServeCommand{ui: ui}, nil },
        "db":       func() (cli.Command, error) { return &DbCommand{ui: ui}, nil },
        "db apply": func() (cli.Command, error) { return &DbApplyCommand{ui: ui}, nil },
    }
    c.CommandAliases = map[string]string{
        "start": "serve",
    }
    c.HiddenCommands = []string{"db"}
    c.BeforeRun = func(name string, args []string) int {
        ui.Info(fmt.Sprintf("→ running: %s", name))
        return 0
    }
    c.AfterRun = func(name string, args []string, code int) {
        if code != 0 {
            ui.Warn(fmt.Sprintf("command %q exited with code %d", name, code))
        }
    }

    exitCode, err := c.RunContext(ctx)
    if err != nil {
        log.Println(err)
    }
    os.Exit(exitCode)
}
```

## Extending gocli

gocli is built around small, composable interfaces. Extending the framework
means implementing one or more of these interfaces on your command structs.

### Adding a new command

Implement the `Command` interface and register it in the `Commands` map:

```go
type BuildCommand struct {
    Ui cli.Ui
}

func (c *BuildCommand) Synopsis() string { return "Build the project" }

func (c *BuildCommand) Help() string {
    return `Usage: myapp build [options]

  Build the project from source.

Options:
  -o, --output PATH   Write binary to PATH (default: ./bin/myapp)
  -v, --verbose       Enable verbose output
`
}

func (c *BuildCommand) Run(args []string) int {
    fs := flag.NewFlagSet("build", flag.ContinueOnError)
    output := fs.String("o", "./bin/myapp", "output path")
    verbose := fs.Bool("v", false, "verbose output")
    if err := fs.Parse(args); err != nil {
        return cli.RunResultHelp
    }

    c.Ui.Info(fmt.Sprintf("Building → %s", *output))
    if *verbose {
        c.Ui.Output("verbose mode enabled")
    }
    return 0
}
```

Register it:

```go
c.Commands = map[string]cli.CommandFactory{
    "build": func() (cli.Command, error) {
        return &BuildCommand{Ui: ui}, nil
    },
}
```

### Making a command context-aware (CommandV2)

Implement `CommandV2` when a command runs a long-lived process and must
support cancellation (e.g. via Ctrl-C or a deadline):

```go
type ServeCommand struct{ Ui cli.Ui }

func (c *ServeCommand) Synopsis() string { return "Run the HTTP server" }
func (c *ServeCommand) Help() string     { return "Usage: myapp serve [--port PORT]" }

// Run satisfies the plain Command interface and delegates to RunContext.
func (c *ServeCommand) Run(args []string) int {
    return c.RunContext(context.Background(), args)
}

func (c *ServeCommand) RunContext(ctx context.Context, args []string) int {
    srv := startHTTPServer()
    c.Ui.Info("server started")

    <-ctx.Done() // blocks until SIGINT, timeout, or parent cancel

    c.Ui.Warn("shutting down…")
    srv.Shutdown(context.Background())
    return 0
}
```

Wire up signal propagation in `main`:

```go
ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer cancel()
exitCode, err := c.RunContext(ctx)
```

### Adding shell autocompletion to a command (CommandAutocomplete)

Implement `CommandAutocomplete` to provide flag and argument completions
beyond the default subcommand completion:

```go
type DeployCommand struct{ Ui cli.Ui }

func (c *DeployCommand) AutocompleteArgs() complete.Predictor {
    // Complete positional args with local directory names
    return complete.PredictDirs("*")
}

func (c *DeployCommand) AutocompleteFlags() complete.Flags {
    return complete.Flags{
        "--env":     complete.PredictSet("staging", "production"),
        "--dry-run": complete.PredictNothing,
        "--tag":     complete.PredictAnything,
    }
}
```

No changes to CLI setup are required — gocli detects the interface automatically.

### Customising per-command help (CommandHelpTemplate)

Implement `CommandHelpTemplate` to control how a command's `--help` output
is rendered. The template uses `text/template` syntax and all
[Sprig](https://masterminds.github.io/sprig/) functions are available:

```go
func (c *DeployCommand) HelpTemplate() string {
    return `{{ .Help }}
{{- if gt (len .Subcommands) 0 }}

Subcommands:
{{ range .Subcommands }}  {{ .NameAligned }}  {{ .Synopsis }}
{{ end -}}
{{- end }}

Examples:
  myapp deploy ./dist --env staging
  myapp deploy ./dist --env production --dry-run
`
}
```

Available template variables:

| Variable | Type | Description |
|----------|------|-------------|
| `.Name` | `string` | CLI binary name |
| `.SubcommandName` | `string` | Matched subcommand key |
| `.Help` | `string` | Output of `command.Help()` |
| `.Subcommands` | `[]map` | Child subcommands (nested CLIs only) |

Each `.Subcommands` entry has `.Name`, `.NameAligned`, `.Help`, `.Synopsis`.

### Implementing a custom top-level HelpFunc

Replace the default help output entirely:

```go
c.HelpFunc = func(commands map[string]cli.CommandFactory) string {
    var b strings.Builder
    fmt.Fprintf(&b, "myapp %s\n\n", version)
    fmt.Fprintln(&b, "USAGE")
    fmt.Fprintf(&b, "  myapp <command> [flags]\n\n")
    fmt.Fprintln(&b, "COMMANDS")
    for name, factory := range commands {
        cmd, _ := factory()
        fmt.Fprintf(&b, "  %-16s %s\n", name, cmd.Synopsis())
    }
    return b.String()
}
```

Use `FilteredHelpFunc` to show only a subset of commands in a particular
context:

```go
c.HelpFunc = cli.FilteredHelpFunc(
    []string{"deploy", "rollback", "status"},
    cli.BasicHelpFunc("myapp"),
)
```

### Implementing a custom Ui

Any type that implements the six-method `Ui` interface works as a drop-in:

```go
type JSONUi struct {
    enc *json.Encoder
}

func NewJSONUi(w io.Writer) *JSONUi {
    return &JSONUi{enc: json.NewEncoder(w)}
}

func (u *JSONUi) Output(msg string) { u.enc.Encode(map[string]string{"level": "output", "msg": msg}) }
func (u *JSONUi) Info(msg string)   { u.enc.Encode(map[string]string{"level": "info",   "msg": msg}) }
func (u *JSONUi) Error(msg string)  { u.enc.Encode(map[string]string{"level": "error",  "msg": msg}) }
func (u *JSONUi) Warn(msg string)   { u.enc.Encode(map[string]string{"level": "warn",   "msg": msg}) }
func (u *JSONUi) Ask(q string) (string, error)       { return "", errors.New("interactive input unsupported in JSON mode") }
func (u *JSONUi) AskSecret(q string) (string, error) { return "", errors.New("interactive input unsupported in JSON mode") }
```

Wrap it in `ConcurrentUi` when goroutines write to it concurrently:

```go
ui := &cli.ConcurrentUi{Ui: NewJSONUi(os.Stdout)}
```

### Using BeforeRun and AfterRun for cross-cutting concerns

These hooks apply to every dispatched command without modifying the
commands themselves — useful for authentication, logging, and metrics:

```go
// Authentication gate
c.BeforeRun = func(name string, args []string) int {
    if name == "login" {
        return 0 // login itself must always be reachable
    }
    if token := os.Getenv("APP_TOKEN"); token == "" {
        fmt.Fprintln(os.Stderr, "error: not authenticated — run 'myapp login'")
        return 1 // non-zero aborts dispatch; AfterRun is NOT called
    }
    return 0
}

// Structured audit log + metrics
c.AfterRun = func(name string, args []string, exitCode int) {
    slog.Info("command finished", "cmd", name, "exit", exitCode)
    metrics.RecordCommand(name, exitCode)
}
```

## Development

### Running tests

```bash
# Standard
make test
go test ./...

# With race detector (recommended before every commit)
make testrace
go test -race ./...
```

### Updating dependencies

```bash
make updatedeps
```

## License

MIT

---

Built on
[armon/go-radix](https://github.com/armon/go-radix) ·
[posener/complete](https://github.com/posener/complete) ·
[Masterminds/sprig](https://github.com/Masterminds/sprig) ·
[fatih/color](https://github.com/fatih/color)
