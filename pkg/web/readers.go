package web

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/blackcoderx/zap/pkg/core"
	"github.com/blackcoderx/zap/pkg/storage"
	"gopkg.in/yaml.v3"
)

func readConfig(zapDir string) (*core.Config, error) {
	// Try YAML first (new format), fall back to JSON (legacy)
	yamlPath := filepath.Join(zapDir, "config.yaml")
	if data, err := os.ReadFile(yamlPath); err == nil {
		var cfg core.Config
		return &cfg, yaml.Unmarshal(data, &cfg)
	}
	data, err := os.ReadFile(filepath.Join(zapDir, "config.json"))
	if err != nil {
		return nil, err
	}
	var cfg core.Config
	return &cfg, json.Unmarshal(data, &cfg)
}

type manifestData struct {
	Version     int            `json:"version"`
	Description string         `json:"description"`
	Counts      map[string]int `json:"counts"`
	LastUpdated string         `json:"last_updated"`
}

func readManifest(zapDir string) (*manifestData, error) {
	data, err := os.ReadFile(filepath.Join(zapDir, "manifest.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return &manifestData{Counts: map[string]int{}}, nil
		}
		return nil, err
	}
	var m manifestData
	return &m, json.Unmarshal(data, &m)
}

func readMemory(zapDir string) ([]core.MemoryEntry, error) {
	data, err := os.ReadFile(filepath.Join(zapDir, "memory.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var mf struct {
		Version int               `json:"version"`
		Entries []core.MemoryEntry `json:"entries"`
	}
	if err := json.Unmarshal(data, &mf); err != nil {
		return nil, err
	}
	if mf.Entries == nil {
		return []core.MemoryEntry{}, nil
	}
	return mf.Entries, nil
}

func readHistory(_ string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func readRequest(zapDir, name string) (*storage.Request, error) {
	path := filepath.Join(zapDir, "requests", name+".yaml")
	return storage.LoadRequest(path)
}

func listRequestNames(zapDir string) ([]string, error) {
	names, err := storage.ListRequests(zapDir)
	if err != nil {
		return nil, err
	}
	// Strip .yaml / .yml extensions returned by ListRequests
	result := make([]string, 0, len(names))
	for _, n := range names {
		n = strings.TrimSuffix(n, ".yaml")
		n = strings.TrimSuffix(n, ".yml")
		result = append(result, n)
	}
	return result, nil
}

func readEnvironment(zapDir, name string) (map[string]string, error) {
	path := filepath.Join(zapDir, "environments", name+".yaml")
	env, err := storage.LoadEnvironment(path)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return map[string]string{}, nil
	}
	return env, nil
}

func listEnvironmentNames(zapDir string) ([]string, error) {
	return storage.ListEnvironments(zapDir)
}

func readVariables(zapDir string) (map[string]string, error) {
	data, err := os.ReadFile(filepath.Join(zapDir, "variables.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	var vars map[string]string
	if err := json.Unmarshal(data, &vars); err != nil {
		return nil, err
	}
	if vars == nil {
		return map[string]string{}, nil
	}
	return vars, nil
}

func readAPIGraph(zapDir string) (json.RawMessage, error) {
	data, err := os.ReadFile(filepath.Join(zapDir, "snapshots", "api-graph.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return json.RawMessage("null"), nil
		}
		return nil, err
	}
	return data, nil
}

func listExportFiles(zapDir string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(zapDir, "exports"))
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	if names == nil {
		return []string{}, nil
	}
	return names, nil
}

func readExportFile(zapDir, name string) ([]byte, error) {
	clean := filepath.Base(name)
	return os.ReadFile(filepath.Join(zapDir, "exports", clean))
}
