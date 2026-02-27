package main

import (
	"fmt"
	"os"

	"github.com/blackcoderx/falcon/pkg/core"
	"github.com/blackcoderx/falcon/pkg/core/tools/persistence"
	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
	"github.com/blackcoderx/falcon/pkg/tui"
	"github.com/blackcoderx/falcon/pkg/web"
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

			// Start Web UI (alongside TUI, disabled only if explicitly set to false)
			var webShutdown func()
			var webPort int
			if !viper.IsSet("web_ui.enabled") || viper.GetBool("web_ui.enabled") {
				port := viper.GetInt("web_ui.port")
				actualPort, shutdown, err := web.Start(core.FalconFolderName, port)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Web UI failed to start: %v\n", err)
				} else {
					fmt.Printf("Falcon Web UI -> http://localhost:%d\n\n", actualPort)
					webPort = actualPort
					webShutdown = shutdown
				}
			}

			// Interactive Mode: Start TUI
			if err := tui.Run(webPort); err != nil {
				fmt.Fprintf(os.Stderr, "Error running Falcon: %v\n", err)
				os.Exit(1)
			}

			// Shut down web server after TUI exits
			if webShutdown != nil {
				webShutdown()
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
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".falcon")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
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
