package cli

// VerbosityLevel controls how much output a [LevelFilterUi] passes through
// to the underlying [Ui].
type VerbosityLevel int

const (
	// VerbosityQuiet suppresses Output, Info, and Warn; only Error passes through.
	VerbosityQuiet VerbosityLevel = iota
	// VerbosityNormal is the default level: all Ui methods pass through unchanged.
	VerbosityNormal
	// VerbosityVerbose is identical to VerbosityNormal at the Ui layer.
	// Commands can inspect [CLI.Verbosity] to emit additional detail when this
	// level is active.
	VerbosityVerbose
)

// LevelFilterUi is a [Ui] implementation that wraps another Ui and suppresses
// output calls that fall below the configured [VerbosityLevel].
//
//   - [VerbosityQuiet]   — only Error passes through
//   - [VerbosityNormal]  — Output, Info, Warn, and Error all pass through (default)
//   - [VerbosityVerbose] — identical to Normal at the Ui layer; use [CLI.Verbosity]
//     in commands to emit extra detail
//
// Ask and AskSecret always pass through regardless of level, because interactive
// prompts are required for the program to function correctly.
type LevelFilterUi struct {
	// Level is the minimum verbosity required for output to be forwarded.
	Level VerbosityLevel
	// Ui is the underlying Ui that receives forwarded calls.
	Ui Ui
}

func (u *LevelFilterUi) Ask(query string) (string, error) {
	return u.Ui.Ask(query)
}

func (u *LevelFilterUi) AskSecret(query string) (string, error) {
	return u.Ui.AskSecret(query)
}

// Output is suppressed when Level is VerbosityQuiet.
func (u *LevelFilterUi) Output(message string) {
	if u.Level >= VerbosityNormal {
		u.Ui.Output(message)
	}
}

// Info is suppressed when Level is VerbosityQuiet.
func (u *LevelFilterUi) Info(message string) {
	if u.Level >= VerbosityNormal {
		u.Ui.Info(message)
	}
}

// Warn is suppressed when Level is VerbosityQuiet.
func (u *LevelFilterUi) Warn(message string) {
	if u.Level >= VerbosityNormal {
		u.Ui.Warn(message)
	}
}

// Error always passes through regardless of Level.
func (u *LevelFilterUi) Error(message string) {
	u.Ui.Error(message)
}
