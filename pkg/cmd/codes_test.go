package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock-cli/pkg/config"
	"stock-cli/pkg/kiwoom"
	"stock-cli/pkg/stocklookup"
)

type fakeStockInfoClient struct {
	rows        []kiwoom.StockInfoRow
	err         error
	marketTypes []string
}

func (c *fakeStockInfoClient) StockInfoRows(_ context.Context, marketTypes []string) ([]kiwoom.StockInfoRow, error) {
	c.marketTypes = marketTypes
	return c.rows, c.err
}

func TestRootCommandContainsCodes(t *testing.T) {
	var found bool
	for _, child := range Command.Commands {
		if child.Name == "codes" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestCodesCommandContainsLookup(t *testing.T) {
	found := map[string]bool{}
	for _, child := range codesCmd.Commands {
		found[child.Name] = true
	}
	assert.True(t, found["lookup"])
}

func TestCodesLookupCommandContainsFlags(t *testing.T) {
	found := map[string]bool{}
	for _, flag := range codesLookupCmd.Flags {
		for _, name := range flag.Names() {
			found[name] = true
		}
	}
	assert.True(t, found["name"])
	assert.True(t, found["limit"])
}

func TestParseCodesLookupOptionsPreservesRepeatedNamesAndParsesLimit(t *testing.T) {
	opts, unusedArgs, showHelp := parseCodesLookupArgs([]string{"--name", " 삼성전자 ", "--name=SK하이닉스", "--limit", "2"})
	assert.False(t, showHelp)
	assert.Empty(t, unusedArgs)

	names, limit, err := parseCodesLookupOptions(opts, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"삼성전자", "SK하이닉스"}, names)
	assert.Equal(t, 2, limit)
}

func TestParseCodesLookupOptionsRejectsBlankNameAndBadLimit(t *testing.T) {
	_, _, err := parseCodesLookupOptions(codesLookupOptions{Names: []string{"  "}, Limit: "10"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stock name must not be blank")

	_, _, err = parseCodesLookupOptions(codesLookupOptions{Names: []string{"삼성전자"}, Limit: "nope"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "limit must be an integer")
}

func TestParseCodesLookupArgsKeepsNoValueFailuresInActionPath(t *testing.T) {
	opts, unusedArgs, showHelp := parseCodesLookupArgs([]string{"--name", "삼성전자", "--limit"})
	assert.False(t, showHelp)
	assert.Empty(t, unusedArgs)

	_, _, err := parseCodesLookupOptions(opts, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "limit must be an integer")

	opts, unusedArgs, showHelp = parseCodesLookupArgs([]string{"--name", "--limit", "10"})
	assert.False(t, showHelp)
	assert.Empty(t, unusedArgs)
	_, _, err = parseCodesLookupOptions(opts, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stock name must not be blank")
}

func TestRunCodesLookupValidationFailureBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	stdout, stderr, err := captureConfigOutput(t, func() error {
		return runCodesLookup(context.Background(), codesLookupOptions{Limit: "nope"}, nil)
	})
	require.Error(t, err)
	assert.Empty(t, stderr)
	assertLookupFailure(t, stdout, "ValidationError", "at least one --name is required")
}

func TestRunCodesLookupSuccessWithRepeatedNames(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".stock"), 0700))
	require.NoError(t, config.Save(config.Credentials{AppKey: "app", SecretKey: "secret"}))

	oldNewClient := newStockInfoClient
	defer func() { newStockInfoClient = oldNewClient }()
	fakeClient := &fakeStockInfoClient{rows: []kiwoom.StockInfoRow{
		{Code: "005930", Name: "삼성전자", MarketName: "거래소"},
		{Code: "000660", Name: "SK하이닉스", MarketName: "거래소"},
	}}
	newStockInfoClient = func(config.Credentials) stockInfoClient {
		return fakeClient
	}

	stdout, stderr, err := captureConfigOutput(t, func() error {
		return runCodesLookup(context.Background(), codesLookupOptions{
			Names: []string{" 삼성전자 ", "SK하이닉스"},
			Limit: "10",
		}, nil)
	})
	require.NoError(t, err)
	assert.Empty(t, stderr)

	var output stocklookup.Envelope
	require.NoError(t, json.Unmarshal([]byte(stdout), &output))
	assert.True(t, output.OK)
	require.Len(t, output.Queries, 2)
	assert.Equal(t, "삼성전자", output.Queries[0].Query)
	assert.Equal(t, "005930", output.Queries[0].Candidates[0].Code)
	assert.Equal(t, "SK하이닉스", output.Queries[1].Query)
	assert.Equal(t, "000660", output.Queries[1].Candidates[0].Code)
	assert.Nil(t, fakeClient.marketTypes)
}

func TestRunCodesLookupAPIErrorUsesEnvelope(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".stock"), 0700))
	require.NoError(t, config.Save(config.Credentials{AppKey: "app", SecretKey: "secret"}))

	oldNewClient := newStockInfoClient
	defer func() { newStockInfoClient = oldNewClient }()
	newStockInfoClient = func(config.Credentials) stockInfoClient {
		return &fakeStockInfoClient{err: errors.New("synthetic API failure")}
	}

	stdout, stderr, err := captureConfigOutput(t, func() error {
		return runCodesLookup(context.Background(), codesLookupOptions{Names: []string{"삼성전자"}, Limit: "10"}, nil)
	})
	require.Error(t, err)
	assert.Empty(t, stderr)
	assertLookupFailure(t, stdout, "KiwoomClientError", "synthetic API failure")
}

func TestRunCodesLookupMissingCredentialsUsesConfigError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	stdout, stderr, err := captureConfigOutput(t, func() error {
		return runCodesLookup(context.Background(), codesLookupOptions{Names: []string{"삼성전자"}, Limit: "10"}, nil)
	})
	require.Error(t, err)
	assert.Empty(t, stderr)
	assertLookupFailure(t, stdout, "ConfigError", "stock config set")
}

func assertLookupFailure(t *testing.T, stdout string, wantType string, wantMessage string) {
	t.Helper()

	var output stocklookup.Envelope
	require.NoError(t, json.Unmarshal([]byte(stdout), &output))
	assert.False(t, output.OK)
	assert.Empty(t, output.Queries)
	require.Len(t, output.Errors, 1)
	assert.Equal(t, wantType, output.Errors[0].Type)
	assert.Contains(t, output.Errors[0].Message, wantMessage)
}
