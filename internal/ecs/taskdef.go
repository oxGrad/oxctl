package ecs

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadAndPatch reads a task definition JSON file, finds the container with the
// given name, replaces its image, and returns the patched definition as a map.
func LoadAndPatch(path, containerName, image string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading task def: %w", err)
	}

	var def map[string]any
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("parsing task def JSON: %w", err)
	}

	containers, ok := def["containerDefinitions"].([]any)
	if !ok {
		return nil, fmt.Errorf("task def missing containerDefinitions array")
	}

	found := false
	for _, c := range containers {
		m, ok := c.(map[string]any)
		if !ok {
			continue
		}
		if m["name"] == containerName {
			m["image"] = image
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("container %q not found in task def", containerName)
	}

	return def, nil
}
