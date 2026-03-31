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

// CommandGroup associates a display title with a set of command names.
// It is used with [BasicHelpFunc] to organize the help output into labelled
// sections. Command names must match the keys in the CLI's Commands map.
//
// Example:
//
//	groups := []cli.CommandGroup{
//	    {Name: "Server Commands", Commands: []string{"start", "stop"}},
//	    {Name: "Database Commands", Commands: []string{"db migrate", "db seed"}},
//	}
//	c.HelpFunc = cli.BasicHelpFunc("myapp", groups...)
type CommandGroup struct {
	// Name is the human-readable heading shown above the group in help output
	// (e.g. "Server Commands", "Database Commands").
	Name string

	// Commands lists the command keys that belong to this group. Keys must
	// match entries in the Commands map on CLI.
	Commands []string
}

// BasicHelpFunc generates some basic help output that is usually good enough
// for most CLI applications.
//
// When one or more [CommandGroup] values are supplied the commands are
// rendered under their respective group headings. Commands not assigned to
// any group are collected under an "Other Commands" heading at the end.
// When no groups are supplied the output is the original flat list, so
// existing callers are unaffected.
func BasicHelpFunc(app string, groups ...CommandGroup) HelpFunc {
	return func(commands map[string]CommandFactory) string {
		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf(
			"Usage: %s [--version] [--help] <command> [<args>]\n\n",
			app))

		// Compute the maximum key length once for aligned columns.
		maxKeyLen := 0
		for key := range commands {
			if len(key) > maxKeyLen {
				maxKeyLen = len(key)
			}
		}

		// printSection writes a titled block of commands. Commands that are
		// not present in the commands map are silently skipped.
		printSection := func(title string, keys []string) {
			var present []string
			for _, k := range keys {
				if _, ok := commands[k]; ok {
					present = append(present, k)
				}
			}
			if len(present) == 0 {
				return
			}
			sort.Strings(present)
			buf.WriteString(title + "\n")
			for _, key := range present {
				commandFunc := commands[key]
				command, err := commandFunc()
				if err != nil {
					fmt.Fprintf(os.Stderr, "[ERR] cli: Command '%s' failed to load: %s\n", key, err)
					continue
				}
				padded := key + strings.Repeat(" ", maxKeyLen-len(key))
				buf.WriteString(fmt.Sprintf("    %s    %s\n", padded, command.Synopsis()))
			}
		}

		// No groups → original flat layout (backward compatible).
		if len(groups) == 0 {
			buf.WriteString("Available commands are:\n")
			keys := make([]string, 0, len(commands))
			for key := range commands {
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

		// Grouped layout: render each named group, then collect leftovers.
		assigned := make(map[string]bool)
		for _, g := range groups {
			for _, k := range g.Commands {
				assigned[k] = true
			}
		}

		for _, g := range groups {
			printSection(g.Name+":", g.Commands)
		}

		// Ungrouped commands appear under a default heading.
		var ungrouped []string
		for key := range commands {
			if !assigned[key] {
				ungrouped = append(ungrouped, key)
			}
		}
		printSection("Other Commands:", ungrouped)

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
