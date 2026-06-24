package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupFailureEnvelopeDoesNotPrintStderr(t *testing.T) {
	bin := t.TempDir() + "/stock"
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Env = os.Environ()
	require.NoError(t, build.Run())

	cmd := exec.Command(bin, "codes", "lookup", "--name", "삼성전자", "--limit", "nope")
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir())

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	assert.Contains(t, stdout.String(), `"ok": false`)
	assert.Contains(t, stdout.String(), `"type": "ValidationError"`)
	assert.Empty(t, stderr.String())
}

func TestLookupMissingLimitValueDoesNotPrintStderr(t *testing.T) {
	bin := t.TempDir() + "/stock"
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Env = os.Environ()
	require.NoError(t, build.Run())

	cmd := exec.Command(bin, "codes", "lookup", "--name", "삼성전자", "--limit")
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir())

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	assert.Contains(t, stdout.String(), `"ok": false`)
	assert.Contains(t, stdout.String(), `"message": "limit must be an integer"`)
	assert.Empty(t, stderr.String())
}
