package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	defaultConfigDirName = "summarizer"
	defaultConfigFile    = "config.toml"
	defaultDataDirName   = "summarizer"
	defaultCredsFile     = "credentials.toml"
)

const DefaultConfigTemplate = `default_provider = "openai"

[providers.openai]
base_url = "https://api.openai.com/v1"
model = "gpt-4o-mini"

[telegram]
allowed_user_ids = []
`

const DefaultCredentialsTemplate = `[providers.openai]
api_key = ""

[telegram]
bot_token = ""
`

// Paths contains resolved config and credentials file paths.
type Paths struct {
	ConfigPath      string
	CredentialsPath string
}

// Config holds non-sensitive runtime configuration.
type Config struct {
	DefaultProvider string                    `toml:"default_provider"`
	Providers       map[string]ProviderConfig `toml:"providers"`
	Telegram        TelegramConfig            `toml:"telegram"`
}

// ProviderConfig contains model endpoint settings for a provider.
type ProviderConfig struct {
	BaseURL string `toml:"base_url"`
	Model   string `toml:"model"`
}

// TelegramConfig contains non-sensitive Telegram settings.
type TelegramConfig struct {
	AllowedUserIDs []int64 `toml:"allowed_user_ids"`
}

// Credentials holds sensitive runtime secrets.
type Credentials struct {
	Providers map[string]ProviderCredentials `toml:"providers"`
	Telegram  TelegramCredentials            `toml:"telegram"`
}

// ProviderCredentials contains provider secrets.
type ProviderCredentials struct {
	APIKey string `toml:"api_key"`
}

// TelegramCredentials contains Telegram secrets.
type TelegramCredentials struct {
	BotToken string `toml:"bot_token"`
}

// AISettings is resolved provider settings used by summarization.
type AISettings struct {
	APIKey  string
	BaseURL string
	Model   string
}

// Loaded holds parsed config and credentials along with source file paths.
type Loaded struct {
	Paths       Paths
	Config      *Config
	Credentials *Credentials
}

// ResolvePaths resolves config and credentials paths using optional overrides.
// If overrides are empty, XDG env vars are used with sensible defaults.
func ResolvePaths(configOverride, credentialsOverride string) (Paths, error) {
	if configOverride != "" && credentialsOverride != "" {
		return Paths{ConfigPath: configOverride, CredentialsPath: credentialsOverride}, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, fmt.Errorf("failed to resolve home directory: %w", err)
	}

	configPath := configOverride
	if configPath == "" {
		configHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
		if configHome == "" {
			configHome = filepath.Join(homeDir, ".config")
		}
		configPath = filepath.Join(configHome, defaultConfigDirName, defaultConfigFile)
	}

	credentialsPath := credentialsOverride
	if credentialsPath == "" {
		dataHome := strings.TrimSpace(os.Getenv("XDG_DATA_HOME"))
		if dataHome == "" {
			dataHome = filepath.Join(homeDir, ".local", "share")
		}
		credentialsPath = filepath.Join(dataHome, defaultDataDirName, defaultCredsFile)
	}

	return Paths{ConfigPath: configPath, CredentialsPath: credentialsPath}, nil
}

// Load resolves paths, parses config files, and returns loaded values.
func Load(configOverride, credentialsOverride string) (*Loaded, error) {
	paths, err := ResolvePaths(configOverride, credentialsOverride)
	if err != nil {
		return nil, err
	}

	cfg, err := loadConfigFile(paths.ConfigPath)
	if err != nil {
		return nil, err
	}

	creds, err := loadCredentialsFile(paths.CredentialsPath)
	if err != nil {
		return nil, err
	}

	return &Loaded{Paths: paths, Config: cfg, Credentials: creds}, nil
}

// AISettings resolves and validates provider settings for summarization.
func (l *Loaded) AISettings(modelOverride string) (AISettings, error) {
	if l == nil || l.Config == nil || l.Credentials == nil {
		return AISettings{}, errors.New("configuration is not loaded")
	}

	providerName := strings.TrimSpace(l.Config.DefaultProvider)
	if providerName == "" {
		return AISettings{}, fmt.Errorf("missing required key 'default_provider' in %s", l.Paths.ConfigPath)
	}

	provider, ok := l.Config.Providers[providerName]
	if !ok {
		return AISettings{}, fmt.Errorf("default provider %q not found under [providers] in %s", providerName, l.Paths.ConfigPath)
	}

	if strings.TrimSpace(provider.BaseURL) == "" {
		return AISettings{}, fmt.Errorf("provider %q is missing base_url in %s", providerName, l.Paths.ConfigPath)
	}
	if strings.TrimSpace(provider.Model) == "" {
		return AISettings{}, fmt.Errorf("provider %q is missing model in %s", providerName, l.Paths.ConfigPath)
	}

	providerCreds, ok := l.Credentials.Providers[providerName]
	if !ok {
		return AISettings{}, fmt.Errorf("provider %q credentials not found under [providers] in %s", providerName, l.Paths.CredentialsPath)
	}
	if strings.TrimSpace(providerCreds.APIKey) == "" {
		return AISettings{}, fmt.Errorf("provider %q is missing api_key in %s", providerName, l.Paths.CredentialsPath)
	}

	resolvedModel := provider.Model
	if strings.TrimSpace(modelOverride) != "" {
		resolvedModel = modelOverride
	}

	return AISettings{
		APIKey:  providerCreds.APIKey,
		BaseURL: provider.BaseURL,
		Model:   resolvedModel,
	}, nil
}

// TelegramBotToken resolves bot token with flag override precedence.
func (l *Loaded) TelegramBotToken(tokenOverride string) (string, error) {
	if strings.TrimSpace(tokenOverride) != "" {
		return tokenOverride, nil
	}
	if l == nil || l.Credentials == nil {
		return "", errors.New("configuration is not loaded")
	}

	token := strings.TrimSpace(l.Credentials.Telegram.BotToken)
	if token == "" {
		return "", fmt.Errorf("telegram bot token is required in %s (key: telegram.bot_token)", l.Paths.CredentialsPath)
	}

	return token, nil
}

func loadConfigFile(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config file not found: %s (run 'summarizer config init')", path)
		}
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg Config
	if err := toml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	if cfg.Providers == nil {
		cfg.Providers = make(map[string]ProviderConfig)
	}

	return &cfg, nil
}

func loadCredentialsFile(path string) (*Credentials, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("credentials file not found: %s (run 'summarizer config init')", path)
		}
		return nil, fmt.Errorf("failed to read credentials file %s: %w", path, err)
	}

	var creds Credentials
	if err := toml.Unmarshal(raw, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file %s: %w", path, err)
	}

	if creds.Providers == nil {
		creds.Providers = make(map[string]ProviderCredentials)
	}

	return &creds, nil
}
