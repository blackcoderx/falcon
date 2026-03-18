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

// LoadGlobalConfig reads the global config. If the file doesn't exist, returns a
// zero-value GlobalConfig with no error. Automatically migrates old single-provider
// format (top-level provider/provider_config/default_model keys) into the new
// Providers map on first read.
func LoadGlobalConfig() (*GlobalConfig, error) {
	data, err := os.ReadFile(GlobalConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &GlobalConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read global config: %w", err)
	}

	var cfg GlobalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse global config: %w", err)
	}

	// Migrate old single-provider format in memory.
	if cfg.LegacyProvider != "" && cfg.Providers == nil {
		if cfg.Providers == nil {
			cfg.Providers = map[string]ProviderEntry{}
		}
		cfg.Providers[cfg.LegacyProvider] = ProviderEntry{
			Model:  cfg.LegacyDefaultModel,
			Config: cfg.LegacyProviderConfig,
		}
		cfg.DefaultProvider = cfg.LegacyProvider
		cfg.LegacyProvider = ""
		cfg.LegacyProviderConfig = nil
		cfg.LegacyDefaultModel = ""
	}

	return &cfg, nil
}

// SaveGlobalConfig writes the GlobalConfig to ~/.falcon/config.yaml (0600 perms).
func SaveGlobalConfig(cfg *GlobalConfig) error {
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

// GetActiveProviderEntry returns the ID, model, and config values for the
// currently active (default) provider. Returns empty strings/nil if nothing configured.
func GetActiveProviderEntry(cfg *GlobalConfig) (providerID, model string, values map[string]string) {
	if cfg == nil || cfg.DefaultProvider == "" || cfg.Providers == nil {
		return "", "", nil
	}
	entry, ok := cfg.Providers[cfg.DefaultProvider]
	if !ok {
		return cfg.DefaultProvider, "", nil
	}
	return cfg.DefaultProvider, entry.Model, entry.Config
}

// SetProviderEntry upserts a provider entry into the config without touching
// any other provider entries.
func SetProviderEntry(cfg *GlobalConfig, providerID, model string, values map[string]string) {
	if cfg.Providers == nil {
		cfg.Providers = map[string]ProviderEntry{}
	}
	cfg.Providers[providerID] = ProviderEntry{Model: model, Config: values}
}

// RunGlobalConfigWizard runs the interactive provider/model setup wizard and saves to global config.
// Presents a sub-action menu: add/update a provider, set default provider, or remove a provider.
func RunGlobalConfigWizard() error {
	fmt.Println()
	fmt.Println("  Falcon Global Configuration")
	fmt.Println()

	existing, err := LoadGlobalConfig()
	if err != nil {
		existing = &GlobalConfig{}
	}

	// Sub-action menu
	var action string
	actionForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What would you like to do?").
				Options(
					huh.NewOption("Add or update a provider", "upsert"),
					huh.NewOption("Set default provider", "set_default"),
					huh.NewOption("Remove a provider", "remove"),
				).
				Value(&action),
		),
	).WithTheme(huh.ThemeDracula())

	if err := actionForm.Run(); err != nil {
		return ErrSetupCancelled
	}

	switch action {
	case "upsert":
		return runUpsertProvider(existing)
	case "set_default":
		return runSetDefaultProvider(existing)
	case "remove":
		return runRemoveProvider(existing)
	}
	return nil
}

