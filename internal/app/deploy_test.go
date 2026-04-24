package app_test

import (
	"context"
	"testing"

	"github.com/oxGrad/oxctl/internal/app"
	"github.com/oxGrad/oxctl/internal/config"
	"github.com/oxGrad/oxctl/internal/ecs"
)

type stubECS struct {
	registerCalled bool
	updateCalled   bool
	waitCalled     bool
	registerARN    string
}

func (s *stubECS) RegisterTaskDefinition(_ context.Context, _ map[string]any) (string, error) {
	s.registerCalled = true
	return s.registerARN, nil
}
func (s *stubECS) UpdateService(_ context.Context, _, _, _ string) error {
	s.updateCalled = true
	return nil
}
func (s *stubECS) WaitStable(_ context.Context, _, _ string) error {
	s.waitCalled = true
	return nil
}
func (s *stubECS) DescribeService(_ context.Context, _, _ string) (ecs.ServiceStatus, error) {
	return ecs.ServiceStatus{}, nil
}

func TestDeploy_RegistersAndUpdates(t *testing.T) {
	stub := &stubECS{registerARN: "arn:aws:ecs:us-east-1:123:task-definition/fam:1"}
	d := app.NewDeployer(stub, nil)

	cfg := &config.DeployConfig{
		Cluster:       "my-cluster",
		Service:       "my-service",
		Image:         "new-image:sha",
		ContainerName: "app",
		TaskDef:       "testdata/task-def.json",
		Wait:          false,
	}

	if err := d.Deploy(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	if !stub.registerCalled {
		t.Error("RegisterTaskDefinition not called")
	}
	if !stub.updateCalled {
		t.Error("UpdateService not called")
	}
	if stub.waitCalled {
		t.Error("WaitStable should not be called when wait=false")
	}
}

func TestDeploy_WaitsWhenFlagSet(t *testing.T) {
	stub := &stubECS{registerARN: "arn:aws:ecs:us-east-1:123:task-definition/fam:1"}
	d := app.NewDeployer(stub, nil)

	cfg := &config.DeployConfig{
		Cluster:       "my-cluster",
		Service:       "my-service",
		Image:         "new-image:sha",
		ContainerName: "app",
		TaskDef:       "testdata/task-def.json",
		Wait:          true,
	}

	if err := d.Deploy(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	if !stub.waitCalled {
		t.Error("WaitStable should be called when wait=true")
	}
}
