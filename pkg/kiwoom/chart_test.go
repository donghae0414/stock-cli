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

func TestStockDayChartSendsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, chartEndpoint, r.URL.Path)
		assert.Equal(t, "Bearer cached-token", r.Header.Get("authorization"))
		assert.Equal(t, chartDayAPIID, r.Header.Get("api-id"))
		assert.Equal(t, "N", r.Header.Get("cont-yn"))
		assert.Equal(t, "", r.Header.Get("next-key"))

		var req StockDayChartRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "005930_AL", req.StockCode)
		assert.Equal(t, "20260618", req.BaseDate)
		assert.Equal(t, "1", req.UpdatedStockPriceType)

		_, _ = w.Write([]byte(`{"stk_cd":"005930","stk_dt_pole_chart_qry":[{"cur_prc":"70100","trde_prica":"648525","dt":"20260618","open_pric":"69800","high_pric":"70500","low_pric":"69600"}],"return_code":0,"return_msg":"ok"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	response, err := c.StockDayChart(context.Background(), StockDayChartRequest{
		StockCode:             "005930_AL",
		BaseDate:              "20260618",
		UpdatedStockPriceType: "1",
	})
	require.NoError(t, err)
	assert.Equal(t, "005930", response.StockCode)
	require.Len(t, response.Rows, 1)
	assert.Equal(t, "20260618", response.Rows[0].Date)
}

func TestStockWeekChartSendsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, chartEndpoint, r.URL.Path)
		assert.Equal(t, "Bearer cached-token", r.Header.Get("authorization"))
		assert.Equal(t, chartWeekAPIID, r.Header.Get("api-id"))

		var req StockWeekChartRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "005930_AL", req.StockCode)
		assert.Equal(t, "20260618", req.BaseDate)
		assert.Equal(t, "1", req.UpdatedStockPriceType)

		_, _ = w.Write([]byte(`{"stk_cd":"005930","stk_stk_pole_chart_qry":[{"cur_prc":"69500","trde_prica":"3922030535087","dt":"20260615","open_pric":"68400","high_pric":"70400","low_pric":"67500"}],"return_code":0,"return_msg":"ok"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	response, err := c.StockWeekChart(context.Background(), StockWeekChartRequest{
		StockCode:             "005930_AL",
		BaseDate:              "20260618",
		UpdatedStockPriceType: "1",
	})
	require.NoError(t, err)
	assert.Equal(t, "005930", response.StockCode)
	require.Len(t, response.Rows, 1)
	assert.Equal(t, "20260615", response.Rows[0].Date)
}

func TestStockMinuteChartSendsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, chartEndpoint, r.URL.Path)
		assert.Equal(t, "Bearer cached-token", r.Header.Get("authorization"))
		assert.Equal(t, chartMinuteAPIID, r.Header.Get("api-id"))

		var req StockMinuteChartRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "005930_AL", req.StockCode)
		assert.Equal(t, "1", req.TickScope)
		assert.Equal(t, "1", req.UpdatedStockPriceType)
		assert.Equal(t, "20260618", req.BaseDate)

		_, _ = w.Write([]byte(`{"stk_cd":"005930","stk_min_pole_chart_qry":[{"cur_prc":"-78800","cntr_tm":"20260618132000","open_pric":"-78850","high_pric":"-78900","low_pric":"-78800"}],"return_code":0,"return_msg":"ok"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	response, err := c.StockMinuteChart(context.Background(), StockMinuteChartRequest{
		StockCode:             "005930_AL",
		TickScope:             "1",
		UpdatedStockPriceType: "1",
		BaseDate:              "20260618",
	})
	require.NoError(t, err)
	assert.Equal(t, "005930", response.StockCode)
	require.Len(t, response.Rows, 1)
	assert.Equal(t, "20260618132000", response.Rows[0].Timestamp)
}

func TestStockChartBusinessErrorIncludesReturnMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"return_code":9,"return_msg":"chart rejected"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	_, err := c.StockDayChart(context.Background(), StockDayChartRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "return_code=9")
	assert.Contains(t, err.Error(), `return_msg="chart rejected"`)
}
