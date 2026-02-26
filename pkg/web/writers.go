package web

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blackcoderx/falcon/pkg/core"
	"github.com/blackcoderx/falcon/pkg/storage"
	"gopkg.in/yaml.v3"
)

// atomicWrite writes data to path via a temp file + rename to avoid partial writes.
func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("atomicWrite: %w", err)
	}
	return os.Rename(tmp, path)
}

func writeConfig(falconDir string, cfg *core.Config) error {
	// Write YAML if config.yaml exists, otherwise fall back to JSON
	yamlPath := filepath.Join(falconDir, "config.yaml")
	if _, err := os.Stat(yamlPath); err == nil {
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		return atomicWrite(yamlPath, data)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(filepath.Join(falconDir, "config.json"), data)
}

func writeMemory(falconDir string, entries []core.MemoryEntry) error {
	mf := struct {
		Version int               `json:"version"`
		Entries []core.MemoryEntry `json:"entries"`
	}{Version: 1, Entries: entries}
	data, err := json.MarshalIndent(mf, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(filepath.Join(falconDir, "memory.json"), data)
}

func writeRequest(falconDir, name string, req *storage.Request) error {
	return storage.SaveRequest(*req, filepath.Join(falconDir, "requests", name+".yaml"))
}

func deleteRequestFile(falconDir, name string) error {
	path := filepath.Join(falconDir, "requests", name+".yaml")
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func writeEnvironment(falconDir, name string, vars map[string]string) error {
	return storage.SaveEnvironment(vars, filepath.Join(falconDir, "environments", name+".yaml"))
}

func deleteEnvironmentFile(falconDir, name string) error {
	path := filepath.Join(falconDir, "environments", name+".yaml")
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func writeVariables(falconDir string, vars map[string]string) error {
	data, err := json.MarshalIndent(vars, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(filepath.Join(falconDir, "variables.json"), data)
}
