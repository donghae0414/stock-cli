package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"
)

const maxMarketTickPrice = int64(1<<63 - 1 - 999)

var marketCmd = cli.Command{
	Name:     "market",
	Usage:    "Use local market rule helpers",
	Category: "LOCAL HELPER",
	Suggest:  true,
	Commands: []*cli.Command{
		&marketTickCmd,
	},
}

var marketTickCmd = cli.Command{
	Name:    "tick",
	Usage:   "Calculate Korean stock tick prices",
	Suggest: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "price",
			Usage: "Positive whole-won stock price",
		},
	},
	Action:          handleMarketTick,
	HideHelpCommand: true,
}

type marketTickOptions struct {
	Price string
}

type marketTickOutput struct {
	Price       int64 `json:"price"`
	TickSize    int64 `json:"tick_size"`
	LowerPrice  int64 `json:"lower_price"`
	UpperPrice  int64 `json:"upper_price"`
	IsValidTick bool  `json:"is_valid_tick"`
}

func handleMarketTick(ctx context.Context, cmd *cli.Command) error {
	return runMarketTick(ctx, marketTickOptions{
		Price: cmd.String("price"),
	}, cmd.Args().Slice())
}

func runMarketTick(_ context.Context, opts marketTickOptions, unusedArgs []string) error {
	if len(unusedArgs) > 0 {
		return fmt.Errorf("unexpected extra arguments: %v", unusedArgs)
	}
	price, err := parseMarketTickPrice(opts.Price)
	if err != nil {
		return err
	}
	output := buildMarketTickOutput(price)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func parseMarketTickPrice(price string) (int64, error) {
	trimmed := strings.TrimSpace(price)
	if trimmed == "" {
		return 0, fmt.Errorf("missing price")
	}
	parsed, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("invalid price %q: expected a positive whole-won integer", price)
	}
	if parsed > maxMarketTickPrice {
		return 0, fmt.Errorf("invalid price %q: expected a value no greater than %d", price, maxMarketTickPrice)
	}
	return parsed, nil
}

func buildMarketTickOutput(price int64) marketTickOutput {
	tickSize := marketTickSize(price)
	lowerPrice := (price / tickSize) * tickSize
	upperPrice := lowerPrice
	if price%tickSize != 0 {
		upperPrice += tickSize
	}
	return marketTickOutput{
		Price:       price,
		TickSize:    tickSize,
		LowerPrice:  lowerPrice,
		UpperPrice:  upperPrice,
		IsValidTick: lowerPrice == price && upperPrice == price,
	}
}

func marketTickSize(price int64) int64 {
	switch {
	case price < 2000:
		return 1
	case price < 5000:
		return 5
	case price < 20000:
		return 10
	case price < 50000:
		return 50
	case price < 200000:
		return 100
	case price < 500000:
		return 500
	default:
		return 1000
	}
}
