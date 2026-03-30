package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/armon/go-radix"
	"github.com/fatih/color"
	"github.com/posener/complete"
)

// CLI contains the state necessary to run subcommands and parse the
// command line arguments.
//
// CLI also supports nested subcommands, such as "cli foo bar". To use
// nested subcommands, the key in the Commands mapping below contains the
// full subcommand. In this example, it would be "foo bar".
//
// If you use a CLI with nested subcommands, some semantics change due to
// ambiguities:
//
//   - We use longest prefix matching to find a matching subcommand. This
//     means if you register "foo bar" and the user executes "cli foo qux",
//     the "foo" command will be executed with the arg "qux". It is up to
//     you to handle these args. One option is to just return the special
//     help return code `RunResultHelp` to display help and exit.
//
//   - The help flag "-h" or "-help" will look at all args to determine
//     the help function. For example: "otto apps list -h" will show the
//     help for "apps list" but "otto apps -h" will show it for "apps".
//     In the normal CLI, only the first subcommand is used.
//
//   - The help flag will list any subcommands that a command takes
//     as well as the command's help itself. If there are no subcommands,
//     it will note this. If the CLI itself has no subcommands, this entire
//     section is omitted.
//
//   - Any parent commands that don't exist are automatically created as
//     no-op commands that just show help for other subcommands. For example,
//     if you only register "foo bar", then "foo" is automatically created.
type CLI struct {
	// Args is the list of command-line arguments received excluding
	// the name of the app. For example, if the command "./cli foo bar"
	// was invoked, then Args should be []string{"foo", "bar"}.
	Args []string

	// Commands is a mapping of subcommand names to a factory function
	// for creating that Command implementation. If there is a command
	// with a blank string "", then it will be used as the default command
	// if no subcommand is specified.
	//
	// If the key has a space in it, this will create a nested subcommand.
	// For example, if the key is "foo bar", then to access it our CLI
	// must be accessed with "./cli foo bar". See the docs for CLI for
	// notes on how this changes some other behavior of the CLI as well.
	//
	// The factory should be as cheap as possible, ideally only allocating
	// a struct. The factory may be called multiple times in the course
	// of a command execution and certain events such as help require the
	// instantiation of all commands. Expensive initialization should be
	// deferred to function calls within the interface implementation.
	Commands map[string]CommandFactory

	// HiddenCommands is a list of commands that are "hidden". Hidden
	// commands are not given to the help function callback and do not
	// show up in autocomplete. The values in the slice should be equivalent
	// to the keys in the command map.
	HiddenCommands []string

	// CommandAliases maps alias names to the canonical command name.
	// Aliases are hidden from help and autocomplete output; they simply
	// register an alternate invocation for an existing command.
	// Example: {"rm": "delete"} lets users type either "rm" or "delete".
	CommandAliases map[string]string

	// Name defines the name of the CLI.
	Name string

	// Version of the CLI. Printed as-is when --version is passed.
	// Use VersionFunc instead when the version string must be computed at
	// runtime (e.g. from embedded build metadata).
	Version string

	// VersionFunc is called to obtain the version string when Version is
	// empty. It is ignored when Version is set. Typical use:
	//
	//   var (
	//       buildVersion = "dev"
	//       buildCommit  = "none"
	//   )
	//   c.VersionFunc = func() string {
	//       return fmt.Sprintf("%s (%s)", buildVersion, buildCommit)
	//   }
	VersionFunc func() string

	// Autocomplete enables or disables subcommand auto-completion support.
	// This is enabled by default when NewCLI is called. Otherwise, this
	// must enabled explicitly.
	//
	// Autocomplete requires the "Name" option to be set on CLI. This name
	// should be set exactly to the binary name that is autocompleted.
	//
	// Autocompletion is supported via the github.com/posener/complete
	// library. This library supports bash, zsh and fish. To add support
	// for other shells, please see that library.
	//
	// AutocompleteInstall and AutocompleteUninstall are the global flag
	// names for installing and uninstalling the autocompletion handlers
	// for the user's shell. The flag should omit the hyphen(s) in front of
	// the value. Both single and double hyphens will automatically be supported
	// for the flag name. These default to `autocomplete-install` and
	// `autocomplete-uninstall` respectively.
	//
	// AutocompleteNoDefaultFlags is a boolean which controls if the default auto-
	// complete flags like -help and -version are added to the output.
	//
	// AutocompleteGlobalFlags are a mapping of global flags for
	// autocompletion. The help and version flags are automatically added.
	Autocomplete               bool
	AutocompleteInstall        string
	AutocompleteUninstall      string
	AutocompleteNoDefaultFlags bool
	AutocompleteGlobalFlags    complete.Flags
	autocompleteInstaller      autocompleteInstaller // For tests

	// HelpFunc is the function called to generate the generic help
	// text that is shown if help must be shown for the CLI that doesn't
	// pertain to a specific command.
	HelpFunc HelpFunc

	// HelpWriter is used to print help text and version when requested.
	// Defaults to os.Stderr for backwards compatibility.
	// It is recommended that you set HelpWriter to os.Stdout, and
	// ErrorWriter to os.Stderr.
	HelpWriter io.Writer

	// ErrorWriter used to output errors when a command can not be run.
	// Defaults to the value of HelpWriter for backwards compatibility.
	// It is recommended that you set HelpWriter to os.Stdout, and
	// ErrorWriter to os.Stderr.
	ErrorWriter io.Writer

	// NoColorFlag is the name of the global flag that disables ANSI color
	// output. Both the single-hyphen and double-hyphen forms are accepted
	// (e.g. -no-color and --no-color). Set to "" to disable the flag.
	// Defaults to "no-color" when constructed via NewCLI.
	//
	// The NO_COLOR environment variable (https://no-color.org) is respected
	// automatically by the underlying color library regardless of this flag.
	NoColorFlag string

	// VerbosityFlag is the name of the global --verbose flag. When non-empty,
	// the CLI also recognises a --quiet flag as the opposite. Both single-hyphen
	// and double-hyphen forms are accepted (e.g. -verbose and --verbose).
	// Set to "" to disable verbosity flags entirely.
	// Defaults to "verbose" when constructed via NewCLI.
	//
	// Callers can read the active level with [CLI.Verbosity] and wrap their Ui
	// with [LevelFilterUi] accordingly.
	VerbosityFlag string

	// BeforeRun is called before a subcommand is dispatched. A non-zero
	// return value aborts execution and is returned as the exit code.
	// BeforeRun is not called when the CLI handles help, version, or
	// autocomplete flags internally.
	BeforeRun func(name string, args []string) int

	// AfterRun is called after a subcommand completes, regardless of the
	// exit code. It receives the subcommand name, its arguments, and the
	// raw exit code returned by the command (which may be RunResultHelp).
	// AfterRun is not called when BeforeRun aborts execution.
	AfterRun func(name string, args []string, exitCode int)

	//---------------------------------------------------------------
	// Internal fields set automatically

	once           sync.Once
	autocomplete   *complete.Complete
	commandTree    *radix.Tree
	commandNested  bool
	commandHidden  map[string]struct{}
	subcommand     string
	subcommandArgs []string
	topFlags       []string
	helpFuncMap    template.FuncMap

	// These are true when special global flags are set. We can/should
	// probably use a bitset for this one day.
	isHelp                  bool
	isVersion               bool
	isNoColor               bool
	isVerbose               bool
	isQuiet                 bool
	isAutocompleteInstall   bool
	isAutocompleteUninstall bool
}

