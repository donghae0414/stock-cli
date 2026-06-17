package cmd

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock-cli/pkg/kiwoom"
)

func TestOrdersCancelCommandContainsCash(t *testing.T) {
	found := map[string]bool{}
	for _, child := range ordersCancelCmd.Commands {
		found[child.Name] = true
	}
	assert.True(t, found["cash"])
	assert.True(t, found["credit"])
}

func TestOrdersCancelCashCommandContainsFlags(t *testing.T) {
	found := map[string]bool{}
	for _, flag := range ordersCancelCashCmd.Flags {
		for _, name := range flag.Names() {
			found[name] = true
		}
	}
	assert.True(t, found["stock-code"])
	assert.True(t, found["original-order-id"])
	assert.True(t, found["quantity"])
	assert.True(t, found["trading-venue"])
}

func TestOrdersCancelCreditCommandContainsFlags(t *testing.T) {
	found := map[string]bool{}
	for _, flag := range ordersCancelCreditCmd.Flags {
		for _, name := range flag.Names() {
			found[name] = true
		}
	}
	assert.True(t, found["stock-code"])
	assert.True(t, found["original-order-id"])
	assert.True(t, found["quantity"])
	assert.True(t, found["trading-venue"])
}

func TestBuildCashCancelRequestDefaultsQuantity(t *testing.T) {
	req, err := buildCashCancelRequest(orderCancelOptions{
		StockCode:       "005930",
		OriginalOrderID: "0000140",
		TradingVenue:    "",
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, "SOR", req.DomesticExchangeType)
	assert.Equal(t, "0000140", req.OriginalOrderID)
	assert.Equal(t, "005930", req.StockCode)
	assert.Equal(t, "0", req.CancelQuantity)
}

func TestBuildCashCancelRequestWithQuantity(t *testing.T) {
	req, err := buildCashCancelRequest(orderCancelOptions{
		StockCode:       "005930",
		OriginalOrderID: "0000140",
		Quantity:        "2",
		TradingVenue:    "nxt",
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, "NXT", req.DomesticExchangeType)
	assert.Equal(t, "0000140", req.OriginalOrderID)
	assert.Equal(t, "005930", req.StockCode)
	assert.Equal(t, "2", req.CancelQuantity)
}

func TestBuildCashCancelRequestRejectsInvalidInputs(t *testing.T) {
	valid := orderCancelOptions{
		StockCode:       "005930",
		OriginalOrderID: "0000140",
		TradingVenue:    "SOR",
	}
	tests := []struct {
		name string
		opts orderCancelOptions
		want string
	}{
		{name: "invalid stock code", opts: withCashCancelStockCode(valid, "00593A"), want: "invalid stock code"},
		{name: "missing original order id", opts: withCashCancelOriginalOrderID(valid, ""), want: "invalid original order id"},
		{name: "non digit original order id", opts: withCashCancelOriginalOrderID(valid, "00001A0"), want: "invalid original order id"},
		{name: "space wrapped original order id", opts: withCashCancelOriginalOrderID(valid, " 0000140 "), want: "invalid original order id"},
		{name: "unicode digit original order id", opts: withCashCancelOriginalOrderID(valid, "٠٠٠٠١٤٠"), want: "invalid original order id"},
		{name: "zero quantity", opts: withCashCancelQuantity(valid, "0"), want: "invalid quantity"},
		{name: "invalid quantity", opts: withCashCancelQuantity(valid, "one"), want: "invalid quantity"},
		{name: "invalid venue", opts: withCashCancelVenue(valid, "NASDAQ"), want: "invalid trading venue"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildCashCancelRequest(tt.opts, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestBuildCashCancelRequestRejectsUnexpectedArgs(t *testing.T) {
	_, err := buildCashCancelRequest(orderCancelOptions{}, []string{"extra"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected extra arguments")
}

func TestBuildCreditCancelRequestDefaultsQuantity(t *testing.T) {
	req, err := buildCreditCancelRequest(orderCancelOptions{
		StockCode:       "005930",
		OriginalOrderID: "0001615",
		TradingVenue:    "",
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, "SOR", req.DomesticExchangeType)
	assert.Equal(t, "0001615", req.OriginalOrderID)
	assert.Equal(t, "005930", req.StockCode)
	assert.Equal(t, "0", req.CancelQuantity)
}

func TestBuildCreditCancelRequestWithQuantity(t *testing.T) {
	req, err := buildCreditCancelRequest(orderCancelOptions{
		StockCode:       "005930",
		OriginalOrderID: "0001615",
		Quantity:        "2",
		TradingVenue:    "krx",
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, "KRX", req.DomesticExchangeType)
	assert.Equal(t, "0001615", req.OriginalOrderID)
	assert.Equal(t, "005930", req.StockCode)
	assert.Equal(t, "2", req.CancelQuantity)
}

func TestBuildCreditCancelRequestRejectsInvalidInputs(t *testing.T) {
	valid := orderCancelOptions{
		StockCode:       "005930",
		OriginalOrderID: "0001615",
		TradingVenue:    "SOR",
	}
	tests := []struct {
		name string
		opts orderCancelOptions
		want string
	}{
		{name: "invalid stock code", opts: withCashCancelStockCode(valid, "00593A"), want: "invalid stock code"},
		{name: "missing original order id", opts: withCashCancelOriginalOrderID(valid, ""), want: "invalid original order id"},
		{name: "non digit original order id", opts: withCashCancelOriginalOrderID(valid, "00016A5"), want: "invalid original order id"},
		{name: "zero quantity", opts: withCashCancelQuantity(valid, "0"), want: "invalid quantity"},
		{name: "invalid quantity", opts: withCashCancelQuantity(valid, "one"), want: "invalid quantity"},
		{name: "invalid venue", opts: withCashCancelVenue(valid, "NASDAQ"), want: "invalid trading venue"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildCreditCancelRequest(tt.opts, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestBuildCreditCancelRequestRejectsUnexpectedArgs(t *testing.T) {
	_, err := buildCreditCancelRequest(orderCancelOptions{}, []string{"extra"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected extra arguments")
}

func TestNormalizeCashCancelResponseJSONSchema(t *testing.T) {
	output, err := normalizeCashCancelResponse(kiwoom.CashCancelResponse{
		OrderID:             "0000141",
		BaseOriginalOrderID: "0000140",
		CancelQuantity:      "000000000001",
	})
	require.NoError(t, err)
	data, err := json.Marshal(output)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, map[string]struct{}{
		"order_id":               {},
		"base_original_order_id": {},
		"cancelled_quantity":     {},
	}, keySet(decoded))
	assert.Equal(t, "0000141", decoded["order_id"])
	assert.Equal(t, "0000140", decoded["base_original_order_id"])
	assert.Equal(t, float64(1), decoded["cancelled_quantity"])
}

func TestNormalizeCreditCancelResponseJSONSchema(t *testing.T) {
	output, err := normalizeCreditCancelResponse(kiwoom.CreditCancelResponse{
		OrderID:             "0001695",
		BaseOriginalOrderID: "0001615",
		CancelQuantity:      "000000000001",
	})
	require.NoError(t, err)
	data, err := json.Marshal(output)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, map[string]struct{}{
		"order_id":               {},
		"base_original_order_id": {},
		"cancelled_quantity":     {},
	}, keySet(decoded))
	assert.Equal(t, "0001695", decoded["order_id"])
	assert.Equal(t, "0001615", decoded["base_original_order_id"])
	assert.Equal(t, float64(1), decoded["cancelled_quantity"])
}

func withCashCancelStockCode(opts orderCancelOptions, stockCode string) orderCancelOptions {
	opts.StockCode = stockCode
	return opts
}

func withCashCancelOriginalOrderID(opts orderCancelOptions, originalOrderID string) orderCancelOptions {
	opts.OriginalOrderID = originalOrderID
	return opts
}

func withCashCancelQuantity(opts orderCancelOptions, quantity string) orderCancelOptions {
	opts.Quantity = quantity
	return opts
}

func withCashCancelVenue(opts orderCancelOptions, tradingVenue string) orderCancelOptions {
	opts.TradingVenue = tradingVenue
	return opts
}
