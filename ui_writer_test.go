package cli

import (
	"io"
	"testing"
)

func TestUiWriter_impl(t *testing.T) {
	var _ io.Writer = new(UiWriter)
}

func TestUiWriter_levels(t *testing.T) {
	cases := []struct {
		level    UiWriterLevel
		wantOut  string
		wantErr  string
	}{
		{LevelInfo,   "hello\n", ""},
		{LevelOutput, "hello\n", ""},
		{LevelWarn,   "",        "hello\n"},
		{LevelError,  "",        "hello\n"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run("", func(t *testing.T) {
			ui := NewMockUi()
			w := &UiWriter{Ui: ui, Level: tc.level}
			w.Write([]byte("hello\n"))

			if got := ui.OutputWriter.String(); got != tc.wantOut {
				t.Errorf("level %d: OutputWriter = %q, want %q", tc.level, got, tc.wantOut)
			}
			if got := ui.ErrorWriter.String(); got != tc.wantErr {
				t.Errorf("level %d: ErrorWriter = %q, want %q", tc.level, got, tc.wantErr)
			}
		})
	}
}

func TestUiWriter(t *testing.T) {
	ui := new(MockUi)
	w := &UiWriter{
		Ui: ui,
	}

	w.Write([]byte("foo\n"))
	w.Write([]byte("bar\n"))

	if ui.OutputWriter.String() != "foo\nbar\n" {
		t.Fatalf("bad: %s", ui.OutputWriter.String())
	}
}

func TestUiWriter_empty(t *testing.T) {
	ui := new(MockUi)
	w := &UiWriter{
		Ui: ui,
	}

	w.Write([]byte(""))

	if ui.OutputWriter.String() != "\n" {
		t.Fatalf("bad: %s", ui.OutputWriter.String())
	}
}
