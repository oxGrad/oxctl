package ecs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oxGrad/oxctl/internal/runner"
)

// ServiceStatus represents the state of an ECS service.
type ServiceStatus struct {
	RunningCount   int
	DesiredCount   int
	TaskDefinition string
	Deployments    []Deployment
}

// Deployment represents a single ECS deployment entry.
type Deployment struct {
	Status       string `json:"status"`
	RolloutState string `json:"rolloutState"`
}

// ECSDeployer performs ECS operations.
type ECSDeployer interface {
	RegisterTaskDefinition(ctx context.Context, def map[string]any) (string, error)
	UpdateService(ctx context.Context, cluster, service, taskDefArn string) error
	WaitStable(ctx context.Context, cluster, service string) error
	DescribeService(ctx context.Context, cluster, service string) (ServiceStatus, error)
}

// AWSCLIDeployer implements ECSDeployer using the aws CLI.
type AWSCLIDeployer struct {
	r runner.CommandRunner
}

// NewAWSCLIDeployer creates a deployer backed by the aws CLI.
func NewAWSCLIDeployer(r runner.CommandRunner) *AWSCLIDeployer {
	return &AWSCLIDeployer{r: r}
}

func (d *AWSCLIDeployer) RegisterTaskDefinition(ctx context.Context, def map[string]any) (string, error) {
	b, err := json.Marshal(def)
	if err != nil {
		return "", fmt.Errorf("marshalling task def: %w", err)
	}
	out, err := d.r.Output(ctx, "aws", "ecs", "register-task-definition", "--cli-input-json", string(b))
	if err != nil {
		return "", fmt.Errorf("register-task-definition: %w", err)
	}
	if out == nil {
		return "", nil // dry-run
	}
	var resp struct {
		TaskDefinition struct {
			TaskDefinitionArn string `json:"taskDefinitionArn"`
		} `json:"taskDefinition"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", fmt.Errorf("parsing register response: %w", err)
	}
	return resp.TaskDefinition.TaskDefinitionArn, nil
}

func (d *AWSCLIDeployer) UpdateService(ctx context.Context, cluster, service, taskDefArn string) error {
	return d.r.Run(ctx, "aws", "ecs", "update-service",
		"--cluster", cluster,
		"--service", service,
		"--task-definition", taskDefArn,
	)
}

func (d *AWSCLIDeployer) WaitStable(ctx context.Context, cluster, service string) error {
	return d.r.Run(ctx, "aws", "ecs", "wait", "services-stable",
		"--cluster", cluster,
		"--services", service,
	)
}

func (d *AWSCLIDeployer) DescribeService(ctx context.Context, cluster, service string) (ServiceStatus, error) {
	out, err := d.r.Output(ctx, "aws", "ecs", "describe-services",
		"--cluster", cluster,
		"--services", service,
	)
	if err != nil {
		return ServiceStatus{}, fmt.Errorf("describe-services: %w", err)
	}
	var resp struct {
		Services []struct {
			RunningCount   int          `json:"runningCount"`
			DesiredCount   int          `json:"desiredCount"`
			TaskDefinition string       `json:"taskDefinition"`
			Deployments    []Deployment `json:"deployments"`
		} `json:"services"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return ServiceStatus{}, fmt.Errorf("parsing describe response: %w", err)
	}
	if len(resp.Services) == 0 {
		return ServiceStatus{}, fmt.Errorf("service %q not found in cluster %q", service, cluster)
	}
	s := resp.Services[0]
	return ServiceStatus{
		RunningCount:   s.RunningCount,
		DesiredCount:   s.DesiredCount,
		TaskDefinition: s.TaskDefinition,
		Deployments:    s.Deployments,
	}, nil
}
