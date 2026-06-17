package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"stock-cli/pkg/config"
	"stock-cli/pkg/kiwoom"
)

var ordersCancelCmd = cli.Command{
	Name:    "cancel",
	Usage:   "Cancel Kiwoom orders",
	Suggest: true,
	Commands: []*cli.Command{
		&ordersCancelCashCmd,
	},
}

var ordersCancelCashCmd = cli.Command{
	Name:    "cash",
	Usage:   "Cancel a cash stock order",
	Suggest: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "stock-code",
			Usage: "Six-digit stock code",
		},
		&cli.StringFlag{
			Name:  "original-order-id",
			Usage: "Original order identifier",
		},
		&cli.StringFlag{
			Name:  "quantity",
			Usage: "Positive cancel quantity; omit to cancel remaining quantity",
		},
		&cli.StringFlag{
			Name:  "trading-venue",
			Usage: "Trading venue: SOR, KRX, or NXT",
			Value: "SOR",
		},
	},
	Action:          handleOrdersCancelCash,
	HideHelpCommand: true,
}

func handleOrdersCancelCash(ctx context.Context, cmd *cli.Command) error {
	return runOrdersCancelCash(ctx, cashCancelOptions{
		StockCode:       cmd.String("stock-code"),
		OriginalOrderID: cmd.String("original-order-id"),
		Quantity:        cmd.String("quantity"),
		TradingVenue:    cmd.String("trading-venue"),
	}, cmd.Args().Slice())
}

type cashCancelOptions struct {
	StockCode       string
	OriginalOrderID string
	Quantity        string
	TradingVenue    string
}

func runOrdersCancelCash(ctx context.Context, opts cashCancelOptions, unusedArgs []string) error {
	request, err := buildCashCancelRequest(opts, unusedArgs)
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
	response, err := client.CashCancelOrder(ctx, request)
	if err != nil {
		return err
	}

	output, err := normalizeCashCancelResponse(response)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func buildCashCancelRequest(opts cashCancelOptions, unusedArgs []string) (kiwoom.CashCancelRequest, error) {
	if len(unusedArgs) > 0 {
		return kiwoom.CashCancelRequest{}, fmt.Errorf("unexpected extra arguments: %v", unusedArgs)
	}

	stockCode, err := parseRequiredStockCode(opts.StockCode)
	if err != nil {
		return kiwoom.CashCancelRequest{}, err
	}
	originalOrderID, err := parseOriginalOrderID(opts.OriginalOrderID)
	if err != nil {
		return kiwoom.CashCancelRequest{}, err
	}
	cancelQuantity, err := parseCancelQuantity(opts.Quantity)
	if err != nil {
		return kiwoom.CashCancelRequest{}, err
	}
	tradingVenue, err := parseTradingVenue(opts.TradingVenue)
	if err != nil {
		return kiwoom.CashCancelRequest{}, err
	}

	return kiwoom.CashCancelRequest{
		DomesticExchangeType: tradingVenue,
		OriginalOrderID:      originalOrderID,
		StockCode:            stockCode,
		CancelQuantity:       cancelQuantity,
	}, nil
}

func parseOriginalOrderID(originalOrderID string) (string, error) {
	if !isDigitsOnly(originalOrderID) {
		return "", fmt.Errorf("invalid original order id %q: expected digits", originalOrderID)
	}
	return originalOrderID, nil
}

func parseCancelQuantity(quantity string) (string, error) {
	trimmed := strings.TrimSpace(quantity)
	if trimmed == "" {
		return "0", nil
	}
	return parsePositiveIntString("quantity", trimmed)
}

type cashCancelOutput struct {
	OrderID             string `json:"order_id"`
	BaseOriginalOrderID string `json:"base_original_order_id"`
	CancelledQuantity   int64  `json:"cancelled_quantity"`
}

func normalizeCashCancelResponse(response kiwoom.CashCancelResponse) (cashCancelOutput, error) {
	cancelledQuantity, err := parseKiwoomInt(response.CancelQuantity, true)
	if err != nil {
		return cashCancelOutput{}, fmt.Errorf("invalid cncl_qty for order %s: %w", response.OrderID, err)
	}
	return cashCancelOutput{
		OrderID:             response.OrderID,
		BaseOriginalOrderID: response.BaseOriginalOrderID,
		CancelledQuantity:   cancelledQuantity,
	}, nil
}
