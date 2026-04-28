package config

import "time"

// DeployConfig holds all parameters required for a deploy operation.
type DeployConfig struct {
	Cluster       string
	Service       string
	Image         string
	ContainerName string
	ContainerPort int
	TaskDef       string
	Wait          bool
	Timeout       time.Duration

	// Service creation (only required when the ECS service does not exist yet)
	DesiredCount     int
	SubnetIDs        []string
	SecurityGroupIDs []string
	TargetGroupARN   string

	// LogGroup is the CloudWatch log group to ensure before deploying.
	// Defaults to /ecs/<service> when empty.
	LogGroup string
}

// StatusConfig holds parameters for a status query.
type StatusConfig struct {
	Cluster string
	Service string
}
