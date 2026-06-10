package config

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFrom_allEmpty(t *testing.T) {
	home := t.TempDir()
	creds, err := LoadFrom(func() (string, error) { return home, nil }, func(string) string { return "" })
	require.NoError(t, err)
	assert.Equal(t, SourceNone, creds.AppKeySource)
	assert.Equal(t, SourceNone, creds.SecretKeySource)
}

func TestLoadFrom_envOnly(t *testing.T) {
	home := t.TempDir()
	env := map[string]string{
		"KIWOOM_APPKEY":    "app-from-env",
		"KIWOOM_SECRETKEY": "secret-from-env",
	}
	creds, err := LoadFrom(func() (string, error) { return home, nil }, func(k string) string { return env[k] })
	require.NoError(t, err)
	assert.Equal(t, "app-from-env", creds.AppKey)
	assert.Equal(t, SourceEnv, creds.AppKeySource)
	assert.Equal(t, "secret-from-env", creds.SecretKey)
	assert.Equal(t, SourceEnv, creds.SecretKeySource)
}

func TestLoadFrom_completeEnvBypassesBrokenToml(t *testing.T) {
	home := t.TempDir()
	configDir := filepath.Join(home, ".stock")
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config"), []byte("not valid toml [[["), 0600))

	env := map[string]string{
		"KIWOOM_APPKEY":    "app-from-env",
		"KIWOOM_SECRETKEY": "secret-from-env",
	}
	creds, err := LoadFrom(func() (string, error) {
		t.Fatal("homeDir should not be called when complete env credentials are set")
		return "", nil
	}, func(k string) string { return env[k] })
	require.NoError(t, err)
	assert.Equal(t, "app-from-env", creds.AppKey)
	assert.Equal(t, "secret-from-env", creds.SecretKey)
}

func TestLoadFrom_fileOnly(t *testing.T) {
	home := t.TempDir()
	writeTestConfig(t, home, "file-app", "file-secret")

	creds, err := LoadFrom(func() (string, error) { return home, nil }, func(string) string { return "" })
	require.NoError(t, err)
	assert.Equal(t, "file-app", creds.AppKey)
	assert.Equal(t, SourceFile, creds.AppKeySource)
	assert.Equal(t, "file-secret", creds.SecretKey)
	assert.Equal(t, SourceFile, creds.SecretKeySource)
}

func TestLoadFrom_envOverridesFile(t *testing.T) {
	home := t.TempDir()
	writeTestConfig(t, home, "file-app", "file-secret")

	env := map[string]string{
		"KIWOOM_APPKEY":    "env-app",
		"KIWOOM_SECRETKEY": "env-secret",
	}
	creds, err := LoadFrom(func() (string, error) { return home, nil }, func(k string) string { return env[k] })
	require.NoError(t, err)
	assert.Equal(t, "env-app", creds.AppKey)
	assert.Equal(t, SourceEnv, creds.AppKeySource)
	assert.Equal(t, "env-secret", creds.SecretKey)
	assert.Equal(t, SourceEnv, creds.SecretKeySource)
}

func TestLoadFrom_mixedSources(t *testing.T) {
	home := t.TempDir()
	writeTestConfig(t, home, "file-app", "file-secret")

	env := map[string]string{"KIWOOM_APPKEY": "env-app"}
	creds, err := LoadFrom(func() (string, error) { return home, nil }, func(k string) string { return env[k] })
	require.NoError(t, err)
	assert.Equal(t, "env-app", creds.AppKey)
	assert.Equal(t, SourceEnv, creds.AppKeySource)
	assert.Equal(t, "file-secret", creds.SecretKey)
	assert.Equal(t, SourceFile, creds.SecretKeySource)
}

func TestLoadFrom_brokenToml(t *testing.T) {
	home := t.TempDir()
	configDir := filepath.Join(home, ".stock")
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config"), []byte("not valid toml [[["), 0600))

	_, err := LoadFrom(func() (string, error) { return home, nil }, func(string) string { return "" })
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
	assert.Contains(t, err.Error(), "stock config set")
}

func TestLoadFrom_loosePermissionsWarning(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not applicable on Windows")
	}
	home := t.TempDir()
	writeTestConfig(t, home, "ak", "sk")
	configPath := filepath.Join(home, ".stock", "config")
	require.NoError(t, os.Chmod(configPath, 0644))

	origStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	_, loadErr := LoadFrom(func() (string, error) { return home, nil }, func(string) string { return "" })

	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	r.Close()

	require.NoError(t, loadErr)
	assert.Contains(t, buf.String(), "warning")
	assert.Contains(t, buf.String(), "loose permissions")
}

func TestSaveRoundtrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	c := Credentials{AppKey: "saved-app", SecretKey: "saved-secret"}
	err := Save(c)
	require.NoError(t, err)

	path := filepath.Join(home, ".stock", "config")
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	dirInfo, err := os.Stat(filepath.Dir(path))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0700), dirInfo.Mode().Perm())

	loaded, err := readFile(path)
	require.NoError(t, err)
	assert.Equal(t, "saved-app", loaded.Kiwoom.AppKey)
	assert.Equal(t, "saved-secret", loaded.Kiwoom.SecretKey)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(data), "# Stock CLI Kiwoom credentials."))
	assert.NotContains(t, string(data), "host")
}

func TestSave_tightensExistingDirectoryPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not applicable on Windows")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".stock")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.Chmod(configDir, 0755))

	err := Save(Credentials{AppKey: "ak", SecretKey: "sk"})
	require.NoError(t, err)

	dirInfo, err := os.Stat(configDir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0700), dirInfo.Mode().Perm())
}

func TestReadFileParsesKiwoomSection(t *testing.T) {
	var buf bytes.Buffer
	cfg := configFile{}
	cfg.Kiwoom.AppKey = "ak"
	cfg.Kiwoom.SecretKey = "sk"
	require.NoError(t, toml.NewEncoder(&buf).Encode(cfg))
	assert.Contains(t, buf.String(), "[kiwoom]")
	assert.NotContains(t, buf.String(), "host")
}

func writeTestConfig(t *testing.T, home, appKey, secretKey string) {
	t.Helper()
	configDir := filepath.Join(home, ".stock")
	require.NoError(t, os.MkdirAll(configDir, 0700))
	content := "[kiwoom]\nappkey = \"" + appKey + "\"\nsecretkey = \"" + secretKey + "\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config"), []byte(content), 0600))
}
