package cli

import (
	"strings"
	"sync"
	"testing"
)

func TestConcurrentUi_impl(t *testing.T) {
	var _ Ui = new(ConcurrentUi)
}

// TestConcurrentUi_concurrentWrites verifies that ConcurrentUi serialises
// concurrent calls correctly: no writes are lost and the race detector
// reports no data races (run with: go test -race).
func TestConcurrentUi_concurrentWrites(t *testing.T) {
	mock := NewMockUi()
	ui := &ConcurrentUi{Ui: mock}

	const goroutines = 50
	const iterations = 10

	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(4)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				ui.Output("output")
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				ui.Info("info")
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				ui.Error("error")
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				ui.Warn("warn")
			}
		}()
	}

	wg.Wait()

	// Output and Info both route to OutputWriter; each call writes exactly one line.
	gotOutput := strings.Count(mock.OutputWriter.String(), "\n")
	wantOutput := goroutines * iterations * 2 // Output + Info
	if gotOutput != wantOutput {
		t.Errorf("OutputWriter: got %d lines, want %d", gotOutput, wantOutput)
	}

	// Error and Warn both route to ErrorWriter; each call writes exactly one line.
	gotError := strings.Count(mock.ErrorWriter.String(), "\n")
	wantError := goroutines * iterations * 2 // Error + Warn
	if gotError != wantError {
		t.Errorf("ErrorWriter: got %d lines, want %d", gotError, wantError)
	}
}
