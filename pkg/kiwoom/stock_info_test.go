package kiwoom

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStockInfoRowsSendsRequestAndFollowsContinuation(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		assert.Equal(t, stockInfoEndpoint, r.URL.Path)
		assert.Equal(t, "Bearer cached-token", r.Header.Get("authorization"))
		assert.Equal(t, stockInfoAPIID, r.Header.Get("api-id"))

		var req StockInfoRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "0", req.MarketType)

		switch requests {
		case 1:
			assert.Equal(t, "N", r.Header.Get("cont-yn"))
			assert.Equal(t, "", r.Header.Get("next-key"))
			w.Header().Set("cont-yn", "Y")
			w.Header().Set("next-key", "page-2")
			_, _ = w.Write([]byte(`{"list":[{"code":"005930","name":"삼성전자","marketName":"거래소","upName":"전기전자"}],"return_code":0,"return_msg":"ok"}`))
		case 2:
			assert.Equal(t, "Y", r.Header.Get("cont-yn"))
			assert.Equal(t, "page-2", r.Header.Get("next-key"))
			_, _ = w.Write([]byte(`{"list":[{"code":"000660","name":"SK하이닉스","marketName":"거래소","upName":"전기전자"}],"return_code":0,"return_msg":"ok"}`))
		default:
			t.Fatalf("unexpected request count %d", requests)
		}
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	rows, err := c.StockInfoRows(context.Background(), []string{"0"})
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "005930", rows[0].Code)
	assert.Equal(t, "000660", rows[1].Code)
}

func TestStockInfoRowsNilMarketTypesUsesDefaults(t *testing.T) {
	marketTypes := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req StockInfoRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		marketTypes = append(marketTypes, req.MarketType)
		_, _ = w.Write([]byte(`{"list":[],"return_code":0,"return_msg":"ok"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	rows, err := c.StockInfoRows(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, rows)
	assert.Equal(t, []string{"0", "10"}, marketTypes)
}

func TestStockInfoRowsBusinessErrorIncludesReturnMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"return_code":9,"return_msg":"stock info rejected"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	_, err := c.StockInfoRows(context.Background(), []string{"0"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "return_code=9")
	assert.Contains(t, err.Error(), `return_msg="stock info rejected"`)
}
