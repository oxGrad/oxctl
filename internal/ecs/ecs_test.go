package ecs_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/oxGrad/oxctl/internal/ecs"
)

// fakeRunner records calls made to it.
type fakeRunner struct {
	calls  []string
	output []byte
	err    error
}

func (f *fakeRunner) Run(ctx context.Context, name string, args ...string) error {
	f.calls = append(f.calls, name+" "+strings.Join(args, " "))
	return f.err
}

func (f *fakeRunner) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, name+" "+strings.Join(args, " "))
	return f.output, f.err
}

func TestRegisterTaskDefinition_CallsAWS(t *testing.T) {
	taskDef := map[string]any{"family": "my-family", "containerDefinitions": []any{}}
	taskDefJSON, _ := json.Marshal(taskDef)

	fakeOutput := []byte(`{"taskDefinition":{"taskDefinitionArn":"arn:aws:ecs:us-east-1:123:task-definition/my-family:42"}}`)
	fr := &fakeRunner{output: fakeOutput}

	d := ecs.NewAWSCLIDeployer(fr)
	arn, err := d.RegisterTaskDefinition(context.Background(), taskDef)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(arn, "my-family:42") {
		t.Errorf("unexpected ARN: %q", arn)
	}
	if len(fr.calls) == 0 || !strings.Contains(fr.calls[0], "register-task-definition") {
		t.Errorf("expected register-task-definition call, got: %v", fr.calls)
	}
	_ = taskDefJSON
}

func TestUpdateService_CallsAWS(t *testing.T) {
	fr := &fakeRunner{}
	d := ecs.NewAWSCLIDeployer(fr)
	err := d.UpdateService(context.Background(), "my-cluster", "my-service", "arn:aws:ecs:us-east-1:123:task-definition/my-family:42")
	if err != nil {
		t.Fatal(err)
	}
	if len(fr.calls) == 0 || !strings.Contains(fr.calls[0], "update-service") {
		t.Errorf("expected update-service call, got: %v", fr.calls)
	}
}

func TestDescribeService_ParsesOutput(t *testing.T) {
	fakeOutput := []byte(`{
		"services": [{
			"runningCount": 2,
			"desiredCount": 2,
			"taskDefinition": "arn:aws:ecs:us-east-1:123:task-definition/my-family:42",
			"deployments": [{"status": "PRIMARY", "rolloutState": "COMPLETED"}]
		}]
	}`)
	fr := &fakeRunner{output: fakeOutput}
	d := ecs.NewAWSCLIDeployer(fr)

	svc, err := d.DescribeService(context.Background(), "my-cluster", "my-service")
	if err != nil {
		t.Fatal(err)
	}
	if svc.RunningCount != 2 {
		t.Errorf("running count: want 2, got %d", svc.RunningCount)
	}
	if svc.DesiredCount != 2 {
		t.Errorf("desired count: want 2, got %d", svc.DesiredCount)
	}
	if svc.TaskDefinition != "arn:aws:ecs:us-east-1:123:task-definition/my-family:42" {
		t.Errorf("task def ARN mismatch: %q", svc.TaskDefinition)
	}
}
