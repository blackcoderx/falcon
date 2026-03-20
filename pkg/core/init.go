package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blackcoderx/falcon/pkg/core/tools/spec_ingester"
	"github.com/blackcoderx/falcon/pkg/llm"
	"github.com/charmbracelet/huh"
	"gopkg.in/yaml.v3"
)

const FalconFolderName = ".falcon"

// ToolLimitsConfig holds per-tool call limits configuration
type ToolLimitsConfig struct {
	DefaultLimit int            `yaml:"default_limit"` // Fallback limit for tools without specific limit
	TotalLimit   int            `yaml:"total_limit"`   // Safety cap on total tool calls per session
	PerTool      map[string]int `yaml:"per_tool"`      // Per-tool limits (tool_name -> max_calls)
}

// Config represents the user's Falcon configuration.
// Provider-specific settings are stored generically in ProviderConfig so that
// new providers can be added without changing this struct.
type Config struct {
	Provider       string            `yaml:"provider"`                  // "ollama", "gemini", "openrouter", …
	ProviderConfig map[string]string `yaml:"provider_config,omitempty"` // provider-specific key/value pairs
	DefaultModel   string            `yaml:"default_model"`
	Theme          string            `yaml:"theme"`
	Framework      string            `yaml:"framework"` // API framework (e.g., gin, fastapi, express)

	// Legacy fields — migrated automatically on first load; do not use in new code.
	OllamaConfig     *OllamaConfig     `yaml:"ollama,omitempty"`
	GeminiConfig     *GeminiConfig     `yaml:"gemini,omitempty"`
	OpenRouterConfig *OpenRouterConfig `yaml:"openrouter,omitempty"`
	OllamaURL        string            `yaml:"ollama_url,omitempty"`
	OllamaAPIKey     string            `yaml:"ollama_api_key,omitempty"`
}

// --- Legacy config structs (kept for migration only) ---

// OllamaConfig holds Ollama-specific configuration (legacy).
type OllamaConfig struct {
	Mode   string `yaml:"mode"`    // "local" or "cloud"
	URL    string `yaml:"url"`     // API URL
	APIKey string `yaml:"api_key"` // API key (for cloud mode)
}

// GeminiConfig holds Gemini-specific configuration (legacy).
type GeminiConfig struct {
	APIKey string `yaml:"api_key"`
}

// OpenRouterConfig holds OpenRouter-specific configuration (legacy).
type OpenRouterConfig struct {
	APIKey string `yaml:"api_key"`
}

// SupportedFrameworks lists frameworks that Falcon recognizes
var SupportedFrameworks = []string{
	"gin",     // Go - Gin
	"echo",    // Go - Echo
	"chi",     // Go - Chi
	"fiber",   // Go - Fiber
	"fastapi", // Python - FastAPI
	"flask",   // Python - Flask
	"django",  // Python - Django REST Framework
	"express", // Node.js - Express
	"nestjs",  // Node.js - NestJS
	"hono",    // Node.js/Bun - Hono
	"spring",  // Java - Spring Boot
	"laravel", // PHP - Laravel
	"rails",   // Ruby - Rails
	"actix",   // Rust - Actix Web
	"axum",    // Rust - Axum
	"other",   // Other/custom framework
}

// ProviderEntry holds the model and credentials for a single configured provider.
type ProviderEntry struct {
	Model  string            `yaml:"model"`
	Config map[string]string `yaml:"config"`
}

// GlobalConfig is the multi-provider global config written to ~/.falcon/config.yaml.
type GlobalConfig struct {
	DefaultProvider string                   `yaml:"default_provider"`
	Theme           string                   `yaml:"theme"`
	Providers       map[string]ProviderEntry `yaml:"providers,omitempty"`

	// Legacy migration fields — present only in old single-provider configs.
	// LoadGlobalConfig migrates them into Providers on first read.
	LegacyProvider       string            `yaml:"provider,omitempty"`
	LegacyProviderConfig map[string]string `yaml:"provider_config,omitempty"`
	LegacyDefaultModel   string            `yaml:"default_model,omitempty"`
}

// SetupResult holds the values collected by the first-run setup wizard.
type SetupResult struct {
	Framework      string
	Provider       string
	ProviderValues map[string]string // keyed by SetupField.Key
	Model          string
}

