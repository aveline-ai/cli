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
	// PrimerVersions is, per API URL, the version of Aveline's base doc
	// (the "primer") last seen in a get-orientation response, so the CLI
	// can say once when it changes. Keyed by URL because a dev server and
	// production can be at different versions. Absent means never seen.
	PrimerVersions map[string]int `toml:"primer_versions,omitempty"`
}

// PrimerVersion is the primer version last seen from apiURL, or 0.
func (c Config) PrimerVersion(apiURL string) int {
	return c.PrimerVersions[apiURL]
}

// WithPrimerVersion returns a copy of c remembering v for apiURL.
func (c Config) WithPrimerVersion(apiURL string, v int) Config {
	m := make(map[string]int, len(c.PrimerVersions)+1)
	for k, val := range c.PrimerVersions {
		m[k] = val
	}
	m[apiURL] = v
	c.PrimerVersions = m
	return c
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
// directory if necessary. The write is atomic (a temp file in the same
// directory, then rename), so a concurrent Load never sees a truncated
// file.
func Save(c Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	f, err := os.CreateTemp(filepath.Dir(p), ".config-*.toml")
	if err != nil {
		return fmt.Errorf("creating temp config: %w", err)
	}
	tmp := f.Name()
	cleanup := func() { f.Close(); os.Remove(tmp) }
	if err := f.Chmod(0o600); err != nil {
		cleanup()
		return fmt.Errorf("setting config mode: %w", err)
	}
	if err := toml.NewEncoder(f).Encode(c); err != nil {
		cleanup()
		return fmt.Errorf("encoding config: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("writing %s: %w", p, err)
	}
	if err := os.Rename(tmp, p); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("replacing %s: %w", p, err)
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
