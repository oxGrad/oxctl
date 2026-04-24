package ecs_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/oxGrad/oxctl/internal/ecs"
)

func writeTaskDef(t *testing.T, dir string, def map[string]any) string {
	t.Helper()
	b, _ := json.Marshal(def)
	path := filepath.Join(dir, "task-def.json")
	os.WriteFile(path, b, 0644)
	return path
}

func TestLoadAndPatch_InjectsImage(t *testing.T) {
	dir := t.TempDir()
	def := map[string]any{
		"family": "my-family",
		"containerDefinitions": []any{
			map[string]any{"name": "app", "image": "old-image:1"},
			map[string]any{"name": "sidecar", "image": "sidecar:latest"},
		},
	}
	path := writeTaskDef(t, dir, def)

	patched, err := ecs.LoadAndPatch(path, "app", "new-image:sha")
	if err != nil {
		t.Fatal(err)
	}

	containers := patched["containerDefinitions"].([]any)
	app := containers[0].(map[string]any)
	sidecar := containers[1].(map[string]any)

	if app["image"] != "new-image:sha" {
		t.Errorf("app image not patched: %v", app["image"])
	}
	if sidecar["image"] != "sidecar:latest" {
		t.Errorf("sidecar image should be unchanged: %v", sidecar["image"])
	}
}

func TestLoadAndPatch_ContainerNotFound(t *testing.T) {
	dir := t.TempDir()
	def := map[string]any{
		"containerDefinitions": []any{
			map[string]any{"name": "app", "image": "old:1"},
		},
	}
	path := writeTaskDef(t, dir, def)

	_, err := ecs.LoadAndPatch(path, "nonexistent", "new:sha")
	if err == nil {
		t.Fatal("expected error for missing container name")
	}
}

func TestLoadAndPatch_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "task-def.json")
	os.WriteFile(path, []byte("{bad json"), 0644)

	_, err := ecs.LoadAndPatch(path, "app", "img:1")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
