package kiwoom

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureTokenReusesValidCache(t *testing.T) {
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	writeTokenCache(t, cachePath, tokenCache{
		Token:     "cached-token",
		TokenType: "Bearer",
		ExpiresDT: "20260611120000",
	})

	c := NewClient("app", "secret",
		WithTokenCachePath(cachePath),
		WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
		WithHTTPClient(roundTripFunc(func(*http.Request) (*http.Response, error) {
			t.Fatal("HTTP client should not be called for a valid cache")
			return nil, nil
		})),
	)

	token, err := c.EnsureToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "cached-token", token)
}

func TestEnsureTokenMissingCredentialsMentionsConfigSetOnly(t *testing.T) {
	c := NewClient("", "")

	_, err := c.EnsureToken(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stock config set")
	assert.NotContains(t, err.Error(), "KIWOOM_APPKEY")
	assert.NotContains(t, err.Error(), "KIWOOM_SECRETKEY")
}

func TestEnsureTokenRepairsValidCachePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows does not support unix permission assertions")
	}
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0755))
	writeTokenCache(t, cachePath, tokenCache{
		Token:     "cached-token",
		TokenType: "Bearer",
		ExpiresDT: "20260611120000",
	})
	require.NoError(t, os.Chmod(filepath.Dir(cachePath), 0755))
	require.NoError(t, os.Chmod(cachePath, 0644))

	c := NewClient("app", "secret",
		WithTokenCachePath(cachePath),
		WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
		WithHTTPClient(roundTripFunc(func(*http.Request) (*http.Response, error) {
			t.Fatal("HTTP client should not be called for a valid cache")
			return nil, nil
		})),
	)

	token, err := c.EnsureToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "cached-token", token)

	info, err := os.Stat(cachePath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	dirInfo, err := os.Stat(filepath.Dir(cachePath))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0700), dirInfo.Mode().Perm())
}

func TestEnsureTokenRefreshesMissingCacheAndWritesPermissions(t *testing.T) {
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/oauth2/token", r.URL.Path)
		assert.Equal(t, "application/json;charset=UTF-8", r.Header.Get("Content-Type"))
		var req tokenIssueRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "client_credentials", req.GrantType)
		assert.Equal(t, "app", req.AppKey)
		assert.Equal(t, "secret", req.SecretKey)
		_, _ = w.Write([]byte(`{"token":"issued-token","token_type":"Bearer","expires_dt":"20260611120000","return_code":0,"return_msg":"ok"}`))
	}))
	defer server.Close()

	c := NewClient("app", "secret",
		WithHost(server.URL),
		WithTokenCachePath(cachePath),
		WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
	)

	token, err := c.EnsureToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "issued-token", token)

	data, err := os.ReadFile(cachePath)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "app")
	assert.NotContains(t, string(data), "secret")

	if runtime.GOOS != "windows" {
		info, err := os.Stat(cachePath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
		dirInfo, err := os.Stat(filepath.Dir(cachePath))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0700), dirInfo.Mode().Perm())
	}
}

func TestEnsureTokenRefreshesMalformedAndExpiredCache(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"malformed", "not json"},
		{"expired", `{"token":"old","token_type":"Bearer","expires_dt":"20260611095900"}`},
		{"near expiry", `{"token":"old","token_type":"Bearer","expires_dt":"20260611100030"}`},
		{"bad expiry", `{"token":"old","token_type":"Bearer","expires_dt":"bad"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			cachePath := filepath.Join(home, ".stock", "token.json")
			require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
			require.NoError(t, os.WriteFile(cachePath, []byte(tc.content), 0600))
			var calls int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				_, _ = w.Write([]byte(`{"token":"new-token","token_type":"Bearer","expires_dt":"20260611120000","return_code":0,"return_msg":"ok"}`))
			}))
			defer server.Close()

			c := NewClient("app", "secret",
				WithHost(server.URL),
				WithTokenCachePath(cachePath),
				WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
			)

			token, err := c.EnsureToken(context.Background())
			require.NoError(t, err)
			assert.Equal(t, "new-token", token)
			assert.Equal(t, 1, calls)
		})
	}
}

func TestPostJSONSendsAuthenticatedRequest(t *testing.T) {
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	writeTokenCache(t, cachePath, tokenCache{
		Token:     "cached-token",
		TokenType: "Bearer",
		ExpiresDT: "20260611120000",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/dostk/acnt", r.URL.Path)
		assert.Equal(t, "Bearer cached-token", r.Header.Get("authorization"))
		assert.Equal(t, "ka10085", r.Header.Get("api-id"))
		assert.Equal(t, "N", r.Header.Get("cont-yn"))
		assert.Equal(t, "", r.Header.Get("next-key"))
		var req AccountProfitRateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "0", req.ExchangeType)
		_, _ = w.Write([]byte(`{"acnt_prft_rt":[],"return_code":0,"return_msg":"ok"}`))
	}))
	defer server.Close()

	c := NewClient("app", "secret",
		WithHost(server.URL),
		WithTokenCachePath(cachePath),
		WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
	)

	res, err := c.AccountProfitRates(context.Background())
	require.NoError(t, err)
	assert.Empty(t, res.Rows)
}

func TestTokenHTTPErrorDoesNotExposeSensitiveBodyFields(t *testing.T) {
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"token":"secret-token","authorization":"Bearer secret-token","return_code":401,"return_msg":"token rejected by Kiwoom"}`))
	}))
	defer server.Close()

	c := NewClient("app", "secret",
		WithHost(server.URL),
		WithTokenCachePath(cachePath),
		WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
	)

	_, err := c.EnsureToken(context.Background())
	require.Error(t, err)
	message := err.Error()
	assert.Contains(t, message, "return_code=401")
	assert.Contains(t, message, `return_msg="token rejected by Kiwoom"`)
	assert.NotContains(t, message, "secret-token")
	assert.NotContains(t, message, "authorization")
}

