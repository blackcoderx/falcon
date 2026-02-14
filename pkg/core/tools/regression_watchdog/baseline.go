package regression_watchdog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/blackcoderx/zap/pkg/core/tools/shared"
)

// BaselineStore manages the persistence of API snapshots.
type BaselineStore struct {
	ZapDir string
}

// APIBaseline represents a stored snapshot of API behavior.
type APIBaseline struct {
	Name      string                         `json:"name"`
	CreatedAt time.Time                      `json:"created_at"`
	Snapshots map[string]shared.HTTPResponse `json:"snapshots"`
}

// NewBaselineStore creates a new baseline manager.
func NewBaselineStore(zapDir string) *BaselineStore {
	return &BaselineStore{ZapDir: zapDir}
}

// Save persists a baseline to disk.
func (s *BaselineStore) Save(b APIBaseline) error {
	dir := filepath.Join(s.ZapDir, "baselines")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, b.Name+".json")
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Load retrieves a baseline from disk.
func (s *BaselineStore) Load(name string) (*APIBaseline, error) {
	path := filepath.Join(s.ZapDir, "baselines", name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("baseline '%s' not found", name)
		}
		return nil, err
	}

	var b APIBaseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}

	return &b, nil
}
