package cli

import (
	"strings"
	"testing"
)

func TestLevelFilterUi_implements(t *testing.T) {
	var _ Ui = new(LevelFilterUi)
}

// TestLevelFilterUi_Normal verifies that all Ui methods pass through at the
// default VerbosityNormal level.
func TestLevelFilterUi_Normal(t *testing.T) {
	inner := NewMockUi()
	ui := &LevelFilterUi{Level: VerbosityNormal, Ui: inner}

	ui.Output("out")
	ui.Info("info")
	ui.Warn("warn")
	ui.Error("err")

	if inner.OutputWriter.String() != "out\ninfo\n" {
		t.Fatalf("unexpected output: %q", inner.OutputWriter.String())
	}
	if inner.ErrorWriter.String() != "warn\nerr\n" {
		t.Fatalf("unexpected errors: %q", inner.ErrorWriter.String())
	}
}

// TestLevelFilterUi_Verbose verifies that VerbosityVerbose behaves identically
// to VerbosityNormal at the Ui layer (all methods pass through).
func TestLevelFilterUi_Verbose(t *testing.T) {
	inner := NewMockUi()
	ui := &LevelFilterUi{Level: VerbosityVerbose, Ui: inner}

	ui.Output("out")
	ui.Info("info")
	ui.Warn("warn")
	ui.Error("err")

	if inner.OutputWriter.String() != "out\ninfo\n" {
		t.Fatalf("unexpected output: %q", inner.OutputWriter.String())
	}
	if inner.ErrorWriter.String() != "warn\nerr\n" {
		t.Fatalf("unexpected errors: %q", inner.ErrorWriter.String())
	}
}

// TestLevelFilterUi_Quiet verifies that VerbosityQuiet suppresses Output,
// Info, and Warn while always forwarding Error.
func TestLevelFilterUi_Quiet(t *testing.T) {
	inner := NewMockUi()
	ui := &LevelFilterUi{Level: VerbosityQuiet, Ui: inner}

	ui.Output("out")
	ui.Info("info")
	ui.Warn("warn")
	ui.Error("err")

	if inner.OutputWriter.String() != "" {
		t.Fatalf("expected no output in quiet mode, got: %q", inner.OutputWriter.String())
	}
	if inner.ErrorWriter.String() != "err\n" {
		t.Fatalf("unexpected errors: %q", inner.ErrorWriter.String())
	}
}

// TestLevelFilterUi_Quiet_ErrorAlwaysPasses verifies that Error is never
// suppressed, even at VerbosityQuiet.
func TestLevelFilterUi_Quiet_ErrorAlwaysPasses(t *testing.T) {
	inner := NewMockUi()
	ui := &LevelFilterUi{Level: VerbosityQuiet, Ui: inner}

	ui.Error("fatal problem")

	if inner.ErrorWriter.String() != "fatal problem\n" {
		t.Fatalf("Error was suppressed in quiet mode: %q", inner.ErrorWriter.String())
	}
}

// TestLevelFilterUi_Ask verifies that Ask always passes through.
func TestLevelFilterUi_Ask(t *testing.T) {
	inner := NewMockUi()
	inner.InputReader = strings.NewReader("answer\n")
	ui := &LevelFilterUi{Level: VerbosityQuiet, Ui: inner}

	result, err := ui.Ask("question?")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if result != "answer" {
		t.Fatalf("unexpected result: %q", result)
	}
}

// TestLevelFilterUi_AskSecret verifies that AskSecret always passes through.
func TestLevelFilterUi_AskSecret(t *testing.T) {
	inner := NewMockUi()
	inner.InputReader = strings.NewReader("secret\n")
	ui := &LevelFilterUi{Level: VerbosityQuiet, Ui: inner}

	result, err := ui.AskSecret("password?")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if result != "secret" {
		t.Fatalf("unexpected result: %q", result)
	}
}

// TestLevelFilterUi_Composable verifies that LevelFilterUi composes correctly
// with PrefixedUi — the level filter suppresses before the prefix is applied.
func TestLevelFilterUi_Composable(t *testing.T) {
	inner := NewMockUi()
	prefixed := &PrefixedUi{OutputPrefix: "[app] ", InfoPrefix: "[app] ", Ui: inner}
	ui := &LevelFilterUi{Level: VerbosityQuiet, Ui: prefixed}

	ui.Output("hello")
	ui.Info("world")

	if inner.OutputWriter.String() != "" {
		t.Fatalf("expected no output through quiet+prefixed ui, got: %q", inner.OutputWriter.String())
	}
}