// frameworkGroup organizes frameworks by language for the setup wizard.
type frameworkGroup struct {
	Language   string
	Frameworks []string
}

// frameworkGroups lists frameworks grouped by language for display in the setup wizard select.
var frameworkGroups = []frameworkGroup{
	{Language: "Go", Frameworks: []string{"gin", "echo", "chi", "fiber"}},
	{Language: "Python", Frameworks: []string{"fastapi", "flask", "django"}},
	{Language: "Node.js", Frameworks: []string{"express", "nestjs", "hono"}},
	{Language: "Java", Frameworks: []string{"spring"}},
	{Language: "PHP", Frameworks: []string{"laravel"}},
	{Language: "Ruby", Frameworks: []string{"rails"}},
	{Language: "Rust", Frameworks: []string{"actix", "axum"}},
	{Language: "Other", Frameworks: []string{"other"}},
}

// buildFrameworkOptions creates huh.Option entries for all supported frameworks,
// labeled by language (e.g., "gin (Go)").
func buildFrameworkOptions() []huh.Option[string] {
	var options []huh.Option[string]
	for _, group := range frameworkGroups {
		for _, fw := range group.Frameworks {
			label := fmt.Sprintf("%s (%s)", fw, group.Language)
			if fw == "other" {
				label = "other (custom/unlisted)"
			}
			options = append(options, huh.NewOption(label, fw))
		}
	}
	return options
}

// providerOptions builds huh.Option entries from the registered provider registry.
// Adding a new provider to llm.Register() automatically surfaces it here.
func providerOptions() []huh.Option[string] {
	var opts []huh.Option[string]
	for _, p := range llm.All() {
		opts = append(opts, huh.NewOption(p.DisplayName(), p.ID()))
	}
	return opts
}

// runSetupWizard displays an interactive setup wizard on first run.
// Provider-specific forms are generated from the provider's SetupFields(),
// so this function never needs to change when a new provider is added.
// If a global config already exists with a provider, provider/model steps are skipped.
func runSetupWizard(frameworkFlag string) (*SetupResult, error) {
	var (
		selectedFramework = frameworkFlag
		selectedProvider  string
	)

	fmt.Println()
	fmt.Println("  Welcome to Falcon - AI-powered API debugging assistant")
	fmt.Println("  Let's configure your setup.")
	fmt.Println()

	// Check if global config already has provider configured
	existingGlobal, _ := LoadGlobalConfig()
	hasGlobalConfig := existingGlobal != nil && existingGlobal.DefaultProvider != ""

	// Phase 1: Framework selection (skip if --framework flag was provided)
	var configGroups []*huh.Group
	if frameworkFlag == "" {
		configGroups = append(configGroups,
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select your API framework").
					Description("Falcon uses this to provide framework-specific debugging hints.").
					Options(buildFrameworkOptions()...).
					Value(&selectedFramework).
					Height(10),
			),
		)
	}

	if hasGlobalConfig {
		// Only ask for framework — provider/model are already in global config
		if len(configGroups) > 0 {
			frameworkForm := huh.NewForm(configGroups...).WithTheme(huh.ThemeDracula())
			if err := frameworkForm.Run(); err != nil {
				return nil, fmt.Errorf("setup cancelled: %w", err)
			}
		}
		provID, model, values := GetActiveProviderEntry(existingGlobal)
		if values == nil {
			values = map[string]string{}
		}
		result := &SetupResult{
			Framework:      selectedFramework,
			Provider:       provID,
			ProviderValues: values,
			Model:          model,
		}
		fmt.Printf("  Using existing global provider: %s\n\n", result.Provider)
		return result, nil
	}

	// Phase 2: Provider selection
	configGroups = append(configGroups,
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select your LLM provider").
				Description("Choose which AI service to use for assistance.").
				Options(providerOptions()...).
				Value(&selectedProvider),
		),
	)

	providerForm := huh.NewForm(configGroups...).WithTheme(huh.ThemeDracula())
	if err := providerForm.Run(); err != nil {
		return nil, fmt.Errorf("setup cancelled: %w", err)
	}

	// Phase 3: Provider-specific fields (driven by the provider's SetupFields)
	result := &SetupResult{
		Framework:      selectedFramework,
		Provider:       selectedProvider,
		ProviderValues: map[string]string{},
	}

	p, ok := llm.Get(selectedProvider)
	if !ok {
		return nil, fmt.Errorf("unknown provider %q", selectedProvider)
	}

	fields := p.SetupFields()
	if len(fields) > 0 {
		modelVar, err := buildProviderForm(p, result.ProviderValues)
		if err != nil {
			return nil, err
		}
		result.Model = modelVar
	}

	// Apply per-field defaults
	for _, f := range fields {
		if result.ProviderValues[f.Key] == "" && f.Default != "" {
			result.ProviderValues[f.Key] = f.Default
		}
	}
	if result.Model == "" {
		result.Model = p.DefaultModel()
	}

	// Phase 4: Confirmation
	confirmDescription := buildConfirmSummary(result, p)

	var confirmed bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Create configuration with these settings?").
				Description(confirmDescription).
				Affirmative("Yes, create config").
				Negative("No, cancel").
				Value(&confirmed),
		),
	).WithTheme(huh.ThemeDracula())

	if err := confirmForm.Run(); err != nil {
		return nil, fmt.Errorf("confirmation cancelled: %w", err)
	}

	if !confirmed {
		return nil, fmt.Errorf("setup cancelled by user")
	}

	return result, nil
}

