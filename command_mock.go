package cli

import (
	"context"

	"github.com/posener/complete"
)

// MockCommand is an implementation of Command that can be used for tests.
// It is publicly exported from this package in case you want to use it
// externally.
type MockCommand struct {
	// Settable
	HelpText     string
	RunResult    int
	SynopsisText string

	// Set by the command
	RunCalled bool
	RunArgs   []string
}

func (c *MockCommand) Help() string {
	return c.HelpText
}

func (c *MockCommand) Run(args []string) int {
	c.RunCalled = true
	c.RunArgs = args

	return c.RunResult
}

func (c *MockCommand) Synopsis() string {
	return c.SynopsisText
}

// MockCommandAutocomplete is an implementation of CommandAutocomplete.
type MockCommandAutocomplete struct {
	MockCommand

	// Settable
	AutocompleteArgsValue  complete.Predictor
	AutocompleteFlagsValue complete.Flags
}

func (c *MockCommandAutocomplete) AutocompleteArgs() complete.Predictor {
	return c.AutocompleteArgsValue
}

func (c *MockCommandAutocomplete) AutocompleteFlags() complete.Flags {
	return c.AutocompleteFlagsValue
}

// MockCommandV2 is an implementation of CommandV2 for tests.
type MockCommandV2 struct {
	MockCommand

	// RunContextResult is the exit code RunContext returns.
	RunContextResult int

	// Set by the command
	RunContextCalled bool
	RunContextArgs   []string
	RunContextCtx    context.Context
}

func (c *MockCommandV2) RunContext(ctx context.Context, args []string) int {
	c.RunContextCalled = true
	c.RunContextCtx = ctx
	c.RunContextArgs = args
	return c.RunContextResult
}

// MockCommandHelpTemplate is an implementation of CommandHelpTemplate.
type MockCommandHelpTemplate struct {
	MockCommand

	// Settable
	HelpTemplateText string
}

func (c *MockCommandHelpTemplate) HelpTemplate() string {
	return c.HelpTemplateText
}
