package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
	"github.com/blackcoderx/falcon/pkg/core/tools/spec_ingester"
	"github.com/charmbracelet/huh"
	"gopkg.in/yaml.v3"
)

const ZapFolderName = ".zap"

// ToolLimitsConfig holds per-tool call limits configuration
type ToolLimitsConfig struct {
	DefaultLimit int            `yaml:"default_limit"` // Fallback limit for tools without specific limit
	TotalLimit   int            `yaml:"total_limit"`   // Safety cap on total tool calls per session
	PerTool      map[string]int `yaml:"per_tool"`      // Per-tool limits (tool_name -> max_calls)
}

// OllamaConfig holds Ollama-specific configuration
type OllamaConfig struct {
	Mode   string `yaml:"mode"`    // "local" or "cloud"
	URL    string `yaml:"url"`     // API URL
	APIKey string `yaml:"api_key"` // API key (for cloud mode)
}

// GeminiConfig holds Gemini-specific configuration
type GeminiConfig struct {
	APIKey string `yaml:"api_key"` // Gemini API key
}

// WebUIConfig controls the embedded web dashboard
type WebUIConfig struct {
	Enabled bool `yaml:"enabled"` // default true
	Port    int  `yaml:"port"`    // 0 = OS-assigned random port
}

// Config represents the user's Falcon configuration
type Config struct {
	Provider     string           `yaml:"provider"` // "ollama" or "gemini"
	OllamaConfig *OllamaConfig    `yaml:"ollama,omitempty"`
	GeminiConfig *GeminiConfig    `yaml:"gemini,omitempty"`
	DefaultModel string           `yaml:"default_model"`
	Theme        string           `yaml:"theme"`
	Framework    string           `yaml:"framework"` // API framework (e.g., gin, fastapi, express)
	ToolLimits   ToolLimitsConfig `yaml:"tool_limits"`
	WebUI        WebUIConfig      `yaml:"web_ui"`

	// Legacy fields for backward compatibility (deprecated)
	OllamaURL    string `yaml:"ollama_url,omitempty"`
	OllamaAPIKey string `yaml:"ollama_api_key,omitempty"`
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

// SetupResult holds the collected values from the first-run setup wizard.
type SetupResult struct {
	Framework  string
	Provider   string // "ollama" or "gemini"
	OllamaMode string // "local" or "cloud" (for Ollama only)
	OllamaURL  string // Ollama API URL
	GeminiKey  string // Gemini API key
	OllamaKey  string // Ollama API key (for cloud mode)
	Model      string
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

// providerOptions returns the available LLM provider options for the setup wizard.
func providerOptions() []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("Ollama (local or cloud)", "ollama"),
		huh.NewOption("Gemini (Google AI)", "gemini"),
	}
}

// ollamaModeOptions returns the Ollama mode options (local vs cloud).
func ollamaModeOptions() []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("Local (run on your machine)", "local"),
		huh.NewOption("Cloud (Ollama Cloud)", "cloud"),
	}
}

