package main

import (
	"fmt"
	"os"

	"github.com/blackcoderx/zap/pkg/core"
	"github.com/blackcoderx/zap/pkg/core/tools/persistence"
	"github.com/blackcoderx/zap/pkg/core/tools/shared"
	"github.com/blackcoderx/zap/pkg/tui"
	"github.com/charmbracelet/glamour"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	rootCmd     = &cobra.Command{
		Use:   "zap",
		Short: "ZAP - AI-powered API testing in your terminal",
		Long: `ZAP is the AI-powered developer assistant that lives where you work—your terminal.
It bridges the gap between coding, testing, and fixing by giving you an autonomous
agent that understands your code and can interact with your APIs naturally.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Load .env file if it exists (optional, warn if malformed)
			if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: Failed to load .env file: %v\n", err)
			}

			// Initialize .zap folder (runs setup wizard on first run)
			if err := core.InitializeZapFolder(framework); err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing config folder: %v\n", err)
				os.Exit(1)
			}

			// Re-read config after initialization (first run creates config.json
			// after Viper's initial read, so values would be stale without this)
			_ = viper.ReadInConfig()

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
				fmt.Fprintf(os.Stderr, "Error running ZAP: %v\n", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .zap/config.json)")

	// CLI Flags
	rootCmd.Flags().StringVarP(&requestFile, "request", "r", "", "Execute a saved request file (YAML)")
	rootCmd.Flags().StringVarP(&envName, "env", "e", "dev", "Environment to use for variable substitution")
	rootCmd.Flags().StringVarP(&framework, "framework", "f", "", "API framework (gin, fastapi, express, etc.)")

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ZAP %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
		},
	})
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".zap")
		viper.SetConfigType("json")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}

func runCLI(requestName, env string) error {
	zapDir := core.ZapFolderName

	// Initialize shared components
	responseManager := shared.NewResponseManager()
	varStore := shared.NewVariableStore(zapDir)

	// Initialize tools
	persistManager := persistence.NewPersistenceManager(zapDir)

	// Set environment if specified (TODO: implement simplified SetEnvironment helper if needed,
	// or use persistence tool directly. For CLI simplicity, let's load env var store directly)
	if envName != "" {
		// In a real CLI runner, we'd need a proper way to load environments.
		// For now, let's skip the explicit tool call wrapper and use the store if possible,
		// or just acknowledge this needs the registry to be fully robust.
		// Simplified:
		// varStore.LoadEnvironment(envName) // Hypothetical
	}

	// Load request
	loadTool := persistence.NewLoadRequestTool(persistManager)
	loadArgs := fmt.Sprintf(`{"name_or_id": "%s"}`, requestName)

	reqArgs, err := loadTool.Execute(loadArgs)
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
