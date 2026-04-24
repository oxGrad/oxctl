package config

import "time"

// DeployConfig holds all parameters required for a deploy operation.
type DeployConfig struct {
	Cluster       string
	Service       string
	Image         string
	ContainerName string
	TaskDef       string
	Wait          bool
	Timeout       time.Duration
}

// StatusConfig holds parameters for a status query.
type StatusConfig struct {
	Cluster string
	Service string
}
