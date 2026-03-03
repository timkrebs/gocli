package cli

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
)

// HelpFunc is the type of the function that is responsible for generating
// the help output when the CLI must show the general help text.
type HelpFunc func(map[string]CommandFactory) string

// BasicHelpFunc generates some basic help output that is usually good enough
// for most CLI applications.
func BasicHelpFunc(app string) HelpFunc {
	return func(commands map[string]CommandFactory) string {
		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf(
			"Usage: %s [--version] [--help] <command> [<args>]\n\n",
			app))
		buf.WriteString("Available commands are:\n")

		// Get the list of keys so we can sort them, and also get the maximum
		// key length so they can be aligned properly.
		keys := make([]string, 0, len(commands))
		maxKeyLen := 0
		for key := range commands {
			if len(key) > maxKeyLen {
				maxKeyLen = len(key)
			}

			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			commandFunc, ok := commands[key]
			if !ok {
				// This should never happen since we JUST built the list of
				// keys, but handle it gracefully instead of crashing.
				fmt.Fprintf(os.Stderr, "[ERR] cli: Command '%s' not found\n", key)
				continue
			}

			command, err := commandFunc()
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERR] cli: Command '%s' failed to load: %s\n", key, err)
				continue
			}

			key = fmt.Sprintf("%s%s", key, strings.Repeat(" ", maxKeyLen-len(key)))
			buf.WriteString(fmt.Sprintf("    %s    %s\n", key, command.Synopsis()))
		}

		return buf.String()
	}
}

// FilteredHelpFunc will filter the commands to only include the keys
// in the include parameter.
func FilteredHelpFunc(include []string, f HelpFunc) HelpFunc {
	return func(commands map[string]CommandFactory) string {
		set := make(map[string]struct{})
		for _, k := range include {
			set[k] = struct{}{}
		}

		filtered := make(map[string]CommandFactory)
		for k, factory := range commands {
			if _, ok := set[k]; ok {
				filtered[k] = factory
			}
		}

		return f(filtered)
	}
}
