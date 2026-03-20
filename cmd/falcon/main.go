package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blackcoderx/falcon/pkg/core"
	"github.com/blackcoderx/falcon/pkg/core/tools/persistence"
	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
	"github.com/blackcoderx/falcon/pkg/tui"
	"github.com/charmbracelet/glamour"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	// Version info (injected by GoReleaser)
	version = "dev"
	commit  = "none"
	date    = "unknown"

	cfgFile     string
	requestFile string
	envName     string
	framework   string
	noIndex     bool
	rootCmd     = &cobra.Command{
		Use:   "falcon",
		Short: "Falcon - AI-powered API testing in your terminal",
		Long: `Falcon is the AI-powered developer assistant that lives where you work—your terminal.
It bridges the gap between coding, testing, and fixing by giving you an autonomous
agent that understands your code and can interact with your APIs naturally.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Load .env file if it exists (optional, warn if malformed)
			if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: Failed to load .env file: %v\n", err)
			}

			// Initialize .falcon folder (runs setup wizard on first run)
			if err := core.InitializeFalconFolder(framework, noIndex); err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing config folder: %v\n", err)
				os.Exit(1)
			}

			// Re-read config after initialization (first run creates config.yaml
			// after Viper's initial read, so values would be stale without this)
			_ = viper.ReadInConfig()
			if gcfg, err := core.LoadGlobalConfig(); err == nil {
				provID, model, values := core.GetActiveProviderEntry(gcfg)
				if provID != "" {
					viper.Set("provider", provID)
					viper.Set("default_model", model)
					viper.Set("provider_config", values)
				}
				if gcfg.Theme != "" {
					viper.Set("theme", gcfg.Theme)
				}
			}

			// CLI Mode: Execute saved request
			if requestFile != "" {
				if err := runCLI(requestFile, envName); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				return
			}

			// Interactive Mode: Start TUI
			if err := tui.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error running Falcon: %v\n", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .falcon/config.yaml)")

	// CLI Flags
	rootCmd.Flags().StringVarP(&requestFile, "request", "r", "", "Execute a saved request file (YAML)")
	rootCmd.Flags().StringVarP(&envName, "env", "e", "dev", "Environment to use for variable substitution")
	rootCmd.Flags().StringVarP(&framework, "framework", "f", "", "API framework (gin, fastapi, express, etc.)")
	rootCmd.Flags().BoolVar(&noIndex, "no-index", false, "Skip automatic API specification indexing")

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Falcon %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
		},
	})

	// Config command — global LLM provider/model setup wizard
	rootCmd.AddCommand(&cobra.Command{
		Use:   "config",
		Short: "Configure global LLM provider and model credentials",
		Long:  "Interactive wizard to set up or update your LLM provider credentials stored in ~/.falcon/config.yaml",
		Run: func(cmd *cobra.Command, args []string) {
			if err := core.RunGlobalConfigWizard(); err != nil {
				if errors.Is(err, core.ErrSetupCancelled) {
					fmt.Println("Setup cancelled.")
					return
				}
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	})
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		viper.AutomaticEnv()
		_ = viper.ReadInConfig()
		return
	}

	// Load global config first (~/.falcon/config.yaml) — provider, model, theme
	globalPath := core.GlobalConfigPath()
	viper.SetConfigFile(globalPath)
	viper.AutomaticEnv()
	_ = viper.ReadInConfig() // no error if not found

	// Inject active provider values as flat viper keys so all downstream code
	// (newLLMClient, collectProviderValues) works without modification.
	if gcfg, err := core.LoadGlobalConfig(); err == nil {
		provID, model, values := core.GetActiveProviderEntry(gcfg)
		if provID != "" {
			viper.Set("provider", provID)
			viper.Set("default_model", model)
			viper.Set("provider_config", values)
		}
		if gcfg.Theme != "" {
			viper.Set("theme", gcfg.Theme)
		}
	}

	// Overlay project config for framework only
	projectPath := filepath.Join(core.FalconFolderName, "config.yaml")
	projectData, err := os.ReadFile(projectPath)
	if err == nil {
		var projectCfg core.Config
		if yaml.Unmarshal(projectData, &projectCfg) == nil {
			if projectCfg.Framework != "" {
				viper.Set("framework", projectCfg.Framework)
			}
		}
	}
}

func runCLI(requestName, env string) error {
	falconDir := core.FalconFolderName

	// Initialize shared components
	responseManager := shared.NewResponseManager()
	varStore := shared.NewVariableStore(falconDir)

	// Initialize tools
	persistManager := persistence.NewPersistenceManager(falconDir)

	// Set environment if specified (TODO: implement simplified SetEnvironment helper if needed,
	// or use persistence tool directly. For CLI simplicity, let's load env var store directly)
	if envName != "" {
		// In a real CLI runner, we'd need a proper way to load environments.
		// For now, let's skip the explicit tool call wrapper and use the store if possible,
		// or just acknowledge this needs the registry to be fully robust.
		// Simplified:
		// varStore.LoadEnvironment(envName) // Hypothetical
	}

	// Load request using unified request tool
	requestTool := persistence.NewRequestTool(persistManager)
	loadArgs := fmt.Sprintf(`{"action":"load","name":"%s"}`, requestName)

	reqArgs, err := requestTool.Execute(loadArgs)
	if err != nil {
		return fmt.Errorf("failed to load request '%s': %w", requestName, err)
	}

	// Execute request
	httpTool := shared.NewHTTPTool(responseManager, varStore)
	resp, err := httpTool.Execute(reqArgs)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	// Render response with Glamour
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		fmt.Println(resp) // Fallback to raw output
		return nil
	}

	out, err := renderer.Render(resp)
	if err != nil {
		fmt.Println(resp) // Fallback
		return nil
	}

	fmt.Print(out)
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
