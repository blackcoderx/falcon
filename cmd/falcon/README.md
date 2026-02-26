# cmd/falcon

This is the entry point for Falcon. It uses [Cobra](https://github.com/spf13/cobra) for CLI parsing and [Viper](https://github.com/spf13/viper) for configuration management.

## Package Overview

```
cmd/falcon/
├── main.go     # CLI setup, flag parsing, initialization, routes to TUI or CLI mode
└── update.go   # Self-update subcommand via go-github-selfupdate
```

## CLI Modes

Falcon supports two execution modes:

### Interactive Mode (Default)

Launches the full TUI for interactive API testing and debugging:

```bash
./falcon
./falcon --framework gin
```

### CLI Mode

Executes a saved request and exits — for automation and CI/CD:

```bash
./falcon --request get-users --env prod
./falcon -r get-users -e dev
```

## Command Line Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--framework` | `-f` | Set/update the API framework (gin, fastapi, express, etc.) |
| `--request` | `-r` | Execute a saved request by name (triggers CLI mode) |
| `--env` | `-e` | Environment to use for variable substitution (dev, prod, staging) |
| `--no-index` | | Skip automatic API spec indexing on first run |
| `--config` | | Path to a custom config file |
| `--help` | `-h` | Show help |

## Subcommands

```bash
falcon version   # Print version, commit hash, and build date
falcon update    # Self-update binary to the latest GitHub release
```

## Initialization Flow

On every run, Falcon:

1. Loads environment variables from `.env` (if present)
2. Initializes the `.falcon` folder (runs setup wizard on first run)
3. Migrates legacy config fields if needed
4. Re-reads `config.yaml` after initialization
5. Updates framework in config if `--framework` flag is set
6. Routes to CLI mode (if `--request` is set) or launches the TUI

## CLI Mode

When `--request` is provided, Falcon runs a saved request non-interactively:

```bash
./falcon --request get-users --env prod
```

This:
1. Loads the environment from `.falcon/environments/prod.yaml`
2. Loads the request from `.falcon/requests/get-users.yaml`
3. Substitutes all `{{VAR}}` placeholders with environment values
4. Executes the HTTP request
5. Renders the response to stdout using Glamour markdown
6. Exits with code `0` (success) or `1` (error)

## Configuration Loading

Falcon reads `config.yaml` from `.falcon/config.yaml` by default. A custom path can be provided via `--config`. The config file is YAML:

```yaml
provider: ollama
ollama:
  mode: local
  url: http://localhost:11434
  api_key: ""
default_model: llama3
framework: gin
```

## Environment Variables

Falcon loads `.env` from the project root at startup:

```env
OLLAMA_API_KEY=your-ollama-cloud-key
GEMINI_API_KEY=your-gemini-key
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (config load failure, request not found, HTTP error) |

## Usage Examples

### First-Time Setup

```bash
# Interactive setup wizard (runs automatically on first launch)
./falcon

# Skip wizard by specifying framework directly
./falcon --framework fastapi
```

### Daily Usage

```bash
# Interactive TUI
./falcon

# Execute a saved request
./falcon -r get-users -e dev
./falcon --request create-user --env prod
```

### CI/CD Integration

```bash
#!/bin/bash
./falcon --request health-check --env staging
if [ $? -ne 0 ]; then
    echo "Health check failed"
    exit 1
fi
./falcon --request get-users --env staging
./falcon --request create-user --env staging
```

### Update Framework

```bash
./falcon --framework express
```

## Development

### Building

```bash
go build -o falcon.exe ./cmd/falcon
```

### Running Locally

```bash
go run ./cmd/falcon
go run ./cmd/falcon --framework gin
go run ./cmd/falcon -r my-request -e dev
```

### Adding New Flags

1. Add flag definition in `main.go`:
   ```go
   rootCmd.Flags().Bool("verbose", false, "Enable verbose output")
   ```

2. Retrieve in the `run()` function:
   ```go
   verbose, _ := cmd.Flags().GetBool("verbose")
   ```

3. Update help text if needed.
