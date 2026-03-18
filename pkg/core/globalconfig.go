package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blackcoderx/falcon/pkg/llm"
	"github.com/charmbracelet/huh"
	"gopkg.in/yaml.v3"
)

// ErrSetupCancelled is returned when the user cancels the setup wizard.
var ErrSetupCancelled = errors.New("setup cancelled by user")

// GlobalFalconDir returns the path to the global ~/.falcon directory.
func GlobalFalconDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".falcon" // fallback
	}
	return filepath.Join(home, ".falcon")
}

// GlobalConfigPath returns the path to the global config file.
func GlobalConfigPath() string {
	return filepath.Join(GlobalFalconDir(), "config.yaml")
}

// EnsureGlobalFalconDir creates ~/.falcon if it doesn't exist (with 0700 permissions).
func EnsureGlobalFalconDir() error {
	dir := GlobalFalconDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create global falcon dir %s: %w", dir, err)
	}
	return nil
}

// LoadGlobalConfig reads the global config. If the file doesn't exist, returns a zero-value Config with no error.
func LoadGlobalConfig() (*Config, error) {
	data, err := os.ReadFile(GlobalConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read global config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse global config: %w", err)
	}
	return &cfg, nil
}

// SaveGlobalConfig writes the config to ~/.falcon/config.yaml with 0600 permissions.
func SaveGlobalConfig(cfg *Config) error {
	if err := EnsureGlobalFalconDir(); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal global config: %w", err)
	}

	if err := os.WriteFile(GlobalConfigPath(), data, 0600); err != nil {
		return fmt.Errorf("failed to write global config: %w", err)
	}
	return nil
}

// RunGlobalConfigWizard runs the interactive provider/model setup wizard and saves to global config.
// Re-uses buildProviderForm and providerOptions from init.go (same package).
func RunGlobalConfigWizard() error {
	var selectedProvider string

	fmt.Println()
	fmt.Println("  Falcon Global Configuration")
	fmt.Println("  Configure your LLM provider credentials.")
	fmt.Println()

	// Phase 1: Provider selection
	providerForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select your LLM provider").
				Description("Choose which AI service to use for assistance.").
				Options(providerOptions()...).
				Value(&selectedProvider),
		),
	).WithTheme(huh.ThemeDracula())

	if err := providerForm.Run(); err != nil {
		return fmt.Errorf("setup cancelled: %w", ErrSetupCancelled)
	}

	p, ok := llm.Get(selectedProvider)
	if !ok {
		return fmt.Errorf("unknown provider %q", selectedProvider)
	}

	// Phase 2: Provider-specific fields
	providerValues := map[string]string{}
	var modelVar string

	fields := p.SetupFields()
	if len(fields) > 0 {
		mv, err := buildProviderForm(p, providerValues)
		if err != nil {
			return err
		}
		modelVar = mv
	}

	// Apply per-field defaults
	for _, f := range fields {
		if providerValues[f.Key] == "" && f.Default != "" {
			providerValues[f.Key] = f.Default
		}
	}
	if modelVar == "" {
		modelVar = p.DefaultModel()
	}

	result := &SetupResult{
		Provider:       selectedProvider,
		ProviderValues: providerValues,
		Model:          modelVar,
	}

	// Phase 3: Confirmation summary
	confirmDescription := buildConfirmSummary(result, p)

	var confirmed bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Save global configuration with these settings?").
				Description(confirmDescription).
				Affirmative("Yes, save config").
				Negative("No, cancel").
				Value(&confirmed),
		),
	).WithTheme(huh.ThemeDracula())

	if err := confirmForm.Run(); err != nil {
		return fmt.Errorf("setup cancelled: %w", ErrSetupCancelled)
	}

	if !confirmed {
		return ErrSetupCancelled
	}

	// Phase 4: Load existing global config to preserve non-provider fields, then save
	existing, err := LoadGlobalConfig()
	if err != nil {
		existing = &Config{}
	}

	existing.Provider = selectedProvider
	existing.ProviderConfig = providerValues
	existing.DefaultModel = modelVar
	if existing.Theme == "" {
		existing.Theme = "dark"
	}

	if err := SaveGlobalConfig(existing); err != nil {
		return err
	}

	fmt.Println("Global configuration saved to ~/.falcon/config.yaml")
	return nil
}

// migrateToGlobalConfig migrates provider/model/theme from .falcon/config.yaml to
// ~/.falcon/config.yaml, rewriting the project config to contain only framework + web_ui.
// No-op if ~/.falcon/config.yaml already exists.
func migrateToGlobalConfig() error {
	// No-op if global config already exists
	if _, err := os.Stat(GlobalConfigPath()); err == nil {
		return nil
	}

	// Try to read project config
	projectPath := filepath.Join(FalconFolderName, "config.yaml")
	data, err := os.ReadFile(projectPath)
	if err != nil {
		return nil // No project config either — nothing to migrate
	}

	var projectCfg Config
	if err := yaml.Unmarshal(data, &projectCfg); err != nil {
		return nil // Malformed project config — skip migration
	}

	// Only migrate if there's something meaningful
	if projectCfg.Provider == "" && projectCfg.DefaultModel == "" {
		return nil
	}

	// Write global config with provider/model/theme/web_ui fields
	globalCfg := &Config{
		Provider:         projectCfg.Provider,
		ProviderConfig:   projectCfg.ProviderConfig,
		DefaultModel:     projectCfg.DefaultModel,
		Theme:            projectCfg.Theme,
		WebUI:            projectCfg.WebUI,
		OllamaConfig:     projectCfg.OllamaConfig,
		GeminiConfig:     projectCfg.GeminiConfig,
		OpenRouterConfig: projectCfg.OpenRouterConfig,
		OllamaURL:        projectCfg.OllamaURL,
		OllamaAPIKey:     projectCfg.OllamaAPIKey,
	}
	if globalCfg.Theme == "" {
		globalCfg.Theme = "dark"
	}

	if err := EnsureGlobalFalconDir(); err != nil {
		return err
	}
	if err := SaveGlobalConfig(globalCfg); err != nil {
		return fmt.Errorf("migration: failed to write global config: %w", err)
	}

	// Rewrite project config with only framework
	projectOnly := projectConfigFile{
		Framework: projectCfg.Framework,
	}
	newData, err := yaml.Marshal(projectOnly)
	if err != nil {
		return fmt.Errorf("migration: failed to marshal project config: %w", err)
	}
	if err := os.WriteFile(projectPath, newData, 0644); err != nil {
		return fmt.Errorf("migration: failed to write project config: %w", err)
	}

	return nil
}
