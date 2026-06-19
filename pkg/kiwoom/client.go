package kiwoom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"stock-cli/pkg/config"
)

const (
	DefaultHost      = "https://api.kiwoom.com"
	tokenEndpoint    = "/oauth2/token"
	tokenRefreshSkew = time.Minute
)

var kiwoomTimeLocation = func() *time.Location {
	location, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return time.FixedZone("Asia/Seoul", 9*60*60)
	}
	return location
}()

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	host           string
	appKey         string
	secretKey      string
	tokenCachePath string
	httpClient     HTTPDoer
	now            func() time.Time
}

type ClientOption func(*Client)

func NewClient(appKey, secretKey string, opts ...ClientOption) *Client {
	c := &Client{
		host:       DefaultHost,
		appKey:     strings.TrimSpace(appKey),
		secretKey:  strings.TrimSpace(secretKey),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		now:        time.Now,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithHost(host string) ClientOption {
	return func(c *Client) {
		c.host = strings.TrimRight(host, "/")
	}
}

func WithHTTPClient(client HTTPDoer) ClientOption {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}

func WithTokenCachePath(path string) ClientOption {
	return func(c *Client) {
		c.tokenCachePath = path
	}
}

func WithNow(now func() time.Time) ClientOption {
	return func(c *Client) {
		if now != nil {
			c.now = now
		}
	}
}

func DefaultTokenCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".stock", "token.json"), nil
}

func (c *Client) EnsureToken(ctx context.Context) (string, error) {
	if c.appKey == "" || c.secretKey == "" {
		return "", fmt.Errorf(config.MissingCredentialsMessage)
	}

	if token, ok := c.readValidCachedToken(); ok {
		return token, nil
	}

	token, err := c.issueToken(ctx)
	if err != nil {
		return "", err
	}
	if err := c.writeTokenCache(token); err != nil {
		return "", err
	}
	return token.Token, nil
}

type continuationRequest struct {
	ContYN  string
	NextKey string
}

func firstContinuationRequest() continuationRequest {
	return continuationRequest{ContYN: "N"}
}

func (r continuationRequest) headers(apiID string) map[string]string {
	return map[string]string{
		"cont-yn":  r.ContYN,
		"next-key": r.NextKey,
		"api-id":   apiID,
	}
}

type continuationResponse struct {
	ContYN  string
	NextKey string
}

func (r continuationResponse) nextRequest(operation string, pages int, maxPages int, seenNextKeys map[string]struct{}) (continuationRequest, bool, error) {
	if r.ContYN != "Y" {
		return continuationRequest{}, false, nil
	}
	if r.NextKey == "" {
		return continuationRequest{}, false, fmt.Errorf("%s requested continuation without next-key", operation)
	}
	if pages >= maxPages {
		return continuationRequest{}, false, fmt.Errorf("%s exceeded continuation page limit (%d)", operation, maxPages)
	}
	if _, ok := seenNextKeys[r.NextKey]; ok {
		return continuationRequest{}, false, fmt.Errorf("%s returned repeated continuation next-key", operation)
	}
	seenNextKeys[r.NextKey] = struct{}{}
	return continuationRequest{ContYN: r.ContYN, NextKey: r.NextKey}, true, nil
}

func postJSONContinuationPages[T any](
	ctx context.Context,
	client *Client,
	endpoint string,
	body any,
	apiID string,
	operation string,
	maxPages int,
	handlePage func(*T) error,
) error {
	seenNextKeys := map[string]struct{}{}
	pages := 0
	continuation := firstContinuationRequest()

	for {
		var page T
		continuationResponse, err := client.postJSONWithContinuation(
			ctx,
			endpoint,
			body,
			continuation.headers(apiID),
			&page,
		)
		if err != nil {
			return err
		}
		if err := handlePage(&page); err != nil {
			return err
		}
		pages++

		nextContinuation, hasNext, err := continuationResponse.nextRequest(operation, pages, maxPages, seenNextKeys)
		if err != nil {
			return err
		}
		if !hasNext {
			return nil
		}
		continuation = nextContinuation
	}
}

func (c *Client) PostJSON(ctx context.Context, endpoint string, body any, headers map[string]string, out any) error {
	// PostJSON is for one-shot requests; paginated endpoints use postJSONContinuationPages.
	_, err := c.postJSONWithContinuation(ctx, endpoint, body, headers, out)
	return err
}

