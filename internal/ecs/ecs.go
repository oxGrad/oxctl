package ecs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oxGrad/oxctl/internal/runner"
)

// ServiceStatus represents the state of an ECS service.
type ServiceStatus struct {
	Status         string // "ACTIVE", "DRAINING", "INACTIVE", or "MISSING"
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

// CreateServiceInput holds all parameters for creating a new ECS Fargate service.
type CreateServiceInput struct {
	Cluster          string
	ServiceName      string
	TaskDefARN       string
	DesiredCount     int
	SubnetIDs        []string
	SecurityGroupIDs []string
	TargetGroupARN   string
	ContainerName    string
	ContainerPort    int
}

// ECSDeployer performs ECS operations.
type ECSDeployer interface {
	RegisterTaskDefinition(ctx context.Context, def map[string]any) (string, error)
	UpdateService(ctx context.Context, cluster, service, taskDefArn string) error
	CreateService(ctx context.Context, in CreateServiceInput) error
	WaitStable(ctx context.Context, cluster, service string) error
	DescribeService(ctx context.Context, cluster, service string) (ServiceStatus, error)
	EnsureLogGroup(ctx context.Context, logGroup string) error
	UpdateClusterInsights(ctx context.Context, cluster string) error
}

// AWSCLIDeployer implements ECSDeployer using the aws CLI.
type AWSCLIDeployer struct {
	r runner.CommandRunner
}

// NewAWSCLIDeployer creates a deployer backed by the aws CLI.
func NewAWSCLIDeployer(r runner.CommandRunner) *AWSCLIDeployer {
	return &AWSCLIDeployer{r: r}
}

const deploymentConfig = "deploymentCircuitBreaker={enable=true,rollback=true},maximumPercent=200,minimumHealthyPercent=100"

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
		"--health-check-grace-period-seconds", "60",
		"--deployment-configuration", deploymentConfig,
		"--enable-execute-command",
	)
}

func (d *AWSCLIDeployer) CreateService(ctx context.Context, in CreateServiceInput) error {
	networkCfg := fmt.Sprintf(
		"awsvpcConfiguration={subnets=[%s],securityGroups=[%s],assignPublicIp=DISABLED}",
		strings.Join(in.SubnetIDs, ","),
		strings.Join(in.SecurityGroupIDs, ","),
	)
	lbCfg := fmt.Sprintf(
		"targetGroupArn=%s,containerName=%s,containerPort=%d",
		in.TargetGroupARN, in.ContainerName, in.ContainerPort,
	)
	return d.r.Run(ctx, "aws", "ecs", "create-service",
		"--cluster", in.Cluster,
		"--service-name", in.ServiceName,
		"--task-definition", in.TaskDefARN,
		"--desired-count", fmt.Sprintf("%d", in.DesiredCount),
		"--launch-type", "FARGATE",
		"--health-check-grace-period-seconds", "60",
		"--deployment-configuration", deploymentConfig,
		"--enable-execute-command",
		"--network-configuration", networkCfg,
		"--load-balancers", lbCfg,
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
			Status         string       `json:"status"`
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
		return ServiceStatus{Status: "MISSING"}, nil
	}
	s := resp.Services[0]
	return ServiceStatus{
		Status:         s.Status,
		RunningCount:   s.RunningCount,
		DesiredCount:   s.DesiredCount,
		TaskDefinition: s.TaskDefinition,
		Deployments:    s.Deployments,
	}, nil
}

func (d *AWSCLIDeployer) EnsureLogGroup(ctx context.Context, logGroup string) error {
	// Ignore "already exists" — mirrors `|| true` in the pipeline script.
	_ = d.r.Run(ctx, "aws", "logs", "create-log-group", "--log-group-name", logGroup)
	return nil
}

func (d *AWSCLIDeployer) UpdateClusterInsights(ctx context.Context, cluster string) error {
	return d.r.Run(ctx, "aws", "ecs", "update-cluster-settings",
		"--cluster", cluster,
		"--settings", "name=containerInsights,value=enhanced",
	)
}
