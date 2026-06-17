package cmd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"stock-cli/pkg/kiwoom"
)

func TestOrdersCommandContainsCreateAndCancel(t *testing.T) {
	found := map[string]bool{}
	for _, child := range ordersCmd.Commands {
		found[child.Name] = true
	}
	assert.True(t, found["create"])
	assert.True(t, found["cancel"])
	assert.True(t, found["list"])
}

func TestOrdersCreateCommandContainsCash(t *testing.T) {
	var found bool
	for _, child := range ordersCreateCmd.Commands {
		if child.Name == "cash" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestOrdersCreateCashCommandContainsFlags(t *testing.T) {
	found := map[string]bool{}
	for _, flag := range ordersCreateCashCmd.Flags {
		for _, name := range flag.Names() {
			found[name] = true
		}
	}
	assert.True(t, found["side"])
	assert.True(t, found["stock-code"])
	assert.True(t, found["order-type"])
	assert.True(t, found["quantity"])
	assert.True(t, found["price"])
	assert.True(t, found["trading-venue"])
}

func TestBuildCashOrderRequestBuyLimit(t *testing.T) {
	req, side, err := buildCashOrderRequest(cashOrderOptions{
		Side:         " BUY ",
		StockCode:    "005930",
		OrderType:    "limit",
		Quantity:     "1",
		Price:        "74100",
		TradingVenue: "",
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, cashOrderSideBuy, side)
	assert.Equal(t, "SOR", req.DomesticExchangeType)
	assert.Equal(t, "005930", req.StockCode)
	assert.Equal(t, "1", req.OrderQuantity)
	assert.Equal(t, "74100", req.OrderUnitPrice)
	assert.Equal(t, "0", req.TradeType)
	assert.Equal(t, "", req.ConditionPrice)
}

func TestBuildCashOrderRequestSellMarket(t *testing.T) {
	req, side, err := buildCashOrderRequest(cashOrderOptions{
		Side:         "sell",
		StockCode:    "005930",
		OrderType:    " MARKET ",
		Quantity:     "3",
		TradingVenue: "krx",
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, cashOrderSideSell, side)
	assert.Equal(t, "KRX", req.DomesticExchangeType)
	assert.Equal(t, "005930", req.StockCode)
	assert.Equal(t, "3", req.OrderQuantity)
	assert.Equal(t, "", req.OrderUnitPrice)
	assert.Equal(t, "3", req.TradeType)
	assert.Equal(t, "", req.ConditionPrice)
}

func TestBuildCashOrderRequestRejectsInvalidInputs(t *testing.T) {
	valid := cashOrderOptions{
		Side:         "buy",
		StockCode:    "005930",
		OrderType:    "limit",
		Quantity:     "1",
		Price:        "74100",
		TradingVenue: "SOR",
	}
	tests := []struct {
		name string
		opts cashOrderOptions
		want string
	}{
		{name: "invalid side", opts: withCashOrderSide(valid, "hold"), want: "invalid side"},
		{name: "invalid stock code", opts: withCashOrderStockCode(valid, "00593A"), want: "invalid stock code"},
		{name: "invalid order type", opts: withCashOrderType(valid, "stop"), want: "invalid order type"},
		{name: "missing quantity", opts: withCashOrderQuantity(valid, ""), want: "missing quantity"},
		{name: "zero quantity", opts: withCashOrderQuantity(valid, "0"), want: "invalid quantity"},
		{name: "limit missing price", opts: withCashOrderPrice(valid, ""), want: "missing price"},
		{name: "limit invalid price", opts: withCashOrderPrice(valid, "nope"), want: "invalid price"},
		{name: "invalid venue", opts: withCashOrderVenue(valid, "NYSE"), want: "invalid trading venue"},
		{name: "market with price", opts: cashOrderOptions{Side: "sell", StockCode: "005930", OrderType: "market", Quantity: "1", Price: "74100", TradingVenue: "SOR"}, want: "price is invalid for market orders"},
		{name: "market with explicit empty price", opts: cashOrderOptions{Side: "sell", StockCode: "005930", OrderType: "market", Quantity: "1", PriceSet: true, TradingVenue: "SOR"}, want: "price is invalid for market orders"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := buildCashOrderRequest(tt.opts, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestBuildCashOrderRequestRejectsUnexpectedArgs(t *testing.T) {
	_, _, err := buildCashOrderRequest(cashOrderOptions{}, []string{"extra"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected extra arguments")
}

func TestRunOrdersCreateCashValidatesBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KIWOOM_APPKEY", "")
	t.Setenv("KIWOOM_SECRETKEY", "")

	err := runOrdersCreateCash(context.Background(), cashOrderOptions{
		Side:         "sell",
		StockCode:    "005930",
		OrderType:    "market",
		Quantity:     "1",
		Price:        "74100",
		TradingVenue: "SOR",
	}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "price is invalid for market orders")
	assert.NotContains(t, err.Error(), "missing Kiwoom credentials")
}

func TestOrdersCreateCashExplicitEmptyPriceValidatesBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KIWOOM_APPKEY", "")
	t.Setenv("KIWOOM_SECRETKEY", "")

	cmd := &cli.Command{}
	cmd.Commands = []*cli.Command{&ordersCmd}
	err := cmd.Run(context.Background(), []string{
		"stock",
		"orders",
		"create",
		"cash",
		"--side",
		"sell",
		"--stock-code",
		"005930",
		"--order-type",
		"market",
		"--quantity",
		"1",
		"--price",
		"",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "price is invalid for market orders")
	assert.NotContains(t, err.Error(), "missing Kiwoom credentials")
}

func TestNormalizeCashOrderResponseJSONSchema(t *testing.T) {
	output := normalizeCashOrderResponse(kiwoom.CashOrderResponse{
		OrderID: "0000024",
	}, "SOR")
	data, err := json.Marshal(output)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, map[string]struct{}{
		"order_id":      {},
		"trading_venue": {},
	}, keySet(decoded))
	assert.Equal(t, "0000024", decoded["order_id"])
	assert.Equal(t, "SOR", decoded["trading_venue"])
}

func TestNormalizeCashOrderResponseCanonicalizesTradingVenue(t *testing.T) {
	tests := []struct {
		name        string
		response    kiwoom.CashOrderResponse
		submitted   string
		wantTrading string
	}{
		{
			name:        "lowercase response",
			response:    kiwoom.CashOrderResponse{OrderID: "0000024", DomesticExchangeType: "krx"},
			submitted:   "SOR",
			wantTrading: "KRX",
		},
		{
			name:        "blank response falls back to submitted",
			response:    kiwoom.CashOrderResponse{OrderID: "0000024"},
			submitted:   "NXT",
			wantTrading: "NXT",
		},
		{
			name:        "unknown response",
			response:    kiwoom.CashOrderResponse{OrderID: "0000024", DomesticExchangeType: " night "},
			submitted:   "SOR",
			wantTrading: "UNKNOWN_NIGHT",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := normalizeCashOrderResponse(tt.response, tt.submitted)
			assert.Equal(t, tt.wantTrading, output.TradingVenue)
		})
	}
}

func withCashOrderSide(opts cashOrderOptions, side string) cashOrderOptions {
	opts.Side = side
	return opts
}

func withCashOrderStockCode(opts cashOrderOptions, stockCode string) cashOrderOptions {
	opts.StockCode = stockCode
	return opts
}

func withCashOrderType(opts cashOrderOptions, orderType string) cashOrderOptions {
	opts.OrderType = orderType
	return opts
}

func withCashOrderQuantity(opts cashOrderOptions, quantity string) cashOrderOptions {
	opts.Quantity = quantity
	return opts
}

func withCashOrderPrice(opts cashOrderOptions, price string) cashOrderOptions {
	opts.Price = price
	return opts
}

func withCashOrderVenue(opts cashOrderOptions, tradingVenue string) cashOrderOptions {
	opts.TradingVenue = tradingVenue
	return opts
}
