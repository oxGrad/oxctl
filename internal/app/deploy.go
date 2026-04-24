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

// Deploy registers a new task definition revision and updates the ECS service.
func (d *Deployer) Deploy(ctx context.Context, cfg *config.DeployConfig) error {
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

	d.logger.Info("updating service", "cluster", cfg.Cluster, "service", cfg.Service)
	if err := d.ecs.UpdateService(ctx, cfg.Cluster, cfg.Service, arn); err != nil {
		return fmt.Errorf("updating service: %w", err)
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