func (c *Client) postJSONWithContinuation(ctx context.Context, endpoint string, body any, headers map[string]string, out any) (continuationResponse, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return continuationResponse{}, err
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return continuationResponse{}, fmt.Errorf("failed to encode Kiwoom request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.host+endpoint, bytes.NewReader(payload))
	if err != nil {
		return continuationResponse{}, fmt.Errorf("failed to create Kiwoom request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("authorization", "Bearer "+token)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return continuationResponse{}, fmt.Errorf("Kiwoom request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return continuationResponse{}, fmt.Errorf("failed to read Kiwoom response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return continuationResponse{}, fmt.Errorf("Kiwoom request failed with HTTP %d: %s", resp.StatusCode, safeResponseSummary(data))
	}
	if err := json.Unmarshal(data, out); err != nil {
		return continuationResponse{}, fmt.Errorf("failed to decode Kiwoom response: %w", err)
	}
	return continuationResponse{
		ContYN:  resp.Header.Get("cont-yn"),
		NextKey: resp.Header.Get("next-key"),
	}, nil
}

type tokenCache struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
	ExpiresDT string `json:"expires_dt"`
}

type tokenIssueRequest struct {
	GrantType string `json:"grant_type"`
	AppKey    string `json:"appkey"`
	SecretKey string `json:"secretkey"`
}

type tokenIssueResponse struct {
	Token      string `json:"token"`
	TokenType  string `json:"token_type"`
	ExpiresDT  string `json:"expires_dt"`
	ReturnCode int    `json:"return_code"`
	ReturnMsg  string `json:"return_msg"`
}

func (c *Client) issueToken(ctx context.Context) (tokenCache, error) {
	body := tokenIssueRequest{
		GrantType: "client_credentials",
		AppKey:    c.appKey,
		SecretKey: c.secretKey,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return tokenCache{}, fmt.Errorf("failed to encode Kiwoom token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.host+tokenEndpoint, bytes.NewReader(payload))
	if err != nil {
		return tokenCache{}, fmt.Errorf("failed to create Kiwoom token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return tokenCache{}, fmt.Errorf("Kiwoom token request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return tokenCache{}, fmt.Errorf("failed to read Kiwoom token response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return tokenCache{}, fmt.Errorf("Kiwoom token request failed with HTTP %d: %s", resp.StatusCode, safeResponseSummary(data))
	}

	var decoded tokenIssueResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		return tokenCache{}, fmt.Errorf("failed to decode Kiwoom token response: %w", err)
	}
	if decoded.ReturnCode != 0 {
		return tokenCache{}, kiwoomReturnCodeError("token request", decoded.ReturnCode, decoded.ReturnMsg)
	}

	cache := tokenCache{
		Token:     decoded.Token,
		TokenType: decoded.TokenType,
		ExpiresDT: decoded.ExpiresDT,
	}
	if !c.isTokenCacheValid(cache) {
		return tokenCache{}, fmt.Errorf("Kiwoom token response did not include a valid token and expires_dt")
	}
	return cache, nil
}

func (c *Client) readValidCachedToken() (string, bool) {
	path, err := c.cachePath()
	if err != nil {
		return "", false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	var cache tokenCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return "", false
	}
	if !c.isTokenCacheValid(cache) {
		return "", false
	}
	if err := ensureTokenCachePermissions(path); err != nil {
		return "", false
	}
	return cache.Token, true
}

func (c *Client) isTokenCacheValid(cache tokenCache) bool {
	if strings.TrimSpace(cache.Token) == "" {
		return false
	}
	expiresAt, err := parseExpiresDT(cache.ExpiresDT)
	if err != nil {
		return false
	}
	return c.now().Add(tokenRefreshSkew).Before(expiresAt)
}

func (c *Client) writeTokenCache(cache tokenCache) error {
	path, err := c.cachePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("could not create token cache directory %s: %w", dir, err)
	}
	if err := ensureTokenCacheDirPermissions(dir); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".token-*.tmp")
	if err != nil {
		return fmt.Errorf("could not create token cache temp file: %w", err)
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if runtime.GOOS != "windows" {
		if err := tmp.Chmod(0600); err != nil {
			_ = tmp.Close()
			return fmt.Errorf("could not set permissions on token cache temp file: %w", err)
		}
	}
	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cache); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("could not write token cache: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("could not close token cache temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("could not replace token cache file: %w", err)
	}
	cleanup = false
	if err := ensureTokenCachePermissions(path); err != nil {
		return err
	}
	return nil
}

func (c *Client) cachePath() (string, error) {
	if c.tokenCachePath != "" {
		return c.tokenCachePath, nil
	}
	return DefaultTokenCachePath()
}

func parseExpiresDT(value string) (time.Time, error) {
	return time.ParseInLocation("20060102150405", value, kiwoomTimeLocation)
}

func safeResponseSummary(data []byte) string {
	var body struct {
		ReturnCode json.RawMessage `json:"return_code"`
		ReturnMsg  json.RawMessage `json:"return_msg"`
	}
	if err := json.Unmarshal(data, &body); err == nil {
		parts := make([]string, 0, 2)
		if len(body.ReturnCode) > 0 {
			parts = append(parts, "return_code="+safeReturnCode(body.ReturnCode))
		}
		if len(body.ReturnMsg) > 0 {
			parts = append(parts, "return_msg="+safeReturnMsg(body.ReturnMsg))
		}
		if len(parts) > 0 {
			return strings.Join(parts, " ")
		}
	}
	return "response body redacted"
}

func kiwoomReturnCodeError(operation string, code int, returnMsg string) error {
	return fmt.Errorf("Kiwoom %s failed return_code=%d return_msg=%q", operation, code, returnMsg)
}

func safeReturnCode(raw json.RawMessage) string {
	var number json.Number
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&number); err != nil {
		return "invalid"
	}
	if _, err := number.Int64(); err != nil {
		return "invalid"
	}
	return number.String()
}

func safeReturnMsg(raw json.RawMessage) string {
	var message string
	if err := json.Unmarshal(raw, &message); err != nil {
		return "invalid"
	}
	return fmt.Sprintf("%q", message)
}

func ensureTokenCacheDirPermissions(dir string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return fmt.Errorf("could not set permissions on token cache directory %s: %w", dir, err)
	}
	return nil
}

func ensureTokenCachePermissions(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	if err := ensureTokenCacheDirPermissions(filepath.Dir(path)); err != nil {
		return err
	}
	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("could not set permissions on token cache file: %w", err)
	}
	return nil
}
