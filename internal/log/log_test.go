package log_test

import (
	"bytes"
	"strings"
	"testing"

	oxlog "github.com/oxGrad/oxctl/internal/log"
)

func TestNew_JSONOutput(t *testing.T) {
	var buf bytes.Buffer
	l := oxlog.New(&buf, true, true)
	l.Info("hello", "key", "val")
	if !strings.Contains(buf.String(), `"key":"val"`) {
		t.Fatalf("expected JSON log line, got: %s", buf.String())
	}
}

func TestNew_TextOutput(t *testing.T) {
	var buf bytes.Buffer
	l := oxlog.New(&buf, false, false)
	l.Info("hello", "key", "val")
	if !strings.Contains(buf.String(), "key=val") {
		t.Fatalf("expected text log line, got: %s", buf.String())
	}
}

func TestNew_DebugSuppressed(t *testing.T) {
	var buf bytes.Buffer
	l := oxlog.New(&buf, false, false)
	l.Debug("secret")
	if strings.Contains(buf.String(), "secret") {
		t.Fatalf("debug line should be suppressed at info level")
	}
}
