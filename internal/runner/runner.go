package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// CommandRunner executes external commands.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) error
	Output(ctx context.Context, name string, args ...string) ([]byte, error)
}

// ExecRunner runs real commands or prints them in dry-run mode.
type ExecRunner struct {
	dryRun bool
	out    io.Writer
}

// NewExecRunner creates an ExecRunner. Pass dryRun=true to suppress execution.
// If out is nil, os.Stdout is used for dry-run output.
func NewExecRunner(dryRun bool, out io.Writer) *ExecRunner {
	if out == nil {
		out = os.Stdout
	}
	return &ExecRunner{dryRun: dryRun, out: out}
}

func (r *ExecRunner) Run(ctx context.Context, name string, args ...string) error {
	if r.dryRun {
		fmt.Fprintln(r.out, "[dry-run]", name, strings.Join(args, " "))
		return nil
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *ExecRunner) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	if r.dryRun {
		fmt.Fprintln(r.out, "[dry-run]", name, strings.Join(args, " "))
		return nil, nil
	}
	return exec.CommandContext(ctx, name, args...).Output()
}