func TestPostJSONHTTPErrorDoesNotExposeSensitiveBodyFields(t *testing.T) {
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	writeTokenCache(t, cachePath, tokenCache{
		Token:     "cached-token",
		TokenType: "Bearer",
		ExpiresDT: "20260611120000",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"acnt_prft_rt":[{"stk_nm":"private"}],"token":"secret-token","return_code":500,"return_msg":"upstream account failure"}`))
	}))
	defer server.Close()

	c := NewClient("app", "secret",
		WithHost(server.URL),
		WithTokenCachePath(cachePath),
		WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
	)

	var out AccountProfitRateResponse
	err := c.PostJSON(context.Background(), "/api/dostk/acnt", AccountProfitRateRequest{ExchangeType: "0"}, nil, &out)
	require.Error(t, err)
	message := err.Error()
	assert.Contains(t, message, "return_code=500")
	assert.Contains(t, message, `return_msg="upstream account failure"`)
	assert.NotContains(t, message, "secret-token")
	assert.NotContains(t, message, "private")
}

func TestPostJSONHTTPErrorDoesNotExposeInvalidReturnCode(t *testing.T) {
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	writeTokenCache(t, cachePath, tokenCache{
		Token:     "cached-token",
		TokenType: "Bearer",
		ExpiresDT: "20260611120000",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"return_code":"secret-token","return_msg":"diagnostic detail"}`))
	}))
	defer server.Close()

	c := NewClient("app", "secret",
		WithHost(server.URL),
		WithTokenCachePath(cachePath),
		WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
	)

	var out AccountProfitRateResponse
	err := c.PostJSON(context.Background(), "/api/dostk/acnt", AccountProfitRateRequest{ExchangeType: "0"}, nil, &out)
	require.Error(t, err)
	message := err.Error()
	assert.Contains(t, message, "return_code=invalid")
	assert.Contains(t, message, `return_msg="diagnostic detail"`)
	assert.NotContains(t, message, "secret-token")
}

func TestTokenBusinessErrorIncludesReturnMessage(t *testing.T) {
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"return_code":7,"return_msg":"token rejected by Kiwoom"}`))
	}))
	defer server.Close()

	c := NewClient("app", "secret",
		WithHost(server.URL),
		WithTokenCachePath(cachePath),
		WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
	)

	_, err := c.EnsureToken(context.Background())
	require.Error(t, err)
	message := err.Error()
	assert.Contains(t, message, "return_code=7")
	assert.Contains(t, message, `return_msg="token rejected by Kiwoom"`)
	assert.NotContains(t, message, "secret-token")
}

func TestAccountBusinessErrorIncludesReturnMessage(t *testing.T) {
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	writeTokenCache(t, cachePath, tokenCache{
		Token:     "cached-token",
		TokenType: "Bearer",
		ExpiresDT: "20260611120000",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"acnt_prft_rt":[],"return_code":9,"return_msg":"account rejected by Kiwoom"}`))
	}))
	defer server.Close()

	c := NewClient("app", "secret",
		WithHost(server.URL),
		WithTokenCachePath(cachePath),
		WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
	)

	_, err := c.AccountProfitRates(context.Background())
	require.Error(t, err)
	message := err.Error()
	assert.Contains(t, message, "return_code=9")
	assert.Contains(t, message, `return_msg="account rejected by Kiwoom"`)
}

func TestParseExpiresDTUsesKiwoomTimeLocation(t *testing.T) {
	original := time.Local
	time.Local = time.UTC
	t.Cleanup(func() { time.Local = original })

	parsed, err := parseExpiresDT("20260611120000")
	require.NoError(t, err)

	expected := time.Date(2026, 6, 11, 12, 0, 0, 0, kiwoomTimeLocation)
	assert.True(t, parsed.Equal(expected))
	assert.False(t, parsed.Equal(time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)))
}

func writeTokenCache(t *testing.T, path string, cache tokenCache) {
	t.Helper()
	data, err := json.Marshal(cache)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0600))
}

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := parseExpiresDT(value)
	require.NoError(t, err)
	return parsed
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}
