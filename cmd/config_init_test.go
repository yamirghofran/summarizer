package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	cfg "github.com/yamirghofran/summarizer/internal/config"
)

func TestInitConfigFilesCreatesFilesAndDirs(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "cfg", "config.toml")
	credentialsPath := filepath.Join(tmp, "data", "credentials.toml")

	paths, warnings, err := initConfigFiles(configPath, credentialsPath, false)
	if err != nil {
		t.Fatalf("initConfigFiles returned error: %v", err)
	}
	if paths.ConfigPath != configPath || paths.CredentialsPath != credentialsPath {
		t.Fatalf("unexpected paths: %+v", paths)
	}
	if runtime.GOOS != "windows" && len(warnings) != 0 {
		t.Fatalf("expected no warnings, got: %v", warnings)
	}

	assertFileContent(t, configPath, cfg.DefaultConfigTemplate)
	assertFileContent(t, credentialsPath, cfg.DefaultCredentialsTemplate)

	if runtime.GOOS != "windows" {
		assertPerm(t, filepath.Dir(credentialsPath), 0o700)
		assertPerm(t, credentialsPath, 0o600)
	}
}

func TestInitConfigFilesNoOverwriteWithoutForce(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.toml")
	credentialsPath := filepath.Join(tmp, "credentials.toml")

	if err := os.WriteFile(configPath, []byte("original-config"), 0o644); err != nil {
		t.Fatalf("failed to seed config file: %v", err)
	}
	if err := os.WriteFile(credentialsPath, []byte("original-creds"), 0o600); err != nil {
		t.Fatalf("failed to seed credentials file: %v", err)
	}

	_, _, err := initConfigFiles(configPath, credentialsPath, false)
	if err == nil || !strings.Contains(err.Error(), "file already exists") {
		t.Fatalf("expected file exists error, got: %v", err)
	}

	assertFileContent(t, configPath, "original-config")
	assertFileContent(t, credentialsPath, "original-creds")
}

func TestInitConfigFilesOverwritesWithForce(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.toml")
	credentialsPath := filepath.Join(tmp, "credentials.toml")

	if err := os.WriteFile(configPath, []byte("old-config"), 0o644); err != nil {
		t.Fatalf("failed to seed config file: %v", err)
	}
	if err := os.WriteFile(credentialsPath, []byte("old-creds"), 0o600); err != nil {
		t.Fatalf("failed to seed credentials file: %v", err)
	}

	_, _, err := initConfigFiles(configPath, credentialsPath, true)
	if err != nil {
		t.Fatalf("initConfigFiles with --force returned error: %v", err)
	}

	assertFileContent(t, configPath, cfg.DefaultConfigTemplate)
	assertFileContent(t, credentialsPath, cfg.DefaultCredentialsTemplate)
}

func assertFileContent(t *testing.T, path, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	if string(content) != expected {
		t.Fatalf("unexpected content in %s\nwant:\n%s\ngot:\n%s", path, expected, string(content))
	}
}

func assertPerm(t *testing.T, path string, expected os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat %s: %v", path, err)
	}
	if info.Mode().Perm() != expected {
		t.Fatalf("unexpected permissions for %s: got %o want %o", path, info.Mode().Perm(), expected)
	}
}
