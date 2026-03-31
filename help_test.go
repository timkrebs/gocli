package cli

import (
	"strings"
	"testing"
)

// helpCommands builds a minimal CommandFactory map for use in help tests.
func helpCommands(names ...string) map[string]CommandFactory {
	m := make(map[string]CommandFactory, len(names))
	for _, n := range names {
		name := n // capture
		m[name] = func() (Command, error) {
			return &MockCommand{SynopsisText: name + " synopsis"}, nil
		}
	}
	return m
}

// ---------------------------------------------------------------------------
// BasicHelpFunc — flat (no groups)
// ---------------------------------------------------------------------------

func TestBasicHelpFunc_flat(t *testing.T) {
	commands := helpCommands("start", "stop", "status")
	out := BasicHelpFunc("myapp")(commands)

	if !strings.Contains(out, "Usage: myapp") {
		t.Fatalf("missing usage line:\n%s", out)
	}
	if !strings.Contains(out, "Available commands are:") {
		t.Fatalf("missing header:\n%s", out)
	}
	for _, name := range []string{"start", "stop", "status"} {
		if !strings.Contains(out, name) {
			t.Fatalf("missing command %q in output:\n%s", name, out)
		}
	}
}

func TestBasicHelpFunc_flat_sorted(t *testing.T) {
	commands := helpCommands("zebra", "alpha", "middle")
	out := BasicHelpFunc("app")(commands)

	alphaIdx := strings.Index(out, "alpha")
	middleIdx := strings.Index(out, "middle")
	zebraIdx := strings.Index(out, "zebra")

	if !(alphaIdx < middleIdx && middleIdx < zebraIdx) {
		t.Fatalf("commands not in alphabetical order:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// BasicHelpFunc — grouped
// ---------------------------------------------------------------------------

func TestBasicHelpFunc_grouped_headings(t *testing.T) {
	commands := helpCommands("start", "stop", "db migrate", "db seed")
	groups := []CommandGroup{
		{Name: "Server Commands", Commands: []string{"start", "stop"}},
		{Name: "Database Commands", Commands: []string{"db migrate", "db seed"}},
	}

	out := BasicHelpFunc("myapp", groups...)(commands)

	if !strings.Contains(out, "Server Commands:") {
		t.Fatalf("missing 'Server Commands:' heading:\n%s", out)
	}
	if !strings.Contains(out, "Database Commands:") {
		t.Fatalf("missing 'Database Commands:' heading:\n%s", out)
	}
	// Should not include the flat "Available commands are:" header.
	if strings.Contains(out, "Available commands are:") {
		t.Fatalf("should not have flat header when groups are provided:\n%s", out)
	}
}

func TestBasicHelpFunc_grouped_ordering(t *testing.T) {
	commands := helpCommands("start", "stop", "db migrate")
	groups := []CommandGroup{
		{Name: "Server Commands", Commands: []string{"start", "stop"}},
		{Name: "Database Commands", Commands: []string{"db migrate"}},
	}

	out := BasicHelpFunc("myapp", groups...)(commands)

	serverIdx := strings.Index(out, "Server Commands:")
	dbIdx := strings.Index(out, "Database Commands:")
	if serverIdx == -1 || dbIdx == -1 {
		t.Fatalf("missing headings:\n%s", out)
	}
	if serverIdx > dbIdx {
		t.Fatalf("groups rendered in wrong order:\n%s", out)
	}
}

func TestBasicHelpFunc_grouped_ungrouped_under_other(t *testing.T) {
	commands := helpCommands("start", "stop", "version")
	groups := []CommandGroup{
		{Name: "Server Commands", Commands: []string{"start", "stop"}},
	}

	out := BasicHelpFunc("myapp", groups...)(commands)

	if !strings.Contains(out, "Other Commands:") {
		t.Fatalf("expected 'Other Commands:' for ungrouped commands:\n%s", out)
	}
	if !strings.Contains(out, "version") {
		t.Fatalf("ungrouped command 'version' missing from output:\n%s", out)
	}
}

func TestBasicHelpFunc_grouped_no_other_when_all_assigned(t *testing.T) {
	commands := helpCommands("start", "stop")
	groups := []CommandGroup{
		{Name: "Server Commands", Commands: []string{"start", "stop"}},
	}

	out := BasicHelpFunc("myapp", groups...)(commands)

	if strings.Contains(out, "Other Commands:") {
		t.Fatalf("'Other Commands:' should not appear when all commands are grouped:\n%s", out)
	}
}

func TestBasicHelpFunc_grouped_skips_unknown_keys(t *testing.T) {
	commands := helpCommands("start")
	groups := []CommandGroup{
		{Name: "Server Commands", Commands: []string{"start", "nonexistent"}},
	}

	// Should not panic and should still render "start".
	out := BasicHelpFunc("myapp", groups...)(commands)

	if !strings.Contains(out, "start") {
		t.Fatalf("expected 'start' in output:\n%s", out)
	}
}

func TestBasicHelpFunc_grouped_empty_group_omitted(t *testing.T) {
	commands := helpCommands("start")
	groups := []CommandGroup{
		{Name: "Empty Group", Commands: []string{"nonexistent"}},
		{Name: "Server Commands", Commands: []string{"start"}},
	}

	out := BasicHelpFunc("myapp", groups...)(commands)

	if strings.Contains(out, "Empty Group:") {
		t.Fatalf("empty group heading should be omitted:\n%s", out)
	}
	if !strings.Contains(out, "Server Commands:") {
		t.Fatalf("non-empty group should appear:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// FilteredHelpFunc works with grouped BasicHelpFunc
// ---------------------------------------------------------------------------

func TestFilteredHelpFunc_with_groups(t *testing.T) {
	commands := helpCommands("start", "stop", "secret")
	groups := []CommandGroup{
		{Name: "Server Commands", Commands: []string{"start", "stop"}},
	}

	filtered := FilteredHelpFunc([]string{"start", "stop"}, BasicHelpFunc("myapp", groups...))
	out := filtered(commands)

	if strings.Contains(out, "secret") {
		t.Fatalf("filtered command 'secret' should not appear:\n%s", out)
	}
	if !strings.Contains(out, "start") {
		t.Fatalf("included command 'start' should appear:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// Backward compatibility: CLI wired with no groups renders flat output
// ---------------------------------------------------------------------------

func TestCLIRun_helpGrouped(t *testing.T) {
	helpBuf := new(strings.Builder)
	commands := map[string]CommandFactory{
		"start":   func() (Command, error) { return &MockCommand{SynopsisText: "start server"}, nil },
		"stop":    func() (Command, error) { return &MockCommand{SynopsisText: "stop server"}, nil },
		"migrate": func() (Command, error) { return &MockCommand{SynopsisText: "run migrations"}, nil },
	}
	groups := []CommandGroup{
		{Name: "Server Commands", Commands: []string{"start", "stop"}},
		{Name: "Database Commands", Commands: []string{"migrate"}},
	}

	c := &CLI{
		Args:        []string{"-h"},
		Commands:    commands,
		HelpFunc:    BasicHelpFunc("myapp", groups...),
		HelpWriter:  helpBuf,
		ErrorWriter: helpBuf,
	}
	c.Run()

	out := helpBuf.String()
	if !strings.Contains(out, "Server Commands:") {
		t.Fatalf("expected 'Server Commands:' in help output:\n%s", out)
	}
	if !strings.Contains(out, "Database Commands:") {
		t.Fatalf("expected 'Database Commands:' in help output:\n%s", out)
	}
}