// runSetupWizard displays an interactive setup wizard on first run using the huh library.
// If frameworkFlag is non-empty, the framework selection step is skipped.
func runSetupWizard(frameworkFlag string) (*SetupResult, error) {
	// Use separate local variables for huh bindings to avoid
	// any pre-initialized value interference with input fields
	var (
		selectedFramework = frameworkFlag
		selectedProvider  string
		ollamaMode        string
		ollamaURL         string
		ollamaKey         string
		geminiKey         string
		modelName         string
	)

	fmt.Println()
	fmt.Println("  Welcome to Falcon - AI-powered API debugging assistant")
	fmt.Println("  Let's configure your setup.")
	fmt.Println()

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

	// Phase 3: Provider-specific configuration
	result := &SetupResult{
		Framework: selectedFramework,
		Provider:  selectedProvider,
	}

	if selectedProvider == "ollama" {
		// Ollama mode selection
		modeForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select Ollama mode").
					Description("Local runs on your machine, Cloud uses Ollama's hosted service.").
					Options(ollamaModeOptions()...).
					Value(&ollamaMode),
			),
		).WithTheme(huh.ThemeDracula())

		if err := modeForm.Run(); err != nil {
			return nil, fmt.Errorf("setup cancelled: %w", err)
		}

		result.OllamaMode = ollamaMode

		if ollamaMode == "local" {
			// Local Ollama configuration
			localForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Ollama URL").
						Description("Local Ollama server URL (default: http://localhost:11434).").
						Placeholder("http://localhost:11434").
						Value(&ollamaURL),
					huh.NewInput().
						Title("Model name").
						Description("The model to use (must be installed locally).").
						Placeholder("llama3").
						Value(&modelName),
				),
			).WithTheme(huh.ThemeDracula())

			if err := localForm.Run(); err != nil {
				return nil, fmt.Errorf("setup cancelled: %w", err)
			}

			// Set defaults for local mode
			if ollamaURL == "" {
				ollamaURL = "http://localhost:11434"
			}
			if modelName == "" {
				modelName = "llama3"
			}

			result.OllamaURL = ollamaURL
			result.Model = modelName

		} else {
			// Cloud Ollama configuration
			cloudForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Ollama Cloud URL").
						Description("Ollama Cloud API endpoint (default: https://ollama.com).").
						Placeholder("https://ollama.com").
						Value(&ollamaURL),
					huh.NewInput().
						Title("Model name").
						Description("The cloud model to use.").
						Placeholder("qwen3-coder:480b-cloud").
						Value(&modelName),
					huh.NewInput().
						Title("API Key").
						Description("Your Ollama Cloud API key.").
						Placeholder("Enter your API key...").
						EchoMode(huh.EchoModePassword).
						Value(&ollamaKey),
				),
			).WithTheme(huh.ThemeDracula())

			if err := cloudForm.Run(); err != nil {
				return nil, fmt.Errorf("setup cancelled: %w", err)
			}

			// Set defaults for cloud mode
			if ollamaURL == "" {
				ollamaURL = "https://ollama.com"
			}
			if modelName == "" {
				modelName = "qwen3-coder:480b-cloud"
			}

			result.OllamaURL = ollamaURL
			result.OllamaKey = ollamaKey
			result.Model = modelName
		}

	} else {
		// Gemini configuration
		geminiForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Gemini API Key").
					Description("Get your API key from aistudio.google.com.").
					Placeholder("Enter your Gemini API key...").
					EchoMode(huh.EchoModePassword).
					Value(&geminiKey),
				huh.NewInput().
					Title("Model name").
					Description("The Gemini model to use (default: gemini-2.5-flash-lite).").
					Placeholder("gemini-2.5-flash-lite").
					Value(&modelName),
			),
		).WithTheme(huh.ThemeDracula())

		if err := geminiForm.Run(); err != nil {
			return nil, fmt.Errorf("setup cancelled: %w", err)
		}

		// Set defaults for Gemini
		if modelName == "" {
			modelName = "gemini-2.5-flash-lite"
		}

		result.GeminiKey = geminiKey
		result.Model = modelName
	}

	// Phase 4: Confirmation with actual entered values
	var confirmDescription string
	if result.Provider == "ollama" {
		if result.OllamaMode == "local" {
			confirmDescription = fmt.Sprintf(
				"Provider:  Ollama (local)\nFramework: %s\nURL:       %s\nModel:     %s",
				result.Framework,
				result.OllamaURL,
				result.Model,
			)
		} else {
			confirmDescription = fmt.Sprintf(
				"Provider:  Ollama (cloud)\nFramework: %s\nURL:       %s\nModel:     %s\nAPI Key:   %s",
				result.Framework,
				result.OllamaURL,
				result.Model,
				maskAPIKey(result.OllamaKey),
			)
		}
	} else {
		confirmDescription = fmt.Sprintf(
			"Provider:  Gemini\nFramework: %s\nModel:     %s\nAPI Key:   %s",
			result.Framework,
			result.Model,
			maskAPIKey(result.GeminiKey),
		)
	}

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

