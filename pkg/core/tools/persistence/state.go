package persistence

import (
	"path/filepath"

	"github.com/blackcoderx/falcon/pkg/storage"
)

// PersistenceManager maintains state for persistence tools (environment, etc.)
// Renamed from PersistenceTool to avoid confusion with the package name or specific tools.
type PersistenceManager struct {
	baseDir     string
	currentEnv  string
	environment map[string]string
}

// NewPersistenceManager creates a new persistence manager
func NewPersistenceManager(baseDir string) *PersistenceManager {
	return &PersistenceManager{
		baseDir:     baseDir,
		currentEnv:  "",
		environment: make(map[string]string),
	}
}

// SetEnvironment sets the current environment by name
func (pm *PersistenceManager) SetEnvironment(name string) error {
	envPath := filepath.Join(storage.GetEnvironmentsDir(pm.baseDir), name+".yaml")
	env, err := storage.LoadEnvironment(envPath)
	if err != nil {
		return err
	}
	pm.currentEnv = name
	pm.environment = env
	return nil
}

// GetEnvironment returns the current environment variables
func (pm *PersistenceManager) GetEnvironment() map[string]string {
	return pm.environment
}

// GetBaseDir returns the base directory
func (pm *PersistenceManager) GetBaseDir() string {
	return pm.baseDir
}

// GetCurrentEnv returns the name of the current environment
func (pm *PersistenceManager) GetCurrentEnv() string {
	return pm.currentEnv
}
