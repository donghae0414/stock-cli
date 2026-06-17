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

func newCachedOrderTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()

	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	writeTokenCache(t, cachePath, tokenCache{
		Token:     "cached-token",
		TokenType: "Bearer",
		ExpiresDT: "20260611120000",
	})

	return NewClient("app", "secret",
		WithHost(serverURL),
		WithTokenCachePath(cachePath),
		WithNow(func() time.Time { return mustTime(t, "20260611100000") }),
	)
}

func TestCashBuyOrderSendsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, orderEndpoint, r.URL.Path)
		assert.Equal(t, "Bearer cached-token", r.Header.Get("authorization"))
		assert.Equal(t, cashBuyOrderAPIID, r.Header.Get("api-id"))
		assert.Equal(t, "N", r.Header.Get("cont-yn"))
		assert.Equal(t, "", r.Header.Get("next-key"))

		var req CashOrderRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "SOR", req.DomesticExchangeType)
		assert.Equal(t, "005930", req.StockCode)
		assert.Equal(t, "1", req.OrderQuantity)
		assert.Equal(t, "74100", req.OrderUnitPrice)
		assert.Equal(t, "0", req.TradeType)
		assert.Equal(t, "", req.ConditionPrice)

		_, _ = w.Write([]byte(`{"ord_no":"0000024","dmst_stex_tp":"SOR","return_code":0,"return_msg":"ok"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	response, err := c.CashBuyOrder(context.Background(), CashOrderRequest{
		DomesticExchangeType: "SOR",
		StockCode:            "005930",
		OrderQuantity:        "1",
		OrderUnitPrice:       "74100",
		TradeType:            "0",
		ConditionPrice:       "",
	})
	require.NoError(t, err)
	assert.Equal(t, "0000024", response.OrderID)
	assert.Equal(t, "SOR", response.DomesticExchangeType)
}

func TestCashSellOrderSendsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, orderEndpoint, r.URL.Path)
		assert.Equal(t, "Bearer cached-token", r.Header.Get("authorization"))
		assert.Equal(t, cashSellOrderAPIID, r.Header.Get("api-id"))

		var req CashOrderRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "KRX", req.DomesticExchangeType)
		assert.Equal(t, "005930", req.StockCode)
		assert.Equal(t, "3", req.OrderQuantity)
		assert.Equal(t, "", req.OrderUnitPrice)
		assert.Equal(t, "3", req.TradeType)
		assert.Equal(t, "", req.ConditionPrice)

		_, _ = w.Write([]byte(`{"ord_no":"0000138","dmst_stex_tp":"KRX","return_code":0,"return_msg":"ok"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	response, err := c.CashSellOrder(context.Background(), CashOrderRequest{
		DomesticExchangeType: "KRX",
		StockCode:            "005930",
		OrderQuantity:        "3",
		OrderUnitPrice:       "",
		TradeType:            "3",
		ConditionPrice:       "",
	})
	require.NoError(t, err)
	assert.Equal(t, "0000138", response.OrderID)
	assert.Equal(t, "KRX", response.DomesticExchangeType)
}

func TestCashCancelOrderSendsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, orderEndpoint, r.URL.Path)
		assert.Equal(t, "Bearer cached-token", r.Header.Get("authorization"))
		assert.Equal(t, cashCancelAPIID, r.Header.Get("api-id"))
		assert.Equal(t, "N", r.Header.Get("cont-yn"))
		assert.Equal(t, "", r.Header.Get("next-key"))

		var req CashCancelRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "NXT", req.DomesticExchangeType)
		assert.Equal(t, "0000140", req.OriginalOrderID)
		assert.Equal(t, "005930", req.StockCode)
		assert.Equal(t, "0", req.CancelQuantity)

		_, _ = w.Write([]byte(`{"ord_no":"0000141","base_orig_ord_no":"0000140","cncl_qty":"000000000001","return_code":0,"return_msg":"ok"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	response, err := c.CashCancelOrder(context.Background(), CashCancelRequest{
		DomesticExchangeType: "NXT",
		OriginalOrderID:      "0000140",
		StockCode:            "005930",
		CancelQuantity:       "0",
	})
	require.NoError(t, err)
	assert.Equal(t, "0000141", response.OrderID)
	assert.Equal(t, "0000140", response.BaseOriginalOrderID)
	assert.Equal(t, "000000000001", response.CancelQuantity)
}

func TestCreditBuyOrderSendsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, creditOrderEndpoint, r.URL.Path)
		assert.Equal(t, "Bearer cached-token", r.Header.Get("authorization"))
		assert.Equal(t, creditBuyOrderAPIID, r.Header.Get("api-id"))
		assert.Equal(t, "N", r.Header.Get("cont-yn"))
		assert.Equal(t, "", r.Header.Get("next-key"))

		var req CreditOrderRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "SOR", req.DomesticExchangeType)
		assert.Equal(t, "005930", req.StockCode)
		assert.Equal(t, "1", req.OrderQuantity)
		assert.Equal(t, "74100", req.OrderUnitPrice)
		assert.Equal(t, "0", req.TradeType)
		assert.Equal(t, "", req.ConditionPrice)

		_, _ = w.Write([]byte(`{"ord_no":"0001615","dmst_stex_tp":"SOR","return_code":0,"return_msg":"ok"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	response, err := c.CreditBuyOrder(context.Background(), CreditOrderRequest{
		DomesticExchangeType: "SOR",
		StockCode:            "005930",
		OrderQuantity:        "1",
		OrderUnitPrice:       "74100",
		TradeType:            "0",
		ConditionPrice:       "",
	})
	require.NoError(t, err)
	assert.Equal(t, "0001615", response.OrderID)
	assert.Equal(t, "SOR", response.DomesticExchangeType)
}

func TestCreditSellOrderSendsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, creditOrderEndpoint, r.URL.Path)
		assert.Equal(t, "Bearer cached-token", r.Header.Get("authorization"))
		assert.Equal(t, creditSellOrderAPIID, r.Header.Get("api-id"))

		var req CreditSellOrderRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "KRX", req.DomesticExchangeType)
		assert.Equal(t, "005930", req.StockCode)
		assert.Equal(t, "3", req.OrderQuantity)
		assert.Equal(t, "6450", req.OrderUnitPrice)
		assert.Equal(t, "0", req.TradeType)
		assert.Equal(t, "99", req.CreditDealType)
		assert.Equal(t, "", req.CreditLoanDate)
		assert.Equal(t, "", req.ConditionPrice)

		_, _ = w.Write([]byte(`{"ord_no":"0001614","dmst_stex_tp":"KRX","return_code":0,"return_msg":"ok"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	response, err := c.CreditSellOrder(context.Background(), CreditSellOrderRequest{
		DomesticExchangeType: "KRX",
		StockCode:            "005930",
		OrderQuantity:        "3",
		OrderUnitPrice:       "6450",
		TradeType:            "0",
		CreditDealType:       "99",
		CreditLoanDate:       "",
		ConditionPrice:       "",
	})
	require.NoError(t, err)
	assert.Equal(t, "0001614", response.OrderID)
	assert.Equal(t, "KRX", response.DomesticExchangeType)
}

func TestCreditCancelOrderSendsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, creditOrderEndpoint, r.URL.Path)
		assert.Equal(t, "Bearer cached-token", r.Header.Get("authorization"))
		assert.Equal(t, creditCancelAPIID, r.Header.Get("api-id"))
		assert.Equal(t, "N", r.Header.Get("cont-yn"))
		assert.Equal(t, "", r.Header.Get("next-key"))

		var req CreditCancelRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "NXT", req.DomesticExchangeType)
		assert.Equal(t, "0001615", req.OriginalOrderID)
		assert.Equal(t, "005930", req.StockCode)
		assert.Equal(t, "0", req.CancelQuantity)

		_, _ = w.Write([]byte(`{"ord_no":"0001695","base_orig_ord_no":"0001615","cncl_qty":"000000000001","return_code":0,"return_msg":"ok"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	response, err := c.CreditCancelOrder(context.Background(), CreditCancelRequest{
		DomesticExchangeType: "NXT",
		OriginalOrderID:      "0001615",
		StockCode:            "005930",
		CancelQuantity:       "0",
	})
	require.NoError(t, err)
	assert.Equal(t, "0001695", response.OrderID)
	assert.Equal(t, "0001615", response.BaseOriginalOrderID)
	assert.Equal(t, "000000000001", response.CancelQuantity)
}

func TestCashOrderBusinessErrorIncludesReturnMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ord_no":"","return_code":9,"return_msg":"cash order rejected"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	_, err := c.CashBuyOrder(context.Background(), CashOrderRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "return_code=9")
	assert.Contains(t, err.Error(), `return_msg="cash order rejected"`)
}

func TestCashCancelBusinessErrorIncludesReturnMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ord_no":"","return_code":8,"return_msg":"cash cancel rejected"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	_, err := c.CashCancelOrder(context.Background(), CashCancelRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "return_code=8")
	assert.Contains(t, err.Error(), `return_msg="cash cancel rejected"`)
}

func TestCreditOrderBusinessErrorIncludesReturnMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ord_no":"","return_code":9,"return_msg":"credit order rejected"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	_, err := c.CreditBuyOrder(context.Background(), CreditOrderRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "return_code=9")
	assert.Contains(t, err.Error(), `return_msg="credit order rejected"`)
}

func TestCreditCancelBusinessErrorIncludesReturnMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ord_no":"","return_code":8,"return_msg":"credit cancel rejected"}`))
	}))
	defer server.Close()

	c := newCachedOrderTestClient(t, server.URL)
	_, err := c.CreditCancelOrder(context.Background(), CreditCancelRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "return_code=8")
	assert.Contains(t, err.Error(), `return_msg="credit cancel rejected"`)
}

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

func TestOrderListBusinessErrorIncludesReturnMessage(t *testing.T) {
	home := t.TempDir()
	cachePath := filepath.Join(home, ".stock", "token.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))
	writeTokenCache(t, cachePath, tokenCache{
		Token:     "cached-token",
		TokenType: "Bearer",
		ExpiresDT: "20260611120000",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"oso":[],"return_code":9,"return_msg":"order list rejected"}`))
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
	assert.Contains(t, err.Error(), `return_msg="order list rejected"`)
}
