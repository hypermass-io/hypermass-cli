package formatters

import (
	"bytes"
	"hypermass-cli/app_common"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func capture(run func()) string {
	original := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer

	run()

	writer.Close()
	os.Stdout = original

	var buffer bytes.Buffer
	io.Copy(&buffer, reader)

	return buffer.String()
}

// A sync with working credentials shows no banner.
func TestNoAlertWhenNothingIsWrong(t *testing.T) {
	output := capture(func() { printAccountAlert(app_common.StatusReport{}) })

	if strings.TrimSpace(output) != "" {
		t.Errorf("expected no output, got %q", output)
	}
}

// The banner exists so that someone reading rows of per-stream states sees the one problem that
// explains all of them, and where to go to fix it.
func TestAlertSaysWhatIsWrongAndWhereToGo(t *testing.T) {
	report := app_common.StatusReport{
		AccountAlert:      "key rejected or account locked",
		AccountAlertSince: time.Date(2026, 9, 23, 17, 13, 0, 0, time.UTC),
	}

	output := capture(func() { printAccountAlert(report) })

	for _, expected := range []string{"key rejected or account locked", "hypermass.io", "2026-09-23 17:13:00"} {
		if !strings.Contains(output, expected) {
			t.Errorf("expected the alert to mention %q, got %q", expected, output)
		}
	}
}