// InitializeZapFolder creates the .zap directory and initializes default files if they don't exist.
// If framework is empty and this is a first-time setup, prompts the user to select one.
// skipIndex determines if we should auto-run the spec ingester.
func InitializeZapFolder(framework string, skipIndex bool) error {
	// Check if .zap exists
	if _, err := os.Stat(ZapFolderName); os.IsNotExist(err) {
		// Run interactive setup wizard on first run
		setup, err := runSetupWizard(framework)
		if err != nil {
			return fmt.Errorf("setup failed: %w", err)
		}

		// Create .zap directory
		if err := os.Mkdir(ZapFolderName, 0755); err != nil {
			return fmt.Errorf("failed to create .zap folder: %w", err)
		}

		// Create config.yaml with wizard results
		if err := createDefaultConfig(setup); err != nil {
			return err
		}

		// Create falcon.md knowledge base template
		if err := createFalconKnowledgeBase(); err != nil {
			return err
		}

		// Create empty memory.json
		if err := createMemoryFile(); err != nil {
			return err
		}

		// Create requests directory for saved requests
		if err := os.Mkdir(filepath.Join(ZapFolderName, "requests"), 0755); err != nil {
			return fmt.Errorf("failed to create requests folder: %w", err)
		}

		// Create environments directory for environment files
		if err := os.Mkdir(filepath.Join(ZapFolderName, "environments"), 0755); err != nil {
			return fmt.Errorf("failed to create environments folder: %w", err)
		}

		// Create new folder structure
		newFolders := []string{"baselines", "flows"}
		for _, folder := range newFolders {
			if err := os.Mkdir(filepath.Join(ZapFolderName, folder), 0755); err != nil {
				return fmt.Errorf("failed to create %s folder: %w", folder, err)
			}
		}

		// Create baselines README
		if err := createBaselinesReadme(); err != nil {
			return err
		}

		// Create default dev environment
		if err := createDefaultEnvironment(); err != nil {
			return err
		}

		// Create manifest.json
		if err := shared.CreateManifest(ZapFolderName); err != nil {
			return err
		}

		fmt.Printf("\nInitialized .zap folder with framework: %s\n", setup.Framework)

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
				// We create a temporary tool instance to run the logic
				// Note: LLM client is nil here as initial indexing (parsing) might not need it yet
				// If parsing needs LLM, we'd need to init it. But our current implementation describes
				// LLM for fusion. Let's stick to parsing for now.
				tool := spec_ingester.NewIngestSpecTool(nil, ZapFolderName)

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
	if err := ensureDir(filepath.Join(ZapFolderName, "requests")); err != nil {
		return err
	}
	if err := ensureDir(filepath.Join(ZapFolderName, "environments")); err != nil {
		return err
	}
	if err := ensureDir(filepath.Join(ZapFolderName, "baselines")); err != nil {
		return err
	}
	if err := ensureDir(filepath.Join(ZapFolderName, "flows")); err != nil {
		return err
	}

	// Ensure manifest exists (for upgrades)
	if _, err := os.Stat(filepath.Join(ZapFolderName, shared.ManifestFilename)); os.IsNotExist(err) {
		shared.CreateManifest(ZapFolderName)
	}

	// Migrate legacy config fields if present
	if err := migrateLegacyConfig(); err != nil {
		return fmt.Errorf("failed to migrate config: %w", err)
	}

	return nil
}

// migrateLegacyConfig moves legacy top-level Ollama fields to the new OllamaConfig struct.
func migrateLegacyConfig() error {
	configPath := filepath.Join(ZapFolderName, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No config to migrate
		}
		return err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err // Malformed config, skip migration
	}

	changed := false

	// Check for legacy fields
	if (config.OllamaURL != "" || config.OllamaAPIKey != "") && config.OllamaConfig == nil {
		config.OllamaConfig = &OllamaConfig{
			Mode:   "local",
			URL:    config.OllamaURL,
			APIKey: config.OllamaAPIKey,
		}
		// If API key is present, assume cloud or authenticated instance
		if config.OllamaAPIKey != "" {
			if config.OllamaURL == "https://ollama.com" || config.OllamaURL == "" {
				config.OllamaConfig.Mode = "cloud"
				if config.OllamaConfig.URL == "" {
					config.OllamaConfig.URL = "https://ollama.com"
				}
			}
		} else if config.OllamaURL == "" {
			config.OllamaConfig.URL = "http://localhost:11434"
		}

		// Clear legacy fields
		config.OllamaURL = ""
		config.OllamaAPIKey = ""
		changed = true
	}

	if changed {
		newData, err := yaml.Marshal(config)
		if err != nil {
			return err
		}
		return os.WriteFile(configPath, newData, 0644)
	}

	return nil
}

