package cli

// UiWriterLevel controls which Ui method UiWriter forwards each write to.
type UiWriterLevel int

const (
	// LevelInfo routes writes to Ui.Info. This is the zero value and the
	// default when UiWriter.Level is not set explicitly.
	LevelInfo UiWriterLevel = iota
	// LevelOutput routes writes to Ui.Output.
	LevelOutput
	// LevelWarn routes writes to Ui.Warn.
	LevelWarn
	// LevelError routes writes to Ui.Error.
	LevelError
)

// UiWriter is an io.Writer implementation that can be used with loggers or
// any code that writes to an io.Writer. Each written line is forwarded to the
// Ui at the configured Level (defaults to LevelInfo).
//
// Example — redirect the standard logger's error output through Ui.Error:
//
//	log.SetOutput(&cli.UiWriter{Ui: ui, Level: cli.LevelError})
type UiWriter struct {
	Ui    Ui
	Level UiWriterLevel
}

func (w *UiWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	if n > 0 && p[n-1] == '\n' {
		p = p[:n-1]
	}

	msg := string(p)
	switch w.Level {
	case LevelOutput:
		w.Ui.Output(msg)
	case LevelWarn:
		w.Ui.Warn(msg)
	case LevelError:
		w.Ui.Error(msg)
	default: // LevelInfo and any unrecognized value
		w.Ui.Info(msg)
	}
	return n, nil
}
