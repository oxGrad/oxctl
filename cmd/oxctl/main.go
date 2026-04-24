package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/oxGrad/oxctl/internal/app"
	"github.com/oxGrad/oxctl/internal/config"
	"github.com/oxGrad/oxctl/internal/ecs"
	oxlog "github.com/oxGrad/oxctl/internal/log"
	"github.com/oxGrad/oxctl/internal/runner"
	"github.com/oxGrad/oxctl/internal/tui"
	"github.com/oxGrad/oxctl/pkg/util"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var (
		jsonLog  bool
		debugLog bool
		dryRun   bool
	)

	root := &cobra.Command{
		Use:   "oxctl",
		Short: "ECS deployment tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			if util.IsTTY() && !util.IsCI() {
				return tui.Run()
			}
			return cmd.Help()
		},
	}
	root.PersistentFlags().BoolVar(&jsonLog, "json-log", false, "Output logs as JSON")
	root.PersistentFlags().BoolVar(&debugLog, "debug", false, "Enable debug logging")
	root.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Print AWS CLI commands without executing")

	root.AddCommand(deployCmd(&jsonLog, &debugLog, &dryRun))
	root.AddCommand(statusCmd(&jsonLog, &debugLog, &dryRun))
	return root
}

func buildLogger(w *os.File, jsonLog, debugLog bool) *slog.Logger {
	return oxlog.New(w, jsonLog, debugLog)
}

func deployCmd(jsonLog, debugLog, dryRun *bool) *cobra.Command {
	var (
		oxconfPath    string
		useOxconf     bool
		cluster       string
		service       string
		image         string
		containerName string
		taskDef       string
		wait          bool
		timeout       int
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Register a new ECS task definition and update the service",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := buildLogger(os.Stdout, *jsonLog, *debugLog)
			ctx := context.Background()

			var cfg *config.DeployConfig
			if useOxconf {
				path := config.DefaultOxconfPath()
				if oxconfPath != "" {
					path = oxconfPath
				}
				var err error
				cfg, err = config.LoadOxconf(path)
				if err != nil {
					return fmt.Errorf("loading oxconf: %w", err)
				}
			} else {
				cfg = &config.DeployConfig{
					Cluster:       cluster,
					Service:       service,
					Image:         image,
					ContainerName: containerName,
					TaskDef:       taskDef,
					Wait:          wait,
					Timeout:       time.Duration(timeout) * time.Second,
				}
			}

			r := runner.NewExecRunner(*dryRun, nil)
			deployer := app.NewDeployer(ecs.NewAWSCLIDeployer(r), logger)
			return deployer.Deploy(ctx, cfg)
		},
	}

	cmd.Flags().BoolVar(&useOxconf, "oxconf", false, "Load configuration from oxconf file (ignores all other flags)")
	cmd.Flags().StringVar(&oxconfPath, "oxconf-path", "", "Path to oxconf file (used with --oxconf; default: ./oxconf)")
	cmd.Flags().StringVar(&cluster, "cluster", "", "ECS cluster name")
	cmd.Flags().StringVar(&service, "service", "", "ECS service name")
	cmd.Flags().StringVar(&image, "image", "", "Container image URI")
	cmd.Flags().StringVar(&containerName, "container-name", "", "Container name in task definition")
	cmd.Flags().StringVar(&taskDef, "task-def", "", "Path to task definition JSON file")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait for service stability")
	cmd.Flags().IntVar(&timeout, "timeout", 300, "Timeout in seconds for stability wait")

	return cmd
}

func statusCmd(jsonLog, debugLog, dryRun *bool) *cobra.Command {
	var (
		oxconfPath string
		useOxconf  bool
		cluster    string
		service    string
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show ECS service status",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			var clust, svc string
			if useOxconf {
				path := config.DefaultOxconfPath()
				if oxconfPath != "" {
					path = oxconfPath
				}
				cfg, err := config.LoadOxconf(path)
				if err != nil {
					return fmt.Errorf("loading oxconf: %w", err)
				}
				clust = cfg.Cluster
				svc = cfg.Service
			} else {
				clust = cluster
				svc = service
			}

			r := runner.NewExecRunner(*dryRun, nil)
			reporter := app.NewStatusReporter(ecs.NewAWSCLIDeployer(r), os.Stdout)
			return reporter.Report(ctx, clust, svc)
		},
	}

	cmd.Flags().BoolVar(&useOxconf, "oxconf", false, "Load cluster and service from oxconf file")
	cmd.Flags().StringVar(&oxconfPath, "oxconf-path", "", "Path to oxconf file (used with --oxconf; default: ./oxconf)")
	cmd.Flags().StringVar(&cluster, "cluster", "", "ECS cluster name")
	cmd.Flags().StringVar(&service, "service", "", "ECS service name")

	return cmd
}
