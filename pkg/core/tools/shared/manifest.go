package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Manifest represents the .zap folder manifest file.
// It helps the agent understand the knowledge base structure.
type Manifest struct {
	Version     int            `json:"version"`
	Description string         `json:"description"`
	Counts      map[string]int `json:"counts"`
	LastUpdated string         `json:"last_updated"`
}

// ManifestFilename is the name of the manifest file
const ManifestFilename = "manifest.json"

// CreateManifest creates a new manifest.json file in the .zap directory.
func CreateManifest(zapDir string) error {
	manifest := &Manifest{
		Version:     1,
		Description: "Falcon knowledge base - saved requests, environments, and test artifacts",
		Counts:      make(map[string]int),
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	// Initialize counts
	manifest.Counts["requests"] = 0
	manifest.Counts["environments"] = 0
	manifest.Counts["baselines"] = 0
	manifest.Counts["variables"] = 0

	return saveManifest(zapDir, manifest)
}

// LoadManifest reads the manifest file from the .zap directory.
// Returns nil if the file doesn't exist.
func LoadManifest(zapDir string) (*Manifest, error) {
	manifestPath := filepath.Join(zapDir, ManifestFilename)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// UpdateManifestCounts scans the .zap directory and updates file counts.
func UpdateManifestCounts(zapDir string) error {
	manifest, err := LoadManifest(zapDir)
	if err != nil {
		return err
	}

	// If manifest doesn't exist, create it
	if manifest == nil {
		if err := CreateManifest(zapDir); err != nil {
			return err
		}
		manifest, _ = LoadManifest(zapDir)
	}

	// Count requests
	requestsDir := filepath.Join(zapDir, "requests")
	manifest.Counts["requests"] = countYAMLFiles(requestsDir)

	// Count environments
	environmentsDir := filepath.Join(zapDir, "environments")
	manifest.Counts["environments"] = countYAMLFiles(environmentsDir)

	// Count baselines
	baselinesDir := filepath.Join(zapDir, "baselines")
	manifest.Counts["baselines"] = countJSONFiles(baselinesDir)

	// Count global variables (if variables.json exists)
	variablesPath := filepath.Join(zapDir, "variables.json")
	if _, err := os.Stat(variablesPath); err == nil {
		data, err := os.ReadFile(variablesPath)
		if err == nil {
			var vars map[string]string
			if json.Unmarshal(data, &vars) == nil {
				manifest.Counts["variables"] = len(vars)
			}
		}
	}

	// Update timestamp
	manifest.LastUpdated = time.Now().Format(time.RFC3339)

	return saveManifest(zapDir, manifest)
}

// saveManifest writes the manifest to disk
func saveManifest(zapDir string, manifest *Manifest) error {
	manifestPath := filepath.Join(zapDir, ManifestFilename)
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// countYAMLFiles counts .yaml and .yml files in a directory
func countYAMLFiles(dir string) int {
	count := 0
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext == ".yaml" || ext == ".yml" {
			count++
		}
	}
	return count
}

// countJSONFiles counts .json files in a directory
func countJSONFiles(dir string) int {
	count := 0
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".json" {
			count++
		}
	}
	return count
}

// GetManifestSummary returns a human-readable summary of the manifest for the agent
func GetManifestSummary(zapDir string) string {
	manifest, err := LoadManifest(zapDir)
	if err != nil || manifest == nil {
		return ""
	}

	return fmt.Sprintf(
		"Knowledge base: %d requests, %d environments, %d baselines, %d global variables (last updated: %s)",
		manifest.Counts["requests"],
		manifest.Counts["environments"],
		manifest.Counts["baselines"],
		manifest.Counts["variables"],
		manifest.LastUpdated,
	)
}
