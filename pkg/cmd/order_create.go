package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"

	"stock-cli/pkg/config"
	"stock-cli/pkg/kiwoom"
)

var ordersCreateCmd = cli.Command{
	Name:    "create",
	Usage:   "Create Kiwoom orders",
	Suggest: true,
	Commands: []*cli.Command{
		&ordersCreateCashCmd,
	},
}

var ordersCreateCashCmd = cli.Command{
	Name:    "cash",
	Usage:   "Create a cash stock order",
	Suggest: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "side",
			Usage: "Order side: buy or sell",
		},
		&cli.StringFlag{
			Name:  "stock-code",
			Usage: "Six-digit stock code",
		},
		&cli.StringFlag{
			Name:  "order-type",
			Usage: "Order type: limit or market",
		},
		&cli.StringFlag{
			Name:  "quantity",
			Usage: "Positive whole-share order quantity",
		},
		&cli.StringFlag{
			Name:  "price",
			Usage: "Per-share limit price; invalid for market orders",
		},
		&cli.StringFlag{
			Name:  "trading-venue",
			Usage: "Trading venue: SOR, KRX, or NXT",
			Value: "SOR",
		},
	},
	Action:          handleOrdersCreateCash,
	HideHelpCommand: true,
}

func handleOrdersCreateCash(ctx context.Context, cmd *cli.Command) error {
	return runOrdersCreateCash(ctx, cashOrderOptions{
		Side:         cmd.String("side"),
		StockCode:    cmd.String("stock-code"),
		OrderType:    cmd.String("order-type"),
		Quantity:     cmd.String("quantity"),
		Price:        cmd.String("price"),
		PriceSet:     cmd.IsSet("price"),
		TradingVenue: cmd.String("trading-venue"),
	}, cmd.Args().Slice())
}

type cashOrderOptions struct {
	Side         string
	StockCode    string
	OrderType    string
	Quantity     string
	Price        string
	PriceSet     bool
	TradingVenue string
}

type cashOrderSide string

const (
	cashOrderSideBuy  cashOrderSide = "buy"
	cashOrderSideSell cashOrderSide = "sell"
)

func runOrdersCreateCash(ctx context.Context, opts cashOrderOptions, unusedArgs []string) error {
	request, side, err := buildCashOrderRequest(opts, unusedArgs)
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
	var response kiwoom.CashOrderResponse
	switch side {
	case cashOrderSideBuy:
		response, err = client.CashBuyOrder(ctx, request)
	case cashOrderSideSell:
		response, err = client.CashSellOrder(ctx, request)
	default:
		return fmt.Errorf("invalid side %q: expected buy or sell", side)
	}
	if err != nil {
		return err
	}

	output := normalizeCashOrderResponse(response, request.DomesticExchangeType)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func buildCashOrderRequest(opts cashOrderOptions, unusedArgs []string) (kiwoom.CashOrderRequest, cashOrderSide, error) {
	if len(unusedArgs) > 0 {
		return kiwoom.CashOrderRequest{}, "", fmt.Errorf("unexpected extra arguments: %v", unusedArgs)
	}

	side, err := parseCashOrderSide(opts.Side)
	if err != nil {
		return kiwoom.CashOrderRequest{}, "", err
	}
	stockCode, err := parseRequiredStockCode(opts.StockCode)
	if err != nil {
		return kiwoom.CashOrderRequest{}, "", err
	}
	quantity, err := parsePositiveIntString("quantity", opts.Quantity)
	if err != nil {
		return kiwoom.CashOrderRequest{}, "", err
	}
	tradingVenue, err := parseTradingVenue(opts.TradingVenue)
	if err != nil {
		return kiwoom.CashOrderRequest{}, "", err
	}
	tradeType, orderUnitPrice, err := parseCashOrderTypeAndPrice(opts.OrderType, opts.Price, opts.PriceSet)
	if err != nil {
		return kiwoom.CashOrderRequest{}, "", err
	}

	return kiwoom.CashOrderRequest{
		DomesticExchangeType: tradingVenue,
		StockCode:            stockCode,
		OrderQuantity:        quantity,
		OrderUnitPrice:       orderUnitPrice,
		TradeType:            tradeType,
		ConditionPrice:       "",
	}, side, nil
}

func parseCashOrderSide(side string) (cashOrderSide, error) {
	switch strings.ToLower(strings.TrimSpace(side)) {
	case "buy":
		return cashOrderSideBuy, nil
	case "sell":
		return cashOrderSideSell, nil
	default:
		return "", fmt.Errorf("invalid side %q: expected buy or sell", side)
	}
}

func parseRequiredStockCode(stockCode string) (string, error) {
	trimmed := strings.TrimSpace(stockCode)
	if !isSixDigitStockCode(trimmed) {
		return "", fmt.Errorf("invalid stock code %q: expected six digits", stockCode)
	}
	return trimmed, nil
}

func parseTradingVenue(tradingVenue string) (string, error) {
	trimmed := strings.ToUpper(strings.TrimSpace(tradingVenue))
	if trimmed == "" {
		trimmed = "SOR"
	}
	switch trimmed {
	case "SOR", "KRX", "NXT":
		return trimmed, nil
	default:
		return "", fmt.Errorf("invalid trading venue %q: expected SOR, KRX, or NXT", tradingVenue)
	}
}

func parseCashOrderTypeAndPrice(orderType string, price string, priceSet bool) (string, string, error) {
	trimmedPrice := strings.TrimSpace(price)
	switch strings.ToLower(strings.TrimSpace(orderType)) {
	case "limit":
		parsedPrice, err := parsePositiveIntString("price", trimmedPrice)
		if err != nil {
			return "", "", err
		}
		return "0", parsedPrice, nil
	case "market":
		if priceSet || trimmedPrice != "" {
			return "", "", fmt.Errorf("price is invalid for market orders")
		}
		return "3", "", nil
	default:
		return "", "", fmt.Errorf("invalid order type %q: expected limit or market", orderType)
	}
}

func parsePositiveIntString(name string, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("missing %s", name)
	}
	parsed, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil || parsed <= 0 {
		return "", fmt.Errorf("invalid %s %q: expected a positive integer", name, value)
	}
	return trimmed, nil
}

type cashOrderOutput struct {
	OrderID      string `json:"order_id"`
	TradingVenue string `json:"trading_venue"`
}

func normalizeCashOrderResponse(response kiwoom.CashOrderResponse, submittedTradingVenue string) cashOrderOutput {
	return cashOrderOutput{
		OrderID:      response.OrderID,
		TradingVenue: normalizeCashOrderTradingVenue(response.DomesticExchangeType, submittedTradingVenue),
	}
}

func normalizeCashOrderTradingVenue(responseTradingVenue string, submittedTradingVenue string) string {
	trimmed := strings.ToUpper(strings.TrimSpace(responseTradingVenue))
	switch trimmed {
	case "SOR", "KRX", "NXT":
		return trimmed
	case "":
		return submittedTradingVenue
	default:
		return "UNKNOWN_" + trimmed
	}
}