// buildProviderForm creates and runs the huh form for a provider's setup fields,
// writing collected values into the provided map. Also collects model name.
// Returns the model name the user entered (may be empty — caller applies default).
func buildProviderForm(p llm.Provider, values map[string]string) (string, error) {
	fields := p.SetupFields()
	// Allocate string variables for each field so huh can bind to them.
	vars := make([]string, len(fields))
	var modelVar string

	var huhFields []huh.Field
	for i, f := range fields {
		fi := i
		switch f.Type {
		case llm.FieldSelect:
			opts := make([]huh.Option[string], len(f.Options))
			for j, o := range f.Options {
				opts[j] = huh.NewOption(o.Label, o.Value)
			}
			huhFields = append(huhFields,
				huh.NewSelect[string]().
					Title(f.Title).
					Description(f.Description).
					Options(opts...).
					Value(&vars[fi]),
			)
		default:
			inp := huh.NewInput().
				Title(f.Title).
				Description(f.Description).
				Placeholder(f.Placeholder).
				Value(&vars[fi])
			if f.Secret {
				inp = inp.EchoMode(huh.EchoModePassword)
			}
			huhFields = append(huhFields, inp)
		}
	}

	// Always append a model name field at the end
	huhFields = append(huhFields,
		huh.NewInput().
			Title("Model name").
			Description(fmt.Sprintf("The model to use. Leave blank to use the default: %s", p.DefaultModel())).
			Placeholder(p.DefaultModel()).
			Value(&modelVar),
	)

	form := huh.NewForm(huh.NewGroup(huhFields...)).WithTheme(huh.ThemeDracula())
	if err := form.Run(); err != nil {
		return "", fmt.Errorf("setup cancelled: %w", err)
	}

	// Write collected values back into the map
	for i, f := range fields {
		values[f.Key] = vars[i]
	}

	return modelVar, nil
}

