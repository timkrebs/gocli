package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"
)

func (c *CLI) commandHelp(out io.Writer, command Command) {
	// Get the template to use
	tpl := strings.TrimSpace(defaultHelpTemplate)
	if t, ok := command.(CommandHelpTemplate); ok {
		tpl = t.HelpTemplate()
	}
	if !strings.HasSuffix(tpl, "\n") {
		tpl += "\n"
	}

	// Parse it
	t, err := template.New("root").Funcs(c.helpFuncMap).Parse(tpl)
	if err != nil {
		t = template.Must(template.New("root").Parse(fmt.Sprintf(
			"Internal error! Failed to parse command help template: %s\n", err)))
	}

	// Template data
	data := map[string]interface{}{
		"Name":           c.Name,
		"SubcommandName": c.Subcommand(),
		"Help":           command.Help(),
	}

	// Build subcommand list if we have it
	var subcommandsTpl []map[string]interface{}
	if c.commandNested {
		// Get the matching keys
		subcommands := c.helpCommands(c.Subcommand())
		keys := make([]string, 0, len(subcommands))
		for k := range subcommands {
			keys = append(keys, k)
		}

		// Sort the keys
		sort.Strings(keys)

		// Figure out the padding length
		var longest int
		for _, k := range keys {
			if v := len(k); v > longest {
				longest = v
			}
		}

		// Go through and create their structures
		subcommandsTpl = make([]map[string]interface{}, 0, len(subcommands))
		for _, k := range keys {
			// Get the command
			raw, ok := subcommands[k]
			if !ok {
				c.ErrorWriter.Write([]byte(fmt.Sprintf(
					"Error getting subcommand %q", k)))
			}
			sub, err := raw()
			if err != nil {
				c.ErrorWriter.Write([]byte(fmt.Sprintf(
					"Error instantiating %q: %s", k, err)))
			}

			// Find the last space and make sure we only include that last part
			name := k
			if idx := strings.LastIndex(k, " "); idx > -1 {
				name = name[idx+1:]
			}

			subcommandsTpl = append(subcommandsTpl, map[string]interface{}{
				"Name":        name,
				"NameAligned": name + strings.Repeat(" ", longest-len(k)),
				"Help":        sub.Help(),
				"Synopsis":    sub.Synopsis(),
			})
		}
	}
	data["Subcommands"] = subcommandsTpl

	// Write
	err = t.Execute(out, data)
	if err == nil {
		return
	}

	// An error, just output...
	c.ErrorWriter.Write([]byte(fmt.Sprintf(
		"Internal error rendering help: %s", err)))
}

// helpCommands returns the subcommands for the HelpFunc argument.
// This will only contain immediate subcommands.
func (c *CLI) helpCommands(prefix string) map[string]CommandFactory {
	// If our prefix isn't empty, make sure it ends in ' '
	if prefix != "" && prefix[len(prefix)-1] != ' ' {
		prefix += " "
	}

	// Get all the subkeys of this command
	var keys []string
	c.commandTree.WalkPrefix(prefix, func(k string, raw interface{}) bool {
		// Ignore any sub-sub keys, i.e. "foo bar baz" when we want "foo bar"
		if !strings.Contains(k[len(prefix):], " ") {
			keys = append(keys, k)
		}

		return false
	})

	// For each of the keys return that in the map
	result := make(map[string]CommandFactory, len(keys))
	for _, k := range keys {
		raw, ok := c.commandTree.Get(k)
		if !ok {
			// This should not happen since we just found this key via
			// WalkPrefix, but handle it gracefully instead of crashing.
			fmt.Fprintf(c.ErrorWriter, "[ERR] cli: Command '%s' disappeared from tree\n", k)
			continue
		}

		// If this is a hidden command, don't show it
		if _, ok := c.commandHidden[k]; ok {
			continue
		}

		f, ok := raw.(CommandFactory)
		if !ok {
			fmt.Fprintf(c.ErrorWriter, "[ERR] cli: unexpected type for command %q in tree\n", k)
			continue
		}
		result[k] = f
	}

	return result
}

// levenshtein computes the edit distance between strings a and b.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	dp := make([][]int, la+1)
	for i := range dp {
		dp[i] = make([]int, lb+1)
		dp[i][0] = i
	}
	for j := range dp[0] {
		dp[0][j] = j
	}
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1]
			} else {
				dp[i][j] = 1 + min(dp[i-1][j], min(dp[i][j-1], dp[i-1][j-1]))
			}
		}
	}
	return dp[la][lb]
}

// closestCommands returns up to limit non-hidden command names within
// Levenshtein distance 2 of input, sorted by distance then alphabetically.
func (c *CLI) closestCommands(input string, limit int) []string {
	const maxDist = 2
	type candidate struct {
		name string
		dist int
	}
	var candidates []candidate
	c.commandTree.Walk(func(k string, _ interface{}) bool {
		if _, hidden := c.commandHidden[k]; hidden {
			return false
		}
		if d := levenshtein(input, k); d > 0 && d <= maxDist {
			candidates = append(candidates, candidate{k, d})
		}
		return false
	})
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].dist != candidates[j].dist {
			return candidates[i].dist < candidates[j].dist
		}
		return candidates[i].name < candidates[j].name
	})
	result := make([]string, 0, min(limit, len(candidates)))
	for i := 0; i < limit && i < len(candidates); i++ {
		result = append(result, candidates[i].name)
	}
	return result
}

const defaultHelpTemplate = `
{{.Help}}{{if gt (len .Subcommands) 0}}

Subcommands:
{{- range $value := .Subcommands }}
    {{ $value.NameAligned }}    {{ $value.Synopsis }}{{ end }}
{{- end }}
`
