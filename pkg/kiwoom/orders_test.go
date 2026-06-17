package kiwoom

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderListSendsRequestAndFollowsContinuation(t *testing.T) {
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	writeTokenCache(t, cachePath, tokenCache{
		Token:     "cached-token",
		TokenType: "Bearer",
		ExpiresDT: "20260611120000",
	})

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		assert.Equal(t, "/api/dostk/acnt", r.URL.Path)
		assert.Equal(t, "Bearer cached-token", r.Header.Get("authorization"))
		assert.Equal(t, orderListAPIID, r.Header.Get("api-id"))

		var req OrderListRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "1", req.AllStockType)
		assert.Equal(t, "2", req.TradeType)
		assert.Equal(t, "005930", req.StockCode)
		assert.Equal(t, "0", req.ExchangeType)

		switch calls {
		case 1:
			assert.Equal(t, "N", r.Header.Get("cont-yn"))
			assert.Equal(t, "", r.Header.Get("next-key"))
			w.Header().Set("cont-yn", "Y")
			w.Header().Set("next-key", "next-page")
			_, _ = w.Write([]byte(`{"oso":[{"ord_no":"0000069","orig_ord_no":"0000000","stk_cd":"005930","stk_nm":"삼성전자","ord_qty":"1","ord_pric":"0","oso_qty":"1","io_tp_nm":"+매수","cntr_qty":"0","cur_prc":"+74100","stex_tp":"0"}],"return_code":0,"return_msg":"ok"}`))
		case 2:
			assert.Equal(t, "Y", r.Header.Get("cont-yn"))
			assert.Equal(t, "next-page", r.Header.Get("next-key"))
			w.Header().Set("cont-yn", "N")
			_, _ = w.Write([]byte(`{"oso":[{"ord_no":"0000070","orig_ord_no":"0000069","stk_cd":"005930","stk_nm":"삼성전자","ord_qty":"2","ord_pric":"74100","oso_qty":"0","io_tp_nm":"신용매수","cntr_qty":"2","cur_prc":"-74100","stex_tp":"1"}],"return_code":0,"return_msg":"ok"}`))
		default:
			t.Fatalf("unexpected call %d", calls)
		}
	}))
	defer server.Close()

	c := NewClient("app", "secret",
		WithHost(server.URL),
		WithTokenCachePath(cachePath),
		WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
	)

	rows, err := c.OrderList(context.Background(), OrderListRequest{
		AllStockType: "1",
		TradeType:    "2",
		StockCode:    "005930",
		ExchangeType: "0",
	})
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
	require.Len(t, rows, 2)
	assert.Equal(t, "0000069", rows[0].OrderID)
	assert.Equal(t, "0", rows[0].ExchangeType)
	assert.Equal(t, "0000070", rows[1].OrderID)
	assert.Equal(t, "1", rows[1].ExchangeType)
}

func TestOrderListContinuationWithoutNextKeyReturnsError(t *testing.T) {
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	writeTokenCache(t, cachePath, tokenCache{
		Token:     "cached-token",
		TokenType: "Bearer",
		ExpiresDT: "20260611120000",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("cont-yn", "Y")
		_, _ = w.Write([]byte(`{"oso":[],"return_code":0,"return_msg":"ok"}`))
	}))
	defer server.Close()

	c := NewClient("app", "secret",
		WithHost(server.URL),
		WithTokenCachePath(cachePath),
		WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
	)

	_, err := c.OrderList(context.Background(), OrderListRequest{AllStockType: "0", TradeType: "0", ExchangeType: "0"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "continuation without next-key")
}

func TestOrderListRepeatedContinuationNextKeyReturnsError(t *testing.T) {
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	writeTokenCache(t, cachePath, tokenCache{
		Token:     "cached-token",
		TokenType: "Bearer",
		ExpiresDT: "20260611120000",
	})

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("cont-yn", "Y")
		w.Header().Set("next-key", "same-page")
		_, _ = w.Write([]byte(`{"oso":[],"return_code":0,"return_msg":"ok"}`))
	}))
	defer server.Close()

	c := NewClient("app", "secret",
		WithHost(server.URL),
		WithTokenCachePath(cachePath),
		WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
	)

	_, err := c.OrderList(context.Background(), OrderListRequest{AllStockType: "0", TradeType: "0", ExchangeType: "0"})
	require.Error(t, err)
	assert.Equal(t, 2, calls)
	assert.Contains(t, err.Error(), "repeated continuation next-key")
}

func TestOrderListContinuationPageLimitReturnsError(t *testing.T) {
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	writeTokenCache(t, cachePath, tokenCache{
		Token:     "cached-token",
		TokenType: "Bearer",
		ExpiresDT: "20260611120000",
	})

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("cont-yn", "Y")
		w.Header().Set("next-key", "page-"+strconv.Itoa(calls))
		_, _ = w.Write([]byte(`{"oso":[],"return_code":0,"return_msg":"ok"}`))
	}))
	defer server.Close()

	c := NewClient("app", "secret",
		WithHost(server.URL),
		WithTokenCachePath(cachePath),
		WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
	)

	_, err := c.OrderList(context.Background(), OrderListRequest{AllStockType: "0", TradeType: "0", ExchangeType: "0"})
	require.Error(t, err)
	assert.Equal(t, orderListMaxPages, calls)
	assert.Contains(t, err.Error(), "exceeded continuation page limit")
}

func TestOrderListBusinessErrorDoesNotExposeReturnMessage(t *testing.T) {
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	writeTokenCache(t, cachePath, tokenCache{
		Token:     "cached-token",
		TokenType: "Bearer",
		ExpiresDT: "20260611120000",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"oso":[],"return_code":9,"return_msg":"private order detail"}`))
	}))
	defer server.Close()

	c := NewClient("app", "secret",
		WithHost(server.URL),
		WithTokenCachePath(cachePath),
		WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
	)

	_, err := c.OrderList(context.Background(), OrderListRequest{AllStockType: "0", TradeType: "0", ExchangeType: "0"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "return_code=9")
	assert.Contains(t, err.Error(), "return_msg=redacted")
	assert.NotContains(t, err.Error(), "private")
}
