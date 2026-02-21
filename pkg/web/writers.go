package web

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blackcoderx/zap/pkg/core"
	"github.com/blackcoderx/zap/pkg/storage"
)

// atomicWrite writes data to path via a temp file + rename to avoid partial writes.
func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("atomicWrite: %w", err)
	}
	return os.Rename(tmp, path)
}

func writeConfig(zapDir string, cfg *core.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(filepath.Join(zapDir, "config.json"), data)
}

func writeMemory(zapDir string, entries []core.MemoryEntry) error {
	mf := struct {
		Version int               `json:"version"`
		Entries []core.MemoryEntry `json:"entries"`
	}{Version: 1, Entries: entries}
	data, err := json.MarshalIndent(mf, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(filepath.Join(zapDir, "memory.json"), data)
}

func writeRequest(zapDir, name string, req *storage.Request) error {
	return storage.SaveRequest(*req, filepath.Join(zapDir, "requests", name+".yaml"))
}

func deleteRequestFile(zapDir, name string) error {
	path := filepath.Join(zapDir, "requests", name+".yaml")
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func writeEnvironment(zapDir, name string, vars map[string]string) error {
	return storage.SaveEnvironment(vars, filepath.Join(zapDir, "environments", name+".yaml"))
}

func deleteEnvironmentFile(zapDir, name string) error {
	path := filepath.Join(zapDir, "environments", name+".yaml")
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func writeVariables(zapDir string, vars map[string]string) error {
	data, err := json.MarshalIndent(vars, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(filepath.Join(zapDir, "variables.json"), data)
}
