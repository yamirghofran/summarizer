package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	cfg "github.com/yamirghofran/summarizer/internal/config"
)

var (
	configInitForce           bool
	configInitConfigPath      string
	configInitCredentialsPath string
)

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config and credentials files",
	Long: `Create default TOML configuration files for summarize and bot commands.

By default files are created at:
  ~/.config/summarizer/config.toml
  ~/.local/share/summarizer/credentials.toml`,
	Run: runConfigInit,
}

func init() {
	configCmd.AddCommand(configInitCmd)

	configInitCmd.Flags().BoolVar(&configInitForce, "force", false, "Overwrite existing files")
	configInitCmd.Flags().StringVar(&configInitConfigPath, "config-path", "", "Custom path for config.toml")
	configInitCmd.Flags().StringVar(&configInitCredentialsPath, "credentials-path", "", "Custom path for credentials.toml")
}

func runConfigInit(cmd *cobra.Command, args []string) {
	paths, warnings, err := initConfigFiles(configInitConfigPath, configInitCredentialsPath, configInitForce)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, warning := range warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
	}

	fmt.Println("✓ Configuration initialized")
	fmt.Printf("  Config:      %s\n", paths.ConfigPath)
	fmt.Printf("  Credentials: %s\n", paths.CredentialsPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit config.toml to choose your provider/model")
	fmt.Println("  2. Add API key and Telegram token to credentials.toml")
}

func initConfigFiles(configPathOverride, credentialsPathOverride string, force bool) (cfg.Paths, []string, error) {
	paths, err := cfg.ResolvePaths(configPathOverride, credentialsPathOverride)
	if err != nil {
		return cfg.Paths{}, nil, err
	}

	var warnings []string

	configDir := filepath.Dir(paths.ConfigPath)
	credentialsDir := filepath.Dir(paths.CredentialsPath)

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return cfg.Paths{}, nil, fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}
	if err := os.MkdirAll(credentialsDir, 0o700); err != nil {
		return cfg.Paths{}, nil, fmt.Errorf("failed to create data directory %s: %w", credentialsDir, err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(credentialsDir, 0o700); err != nil {
			warnings = append(warnings, fmt.Sprintf("could not enforce permissions on %s: %v", credentialsDir, err))
		}
	} else {
		warnings = append(warnings, "permission hardening checks are best-effort on Windows")
	}

	if err := writeTemplateFile(paths.ConfigPath, cfg.DefaultConfigTemplate, 0o644, force); err != nil {
		return cfg.Paths{}, warnings, err
	}
	if err := writeTemplateFile(paths.CredentialsPath, cfg.DefaultCredentialsTemplate, 0o600, force); err != nil {
		return cfg.Paths{}, warnings, err
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(paths.CredentialsPath, 0o600); err != nil {
			warnings = append(warnings, fmt.Sprintf("could not enforce permissions on %s: %v", paths.CredentialsPath, err))
		}
	}

	return paths, warnings, nil
}

func writeTemplateFile(path, content string, mode os.FileMode, force bool) error {
	if _, err := os.Stat(path); err == nil && !force {
		return fmt.Errorf("file already exists: %s (use --force to overwrite)", path)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed checking file %s: %w", path, err)
	}

	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		return fmt.Errorf("failed writing file %s: %w", path, err)
	}

	return nil
}