// runUpsertProvider adds or updates a single provider config without touching others.
func runUpsertProvider(existing *GlobalConfig) error {
	var selectedProvider string

	providerForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select provider to configure").
				Options(providerOptions()...).
				Value(&selectedProvider),
		),
	).WithTheme(huh.ThemeDracula())

	if err := providerForm.Run(); err != nil {
		return ErrSetupCancelled
	}

	p, ok := llm.Get(selectedProvider)
	if !ok {
		return fmt.Errorf("unknown provider %q", selectedProvider)
	}

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

	confirmDescription := buildConfirmSummary(result, p)

	// Ask if this should become the default provider
	isFirst := existing.DefaultProvider == "" || len(existing.Providers) == 0
	var setAsDefault bool
	if isFirst {
		setAsDefault = true
	} else {
		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Set as default provider?").
					Description(confirmDescription).
					Affirmative("Yes, set as default").
					Negative("No, keep current default").
					Value(&setAsDefault),
			),
		).WithTheme(huh.ThemeDracula())
		if err := confirmForm.Run(); err != nil {
			return ErrSetupCancelled
		}
	}

	SetProviderEntry(existing, selectedProvider, modelVar, providerValues)
	if setAsDefault {
		existing.DefaultProvider = selectedProvider
	}
	if existing.Theme == "" {
		existing.Theme = "dark"
	}

	if err := SaveGlobalConfig(existing); err != nil {
		return err
	}

	fmt.Printf("Provider %q saved to ~/.falcon/config.yaml\n", p.DisplayName())
	if setAsDefault {
		fmt.Printf("Default provider set to %q\n", p.DisplayName())
	}
	return nil
}

// runSetDefaultProvider lets the user pick which configured provider to use as default.
func runSetDefaultProvider(existing *GlobalConfig) error {
	if len(existing.Providers) == 0 {
		fmt.Println("No providers configured yet. Run 'falcon config' and add a provider first.")
		return nil
	}

	var opts []huh.Option[string]
	for _, p := range llm.All() {
		if _, ok := existing.Providers[p.ID()]; ok {
			opts = append(opts, huh.NewOption(p.DisplayName(), p.ID()))
		}
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select the default provider").
				Options(opts...).
				Value(&selected),
		),
	).WithTheme(huh.ThemeDracula())

	if err := form.Run(); err != nil {
		return ErrSetupCancelled
	}

	existing.DefaultProvider = selected
	if err := SaveGlobalConfig(existing); err != nil {
		return err
	}

	p, _ := llm.Get(selected)
	fmt.Printf("Default provider set to %q\n", p.DisplayName())
	return nil
}

// runRemoveProvider lets the user remove a configured provider.
func runRemoveProvider(existing *GlobalConfig) error {
	if len(existing.Providers) == 0 {
		fmt.Println("No providers configured.")
		return nil
	}

	var opts []huh.Option[string]
	for _, p := range llm.All() {
		if _, ok := existing.Providers[p.ID()]; ok {
			opts = append(opts, huh.NewOption(p.DisplayName(), p.ID()))
		}
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select provider to remove").
				Options(opts...).
				Value(&selected),
		),
	).WithTheme(huh.ThemeDracula())

	if err := form.Run(); err != nil {
		return ErrSetupCancelled
	}

	delete(existing.Providers, selected)

	// If the removed provider was the default, clear or reassign default
	if existing.DefaultProvider == selected {
		existing.DefaultProvider = ""
		for _, p := range llm.All() {
			if _, ok := existing.Providers[p.ID()]; ok {
				existing.DefaultProvider = p.ID()
				break
			}
		}
	}

	if err := SaveGlobalConfig(existing); err != nil {
		return err
	}

	p, _ := llm.Get(selected)
	fmt.Printf("Provider %q removed.\n", p.DisplayName())
	if existing.DefaultProvider != "" {
		dp, _ := llm.Get(existing.DefaultProvider)
		fmt.Printf("Default provider is now %q\n", dp.DisplayName())
	}
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

	globalCfg := &GlobalConfig{
		DefaultProvider: projectCfg.Provider,
		Theme:           projectCfg.Theme,
		WebUI:           projectCfg.WebUI,
		Providers: map[string]ProviderEntry{
			projectCfg.Provider: {
				Model:  projectCfg.DefaultModel,
				Config: projectCfg.ProviderConfig,
			},
		},
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
