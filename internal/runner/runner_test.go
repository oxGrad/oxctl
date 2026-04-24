package runner_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/oxGrad/oxctl/internal/runner"
)

func TestExecRunner_Output(t *testing.T) {
	r := runner.NewExecRunner(false, nil)
	out, err := r.Output(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "hello") {
		t.Fatalf("expected 'hello', got %q", out)
	}
}

func TestExecRunner_Run(t *testing.T) {
	r := runner.NewExecRunner(false, nil)
	err := r.Run(context.Background(), "true")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestExecRunner_DryRun_DoesNotExecute(t *testing.T) {
	var buf bytes.Buffer
	r := runner.NewExecRunner(true, &buf)
	err := r.Run(context.Background(), "false")
	if err != nil {
		t.Fatalf("dry-run should not return error, got %v", err)
	}
	if !strings.Contains(buf.String(), "false") {
		t.Fatalf("expected command printed in dry-run, got %q", buf.String())
	}
}