// NewCLI returns a new CLI instance with sensible defaults.
func NewCLI(app, version string) *CLI {
	return &CLI{
		Name:          app,
		Version:       version,
		HelpFunc:      BasicHelpFunc(app),
		Autocomplete:  true,
		NoColorFlag:   "no-color",
		VerbosityFlag: "verbose",
	}
}

// IsHelp returns whether or not the help flag is present within the
// arguments.
func (c *CLI) IsHelp() bool {
	c.once.Do(c.init)
	return c.isHelp
}

// IsVersion returns whether or not the version flag is present within the
// arguments.
func (c *CLI) IsVersion() bool {
	c.once.Do(c.init)
	return c.isVersion
}

// Verbosity returns the active VerbosityLevel derived from the --verbose and
// --quiet global flags. It returns VerbosityNormal when neither flag was passed.
//
// Typical usage: wrap the Ui passed to commands with a [LevelFilterUi] so that
// --quiet suppresses informational output automatically:
//
//	ui = &cli.LevelFilterUi{Level: myCLI.Verbosity(), Ui: ui}
func (c *CLI) Verbosity() VerbosityLevel {
	c.once.Do(c.init)
	switch {
	case c.isQuiet:
		return VerbosityQuiet
	case c.isVerbose:
		return VerbosityVerbose
	default:
		return VerbosityNormal
	}
}

