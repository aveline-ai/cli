// Package config loads and persists the aveline CLI config file.
//
// Resolution order for the config path:
//
//  1. $XDG_CONFIG_HOME/aveline/config.toml
//  2. $HOME/.config/aveline/config.toml
//
// The file is written with 0600 permissions because it holds a bearer token.
//
// Effective settings combine the on-disk config with environment + flag
// overrides via Resolve.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// DefaultAPIURL is the production Aveline API URL.
const DefaultAPIURL = "https://app.aveline.ai"

// Config is the on-disk shape of ~/.config/aveline/config.toml.
type Config struct {
	APIURL    string `toml:"api_url,omitempty"`
	Token     string `toml:"token,omitempty"`
	Workspace string `toml:"workspace,omitempty"`
}

// Path returns the resolved absolute path to the config file. It does not
// require the file to exist.
func Path() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "aveline", "config.toml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".config", "aveline", "config.toml"), nil
}

// Load reads the config file. A missing file is not an error — an empty
// Config is returned.
func Load() (Config, error) {
	p, err := Path()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("reading %s: %w", p, err)
	}
	var c Config
	if err := toml.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("parsing %s: %w", p, err)
	}
	return c, nil
}

// Save writes the config file with 0600 permissions, creating the parent
// directory if necessary.
func Save(c Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("opening %s: %w", p, err)
	}
	defer f.Close()
	enc := toml.NewEncoder(f)
	if err := enc.Encode(c); err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}
	return nil
}

// Resolve produces the effective API URL given precedence:
// flag > AVELINE_API_URL > config > DefaultAPIURL.
func (c Config) Resolve(flagAPIURL string) string {
	if flagAPIURL != "" {
		return flagAPIURL
	}
	if env := os.Getenv("AVELINE_API_URL"); env != "" {
		return env
	}
	if c.APIURL != "" {
		return c.APIURL
	}
	return DefaultAPIURL
}
