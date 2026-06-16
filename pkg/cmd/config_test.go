package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock-cli/pkg/config"
)

func TestMaskKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "(not set)"},
		{"short", "abcd", "****"},
		{"eight", "abcdefgh", "****"},
		{"long", "ABCDefghIJKLmnopQRSTuvwxYZABCDEF", "ABCD****CDEF"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, maskKey(tc.input))
		})
	}
}

func TestConfigSetModel_abortOnCtrlC(t *testing.T) {
	m := newConfigSetModel()
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	final := result.(configSetModel)
	assert.True(t, final.aborted)
}

func TestConfigSetModel_emptyAppKeyRejected(t *testing.T) {
	m := newConfigSetModel()
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	final := result.(configSetModel)
	assert.False(t, final.done)
	assert.Contains(t, final.errMsg, "app key cannot be empty")
}

func TestConfigSetModel_emptySecretKeyRejected(t *testing.T) {
	m := newConfigSetModel()
	for _, r := range "myappkey" {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(configSetModel)
	}

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mid := result.(configSetModel)
	assert.Equal(t, 1, mid.current)

	result2, _ := mid.Update(tea.KeyMsg{Type: tea.KeyEnter})
	final := result2.(configSetModel)
	assert.False(t, final.done)
	assert.Contains(t, final.errMsg, "secret key cannot be empty")
}

func TestConfigSetModel_completesAfterSecretKey(t *testing.T) {
	m := newConfigSetModel()
	for _, value := range []string{"app123", "secret456"} {
		for _, r := range value {
			updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			m = updated.(configSetModel)
		}
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = updated.(configSetModel)
	}

	assert.True(t, m.done)
}

func TestConfigSetModel_nonTTY(t *testing.T) {
	orig := isTerminalFn
	defer func() { isTerminalFn = orig }()
	isTerminalFn = func(uintptr) bool { return false }

	err := runConfigSet()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interactive terminal")
}

func TestInvalidateDefaultTokenCache_removesExistingCache(t *testing.T) {
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	require.NoError(t, os.WriteFile(cachePath, []byte(`{"token":"sentinel-token"}`), 0600))
	restoreConfigSetSeams(t)
	defaultTokenCachePathFn = func() (string, error) { return cachePath, nil }

	require.NoError(t, invalidateDefaultTokenCache())
	_, err := os.Stat(cachePath)
	assert.True(t, os.IsNotExist(err))
}

func TestInvalidateDefaultTokenCache_missingCacheSucceeds(t *testing.T) {
	restoreConfigSetSeams(t)
	defaultTokenCachePathFn = func() (string, error) { return filepath.Join(t.TempDir(), ".stock", "token.json"), nil }

	require.NoError(t, invalidateDefaultTokenCache())
}

func TestRunConfigSet_removesTokenCacheAfterSuccessfulSave(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	require.NoError(t, os.WriteFile(cachePath, []byte(`{"token":"sentinel-token"}`), 0600))
	restoreConfigSetSeams(t)
	isTerminalFn = func(uintptr) bool { return true }
	defaultTokenCachePathFn = func() (string, error) { return cachePath, nil }
	runConfigSetTUIFn = func() (configSetModel, error) {
		m := newConfigSetModel()
		m.inputs[0].SetValue("app")
		m.inputs[1].SetValue("secret")
		m.done = true
		return m, nil
	}

	require.NoError(t, runConfigSet())
	_, err := os.Stat(cachePath)
	assert.True(t, os.IsNotExist(err))
}

func TestRunConfigSet_preservesTokenCacheWhenSaveFails(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	require.NoError(t, os.WriteFile(cachePath, []byte(`{"token":"sentinel-token"}`), 0600))
	restoreConfigSetSeams(t)
	isTerminalFn = func(uintptr) bool { return true }
	configSaveFn = func(config.Credentials) error { return errors.New("save failed") }
	defaultTokenCachePathFn = func() (string, error) { return cachePath, nil }
	runConfigSetTUIFn = func() (configSetModel, error) {
		m := newConfigSetModel()
		m.inputs[0].SetValue("app")
		m.inputs[1].SetValue("secret")
		m.done = true
		return m, nil
	}

	err := runConfigSet()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save credentials")
	_, statErr := os.Stat(cachePath)
	assert.NoError(t, statErr)
}

func TestRunConfigSet_tokenCacheRemoveFailureReturnsPartialSuccessError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	require.NoError(t, os.WriteFile(cachePath, []byte(`{"token":"sentinel-token"}`), 0600))
	restoreConfigSetSeams(t)
	isTerminalFn = func(uintptr) bool { return true }
	defaultTokenCachePathFn = func() (string, error) { return cachePath, nil }
	removeFileFn = func(string) error { return errors.New("permission denied") }
	runConfigSetTUIFn = func() (configSetModel, error) {
		m := newConfigSetModel()
		m.inputs[0].SetValue("app")
		m.inputs[1].SetValue("secret")
		m.done = true
		return m, nil
	}

	err := runConfigSet()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "credentials saved, but failed to remove token cache")
	assert.Contains(t, err.Error(), "permission denied")
	assert.NotContains(t, err.Error(), "sentinel-token")
	_, statErr := os.Stat(cachePath)
	assert.NoError(t, statErr)
}

func TestRunConfigSet_preservesTokenCacheWhenTUIAborts(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	require.NoError(t, os.WriteFile(cachePath, []byte(`{"token":"sentinel-token"}`), 0600))
	restoreConfigSetSeams(t)
	isTerminalFn = func(uintptr) bool { return true }
	defaultTokenCachePathFn = func() (string, error) { return cachePath, nil }
	runConfigSetTUIFn = func() (configSetModel, error) {
		return configSetModel{aborted: true}, nil
	}

	err := runConfigSet()
	require.Error(t, err)
	_, statErr := os.Stat(cachePath)
	assert.NoError(t, statErr)
}

func restoreConfigSetSeams(t *testing.T) {
	t.Helper()
	origIsTerminalFn := isTerminalFn
	origRunConfigSetTUIFn := runConfigSetTUIFn
	origConfigSaveFn := configSaveFn
	origDefaultTokenCachePathFn := defaultTokenCachePathFn
	origRemoveFileFn := removeFileFn
	t.Cleanup(func() {
		isTerminalFn = origIsTerminalFn
		runConfigSetTUIFn = origRunConfigSetTUIFn
		configSaveFn = origConfigSaveFn
		defaultTokenCachePathFn = origDefaultTokenCachePathFn
		removeFileFn = origRemoveFileFn
	})
}

func TestConfigSetModel_viewContainsPrompts(t *testing.T) {
	m := newConfigSetModel()
	view := m.View()
	assert.True(t, strings.Contains(view, "App Key"))
}

func TestRootCommandContainsConfig(t *testing.T) {
	var found bool
	for _, child := range Command.Commands {
		if child.Name == "config" {
			found = true
		}
	}
	assert.True(t, found)
}
