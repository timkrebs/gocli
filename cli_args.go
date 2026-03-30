package cli

import (
	"strings"
)

func (c *CLI) processArgs() {
	for i, arg := range c.Args {
		if arg == "--" {
			break
		}

		// Check for help flags.
		if arg == "-h" || arg == "-help" || arg == "--help" {
			c.isHelp = true
			continue
		}

		// Check for autocomplete flags
		if c.Autocomplete {
			if arg == "-"+c.AutocompleteInstall || arg == "--"+c.AutocompleteInstall {
				c.isAutocompleteInstall = true
				continue
			}

			if arg == "-"+c.AutocompleteUninstall || arg == "--"+c.AutocompleteUninstall {
				c.isAutocompleteUninstall = true
				continue
			}
		}

		// Check for the --no-color flag.
		if c.NoColorFlag != "" {
			if arg == "-"+c.NoColorFlag || arg == "--"+c.NoColorFlag {
				c.isNoColor = true
				continue
			}
		}

		if c.subcommand == "" {
			// Check for version flags if not in a subcommand.
			if arg == "-v" || arg == "-version" || arg == "--version" {
				c.isVersion = true
				continue
			}

			if arg != "" && arg[0] == '-' {
				// Record the arg...
				c.topFlags = append(c.topFlags, arg)
			}
		}

		// If we didn't find a subcommand yet and this is the first non-flag
		// argument, then this is our subcommand.
		if c.subcommand == "" && arg != "" && arg[0] != '-' {
			c.subcommand = arg
			if c.commandNested {
				// If the command has a space in it, then it is invalid.
				// Set a blank command so that it fails.
				if strings.ContainsRune(arg, ' ') {
					c.subcommand = ""
					return
				}

				// Determine the argument we look to to end subcommands.
				// We look at all arguments until one is a flag or has a space.
				// This disallows commands like: ./cli foo "bar baz". An
				// argument with a space is always an argument. A blank
				// argument is always an argument.
				j := 0
				for k, v := range c.Args[i:] {
					if strings.ContainsRune(v, ' ') || v == "" || v[0] == '-' {
						break
					}

					j = i + k + 1
				}

				// Nested CLI, the subcommand is actually the entire
				// arg list up to a flag that is still a valid subcommand.
				searchKey := strings.Join(c.Args[i:j], " ")
				k, _, ok := c.commandTree.LongestPrefix(searchKey)
				if ok {
					// k could be a prefix that doesn't contain the full
					// command such as "foo" instead of "foobar", so we
					// need to verify that we have an entire key. To do that,
					// we look for an ending in a space or an end of string.
					if searchKey == k || strings.HasPrefix(searchKey, k+" ") {
						c.subcommand = k
						i += strings.Count(k, " ")
					}
				}
			}

			// The remaining args the subcommand arguments
			c.subcommandArgs = c.Args[i+1:]
		}
	}

	// If we never found a subcommand and support a default command, then
	// switch to using that.
	if c.subcommand == "" {
		if _, ok := c.Commands[""]; ok {
			args := c.topFlags
			args = append(args, c.subcommandArgs...)
			c.topFlags = nil
			c.subcommandArgs = args
		}
	}
}
