package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/posener/complete"
)

const (
	// RunResultHelp is a value that can be returned from Run to signal
	// to the CLI to render the help output.
	RunResultHelp = -18511

	// Exit code constants follow the POSIX / GNU convention and the
	// de-facto standard used by shells and most CLI tools.

	// ExitCodeSuccess indicates the command completed without error.
	ExitCodeSuccess = 0

	// ExitCodeError indicates a general runtime error (command failed).
	ExitCodeError = 1

	// ExitCodeUsage indicates incorrect usage: bad flags, missing arguments,
	// or invalid input. Conventionally exit code 2 in POSIX tools.
	ExitCodeUsage = 2

	// ExitCodeNotFound is returned by the CLI itself when no matching
	// subcommand is found. Matches the shell convention for "command not found".
	ExitCodeNotFound = 127
)

// ExitError carries a human-readable message and an exit code. Commands and
// helper functions can return *ExitError to propagate both pieces of
// information up to main.
//
// Typical use in main:
//
//	if err := run(); err != nil {
//	    fmt.Fprintln(os.Stderr, err)
//	    os.Exit(cli.ExitCodeOf(err))
//	}
type ExitError struct {
	Code int
	Err  error
}

// NewExitError creates an ExitError with the given code and a formatted
// message. Use ExitCodeError for general failures and ExitCodeUsage for
// bad-argument errors.
func NewExitError(code int, format string, args ...any) *ExitError {
	return &ExitError{Code: code, Err: fmt.Errorf(format, args...)}
}

func (e *ExitError) Error() string { return e.Err.Error() }
func (e *ExitError) Unwrap() error { return e.Err }

// ExitCodeOf extracts the exit code from an error:
//   - nil  → ExitCodeSuccess (0)
//   - *ExitError → its Code field
//   - anything else → ExitCodeError (1)
func ExitCodeOf(err error) int {
	if err == nil {
		return ExitCodeSuccess
	}
	var ee *ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return ExitCodeError
}

// A command is a runnable sub-command of a CLI.
type Command interface {
	// Help should return long-form help text that includes the command-line
	// usage, a brief few sentences explaining the function of the command,
	// and the complete list of flags the command accepts.
	Help() string

	// Run should run the actual command with the given CLI instance and
	// command-line arguments. It should return the exit status when it is
	// finished.
	//
	// There are a handful of special exit codes this can return documented
	// above that change behavior.
	Run(args []string) int

	// Synopsis should return a one-line, short synopsis of the command.
	// This should be less than 50 characters ideally.
	Synopsis() string
}

// CommandAutocomplete is an extension of Command that enables fine-grained
// autocompletion. Subcommand autocompletion will work even if this interface
// is not implemented. By implementing this interface, more advanced
// autocompletion is enabled.
type CommandAutocomplete interface {
	// AutocompleteArgs returns the argument predictor for this command.
	// If argument completion is not supported, this should return
	// complete.PredictNothing.
	AutocompleteArgs() complete.Predictor

	// AutocompleteFlags returns a mapping of supported flags and autocomplete
	// options for this command. The map key for the Flags map should be the
	// complete flag such as "-foo" or "--foo".
	AutocompleteFlags() complete.Flags
}

// CommandHelpTemplate is an extension of Command that also has a function
// for returning a template for the help rather than the help itself. In
// this scenario, both Help and HelpTemplate should be implemented.
//
// If CommandHelpTemplate isn't implemented, the Help is output as-is.
type CommandHelpTemplate interface {
	// HelpTemplate is the template in text/template format to use for
	// displaying the Help. The keys available are:
	//
	//   * ".Help" - The help text itself
	//   * ".Subcommands"
	//
	HelpTemplate() string
}

// CommandV2 extends Command with context-aware execution. Implement this
// interface instead of (or in addition to) Command when the command needs
// to respect cancellation signals or deadlines. The CLI will call
// RunContext when the command implements CommandV2, falling back to Run
// for plain Command implementations.
type CommandV2 interface {
	Command

	// RunContext is like Run but receives a context that carries cancellation
	// signals and deadlines. Commands should respect ctx.Done() to support
	// clean shutdown (e.g. on SIGINT).
	RunContext(ctx context.Context, args []string) int
}

// CommandDeprecated is an optional interface a Command can implement to
// signal that it is deprecated. When a user invokes a deprecated command the
// CLI prints the deprecation message to ErrorWriter before dispatching.
//
// Example:
//
//	func (c *OldPushCommand) DeprecationMessage() string {
//	    return "Use 'deploy' instead. 'push' will be removed in v2."
//	}
type CommandDeprecated interface {
	// DeprecationMessage returns a short, human-readable message that
	// explains why the command is deprecated and what to use instead.
	// Returning an empty string suppresses the warning.
	DeprecationMessage() string
}

// CommandFactory is a type of function that is a factory for commands.
// We need a factory because we may need to setup some state on the
// struct that implements the command itself.
type CommandFactory func() (Command, error)

// noopParentCommand is used internally when a nested subcommand is registered
// (e.g. "server start") but the parent ("server") has no explicit factory.
// It returns RunResultHelp so the CLI automatically shows the subcommand list.
type noopParentCommand struct{}

func (c *noopParentCommand) Help() string {
	return "This command is accessed by using one of the subcommands below."
}

func (c *noopParentCommand) Run(_ []string) int { return RunResultHelp }

func (c *noopParentCommand) Synopsis() string { return "" }
