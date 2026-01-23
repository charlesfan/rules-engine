package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	LLM LLMConfig `yaml:"llm"`
	App AppConfig `yaml:"app"`
}

// LLMConfig holds LLM provider configurations
type LLMConfig struct {
	DefaultProvider string       `yaml:"default_provider"`
	Ollama          OllamaConfig `yaml:"ollama"`
	Claude          ClaudeConfig `yaml:"claude"`
}

// OllamaConfig holds Ollama-specific configuration
type OllamaConfig struct {
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
	Timeout int    `yaml:"timeout"` // seconds
}

// ClaudeConfig holds Claude-specific configuration
type ClaudeConfig struct {
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
	Timeout int    `yaml:"timeout"` // seconds
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Debug    bool   `yaml:"debug"`
	LogLevel string `yaml:"log_level"`
}

// Default returns default configuration
func Default() *Config {
	return &Config{
		LLM: LLMConfig{
			DefaultProvider: "ollama",
			Ollama: OllamaConfig{
				BaseURL: "http://localhost:11434",
				Model:   "llama3.1:8b",
				Timeout: 120,
			},
			Claude: ClaudeConfig{
				Model:   "claude-sonnet-4-20250514",
				Timeout: 120,
			},
		},
		App: AppConfig{
			Debug:    true,
			LogLevel: "info",
		},
	}
}

// Load reads configuration from the specified file path
func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Expand environment variables
	expanded := os.ExpandEnv(string(data))

	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadFromDefaultPaths tries to load config from common locations
func LoadFromDefaultPaths() (*Config, error) {
	// Try multiple paths
	paths := []string{
		"config/config.yaml",
		"../config/config.yaml",
		"../../config/config.yaml",
	}

	// Also try relative to executable
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		paths = append(paths,
			filepath.Join(execDir, "config/config.yaml"),
			filepath.Join(execDir, "../config/config.yaml"),
		)
	}

	// Also try relative to working directory
	if wd, err := os.Getwd(); err == nil {
		// If we're in cmd/chatbot, go up to find config
		if strings.Contains(wd, "cmd") {
			paths = append(paths, filepath.Join(wd, "../../config/config.yaml"))
		}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return Load(path)
		}
	}

	// Return default config if no file found
	return Default(), nil
}
