package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnvConfigPath is the environment variable that overrides the config path.
const EnvConfigPath = "REPORA_CONFIG"

// DefaultDirName is the directory name used under the config home.
const DefaultDirName = "repora"

// DefaultFileName is the configuration file name.
const DefaultFileName = "config.yaml"

// ResolvePath returns the configuration file path using, in order: the explicit
// flag value, the REPORA_CONFIG environment variable, $XDG_CONFIG_HOME, and
// finally ~/.config.
func ResolvePath(flagPath string) (string, error) {
	if strings.TrimSpace(flagPath) != "" {
		return absolute(flagPath)
	}
	if env := strings.TrimSpace(os.Getenv(EnvConfigPath)); env != "" {
		return absolute(env)
	}
	if xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdg != "" {
		return filepath.Join(xdg, DefaultDirName, DefaultFileName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".config", DefaultDirName, DefaultFileName), nil
}

// absolute expands a leading tilde and resolves the path against the working
// directory. Used for the configuration file itself.
func absolute(path string) (string, error) {
	expanded, err := expandTilde(path)
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("resolving %s: %w", path, err)
	}
	return abs, nil
}

// expandTilde expands a leading ~ or ~/ to the user's home directory. Paths
// that do not start with ~ are returned unchanged.
func expandTilde(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("empty path")
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}
