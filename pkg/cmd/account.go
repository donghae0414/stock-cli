package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"

	"stock-cli/pkg/config"
	"stock-cli/pkg/kiwoom"
)

var accountsCmd = cli.Command{
	Name:     "accounts",
	Usage:    "Manage Kiwoom account resources",
	Category: "API RESOURCE",
	Suggest:  true,
	Commands: []*cli.Command{
		&accountsListCmd,
	},
}

var accountsListCmd = cli.Command{
	Name:            "list",
	Usage:           "Get current account holdings",
	Suggest:         true,
	Action:          handleAccountsList,
	HideHelpCommand: true,
}

func handleAccountsList(ctx context.Context, cmd *cli.Command) error {
	return runAccountsList(ctx, cmd.Args().Slice())
}

func runAccountsList(ctx context.Context, unusedArgs []string) error {
	if len(unusedArgs) > 0 {
		return fmt.Errorf("unexpected extra arguments: %v", unusedArgs)
	}

	creds, err := config.Load()
	if err != nil {
		return err
	}
	if creds.AppKey == "" || creds.SecretKey == "" {
		return fmt.Errorf("missing Kiwoom credentials: run 'stock config set' or set KIWOOM_APPKEY / KIWOOM_SECRETKEY")
	}

	client := kiwoom.NewClient(creds.AppKey, creds.SecretKey)
	response, err := client.AccountProfitRates(ctx)
	if err != nil {
		return err
	}

	holdings, err := normalizeAccountHoldings(response.Rows)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(holdings)
}

type accountHolding struct {
	StockCode         string   `json:"stock_code"`
	StockName         string   `json:"stock_name"`
	CurrentPrice      int64    `json:"current_price"`
	PurchasePrice     int64    `json:"purchase_price"`
	ProfitRate        *percent `json:"profit_rate"`
	PurchaseAmount    int64    `json:"purchase_amount"`
	HoldingQuantity   int64    `json:"holding_quantity"`
	OrderableQuantity int64    `json:"orderable_quantity"`
	IsCredit          bool     `json:"is_credit"`
}

type percent float64

func (p percent) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(p), 'f', 2, 64)), nil
}

func normalizeAccountHoldings(rows []kiwoom.AccountProfitRateRow) ([]accountHolding, error) {
	holdings := make([]accountHolding, 0, len(rows))
	for _, row := range rows {
		if strings.HasPrefix(row.StockName, "*") {
			continue
		}

		holdingQuantity, err := parseKiwoomInt(row.RemainingQuantity, false)
		if err != nil {
			return nil, fmt.Errorf("invalid rmnd_qty for %s: %w", row.StockCode, err)
		}
		if holdingQuantity == 0 {
			continue
		}

		currentPrice, err := parseKiwoomInt(row.CurrentPrice, true)
		if err != nil {
			return nil, fmt.Errorf("invalid cur_prc for %s: %w", row.StockCode, err)
		}
		purchasePrice, err := parseKiwoomInt(row.PurchasePrice, true)
		if err != nil {
			return nil, fmt.Errorf("invalid pur_pric for %s: %w", row.StockCode, err)
		}
		purchaseAmount, err := parseKiwoomInt(row.PurchaseAmount, false)
		if err != nil {
			return nil, fmt.Errorf("invalid pur_amt for %s: %w", row.StockCode, err)
		}
		orderableQuantity, err := parseKiwoomInt(row.OrderableQuantity, false)
		if err != nil {
			return nil, fmt.Errorf("invalid clrn_alow_qty for %s: %w", row.StockCode, err)
		}

		var profitRate *percent
		if purchasePrice != 0 {
			rawRate := (float64(currentPrice) - float64(purchasePrice)) / float64(purchasePrice) * 100
			rate := percent(math.Round(rawRate*100) / 100)
			profitRate = &rate
		}

		holdings = append(holdings, accountHolding{
			StockCode:         row.StockCode,
			StockName:         row.StockName,
			CurrentPrice:      currentPrice,
			PurchasePrice:     purchasePrice,
			ProfitRate:        profitRate,
			PurchaseAmount:    purchaseAmount,
			HoldingQuantity:   holdingQuantity,
			OrderableQuantity: orderableQuantity,
			IsCredit:          row.CreditType != "00",
		})
	}
	return holdings, nil
}

func parseKiwoomInt(value string, absolute bool) (int64, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return 0, nil
	}
	if strings.HasPrefix(normalized, "+") {
		normalized = strings.TrimPrefix(normalized, "+")
	} else if absolute && strings.HasPrefix(normalized, "-") {
		normalized = strings.TrimPrefix(normalized, "-")
	}
	return strconv.ParseInt(normalized, 10, 64)
}
