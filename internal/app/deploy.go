package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/oxGrad/oxctl/internal/config"
	"github.com/oxGrad/oxctl/internal/ecs"
)

// Deployer orchestrates an ECS deployment.
type Deployer struct {
	ecs    ecs.ECSDeployer
	logger *slog.Logger
}

// NewDeployer creates a Deployer. If logger is nil, a default stdout logger is used.
func NewDeployer(e ecs.ECSDeployer, logger *slog.Logger) *Deployer {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	return &Deployer{ecs: e, logger: logger}
}

// Deploy registers a new task definition revision, then creates or updates the ECS service.
func (d *Deployer) Deploy(ctx context.Context, cfg *config.DeployConfig) error {
	logGroup := cfg.LogGroup
	if logGroup == "" {
		logGroup = "/ecs/" + cfg.Service
	}

	d.logger.Info("enabling cluster container insights", "cluster", cfg.Cluster)
	if err := d.ecs.UpdateClusterInsights(ctx, cfg.Cluster); err != nil {
		return fmt.Errorf("updating cluster insights: %w", err)
	}

	d.logger.Info("ensuring log group", "log-group", logGroup)
	if err := d.ecs.EnsureLogGroup(ctx, logGroup); err != nil {
		return fmt.Errorf("ensuring log group: %w", err)
	}

	d.logger.Info("patching task definition", "container", cfg.ContainerName, "image", cfg.Image)
	def, err := ecs.LoadAndPatch(cfg.TaskDef, cfg.ContainerName, cfg.Image)
	if err != nil {
		return fmt.Errorf("loading task definition: %w", err)
	}

	d.logger.Info("registering task definition")
	arn, err := d.ecs.RegisterTaskDefinition(ctx, def)
	if err != nil {
		return fmt.Errorf("registering task definition: %w", err)
	}
	d.logger.Info("task definition registered", "arn", arn)

	d.logger.Info("checking service status", "cluster", cfg.Cluster, "service", cfg.Service)
	svcStatus, err := d.ecs.DescribeService(ctx, cfg.Cluster, cfg.Service)
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}

	if svcStatus.Status == "ACTIVE" {
		d.logger.Info("service exists — updating", "cluster", cfg.Cluster, "service", cfg.Service)
		if err := d.ecs.UpdateService(ctx, cfg.Cluster, cfg.Service, arn); err != nil {
			return fmt.Errorf("updating service: %w", err)
		}
	} else {
		d.logger.Info("service not found — creating", "cluster", cfg.Cluster, "service", cfg.Service)
		desiredCount := cfg.DesiredCount
		if desiredCount == 0 {
			desiredCount = 1
		}
		containerPort := cfg.ContainerPort
		if containerPort == 0 {
			containerPort = 3000
		}
		if err := d.ecs.CreateService(ctx, ecs.CreateServiceInput{
			Cluster:          cfg.Cluster,
			ServiceName:      cfg.Service,
			TaskDefARN:       arn,
			DesiredCount:     desiredCount,
			SubnetIDs:        cfg.SubnetIDs,
			SecurityGroupIDs: cfg.SecurityGroupIDs,
			TargetGroupARN:   cfg.TargetGroupARN,
			ContainerName:    cfg.ContainerName,
			ContainerPort:    containerPort,
		}); err != nil {
			return fmt.Errorf("creating service: %w", err)
		}
	}

	if cfg.Wait {
		d.logger.Info("waiting for service stability")
		if err := d.ecs.WaitStable(ctx, cfg.Cluster, cfg.Service); err != nil {
			return fmt.Errorf("waiting for stability: %w", err)
		}
		d.logger.Info("service is stable")
	}

	return nil
}
