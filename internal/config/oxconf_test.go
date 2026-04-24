package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oxGrad/oxctl/internal/config"
)

func writeOxconf(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "oxconf")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadOxconf_AllFields(t *testing.T) {
	dir := t.TempDir()
	writeOxconf(t, dir, `
cluster: my-cluster
service: my-service
image: 123.dkr.ecr.us-east-1.amazonaws.com/app:abc123
container-name: app
task-def: ./task-def.json
wait: true
timeout: 300
`)
	cfg, err := config.LoadOxconf(filepath.Join(dir, "oxconf"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Cluster != "my-cluster" {
		t.Errorf("cluster: want my-cluster, got %q", cfg.Cluster)
	}
	if cfg.Service != "my-service" {
		t.Errorf("service: want my-service, got %q", cfg.Service)
	}
	if cfg.Image != "123.dkr.ecr.us-east-1.amazonaws.com/app:abc123" {
		t.Errorf("image mismatch: %q", cfg.Image)
	}
	if cfg.ContainerName != "app" {
		t.Errorf("container-name mismatch: %q", cfg.ContainerName)
	}
	if cfg.TaskDef != "./task-def.json" {
		t.Errorf("task-def mismatch: %q", cfg.TaskDef)
	}
	if !cfg.Wait {
		t.Error("wait should be true")
	}
	if cfg.Timeout != 300*time.Second {
		t.Errorf("timeout: want 300s, got %v", cfg.Timeout)
	}
}

func TestLoadOxconf_MissingFile(t *testing.T) {
	_, err := config.LoadOxconf("/tmp/does-not-exist-oxconf")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestDefaultOxconfPath(t *testing.T) {
	want := "./oxconf"
	got := config.DefaultOxconfPath()
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}
