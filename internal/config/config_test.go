package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathXDG(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	p, err := Path()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "aveline", "config.toml"), p)
}

func TestPathHomeFallback(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	p, err := Path()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "aveline", "config.toml"), p)
}

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	in := Config{APIURL: "http://localhost:4000", Token: "avl_x", Workspace: "stable-pod"}
	require.NoError(t, Save(in))

	p, _ := Path()
	info, err := os.Stat(p)
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}

	out, err := Load()
	require.NoError(t, err)
	assert.Equal(t, in, out)
}

func TestLoadMissingIsZero(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	c, err := Load()
	require.NoError(t, err)
	assert.Equal(t, Config{}, c)
}

func TestResolvePrecedence(t *testing.T) {
	c := Config{APIURL: "http://from-config"}
	t.Setenv("AVELINE_API_URL", "")
	assert.Equal(t, "http://from-config", c.Resolve(""))

	t.Setenv("AVELINE_API_URL", "http://from-env")
	assert.Equal(t, "http://from-env", c.Resolve(""))
	assert.Equal(t, "http://from-flag", c.Resolve("http://from-flag"))

	t.Setenv("AVELINE_API_URL", "")
	empty := Config{}
	assert.Equal(t, DefaultAPIURL, empty.Resolve(""))
}