// Run runs the actual CLI based on the arguments given.
// It is equivalent to RunContext(context.Background()).
func (c *CLI) Run() (int, error) {
	return c.RunContext(context.Background())
}

// RunContext runs the CLI like Run, but passes ctx to any command that
// implements CommandV2. Commands that only implement Command receive the
// same behavior as before.
func (c *CLI) RunContext(ctx context.Context) (int, error) {
	c.once.Do(c.init)

	// Disable ANSI color output when --no-color was passed. This must be
	// applied before any output is written so that even help/version output
	// respects the flag.
	if c.isNoColor {
		color.NoColor = true
	}

	// If this is a autocompletion request, satisfy it. This must be called
	// first before anything else since its possible to be autocompleting
	// -help or -version or other flags and we want to show completions
	// and not actually write the help or version.
	if c.Autocomplete && c.autocomplete.Complete() {
		return 0, nil
	}

	// Just show the version and exit if instructed.
	if c.IsVersion() {
		version := c.Version
		if version == "" && c.VersionFunc != nil {
			version = c.VersionFunc()
		}
		if version != "" {
			c.HelpWriter.Write([]byte(version + "\n"))
			return 0, nil
		}
	}

	// Just print the help when only '-h' or '--help' is passed.
	if c.IsHelp() && c.Subcommand() == "" {
		c.HelpWriter.Write([]byte(c.HelpFunc(c.helpCommands(c.Subcommand())) + "\n"))
		return 0, nil
	}

	// If we're attempting to install or uninstall autocomplete then handle
	if c.Autocomplete {
		// Autocomplete requires the "Name" to be set so that we know what
		// command to setup the autocomplete on.
		if c.Name == "" {
			return 1, fmt.Errorf(
				"internal error: CLI.Name must be specified for autocomplete to work")
		}

		// If both install and uninstall flags are specified, then error
		if c.isAutocompleteInstall && c.isAutocompleteUninstall {
			return 1, fmt.Errorf(
				"Either the autocomplete install or uninstall flag may " +
					"be specified, but not both.")
		}

		// If the install flag is specified, perform the install or uninstall
		if c.isAutocompleteInstall {
			if err := c.autocompleteInstaller.Install(c.Name); err != nil {
				return 1, err
			}

			return 0, nil
		}

		if c.isAutocompleteUninstall {
			if err := c.autocompleteInstaller.Uninstall(c.Name); err != nil {
				return 1, err
			}

			return 0, nil
		}
	}

	// Attempt to get the factory function for creating the command
	// implementation. If the command is invalid or blank, it is an error.
	raw, ok := c.commandTree.Get(c.Subcommand())
	if !ok {
		c.ErrorWriter.Write([]byte(c.HelpFunc(c.helpCommands(c.subcommandParent())) + "\n"))
		// Suggest close matches to help with typos.
		if suggestions := c.closestCommands(c.Subcommand(), 3); len(suggestions) > 0 {
			fmt.Fprintf(c.ErrorWriter, "\nDid you mean one of these?\n")
			for _, s := range suggestions {
				fmt.Fprintf(c.ErrorWriter, "    %s\n", s)
			}
		}
		return 127, nil
	}

	factory, ok := raw.(CommandFactory)
	if !ok {
		return 1, fmt.Errorf("cli: command %q has unexpected type in registry", c.Subcommand())
	}
	command, err := factory()
	if err != nil {
		return 1, err
	}

	// Warn the user when the command implements CommandDeprecated.
	if dep, ok := command.(CommandDeprecated); ok {
		if msg := dep.DeprecationMessage(); msg != "" {
			fmt.Fprintf(c.ErrorWriter, "DEPRECATED: %q is deprecated. %s\n\n", c.Subcommand(), msg)
		}
	}

	// If we've been instructed to just print the help, then print it
	if c.IsHelp() {
		c.commandHelp(c.HelpWriter, command)
		return 0, nil
	}

	// If there is an invalid flag, then error
	if len(c.topFlags) > 0 {
		c.ErrorWriter.Write([]byte(
			"Invalid flags before the subcommand. If these flags are for\n" +
				"the subcommand, please put them after the subcommand.\n\n"))
		c.commandHelp(c.ErrorWriter, command)
		return 1, nil
	}

	// Before-run hook: a non-zero return aborts execution.
	if c.BeforeRun != nil {
		if code := c.BeforeRun(c.Subcommand(), c.SubcommandArgs()); code != 0 {
			return code, nil
		}
	}

	// Dispatch to the command. Prefer CommandV2.RunContext when available so
	// that context cancellation and deadlines are propagated.
	var code int
	if cv2, ok := command.(CommandV2); ok {
		code = cv2.RunContext(ctx, c.SubcommandArgs())
	} else {
		code = command.Run(c.SubcommandArgs())
	}

	// After-run hook: called with the raw exit code from the command.
	if c.AfterRun != nil {
		c.AfterRun(c.Subcommand(), c.SubcommandArgs(), code)
	}

	if code == RunResultHelp {
		// Requesting help
		c.commandHelp(c.ErrorWriter, command)
		return 1, nil
	}

	return code, nil
}

