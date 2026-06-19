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
	found := map[string]bool{}
	for _, child := range ordersCreateCmd.Commands {
		found[child.Name] = true
	}
	assert.True(t, found["cash"])
	assert.True(t, found["credit"])
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

func TestOrdersCreateCreditCommandContainsFlags(t *testing.T) {
	found := map[string]bool{}
	for _, flag := range ordersCreateCreditCmd.Flags {
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
	assert.True(t, found["loan-selection"])
	assert.True(t, found["loan-date"])
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
	assert.Equal(t, orderSideBuy, side)
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
	assert.Equal(t, orderSideSell, side)
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

func TestBuildCreditOrderRequestBuyLimit(t *testing.T) {
	req, side, err := buildCreditOrderRequest(creditOrderOptions{
		Side:         " BUY ",
		StockCode:    "005930",
		OrderType:    "limit",
		Quantity:     "1",
		Price:        "74100",
		TradingVenue: "",
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, orderSideBuy, side)
	assert.Equal(t, "SOR", req.Buy.DomesticExchangeType)
	assert.Equal(t, "005930", req.Buy.StockCode)
	assert.Equal(t, "1", req.Buy.OrderQuantity)
	assert.Equal(t, "74100", req.Buy.OrderUnitPrice)
	assert.Equal(t, "0", req.Buy.TradeType)
	assert.Equal(t, "", req.Buy.ConditionPrice)
}

func TestBuildCreditOrderRequestSellAggregateLimit(t *testing.T) {
	req, side, err := buildCreditOrderRequest(creditOrderOptions{
		Side:             "sell",
		StockCode:        "005930",
		OrderType:        "limit",
		Quantity:         "3",
		Price:            "6450",
		TradingVenue:     "krx",
		LoanSelection:    " aggregate ",
		LoanSelectionSet: true,
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, orderSideSell, side)
	assert.Equal(t, "KRX", req.Sell.DomesticExchangeType)
	assert.Equal(t, "005930", req.Sell.StockCode)
	assert.Equal(t, "3", req.Sell.OrderQuantity)
	assert.Equal(t, "6450", req.Sell.OrderUnitPrice)
	assert.Equal(t, "0", req.Sell.TradeType)
	assert.Equal(t, "99", req.Sell.CreditDealType)
	assert.Equal(t, "", req.Sell.CreditLoanDate)
	assert.Equal(t, "", req.Sell.ConditionPrice)
}

func TestBuildCreditOrderRequestSellIndividualMarket(t *testing.T) {
	req, side, err := buildCreditOrderRequest(creditOrderOptions{
		Side:             "sell",
		StockCode:        "005930",
		OrderType:        " MARKET ",
		Quantity:         "3",
		TradingVenue:     "nxt",
		LoanSelection:    "individual",
		LoanSelectionSet: true,
		LoanDate:         "20260601",
		LoanDateSet:      true,
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, orderSideSell, side)
	assert.Equal(t, "NXT", req.Sell.DomesticExchangeType)
	assert.Equal(t, "005930", req.Sell.StockCode)
	assert.Equal(t, "3", req.Sell.OrderQuantity)
	assert.Equal(t, "", req.Sell.OrderUnitPrice)
	assert.Equal(t, "3", req.Sell.TradeType)
	assert.Equal(t, "33", req.Sell.CreditDealType)
	assert.Equal(t, "20260601", req.Sell.CreditLoanDate)
	assert.Equal(t, "", req.Sell.ConditionPrice)
}

func TestBuildCreditOrderRequestRejectsInvalidInputs(t *testing.T) {
	valid := creditOrderOptions{
		Side:             "sell",
		StockCode:        "005930",
		OrderType:        "limit",
		Quantity:         "1",
		Price:            "74100",
		TradingVenue:     "SOR",
		LoanSelection:    "aggregate",
		LoanSelectionSet: true,
	}
	tests := []struct {
		name string
		opts creditOrderOptions
		want string
	}{
		{name: "invalid side", opts: withCreditOrderSide(valid, "hold"), want: "invalid side"},
		{name: "invalid stock code", opts: withCreditOrderStockCode(valid, "00593A"), want: "invalid stock code"},
		{name: "invalid order type", opts: withCreditOrderType(valid, "stop"), want: "invalid order type"},
		{name: "missing quantity", opts: withCreditOrderQuantity(valid, ""), want: "missing quantity"},
		{name: "zero quantity", opts: withCreditOrderQuantity(valid, "0"), want: "invalid quantity"},
		{name: "limit missing price", opts: withCreditOrderPrice(valid, ""), want: "missing price"},
		{name: "market with price", opts: creditOrderOptions{Side: "sell", StockCode: "005930", OrderType: "market", Quantity: "1", Price: "74100", TradingVenue: "SOR", LoanSelection: "aggregate", LoanSelectionSet: true}, want: "price is invalid for market orders"},
		{name: "invalid venue", opts: withCreditOrderVenue(valid, "NYSE"), want: "invalid trading venue"},
		{name: "buy with loan selection", opts: creditOrderOptions{Side: "buy", StockCode: "005930", OrderType: "limit", Quantity: "1", Price: "74100", TradingVenue: "SOR", LoanSelection: "aggregate", LoanSelectionSet: true}, want: "loan selection is invalid for credit buy orders"},
		{name: "buy with loan date", opts: creditOrderOptions{Side: "buy", StockCode: "005930", OrderType: "limit", Quantity: "1", Price: "74100", TradingVenue: "SOR", LoanDate: "20260601", LoanDateSet: true}, want: "loan date is invalid for credit buy orders"},
		{name: "sell missing loan selection", opts: withCreditOrderLoanSelection(valid, "", false), want: "missing loan selection"},
		{name: "sell invalid loan selection", opts: withCreditOrderLoanSelection(valid, "lot", true), want: "invalid loan selection"},
		{name: "individual missing loan date", opts: withCreditOrderLoanSelection(valid, "individual", true), want: "invalid loan date"},
		{name: "individual invalid loan date", opts: withCreditOrderLoanDate(withCreditOrderLoanSelection(valid, "individual", true), "2026-06-01", true), want: "invalid loan date"},
		{name: "aggregate with loan date", opts: withCreditOrderLoanDate(valid, "20260601", true), want: "loan date is invalid for aggregate loan selection"},
		{name: "aggregate with explicit empty loan date", opts: withCreditOrderLoanDate(valid, "", true), want: "loan date is invalid for aggregate loan selection"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := buildCreditOrderRequest(tt.opts, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestBuildCreditOrderRequestRejectsUnexpectedArgs(t *testing.T) {
	_, _, err := buildCreditOrderRequest(creditOrderOptions{}, []string{"extra"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected extra arguments")
}

func TestRunOrdersCreateCashValidatesBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

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

func TestRunOrdersCreateCreditValidatesBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	err := runOrdersCreateCredit(context.Background(), creditOrderOptions{
		Side:             "sell",
		StockCode:        "005930",
		OrderType:        "market",
		Quantity:         "1",
		TradingVenue:     "SOR",
		LoanSelection:    "aggregate",
		LoanSelectionSet: true,
		LoanDate:         "20260601",
		LoanDateSet:      true,
	}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loan date is invalid for aggregate loan selection")
	assert.NotContains(t, err.Error(), "missing Kiwoom credentials")
}

func TestOrdersCreateCreditExplicitEmptyPriceValidatesBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cmd := &cli.Command{}
	cmd.Commands = []*cli.Command{&ordersCmd}
	err := cmd.Run(context.Background(), []string{
		"stock",
		"orders",
		"create",
		"credit",
		"--side",
		"buy",
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

func TestNormalizeCreditOrderResponseJSONSchema(t *testing.T) {
	output := normalizeCreditOrderResponse(kiwoom.CreditOrderResponse{
		OrderID: "0001615",
	}, "SOR")
	data, err := json.Marshal(output)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, map[string]struct{}{
		"order_id":      {},
		"trading_venue": {},
	}, keySet(decoded))
	assert.Equal(t, "0001615", decoded["order_id"])
	assert.Equal(t, "SOR", decoded["trading_venue"])
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

func withCreditOrderSide(opts creditOrderOptions, side string) creditOrderOptions {
	opts.Side = side
	return opts
}

func withCreditOrderStockCode(opts creditOrderOptions, stockCode string) creditOrderOptions {
	opts.StockCode = stockCode
	return opts
}

func withCreditOrderType(opts creditOrderOptions, orderType string) creditOrderOptions {
	opts.OrderType = orderType
	return opts
}

func withCreditOrderQuantity(opts creditOrderOptions, quantity string) creditOrderOptions {
	opts.Quantity = quantity
	return opts
}

func withCreditOrderPrice(opts creditOrderOptions, price string) creditOrderOptions {
	opts.Price = price
	return opts
}

func withCreditOrderVenue(opts creditOrderOptions, tradingVenue string) creditOrderOptions {
	opts.TradingVenue = tradingVenue
	return opts
}

func withCreditOrderLoanSelection(opts creditOrderOptions, loanSelection string, loanSelectionSet bool) creditOrderOptions {
	opts.LoanSelection = loanSelection
	opts.LoanSelectionSet = loanSelectionSet
	opts.LoanDate = ""
	opts.LoanDateSet = false
	return opts
}

func withCreditOrderLoanDate(opts creditOrderOptions, loanDate string, loanDateSet bool) creditOrderOptions {
	opts.LoanDate = loanDate
	opts.LoanDateSet = loanDateSet
	return opts
}
