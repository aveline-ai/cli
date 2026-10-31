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
// directory if necessary. It prefers an atomic write (a temp file in the
// same directory, then rename, through any symlink) so a concurrent Load
// never sees a truncated file; where the directory itself is not
// writable but the file is, it falls back to writing in place.
func Save(c Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil && !os.IsExist(err) && !os.IsPermission(err) {
		return fmt.Errorf("creating config dir: %w", err)
	}
	if real, err := filepath.EvalSymlinks(p); err == nil {
		p = real
	}
	sweepTemps(filepath.Dir(p))
	if err := saveAtomic(p, c); err == nil {
		return nil
	}
	return saveInPlace(p, c)
}

func saveAtomic(p string, c Config) error {
	f, err := os.CreateTemp(filepath.Dir(p), tempPrefix+"*.toml")
	if err != nil {
		return err
	}
	tmp := f.Name()
	if err := f.Chmod(0o600); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := toml.NewEncoder(f).Encode(c); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("encoding config: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, p); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

func saveInPlace(p string, c Config) error {
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("opening %s: %w", p, err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(c); err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}
	return nil
}

const tempPrefix = ".config-"

// sweepTemps removes temp files a killed Save may have left behind.
func sweepTemps(dir string) {
	matches, _ := filepath.Glob(filepath.Join(dir, tempPrefix+"*.toml"))
	for _, m := range matches {
		_ = os.Remove(m)
	}
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
