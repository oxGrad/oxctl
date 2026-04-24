package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultOxconfPath returns the default oxconf file path.
func DefaultOxconfPath() string {
	return "./oxconf"
}

type oxconfFile struct {
	Cluster       string `yaml:"cluster"`
	Service       string `yaml:"service"`
	Image         string `yaml:"image"`
	ContainerName string `yaml:"container-name"`
	TaskDef       string `yaml:"task-def"`
	Wait          bool   `yaml:"wait"`
	Timeout       int    `yaml:"timeout"` // seconds
}

// LoadOxconf reads and parses an oxconf YAML file at the given path.
func LoadOxconf(path string) (*DeployConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading oxconf: %w", err)
	}
	var f oxconfFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing oxconf: %w", err)
	}
	return &DeployConfig{
		Cluster:       f.Cluster,
		Service:       f.Service,
		Image:         f.Image,
		ContainerName: f.ContainerName,
		TaskDef:       f.TaskDef,
		Wait:          f.Wait,
		Timeout:       time.Duration(f.Timeout) * time.Second,
	}, nil
}
