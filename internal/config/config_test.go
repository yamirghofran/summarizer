package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePathsUsesXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/custom-config")
	t.Setenv("XDG_DATA_HOME", "/tmp/custom-data")

	paths, err := ResolvePaths("", "")
	if err != nil {
		t.Fatalf("ResolvePaths returned error: %v", err)
	}

	expectedConfig := filepath.Join("/tmp/custom-config", "summarizer", "config.toml")
	expectedCreds := filepath.Join("/tmp/custom-data", "summarizer", "credentials.toml")

	if paths.ConfigPath != expectedConfig {
		t.Fatalf("unexpected config path: got %q, want %q", paths.ConfigPath, expectedConfig)
	}
	if paths.CredentialsPath != expectedCreds {
		t.Fatalf("unexpected credentials path: got %q, want %q", paths.CredentialsPath, expectedCreds)
	}
}

func TestResolvePathsFallbackToHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")

	paths, err := ResolvePaths("", "")
	if err != nil {
		t.Fatalf("ResolvePaths returned error: %v", err)
	}

	expectedConfig := filepath.Join(home, ".config", "summarizer", "config.toml")
	expectedCreds := filepath.Join(home, ".local", "share", "summarizer", "credentials.toml")

	if paths.ConfigPath != expectedConfig {
		t.Fatalf("unexpected config path: got %q, want %q", paths.ConfigPath, expectedConfig)
	}
	if paths.CredentialsPath != expectedCreds {
		t.Fatalf("unexpected credentials path: got %q, want %q", paths.CredentialsPath, expectedCreds)
	}
}

func TestLoadAndResolveAISettings(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.toml")
	credentialsPath := filepath.Join(tmp, "credentials.toml")

	writeFile(t, configPath, `default_provider = "openai"

[providers.openai]
base_url = "https://api.openai.com/v1"
model = "gpt-4o-mini"

[telegram]
allowed_user_ids = [123]
`)
	writeFile(t, credentialsPath, `[providers.openai]
api_key = "sk-test"

[telegram]
bot_token = "123:abc"
`)

	loaded, err := Load(configPath, credentialsPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	settings, err := loaded.AISettings("")
	if err != nil {
		t.Fatalf("AISettings returned error: %v", err)
	}

	if settings.APIKey != "sk-test" || settings.BaseURL != "https://api.openai.com/v1" || settings.Model != "gpt-4o-mini" {
		t.Fatalf("unexpected settings: %+v", settings)
	}

	overridden, err := loaded.AISettings("gpt-5-mini")
	if err != nil {
		t.Fatalf("AISettings override returned error: %v", err)
	}
	if overridden.Model != "gpt-5-mini" {
		t.Fatalf("model override did not apply: got %q", overridden.Model)
	}
}

func TestLoadValidationErrors(t *testing.T) {
	tmp := t.TempDir()

	t.Run("missing default_provider", func(t *testing.T) {
		configPath := filepath.Join(tmp, "missing-default.toml")
		credentialsPath := filepath.Join(tmp, "missing-default-creds.toml")

		writeFile(t, configPath, `[providers.openai]
base_url = "https://api.openai.com/v1"
model = "gpt-4o-mini"
`)
		writeFile(t, credentialsPath, `[providers.openai]
api_key = "sk-test"
`)

		loaded, err := Load(configPath, credentialsPath)
		if err != nil {
			t.Fatalf("Load returned unexpected error: %v", err)
		}

		_, err = loaded.AISettings("")
		if err == nil || !strings.Contains(err.Error(), "default_provider") {
			t.Fatalf("expected default_provider error, got: %v", err)
		}
	})

	t.Run("missing provider entry", func(t *testing.T) {
		configPath := filepath.Join(tmp, "missing-provider.toml")
		credentialsPath := filepath.Join(tmp, "missing-provider-creds.toml")

		writeFile(t, configPath, `default_provider = "openai"

[providers.other]
base_url = "https://example.com/v1"
model = "x"
`)
		writeFile(t, credentialsPath, `[providers.openai]
api_key = "sk-test"
`)

		loaded, err := Load(configPath, credentialsPath)
		if err != nil {
			t.Fatalf("Load returned unexpected error: %v", err)
		}

		_, err = loaded.AISettings("")
		if err == nil || !strings.Contains(err.Error(), "default provider") {
			t.Fatalf("expected missing provider error, got: %v", err)
		}
	})

	t.Run("missing api_key", func(t *testing.T) {
		configPath := filepath.Join(tmp, "missing-key.toml")
		credentialsPath := filepath.Join(tmp, "missing-key-creds.toml")

		writeFile(t, configPath, `default_provider = "openai"

[providers.openai]
base_url = "https://api.openai.com/v1"
model = "gpt-4o-mini"
`)
		writeFile(t, credentialsPath, `[providers.openai]
api_key = ""
`)

		loaded, err := Load(configPath, credentialsPath)
		if err != nil {
			t.Fatalf("Load returned unexpected error: %v", err)
		}

		_, err = loaded.AISettings("")
		if err == nil || !strings.Contains(err.Error(), "api_key") {
			t.Fatalf("expected api_key error, got: %v", err)
		}
	})

	t.Run("missing provider model", func(t *testing.T) {
		configPath := filepath.Join(tmp, "missing-model.toml")
		credentialsPath := filepath.Join(tmp, "missing-model-creds.toml")

		writeFile(t, configPath, `default_provider = "openai"

[providers.openai]
base_url = "https://api.openai.com/v1"
model = ""
`)
		writeFile(t, credentialsPath, `[providers.openai]
api_key = "sk-test"
`)

		loaded, err := Load(configPath, credentialsPath)
		if err != nil {
			t.Fatalf("Load returned unexpected error: %v", err)
		}

		_, err = loaded.AISettings("")
		if err == nil || !strings.Contains(err.Error(), "missing model") {
			t.Fatalf("expected missing model error, got: %v", err)
		}
	})

	t.Run("malformed TOML", func(t *testing.T) {
		configPath := filepath.Join(tmp, "bad-config.toml")
		credentialsPath := filepath.Join(tmp, "bad-creds.toml")

		writeFile(t, configPath, `default_provider = "openai"
[providers.openai
base_url = "https://api.openai.com/v1"`)
		writeFile(t, credentialsPath, `[providers.openai]
api_key = "sk-test"
`)

		_, err := Load(configPath, credentialsPath)
		if err == nil || !strings.Contains(err.Error(), "parse config file") {
			t.Fatalf("expected parse error, got: %v", err)
		}
	})
}

func TestTelegramBotToken(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.toml")
	credentialsPath := filepath.Join(tmp, "credentials.toml")

	writeFile(t, configPath, `default_provider = "openai"

[providers.openai]
base_url = "https://api.openai.com/v1"
model = "gpt-4o-mini"

[telegram]
allowed_user_ids = []
`)
	writeFile(t, credentialsPath, `[providers.openai]
api_key = "sk-test"

[telegram]
bot_token = ""
`)

	loaded, err := Load(configPath, credentialsPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	token, err := loaded.TelegramBotToken("override")
	if err != nil {
		t.Fatalf("override token failed: %v", err)
	}
	if token != "override" {
		t.Fatalf("unexpected override token %q", token)
	}

	_, err = loaded.TelegramBotToken("")
	if err == nil || !strings.Contains(err.Error(), "telegram.bot_token") {
		t.Fatalf("expected missing bot token error, got: %v", err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}