// Subcommand returns the subcommand that the CLI would execute. For
// example, a CLI from "--version version --help" would return a Subcommand
// of "version"
func (c *CLI) Subcommand() string {
	c.once.Do(c.init)
	return c.subcommand
}

// SubcommandArgs returns the arguments that will be passed to the
// subcommand.
func (c *CLI) SubcommandArgs() []string {
	c.once.Do(c.init)
	return c.subcommandArgs
}

// subcommandParent returns the parent of this subcommand, if there is one.
// If there isn't on, "" is returned.
func (c *CLI) subcommandParent() string {
	// Get the subcommand, if it is "" alread just return
	sub := c.Subcommand()
	if sub == "" {
		return sub
	}

	// Clear any trailing spaces and find the last space
	sub = strings.TrimRight(sub, " ")
	idx := strings.LastIndex(sub, " ")

	if idx == -1 {
		// No space means our parent is root
		return ""
	}

	return sub[:idx]
}

func (c *CLI) init() {
	if c.HelpFunc == nil {
		c.HelpFunc = BasicHelpFunc("app")

		if c.Name != "" {
			c.HelpFunc = BasicHelpFunc(c.Name)
		}
	}

	if c.HelpWriter == nil {
		c.HelpWriter = os.Stderr
	}
	if c.ErrorWriter == nil {
		c.ErrorWriter = c.HelpWriter
	}

	// Build our hidden commands
	if len(c.HiddenCommands) > 0 {
		c.commandHidden = make(map[string]struct{})
		for _, h := range c.HiddenCommands {
			c.commandHidden[h] = struct{}{}
		}
	}

	// Build our command tree
	c.commandTree = radix.New()
	c.commandNested = false
	for k, v := range c.Commands {
		k = strings.TrimSpace(k)
		c.commandTree.Insert(k, v)
		if strings.ContainsRune(k, ' ') {
			c.commandNested = true
		}
	}

	// Go through the key and fill in any missing parent commands
	if c.commandNested {
		var walkFn radix.WalkFn
		toInsert := make(map[string]struct{})
		walkFn = func(k string, raw interface{}) bool {
			idx := strings.LastIndex(k, " ")
			if idx == -1 {
				// If there is no space, just ignore top level commands
				return false
			}

			// Trim up to that space so we can get the expected parent
			k = k[:idx]
			if _, ok := c.commandTree.Get(k); ok {
				// Yay we have the parent!
				return false
			}

			// We're missing the parent, so let's insert this
			toInsert[k] = struct{}{}

			// Call the walk function recursively so we check this one too
			return walkFn(k, nil)
		}

		// Walk!
		c.commandTree.Walk(walkFn)

		// Insert any that we're missing
		for k := range toInsert {
			var f CommandFactory = func() (Command, error) {
				return &noopParentCommand{}, nil
			}

			c.commandTree.Insert(k, f)
		}
	}

	// Register command aliases: insert the canonical factory under the alias
	// key and immediately hide it so it doesn't appear in help or autocomplete.
	for alias, canonical := range c.CommandAliases {
		alias = strings.TrimSpace(alias)
		f, ok := c.Commands[canonical]
		if !ok {
			continue // skip aliases pointing to non-existent commands
		}
		c.commandTree.Insert(alias, f)
		if c.commandHidden == nil {
			c.commandHidden = make(map[string]struct{})
		}
		c.commandHidden[alias] = struct{}{}
	}

	// Cache the sprig template function map once so commandHelp doesn't
	// rebuild it on every --help invocation.
	c.helpFuncMap = sprig.TxtFuncMap()

	// Setup autocomplete if we have it enabled. We have to do this after
	// the command tree is setup so we can use the radix tree to easily find
	// all subcommands.
	if c.Autocomplete {
		c.initAutocomplete()
	}

	// Process the args
	c.processArgs()
}
