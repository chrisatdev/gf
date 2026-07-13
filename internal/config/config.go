package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Repo RepoConfig `toml:"repo"`
	Flow FlowConfig `toml:"flow"`
}

type RepoConfig struct {
	Platform    string `toml:"platform"`
	MainBranch  string `toml:"main_branch"`
	ProjectPath string `toml:"project_path"`
}

type FlowConfig struct {
	MFAActive bool `toml:"mfa_active"`
}

const configFile = "gf.toml"

// ErrNotFound is returned when no gf config exists in the current git tree.
var ErrNotFound = errors.New("gf config not found")

// FindGitDir walks upward from the current working directory (max 20 levels)
// looking for a .git directory and returns its absolute path.
func FindGitDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("gf config: %w", err)
	}
	for i := 0; i < 20; i++ {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return gitDir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", ErrNotFound
}

// Load reads .git/gf.toml from the nearest git repo root.
func Load() (*Config, error) {
	gitDir, err := FindGitDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(gitDir, configFile)
	cfg := &Config{
		Repo: RepoConfig{MainBranch: "main"},
	}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("gf config: %w", err)
	}
	return cfg, nil
}

// Write encodes cfg to .git/gf.toml in the nearest git repo root.
func Write(cfg *Config) error {
	gitDir, err := FindGitDir()
	if err != nil {
		return err
	}
	path := filepath.Join(gitDir, configFile)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("gf config: %w", err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("gf config: %w", err)
	}
	return nil
}

// Exists reports whether a gf config file exists in the nearest git repo root.
func Exists() bool {
	gitDir, err := FindGitDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(gitDir, configFile))
	return err == nil
}
