package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/urfave/cli/v3"

	"stock-cli/pkg/config"
	"stock-cli/pkg/kiwoom"
)

var ordersCmd = cli.Command{
	Name:     "orders",
	Usage:    "Manage Kiwoom order resources",
	Category: "API RESOURCE",
	Suggest:  true,
	Commands: []*cli.Command{
		&ordersListCmd,
	},
}

var ordersListCmd = cli.Command{
	Name:    "list",
	Usage:   "Get current open orders",
	Suggest: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "side",
			Usage: "Filter by order side: all, buy, or sell",
			Value: "all",
		},
		&cli.StringFlag{
			Name:  "stock-code",
			Usage: "Filter by six-digit stock code",
		},
	},
	Action:          handleOrdersList,
	HideHelpCommand: true,
}

func handleOrdersList(ctx context.Context, cmd *cli.Command) error {
	return runOrdersList(ctx, orderListOptions{
		Side:      cmd.String("side"),
		StockCode: cmd.String("stock-code"),
	}, cmd.Args().Slice())
}

type orderListOptions struct {
	Side      string
	StockCode string
}

func runOrdersList(ctx context.Context, opts orderListOptions, unusedArgs []string) error {
	if len(unusedArgs) > 0 {
		return fmt.Errorf("unexpected extra arguments: %v", unusedArgs)
	}

	request, err := buildOrderListRequest(opts)
	if err != nil {
		return err
	}

	creds, err := config.Load()
	if err != nil {
		return err
	}
	if creds.AppKey == "" || creds.SecretKey == "" {
		return fmt.Errorf("missing Kiwoom credentials: run 'stock config set' or set KIWOOM_APPKEY / KIWOOM_SECRETKEY")
	}

	client := kiwoom.NewClient(creds.AppKey, creds.SecretKey)
	rows, err := client.OrderList(ctx, request)
	if err != nil {
		return err
	}

	orders, err := normalizeOrderList(rows)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(orders)
}

func buildOrderListRequest(opts orderListOptions) (kiwoom.OrderListRequest, error) {
	tradeType, err := orderSideTradeType(opts.Side)
	if err != nil {
		return kiwoom.OrderListRequest{}, err
	}

	stockCode := strings.TrimSpace(opts.StockCode)
	allStockType := "0"
	if stockCode != "" {
		if !isSixDigitStockCode(stockCode) {
			return kiwoom.OrderListRequest{}, fmt.Errorf("invalid stock code %q: expected six digits", opts.StockCode)
		}
		allStockType = "1"
	}

	return kiwoom.OrderListRequest{
		AllStockType: allStockType,
		TradeType:    tradeType,
		StockCode:    stockCode,
		ExchangeType: "0",
	}, nil
}

func orderSideTradeType(side string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(side)) {
	case "", "all":
		return "0", nil
	case "sell":
		return "1", nil
	case "buy":
		return "2", nil
	default:
		return "", fmt.Errorf("invalid side %q: expected all, buy, or sell", side)
	}
}

func isSixDigitStockCode(stockCode string) bool {
	if len(stockCode) != 6 {
		return false
	}
	for _, r := range stockCode {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

type orderListItem struct {
	OrderID          string      `json:"order_id"`
	OriginalOrderID  string      `json:"original_order_id"`
	StockCode        string      `json:"stock_code"`
	StockName        string      `json:"stock_name"`
	TradingVenue     string      `json:"trading_venue"`
	OrderedQuantity  int64       `json:"ordered_quantity"`
	OrderedPrice     int64       `json:"ordered_price"`
	UnfilledQuantity int64       `json:"unfilled_quantity"`
	FundingType      FundingType `json:"funding_type"`
	FilledQuantity   int64       `json:"filled_quantity"`
	CurrentPrice     int64       `json:"current_price"`
}

func normalizeOrderList(rows []kiwoom.OrderListRow) ([]orderListItem, error) {
	orders := make([]orderListItem, 0, len(rows))
	for _, row := range rows {
		order, err := buildOrderListItem(row)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func buildOrderListItem(row kiwoom.OrderListRow) (orderListItem, error) {
	tradingVenue := tradingVenueFromKiwoomExchangeType(row.ExchangeType)
	orderedQuantity, err := parseKiwoomInt(row.OrderedQuantity, true)
	if err != nil {
		return orderListItem{}, fmt.Errorf("invalid ord_qty for order %s: %w", row.OrderID, err)
	}
	orderedPrice, err := parseKiwoomInt(row.OrderedPrice, true)
	if err != nil {
		return orderListItem{}, fmt.Errorf("invalid ord_pric for order %s: %w", row.OrderID, err)
	}
	unfilledQuantity, err := parseKiwoomInt(row.UnfilledQuantity, true)
	if err != nil {
		return orderListItem{}, fmt.Errorf("invalid oso_qty for order %s: %w", row.OrderID, err)
	}
	filledQuantity, err := parseKiwoomInt(row.FilledQuantity, true)
	if err != nil {
		return orderListItem{}, fmt.Errorf("invalid cntr_qty for order %s: %w", row.OrderID, err)
	}
	currentPrice, err := parseKiwoomInt(row.CurrentPrice, true)
	if err != nil {
		return orderListItem{}, fmt.Errorf("invalid cur_prc for order %s: %w", row.OrderID, err)
	}

	return orderListItem{
		OrderID:          row.OrderID,
		OriginalOrderID:  row.OriginalOrderID,
		StockCode:        row.StockCode,
		StockName:        row.StockName,
		TradingVenue:     tradingVenue,
		OrderedQuantity:  orderedQuantity,
		OrderedPrice:     orderedPrice,
		UnfilledQuantity: unfilledQuantity,
		FundingType:      provisionalFundingTypeFromKiwoomOrderKind(row.OrderKind),
		FilledQuantity:   filledQuantity,
		CurrentPrice:     currentPrice,
	}, nil
}

func tradingVenueFromKiwoomExchangeType(exchangeType string) string {
	trimmed := strings.TrimSpace(exchangeType)
	switch trimmed {
	case "0":
		return "SOR"
	case "1":
		return "KRX"
	case "2":
		return "NXT"
	case "":
		return "UNKNOWN"
	default:
		return "UNKNOWN_" + trimmed
	}
}

func provisionalFundingTypeFromKiwoomOrderKind(orderKind string) FundingType {
	// Kiwoom live rows were unavailable when this was added; docs mark this heuristic provisional.
	if strings.Contains(orderKind, "신용") {
		return FundingTypeCredit
	}
	return FundingTypeCash
}