// updateConfigFramework updates the framework in an existing config file
func updateConfigFramework(framework string) error {
	configPath := filepath.Join(ZapFolderName, "config.yaml")

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
	configPath := filepath.Join(ZapFolderName, "config.yaml")

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
// Returns an error if directory creation fails.
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
	envDir := filepath.Join(ZapFolderName, "environments")

	environments := []struct {
		filename string
		content  string
	}{
		{
			"dev.yaml",
			`# Development environment
BASE_URL: http://localhost:3000
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
	// High-risk tools (external I/O)
	"http_request":     25,
	"performance_test": 5,
	"webhook_listener": 10,
	"auth_oauth2":      10,
	// Medium-risk tools (file system)
	"read_file":    50,
	"list_files":   50,
	"search_code":  30,
	"save_request": 20,
	"load_request": 30,
	// Low-risk tools (in-memory)
	"variable":             100,
	"assert_response":      100,
	"extract_value":        100,
	"auth_bearer":          50,
	"auth_basic":           50,
	"auth_helper":          50,
	"validate_json_schema": 50,
	"compare_responses":    30,
	// Special tools
	"retry":      15,
	"wait":       20,
	"test_suite": 10,
	// AI Analysis & Orchestration
	"analyze_endpoint": 15,
	"analyze_failure":  15,
	"generate_tests":   10,
	"run_tests":        10,
	"run_single_test":  20,
	"auto_test":        5,
	// Sprint 3: Codebase Intelligence & Fixing
	"find_handler":     20,
	"propose_fix":      10,
	"create_test_file": 10,
	// Sprint 4: Reporting
	"security_report": 20,
	"export_results":  20,
	// Memory tool
	"memory": 50,
}

// createDefaultConfig creates a default configuration file with the setup wizard results.
func createDefaultConfig(setup *SetupResult) error {
	config := Config{
		Provider:     setup.Provider,
		DefaultModel: setup.Model,
		Theme:        "dark",
		Framework:    setup.Framework,
		ToolLimits: ToolLimitsConfig{
			DefaultLimit: 50,  // Default: 50 calls per tool
			TotalLimit:   200, // Safety cap: 200 total calls per session
			PerTool:      DefaultToolLimits,
		},
	}

	// Set provider-specific config (only for the selected provider)
	if setup.Provider == "ollama" {
		config.OllamaConfig = &OllamaConfig{
			Mode:   setup.OllamaMode,
			URL:    setup.OllamaURL,
			APIKey: setup.OllamaKey,
		}
		// Don't set GeminiConfig - it will be omitted from YAML
	} else {
		config.GeminiConfig = &GeminiConfig{
			APIKey: setup.GeminiKey,
		}
		// Don't set OllamaConfig - it will be omitted from YAML
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(ZapFolderName, "config.yaml")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// createMemoryFile creates a memory.json file with versioned format
func createMemoryFile() error {
	memory := map[string]interface{}{
		"version": 1,
		"entries": []interface{}{},
	}
	data, err := yaml.Marshal(memory)
	if err != nil {
		return fmt.Errorf("failed to marshal memory: %w", err)
	}

	memoryPath := filepath.Join(ZapFolderName, "memory.json")
	if err := os.WriteFile(memoryPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write memory file: %w", err)
	}

	return nil
}

// createFile creates an empty file
func createFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", path, err)
	}
	defer file.Close()
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
	path := filepath.Join(ZapFolderName, "falcon.md")
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
	path := filepath.Join(ZapFolderName, "baselines", "README.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write baselines README: %w", err)
	}
	return nil
}
