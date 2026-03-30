package cli

import (
	"io"
	"strings"

	"github.com/posener/complete"
)

// defaultAutocompleteInstall and defaultAutocompleteUninstall are the
// default values for the autocomplete install and uninstall flags.
const defaultAutocompleteInstall = "autocomplete-install"
const defaultAutocompleteUninstall = "autocomplete-uninstall"

func (c *CLI) initAutocomplete() {
	if c.AutocompleteInstall == "" {
		c.AutocompleteInstall = defaultAutocompleteInstall
	}

	if c.AutocompleteUninstall == "" {
		c.AutocompleteUninstall = defaultAutocompleteUninstall
	}

	if c.autocompleteInstaller == nil {
		c.autocompleteInstaller = &realAutocompleteInstaller{}
	}

	// We first set c.autocomplete to a noop autocompleter that outputs
	// to nul so that we can detect if we're autocompleting or not. If we're
	// not, then we do nothing. This saves a LOT of compute cycles since
	// initAutoCompleteSub has to walk every command.
	c.autocomplete = complete.New(c.Name, complete.Command{})
	c.autocomplete.Out = io.Discard
	if !c.autocomplete.Complete() {
		return
	}

	// Build the root command
	cmd := c.initAutocompleteSub("")

	// For the root, we add the global flags to the "Flags". This way
	// they don't show up on every command.
	if !c.AutocompleteNoDefaultFlags {
		cmd.Flags = map[string]complete.Predictor{
			"-" + c.AutocompleteInstall:   complete.PredictNothing,
			"-" + c.AutocompleteUninstall: complete.PredictNothing,
			"-help":                       complete.PredictNothing,
			"-version":                    complete.PredictNothing,
		}
		if c.NoColorFlag != "" {
			cmd.Flags["-"+c.NoColorFlag] = complete.PredictNothing
		}
	}
	cmd.GlobalFlags = c.AutocompleteGlobalFlags

	c.autocomplete = complete.New(c.Name, cmd)
}

// initAutocompleteSub creates the complete.Command for a subcommand with
// the given prefix. This will continue recursively for all subcommands.
// The prefix "" (empty string) can be used for the root command.
func (c *CLI) initAutocompleteSub(prefix string) complete.Command {
	var cmd complete.Command
	walkFn := func(k string, raw interface{}) bool {
		// Ignore the empty key which can be present for default commands.
		if k == "" {
			return false
		}

		// Keep track of the full key so that we can nest further if necessary
		fullKey := k

		if len(prefix) > 0 {
			// If we have a prefix, trim the prefix + 1 (for the space)
			// Example: turns "sub one" to "one" with prefix "sub"
			k = k[len(prefix)+1:]
		}

		if idx := strings.Index(k, " "); idx >= 0 {
			// If there is a space, we trim up to the space. This turns
			// "sub sub2 sub3" into "sub". The prefix trim above will
			// trim our current depth properly.
			k = k[:idx]
		}

		if _, ok := cmd.Sub[k]; ok {
			// If we already tracked this subcommand then ignore
			return false
		}

		// If the command is hidden, don't record it at all
		if _, ok := c.commandHidden[fullKey]; ok {
			return false
		}

		if cmd.Sub == nil {
			cmd.Sub = complete.Commands(make(map[string]complete.Command))
		}
		subCmd := c.initAutocompleteSub(fullKey)

		// Instantiate the command so that we can check if the command is
		// a CommandAutocomplete implementation. If there is an error
		// creating the command, we just ignore it since that will be caught
		// later.
		impl, err := raw.(CommandFactory)()
		if err != nil {
			impl = nil
		}

		// Check if it implements CommandAutocomplete. If so, setup the autocomplete
		if c, ok := impl.(CommandAutocomplete); ok {
			subCmd.Args = c.AutocompleteArgs()
			subCmd.Flags = c.AutocompleteFlags()
		}

		cmd.Sub[k] = subCmd
		return false
	}

	walkPrefix := prefix
	if walkPrefix != "" {
		walkPrefix += " "
	}

	c.commandTree.WalkPrefix(walkPrefix, walkFn)
	return cmd
}
