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

func TestRootCommandContainsOrders(t *testing.T) {
	var found bool
	for _, child := range Command.Commands {
		if child.Name == "orders" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestOrdersCommandContainsList(t *testing.T) {
	var found bool
	for _, child := range ordersCmd.Commands {
		if child.Name == "list" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestOrdersListCommandContainsFlags(t *testing.T) {
	found := map[string]bool{}
	for _, flag := range ordersListCmd.Flags {
		for _, name := range flag.Names() {
			found[name] = true
		}
	}
	assert.True(t, found["side"])
	assert.True(t, found["stock-code"])
}

func TestOrderSideTradeType(t *testing.T) {
	tests := []struct {
		name string
		side string
		want string
	}{
		{name: "default empty", side: "", want: "0"},
		{name: "all", side: "all", want: "0"},
		{name: "sell", side: "sell", want: "1"},
		{name: "buy", side: "buy", want: "2"},
		{name: "trim and case", side: " BUY ", want: "2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := orderSideTradeType(tt.side)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	_, err := orderSideTradeType("hold")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid side")
}

func TestBuildOrderListRequest(t *testing.T) {
	req, err := buildOrderListRequest(orderListOptions{Side: "all"})
	require.NoError(t, err)
	assert.Equal(t, "0", req.AllStockType)
	assert.Equal(t, "0", req.TradeType)
	assert.Equal(t, "", req.StockCode)
	assert.Equal(t, "0", req.ExchangeType)

	req, err = buildOrderListRequest(orderListOptions{Side: "buy", StockCode: "005930"})
	require.NoError(t, err)
	assert.Equal(t, "1", req.AllStockType)
	assert.Equal(t, "2", req.TradeType)
	assert.Equal(t, "005930", req.StockCode)
	assert.Equal(t, "0", req.ExchangeType)
}

func TestBuildOrderListRequestRejectsInvalidStockCode(t *testing.T) {
	tests := []string{"5930", "00593A"}
	for _, stockCode := range tests {
		t.Run(stockCode, func(t *testing.T) {
			_, err := buildOrderListRequest(orderListOptions{Side: "all", StockCode: stockCode})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid stock code")
		})
	}
}

func TestNormalizeOrderListJSONSchema(t *testing.T) {
	orders, err := normalizeOrderList([]kiwoom.OrderListRow{
		{
			OrderID:          "0000069",
			OriginalOrderID:  "0000000",
			StockCode:        "005930",
			StockName:        "삼성전자",
			ExchangeType:     "0",
			OrderedQuantity:  "1",
			OrderedPrice:     "0",
			UnfilledQuantity: "1",
			OrderKind:        "+매수",
			FilledQuantity:   "0",
			CurrentPrice:     "+74100",
		},
		{
			OrderID:          "0000070",
			OriginalOrderID:  "0000069",
			StockCode:        "005930",
			StockName:        "삼성전자",
			ExchangeType:     "1",
			OrderedQuantity:  "+2",
			OrderedPrice:     "-74100",
			UnfilledQuantity: "-0",
			OrderKind:        "신용매수",
			FilledQuantity:   "+2",
			CurrentPrice:     "-74100",
		},
	})
	require.NoError(t, err)
	require.Len(t, orders, 2)
	assert.Equal(t, "0000069", orders[0].OrderID)
	assert.Equal(t, "0000000", orders[0].OriginalOrderID)
	assert.Equal(t, "SOR", orders[0].TradingVenue)
	assert.Equal(t, orderListSideBuy, orders[0].Side)
	assert.Equal(t, FundingTypeCash, orders[0].FundingType)
	assert.Equal(t, "KRX", orders[1].TradingVenue)
	assert.Equal(t, orderListSideBuy, orders[1].Side)
	assert.Equal(t, FundingTypeCredit, orders[1].FundingType)
	assert.Equal(t, int64(74100), orders[1].OrderedPrice)
	assert.Equal(t, int64(74100), orders[1].CurrentPrice)

	data, err := json.Marshal(orders)
	require.NoError(t, err)
	var decoded []map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))

	expected := map[string]struct{}{
		"order_id":          {},
		"original_order_id": {},
		"stock_code":        {},
		"stock_name":        {},
		"side":              {},
		"trading_venue":     {},
		"ordered_quantity":  {},
		"ordered_price":     {},
		"unfilled_quantity": {},
		"funding_type":      {},
		"filled_quantity":   {},
		"current_price":     {},
	}
	for _, row := range decoded {
		assert.Equal(t, expected, keySet(row))
		assert.IsType(t, "", row["order_id"])
		assert.IsType(t, "", row["original_order_id"])
		assert.IsType(t, "", row["side"])
		assert.IsType(t, "", row["trading_venue"])
		assert.IsType(t, "", row["funding_type"])
		assert.NotContains(t, row, "stex_tp")
		assert.NotContains(t, row, "io_tp_nm")
	}
}

func TestTradingVenueFromKiwoomExchangeType(t *testing.T) {
	tests := []struct {
		exchangeType string
		want         string
	}{
		{exchangeType: "0", want: "SOR"},
		{exchangeType: "1", want: "KRX"},
		{exchangeType: "2", want: "NXT"},
		{exchangeType: " 1 ", want: "KRX"},
		{exchangeType: "", want: "UNKNOWN"},
		{exchangeType: "9", want: "UNKNOWN_9"},
	}
	for _, tt := range tests {
		t.Run(tt.exchangeType, func(t *testing.T) {
			got := tradingVenueFromKiwoomExchangeType(tt.exchangeType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildOrderListItemPreservesUnknownExchangeType(t *testing.T) {
	order, err := buildOrderListItem(kiwoom.OrderListRow{
		OrderID:          "0000069",
		OriginalOrderID:  "0000000",
		StockCode:        "005930",
		StockName:        "삼성전자",
		ExchangeType:     "9",
		OrderedQuantity:  "1",
		OrderedPrice:     "0",
		UnfilledQuantity: "1",
		OrderKind:        "+매수",
		FilledQuantity:   "0",
		CurrentPrice:     "+74100",
	})
	require.NoError(t, err)
	assert.Equal(t, "UNKNOWN_9", order.TradingVenue)
}

func TestClassifyKiwoomOrderKindSide(t *testing.T) {
	tests := []struct {
		name      string
		orderKind string
		want      orderListSide
	}{
		{name: "buy marker", orderKind: "+매수", want: orderListSideBuy},
		{name: "buy word", orderKind: "매수", want: orderListSideBuy},
		{name: "credit buy", orderKind: "+매수신용", want: orderListSideBuy},
		{name: "whitespace buy", orderKind: "  +매수신용  ", want: orderListSideBuy},
		{name: "sell marker", orderKind: "-매도", want: orderListSideSell},
		{name: "sell word", orderKind: "매도", want: orderListSideSell},
		{name: "credit sell", orderKind: "-매도신용", want: orderListSideSell},
		{name: "whitespace sell", orderKind: "  -매도신용  ", want: orderListSideSell},
		{name: "empty", orderKind: "", want: orderListSideUnknown},
		{name: "unrecognized", orderKind: "정정", want: orderListSideUnknown},
		{name: "ambiguous", orderKind: "매수매도", want: orderListSideUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, classifyKiwoomOrderKind(tt.orderKind).Side)
		})
	}
}

func TestOrdersListRejectsUnexpectedArgs(t *testing.T) {
	cmd := &cli.Command{}
	cmd.Commands = []*cli.Command{&ordersCmd}
	err := cmd.Run(context.Background(), []string{"stock", "orders", "list", "extra"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected extra arguments")
}

func TestRunOrdersListMissingCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	err := runOrdersList(context.Background(), orderListOptions{Side: "all"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stock config set")
	assert.NotContains(t, err.Error(), "KIWOOM_APPKEY")
	assert.NotContains(t, err.Error(), "KIWOOM_SECRETKEY")
}
