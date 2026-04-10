package config

import (
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultBranch string `yaml:"default_branch"`
	Editor        string `yaml:"editor"`
	AutoMR        bool   `yaml:"auto_mr"`
	PullRebase    bool   `yaml:"pull_rebase"`
	GitLabToken   string `yaml:"gitlab_token"`
}

func DefaultConfig() *Config {
	return &Config{
		DefaultBranch: "main",
		Editor:        getDefaultEditor(),
		AutoMR:        true,
		PullRebase:    true,
	}
}

func getDefaultEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	return "vim"
}

func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig(), nil
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), nil
	}

	if cfg.DefaultBranch == "" {
		cfg.DefaultBranch = "main"
	}

	return &cfg, nil
}

func getConfigPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(usr.HomeDir, ".config", "gf")
	return filepath.Join(configDir, "config.yaml"), nil
}

func EnsureConfigDir() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(configPath)
	return os.MkdirAll(dir, 0755)
}

func SaveConfig(cfg *Config) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func ExampleConfig() string {
	return `# gf configuration file
# Path: ~/.config/gf/config.yaml

default_branch: main
editor: vim
auto_mr: true
pull_rebase: true

# GitLab token (optional, for API-based MR creation)
# gitlab_token: your_token_here
`
}
