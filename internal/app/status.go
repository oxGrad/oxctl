package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/oxGrad/oxctl/internal/ecs"
)

// StatusReporter queries and prints ECS service status.
type StatusReporter struct {
	ecs ecs.ECSDeployer
	out io.Writer
}

// NewStatusReporter creates a StatusReporter. If out is nil, os.Stdout is used.
func NewStatusReporter(e ecs.ECSDeployer, out io.Writer) *StatusReporter {
	if out == nil {
		out = os.Stdout
	}
	return &StatusReporter{ecs: e, out: out}
}

// Report fetches and prints service status to the configured writer.
func (s *StatusReporter) Report(ctx context.Context, cluster, service string) error {
	status, err := s.ecs.DescribeService(ctx, cluster, service)
	if err != nil {
		return fmt.Errorf("fetching status: %w", err)
	}

	fmt.Fprintf(s.out, "Cluster:         %s\n", cluster)
	fmt.Fprintf(s.out, "Service:         %s\n", service)
	fmt.Fprintf(s.out, "Running count:   %d\n", status.RunningCount)
	fmt.Fprintf(s.out, "Desired count:   %d\n", status.DesiredCount)
	fmt.Fprintf(s.out, "Task definition: %s\n", status.TaskDefinition)
	fmt.Fprintf(s.out, "Deployments:\n")
	for i, dep := range status.Deployments {
		fmt.Fprintf(s.out, "  [%d] status=%-10s rollout=%s\n", i+1, dep.Status, dep.RolloutState)
	}
	return nil
}
