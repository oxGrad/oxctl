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
	createCalled   bool
	waitCalled     bool
	logGroupCalled bool
	insightsCalled bool
	registerARN    string
	serviceStatus  ecs.ServiceStatus
}

func (s *stubECS) RegisterTaskDefinition(_ context.Context, _ map[string]any) (string, error) {
	s.registerCalled = true
	return s.registerARN, nil
}
func (s *stubECS) UpdateService(_ context.Context, _, _, _ string) error {
	s.updateCalled = true
	return nil
}
func (s *stubECS) CreateService(_ context.Context, _ ecs.CreateServiceInput) error {
	s.createCalled = true
	return nil
}
func (s *stubECS) WaitStable(_ context.Context, _, _ string) error {
	s.waitCalled = true
	return nil
}
func (s *stubECS) DescribeService(_ context.Context, _, _ string) (ecs.ServiceStatus, error) {
	return s.serviceStatus, nil
}
func (s *stubECS) EnsureLogGroup(_ context.Context, _ string) error {
	s.logGroupCalled = true
	return nil
}
func (s *stubECS) UpdateClusterInsights(_ context.Context, _ string) error {
	s.insightsCalled = true
	return nil
}

func baseCfg() *config.DeployConfig {
	return &config.DeployConfig{
		Cluster:       "my-cluster",
		Service:       "my-service",
		Image:         "new-image:sha",
		ContainerName: "app",
		TaskDef:       "testdata/task-def.json",
	}
}

func TestDeploy_UpdatesExistingService(t *testing.T) {
	stub := &stubECS{
		registerARN:   "arn:aws:ecs:us-east-1:123:task-definition/fam:1",
		serviceStatus: ecs.ServiceStatus{Status: "ACTIVE"},
	}
	d := app.NewDeployer(stub, nil)

	if err := d.Deploy(context.Background(), baseCfg()); err != nil {
		t.Fatal(err)
	}
	if !stub.registerCalled {
		t.Error("RegisterTaskDefinition not called")
	}
	if !stub.updateCalled {
		t.Error("UpdateService not called for ACTIVE service")
	}
	if stub.createCalled {
		t.Error("CreateService should not be called for ACTIVE service")
	}
	if !stub.logGroupCalled {
		t.Error("EnsureLogGroup not called")
	}
	if !stub.insightsCalled {
		t.Error("UpdateClusterInsights not called")
	}
}

func TestDeploy_CreatesNewService(t *testing.T) {
	stub := &stubECS{
		registerARN:   "arn:aws:ecs:us-east-1:123:task-definition/fam:1",
		serviceStatus: ecs.ServiceStatus{Status: "MISSING"},
	}
	d := app.NewDeployer(stub, nil)

	cfg := baseCfg()
	cfg.SubnetIDs = []string{"subnet-aaa"}
	cfg.SecurityGroupIDs = []string{"sg-bbb"}
	cfg.TargetGroupARN = "arn:aws:elasticloadbalancing:us-east-1:123:targetgroup/tg/abc"
	cfg.DesiredCount = 1

	if err := d.Deploy(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	if !stub.createCalled {
		t.Error("CreateService not called for missing service")
	}
	if stub.updateCalled {
		t.Error("UpdateService should not be called for missing service")
	}
}

func TestDeploy_WaitsWhenFlagSet(t *testing.T) {
	stub := &stubECS{
		registerARN:   "arn:aws:ecs:us-east-1:123:task-definition/fam:1",
		serviceStatus: ecs.ServiceStatus{Status: "ACTIVE"},
	}
	d := app.NewDeployer(stub, nil)

	cfg := baseCfg()
	cfg.Wait = true

	if err := d.Deploy(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	if !stub.waitCalled {
		t.Error("WaitStable should be called when wait=true")
	}
}

func TestDeploy_NoWaitByDefault(t *testing.T) {
	stub := &stubECS{
		registerARN:   "arn:aws:ecs:us-east-1:123:task-definition/fam:1",
		serviceStatus: ecs.ServiceStatus{Status: "ACTIVE"},
	}
	d := app.NewDeployer(stub, nil)

	if err := d.Deploy(context.Background(), baseCfg()); err != nil {
		t.Fatal(err)
	}
	if stub.waitCalled {
		t.Error("WaitStable should not be called when wait=false")
	}
}