// buildConfirmSummary builds the confirmation text shown before writing config.
func buildConfirmSummary(result *SetupResult, p llm.Provider) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Provider:  %s\n", p.DisplayName())
	if result.Framework != "" {
		fmt.Fprintf(&sb, "Framework: %s\n", result.Framework)
	}
	fmt.Fprintf(&sb, "Model:     %s\n", result.Model)
	for _, f := range p.SetupFields() {
		v := result.ProviderValues[f.Key]
		if f.Secret {
			v = maskAPIKey(v)
		}
		if v != "" {
			label := strings.ReplaceAll(f.Key, "_", " ")
			label = strings.ToUpper(label[:1]) + label[1:]
			fmt.Fprintf(&sb, "%-10s %s\n", label+":", v)
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

// maskAPIKey returns a masked version of the API key for display.
func maskAPIKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// InitializeFalconFolder creates the .falcon directory and initializes default files if they don't exist.
// If framework is empty and this is a first-time setup, prompts the user to select one.
// skipIndex determines if we should auto-run the spec ingester.
func InitializeFalconFolder(framework string, skipIndex bool) error {
	// Check if .falcon exists
	if _, err := os.Stat(FalconFolderName); os.IsNotExist(err) {
		// Run interactive setup wizard on first run
		setup, err := runSetupWizard(framework)
		if err != nil {
			return fmt.Errorf("setup failed: %w", err)
		}

		// Create .falcon directory
		if err := os.Mkdir(FalconFolderName, 0755); err != nil {
			return fmt.Errorf("failed to create .falcon folder: %w", err)
		}

		// Write global config (provider, model, theme) to ~/.falcon/config.yaml
		if err := EnsureGlobalFalconDir(); err != nil {
			return err
		}
		if err := writeGlobalConfig(setup); err != nil {
			return err
		}

		// Write project config (framework only) to .falcon/config.yaml
		if err := writeProjectConfig(setup.Framework); err != nil {
			return err
		}

		// Create falcon.md knowledge base template
		if err := createFalconKnowledgeBase(); err != nil {
			return err
		}

		// Create requests directory for saved requests
		if err := os.Mkdir(filepath.Join(FalconFolderName, "requests"), 0755); err != nil {
			return fmt.Errorf("failed to create requests folder: %w", err)
		}

		// Create environments directory for environment files
		if err := os.Mkdir(filepath.Join(FalconFolderName, "environments"), 0755); err != nil {
			return fmt.Errorf("failed to create environments folder: %w", err)
		}

		// Create folder structure
		newFolders := []string{"baselines", "flows", "reports"}
		for _, folder := range newFolders {
			if err := os.Mkdir(filepath.Join(FalconFolderName, folder), 0755); err != nil {
				return fmt.Errorf("failed to create %s folder: %w", folder, err)
			}
		}

		// Create empty spec.yaml (populated by ingest_spec)
		specPath := filepath.Join(FalconFolderName, "spec.yaml")
		if err := os.WriteFile(specPath, []byte("# Falcon API Specification\n# Populated automatically by ingest_spec.\n"), 0644); err != nil {
			return fmt.Errorf("failed to create spec.yaml: %w", err)
		}

		// Create baselines README
		if err := createBaselinesReadme(); err != nil {
			return err
		}

		// Create default dev environment
		if err := createDefaultEnvironment(); err != nil {
			return err
		}

		fmt.Printf("\nInitialized .falcon folder with framework: %s\n", setup.Framework)

		// Auto-Index if not skipped
		if !skipIndex {
			fmt.Println("\nScanning for API specifications...")
			// TODO: Find specs. For now we look for common files
			specFiles := []string{"openapi.yaml", "openapi.json", "swagger.yaml", "swagger.json", "postman_collection.json"}
			var foundSpec string
			for _, f := range specFiles {
				if _, err := os.Stat(f); err == nil {
					foundSpec = f
					break
				}
			}

			if foundSpec != "" {
				fmt.Printf("Found spec file: %s. Indexing...\n", foundSpec)
				tool := spec_ingester.NewIngestSpecTool(nil, FalconFolderName)

				params := fmt.Sprintf(`{"action":"index", "source":"%s"}`, foundSpec)
				if out, err := tool.Execute(params); err != nil {
					fmt.Printf("Warning: Failed to index spec: %v\n", err)
				} else {
					fmt.Println(out)
				}
			} else {
				fmt.Println("No common API spec files found. Skipping auto-index.")
			}
		}

	} else if framework != "" {
		// Update framework in existing config if provided via flag
		if err := updateConfigFramework(framework); err != nil {
			return fmt.Errorf("failed to update framework: %w", err)
		}
		fmt.Printf("Updated framework to: %s\n", framework)
	}

	// Ensure subdirectories exist (for upgrades from older versions)
	for _, dir := range []string{"requests", "environments", "baselines", "flows", "reports"} {
		if err := ensureDir(filepath.Join(FalconFolderName, dir)); err != nil {
			return err
		}
	}

	// Migrate provider/model config from project to global location (no-op if already done)
	if err := migrateToGlobalConfig(); err != nil {
		return fmt.Errorf("failed to migrate to global config: %w", err)
	}

	// Migrate legacy config fields if present
	if err := migrateLegacyConfig(); err != nil {
		return fmt.Errorf("failed to migrate config: %w", err)
	}

	return nil
}

// migrateLegacyConfig converts old per-provider typed config sub-structs to the
// new generic provider_config map. Safe to run on every startup — it no-ops when
// there is nothing to migrate.
func migrateLegacyConfig() error {
	configPath := filepath.Join(FalconFolderName, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil // Malformed config, skip migration
	}

	changed := false

	// Skip migration if provider_config already populated
	if config.ProviderConfig != nil {
		// Still clear any leftover legacy sub-structs
		if config.OllamaConfig != nil || config.GeminiConfig != nil || config.OpenRouterConfig != nil ||
			config.OllamaURL != "" || config.OllamaAPIKey != "" {
			config.OllamaConfig = nil
			config.GeminiConfig = nil
			config.OpenRouterConfig = nil
			config.OllamaURL = ""
			config.OllamaAPIKey = ""
			changed = true
		}
		if !changed {
			return nil
		}
	} else {
		// Migrate flat legacy fields → OllamaConfig first
		if (config.OllamaURL != "" || config.OllamaAPIKey != "") && config.OllamaConfig == nil {
			config.OllamaConfig = &OllamaConfig{
				Mode:   "local",
				URL:    config.OllamaURL,
				APIKey: config.OllamaAPIKey,
			}
			if config.OllamaAPIKey != "" {
				config.OllamaConfig.Mode = "cloud"
				if config.OllamaConfig.URL == "" {
					config.OllamaConfig.URL = "https://ollama.com"
				}
			} else if config.OllamaConfig.URL == "" {
				config.OllamaConfig.URL = "http://localhost:11434"
			}
			config.OllamaURL = ""
			config.OllamaAPIKey = ""
			changed = true
		}

		// Migrate typed sub-structs → provider_config map
		switch config.Provider {
		case "ollama":
			if config.OllamaConfig != nil {
				config.ProviderConfig = map[string]string{
					"mode":    config.OllamaConfig.Mode,
					"url":     config.OllamaConfig.URL,
					"api_key": config.OllamaConfig.APIKey,
				}
				config.OllamaConfig = nil
				changed = true
			}
		case "gemini":
			if config.GeminiConfig != nil {
				config.ProviderConfig = map[string]string{
					"api_key": config.GeminiConfig.APIKey,
				}
				config.GeminiConfig = nil
				changed = true
			}
		case "openrouter":
			if config.OpenRouterConfig != nil {
				config.ProviderConfig = map[string]string{
					"api_key": config.OpenRouterConfig.APIKey,
				}
				config.OpenRouterConfig = nil
				changed = true
			}
		}
	}

	if !changed {
		return nil
	}

	newData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, newData, 0644)
}

// updateConfigFramework updates the framework in an existing config file
func updateConfigFramework(framework string) error {
	configPath := filepath.Join(FalconFolderName, "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	config.Framework = framework

	newData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, newData, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetConfigFramework reads the framework from the config file
func GetConfigFramework() string {
	configPath := filepath.Join(FalconFolderName, "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return ""
	}

	return config.Framework
}

// ensureDir creates a directory if it doesn't exist.
func ensureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	return nil
}

// createDefaultEnvironment creates default dev, staging, and prod environment files
func createDefaultEnvironment() error {
	envDir := filepath.Join(FalconFolderName, "environments")

	environments := []struct {
		filename string
		content  string
	}{
		{
			"dev.yaml",
			`# Development environment
BASE_URL: http://localhost:8000
API_KEY: your-dev-api-key
`,
		},
		{
			"staging.yaml",
			`# Staging environment
BASE_URL: https://staging.example.com
API_KEY: your-staging-api-key
`,
		},
		{
			"prod.yaml",
			`# Production environment
BASE_URL: https://api.example.com
API_KEY: your-prod-api-key
`,
		},
	}

	for _, env := range environments {
		envPath := filepath.Join(envDir, env.filename)
		if err := os.WriteFile(envPath, []byte(env.content), 0644); err != nil {
			return fmt.Errorf("failed to write %s environment: %w", env.filename, err)
		}
	}
	return nil
}

// DefaultToolLimits defines the default per-tool call limits.
var DefaultToolLimits = map[string]int{
	// Core HTTP
	"http_request":     25,
	"webhook_listener": 10,
	// Unified auth (replaces auth_bearer, auth_basic, auth_oauth2, auth_helper)
	"auth": 50,
	// Utilities
	"wait":  20,
	"retry": 15,
	// Assertions & extraction
	"assert_response":      100,
	"extract_value":        100,
	"validate_json_schema": 50,
	"compare_responses":    30,
	// Unified persistence (replaces save_request, load_request, list_requests)
	"request": 50,
	// Unified environment (replaces set_environment, list_environments)
	"environment": 30,
	// Variables
	"variable": 100,
	// .falcon-scoped tools
	"falcon_write": 30,
	"falcon_read":  50,
	"session_log":  20,
	// Memory
	"memory": 50,
	// Test suites & orchestration
	"test_suite":              10,
	"run_tests":               10,
	"auto_test":               5,
	"orchestrate_integration": 5,
	// Spec & test generation
	"ingest_spec":               5,
	"generate_functional_tests": 5,
	// Specialized testing modules
	"run_smoke":          15,
	"run_data_driven":    10,
	"verify_idempotency": 10,
	"check_regression":   10,
	"run_performance":    5,
	"scan_security":      3,
	// Debugging (file system reads)
	"read_file":        50,
	"list_files":       50,
	"search_code":      30,
	"find_handler":     20,
	"analyze_endpoint": 15,
	"analyze_failure":  15,
	"propose_fix":      10,
	"create_test_file": 10,
	"write_file":       10,
}

// writeGlobalConfig writes provider/model/theme from wizard results to ~/.falcon/config.yaml.
// Upserts the provider entry without wiping other configured providers.
func writeGlobalConfig(setup *SetupResult) error {
	existing, err := LoadGlobalConfig()
	if err != nil {
		existing = &GlobalConfig{}
	}

	SetProviderEntry(existing, setup.Provider, setup.Model, setup.ProviderValues)
	existing.DefaultProvider = setup.Provider
	if existing.Theme == "" {
		existing.Theme = "dark"
	}

	return SaveGlobalConfig(existing)
}

// projectConfigFile is the minimal on-disk representation of .falcon/config.yaml.
// It intentionally contains only framework so the file stays clean.
type projectConfigFile struct {
	Framework string `yaml:"framework"`
}

// writeProjectConfig writes only the framework to .falcon/config.yaml.
func writeProjectConfig(framework string) error {
	config := projectConfigFile{
		Framework: framework,
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %w", err)
	}

	configPath := filepath.Join(FalconFolderName, "config.yaml")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write project config file: %w", err)
	}

	return nil
}

// createFalconKnowledgeBase writes the initial structured template for falcon.md.
func createFalconKnowledgeBase() error {
	content := `# Falcon API Knowledge Base

This file is automatically maintained by Falcon. It records durable facts about
the APIs, codebases, and projects you work with. Edit it freely — Falcon reads
it at the start of every session and updates it as it learns new information.

## Base URLs

<!-- Falcon will fill this in as it discovers your API -->

## Authentication

<!-- Auth method, token endpoints, and patterns Falcon discovers -->

## Known Endpoints

| Method | Path | Description | Notes |
|--------|------|-------------|-------|

## Data Models

<!-- JSON shapes of request/response bodies Falcon has seen -->

## Known Errors

<!-- Error codes and their causes discovered during testing -->

## Project Notes

<!-- Framework, architecture notes, and other project-specific facts -->
`
	path := filepath.Join(FalconFolderName, "falcon.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write falcon.md: %w", err)
	}
	return nil
}

// createBaselinesReadme writes an explanatory README.md into the baselines folder.
func createBaselinesReadme() error {
	content := `# Baselines

This folder stores reference snapshots for API regression testing.

## What Are Baselines?

Baselines capture the expected state of an API endpoint — its status code,
headers, and response body shape — at a known-good point in time. Falcon
compares future responses against these snapshots to detect regressions.

## How to Use

Ask Falcon to create a baseline:
  "Save a baseline for GET /users"

Ask Falcon to check for regressions:
  "Check GET /users against the baseline"

Falcon uses the breaking_change_detector and regression_watchdog tools
to compare live responses against baselines stored here.

## Example Baseline File

` + "```yaml" + `
endpoint: GET /users
captured_at: 2026-01-01T00:00:00Z
status_code: 200
headers:
  content-type: application/json
body_schema:
  type: array
  items:
    type: object
    required: [id, name, email]
    properties:
      id:
        type: integer
      name:
        type: string
      email:
        type: string
` + "```" + `

Baselines are created and updated automatically by Falcon.
`
	path := filepath.Join(FalconFolderName, "baselines", "README.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write baselines README: %w", err)
	}
	return nil
}
