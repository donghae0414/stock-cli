package cmd

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
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
