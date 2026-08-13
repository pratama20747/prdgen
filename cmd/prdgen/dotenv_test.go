package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv_SetsUnsetVars(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "DEEPSEEK_API_KEY=sk-test123\n# komentar diabaikan\n\nDEEPSEEK_MODEL=\"deepseek-chat\"\nQUOTED_SINGLE='hello world'\n"
	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	os.Unsetenv("DEEPSEEK_API_KEY")
	os.Unsetenv("DEEPSEEK_MODEL")
	os.Unsetenv("QUOTED_SINGLE")

	loadDotEnv(envPath)

	if got := os.Getenv("DEEPSEEK_API_KEY"); got != "sk-test123" {
		t.Errorf("DEEPSEEK_API_KEY = %q, want sk-test123", got)
	}
	if got := os.Getenv("DEEPSEEK_MODEL"); got != "deepseek-chat" {
		t.Errorf("DEEPSEEK_MODEL = %q, want deepseek-chat (quotes should be stripped)", got)
	}
	if got := os.Getenv("QUOTED_SINGLE"); got != "hello world" {
		t.Errorf("QUOTED_SINGLE = %q, want %q", got, "hello world")
	}
}

func TestLoadDotEnv_DoesNotOverrideExistingEnv(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("DEEPSEEK_API_KEY=from-dotenv\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("DEEPSEEK_API_KEY", "from-shell-export")
	defer os.Unsetenv("DEEPSEEK_API_KEY")

	loadDotEnv(envPath)

	if got := os.Getenv("DEEPSEEK_API_KEY"); got != "from-shell-export" {
		t.Errorf("shell export should win, got %q", got)
	}
}

func TestLoadDotEnv_MissingFileIsNotError(t *testing.T) {
	loadDotEnv("/nonexistent/path/.env")
}
