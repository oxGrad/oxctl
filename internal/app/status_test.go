package app_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/oxGrad/oxctl/internal/app"
	"github.com/oxGrad/oxctl/internal/ecs"
)

type stubStatusECS struct {
	status ecs.ServiceStatus
}

func (s *stubStatusECS) RegisterTaskDefinition(_ context.Context, _ map[string]any) (string, error) {
	return "", nil
}
func (s *stubStatusECS) UpdateService(_ context.Context, _, _, _ string) error { return nil }
func (s *stubStatusECS) WaitStable(_ context.Context, _, _ string) error       { return nil }
func (s *stubStatusECS) DescribeService(_ context.Context, _, _ string) (ecs.ServiceStatus, error) {
	return s.status, nil
}

func TestStatus_PrintsAllFields(t *testing.T) {
	stub := &stubStatusECS{
		status: ecs.ServiceStatus{
			RunningCount:   2,
			DesiredCount:   2,
			TaskDefinition: "arn:aws:ecs:us-east-1:123:task-definition/fam:42",
			Deployments: []ecs.Deployment{
				{Status: "PRIMARY", RolloutState: "COMPLETED"},
			},
		},
	}
	var buf bytes.Buffer
	sr := app.NewStatusReporter(stub, &buf)

	if err := sr.Report(context.Background(), "my-cluster", "my-service"); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	for _, want := range []string{"2", "fam:42", "PRIMARY", "COMPLETED"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}
