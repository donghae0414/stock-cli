package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"stock-cli/pkg/config"
	"stock-cli/pkg/kiwoom"
)

const (
	defaultChartCount = 120
	maxChartCount     = 600
)

var chartNow = time.Now

var chartCmd = cli.Command{
	Name:     "chart",
	Usage:    "Get Kiwoom stock chart candles",
	Category: "API RESOURCE",
	Suggest:  true,
	Commands: []*cli.Command{
		&chartDayCmd,
		&chartWeekCmd,
		&chartMinuteCmd,
	},
}

var chartDayCmd = cli.Command{
	Name:    "day",
	Usage:   "Get daily stock chart candles",
	Suggest: true,
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "stock-code", Usage: "Six-digit stock code"},
		&cli.IntFlag{Name: "count", Usage: "Number of candles to return, 1 through 600", Value: defaultChartCount},
		&cli.StringFlag{Name: "to", Usage: "End date in YYYYMMDD; defaults to current date"},
	},
	Action:          handleChartDay,
	HideHelpCommand: true,
}

var chartWeekCmd = cli.Command{
	Name:    "week",
	Usage:   "Get weekly stock chart candles",
	Suggest: true,
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "stock-code", Usage: "Six-digit stock code"},
		&cli.IntFlag{Name: "count", Usage: "Number of candles to return, 1 through 600", Value: defaultChartCount},
		&cli.StringFlag{Name: "to", Usage: "End date in YYYYMMDD; defaults to current date"},
	},
	Action:          handleChartWeek,
	HideHelpCommand: true,
}

var chartMinuteCmd = cli.Command{
	Name:    "minute",
	Usage:   "Get minute stock chart candles",
	Suggest: true,
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "stock-code", Usage: "Six-digit stock code"},
		&cli.IntFlag{Name: "interval", Usage: "Minute candle interval: 1, 3, 5, 10, 15, 30, 45, or 60"},
		&cli.IntFlag{Name: "count", Usage: "Number of candles to return, 1 through 600", Value: defaultChartCount},
		&cli.StringFlag{Name: "to", Usage: "End time in YYYYMMDD or YYYYMMDDHHmmss; defaults to current date"},
	},
	Action:          handleChartMinute,
	HideHelpCommand: true,
}

type chartOptions struct {
	StockCode string
	Count     int
	To        string
}

type minuteChartOptions struct {
	chartOptions
	Interval    int
	IntervalSet bool
}

type dateChartOutput struct {
	StockCode string            `json:"stock_code"`
	Chart     string            `json:"chart"`
	To        string            `json:"to"`
	Count     int               `json:"count"`
	Candles   []dateChartCandle `json:"candles"`
}

type minuteChartOutput struct {
	StockCode string              `json:"stock_code"`
	Chart     string              `json:"chart"`
	To        string              `json:"to"`
	Count     int                 `json:"count"`
	Candles   []minuteChartCandle `json:"candles"`
	Interval  int                 `json:"interval"`
}

type dateChartCandle struct {
	Date        string `json:"date"`
	ClosePrice  int64  `json:"close_price"`
	OpenPrice   int64  `json:"open_price"`
	HighPrice   int64  `json:"high_price"`
	LowPrice    int64  `json:"low_price"`
	TradeAmount int64  `json:"trade_amount"`
}

type minuteChartCandle struct {
	Timestamp  string `json:"timestamp"`
	ClosePrice int64  `json:"close_price"`
	OpenPrice  int64  `json:"open_price"`
	HighPrice  int64  `json:"high_price"`
	LowPrice   int64  `json:"low_price"`
}

func handleChartDay(ctx context.Context, cmd *cli.Command) error {
	return runChartDay(ctx, chartOptions{
		StockCode: cmd.String("stock-code"),
		Count:     cmd.Int("count"),
		To:        cmd.String("to"),
	}, cmd.Args().Slice())
}

func handleChartWeek(ctx context.Context, cmd *cli.Command) error {
	return runChartWeek(ctx, chartOptions{
		StockCode: cmd.String("stock-code"),
		Count:     cmd.Int("count"),
		To:        cmd.String("to"),
	}, cmd.Args().Slice())
}

func handleChartMinute(ctx context.Context, cmd *cli.Command) error {
	return runChartMinute(ctx, minuteChartOptions{
		chartOptions: chartOptions{
			StockCode: cmd.String("stock-code"),
			Count:     cmd.Int("count"),
			To:        cmd.String("to"),
		},
		Interval:    cmd.Int("interval"),
		IntervalSet: cmd.IsSet("interval"),
	}, cmd.Args().Slice())
}

func runChartDay(ctx context.Context, opts chartOptions, unusedArgs []string) error {
	stockCode, count, to, err := parseDateChartOptions(opts, unusedArgs)
	if err != nil {
		return err
	}

	creds, err := config.Load()
	if err != nil {
		return err
	}
	if creds.AppKey == "" || creds.SecretKey == "" {
		return fmt.Errorf(config.MissingCredentialsMessage)
	}

	client := kiwoom.NewClient(creds.AppKey, creds.SecretKey)
	response, err := client.StockDayChart(ctx, kiwoom.StockDayChartRequest{
		StockCode:             stockCode + "_AL",
		BaseDate:              to,
		UpdatedStockPriceType: "1",
	})
	if err != nil {
		return err
	}

	output, err := normalizeDayChartOutput(stockCode, to, count, response.Rows)
	if err != nil {
		return err
	}
	return encodeChartOutput(output)
}

func runChartWeek(ctx context.Context, opts chartOptions, unusedArgs []string) error {
	stockCode, count, to, err := parseDateChartOptions(opts, unusedArgs)
	if err != nil {
		return err
	}

	creds, err := config.Load()
	if err != nil {
		return err
	}
	if creds.AppKey == "" || creds.SecretKey == "" {
		return fmt.Errorf(config.MissingCredentialsMessage)
	}

	client := kiwoom.NewClient(creds.AppKey, creds.SecretKey)
	response, err := client.StockWeekChart(ctx, kiwoom.StockWeekChartRequest{
		StockCode:             stockCode + "_AL",
		BaseDate:              to,
		UpdatedStockPriceType: "1",
	})
	if err != nil {
		return err
	}

	output, err := normalizeWeekChartOutput(stockCode, to, count, response.Rows)
	if err != nil {
		return err
	}
	return encodeChartOutput(output)
}

func runChartMinute(ctx context.Context, opts minuteChartOptions, unusedArgs []string) error {
	stockCode, count, to, baseDate, interval, err := parseMinuteChartOptions(opts, unusedArgs)
	if err != nil {
		return err
	}

	creds, err := config.Load()
	if err != nil {
		return err
	}
	if creds.AppKey == "" || creds.SecretKey == "" {
		return fmt.Errorf(config.MissingCredentialsMessage)
	}

	client := kiwoom.NewClient(creds.AppKey, creds.SecretKey)
	response, err := client.StockMinuteChart(ctx, kiwoom.StockMinuteChartRequest{
		StockCode:             stockCode + "_AL",
		TickScope:             strconv.Itoa(interval),
		UpdatedStockPriceType: "1",
		BaseDate:              baseDate,
	})
	if err != nil {
		return err
	}

	output, err := normalizeMinuteChartOutput(stockCode, to, count, interval, response.Rows)
	if err != nil {
		return err
	}
	return encodeChartOutput(output)
}

func parseDateChartOptions(opts chartOptions, unusedArgs []string) (string, int, string, error) {
	if len(unusedArgs) > 0 {
		return "", 0, "", fmt.Errorf("unexpected extra arguments: %v", unusedArgs)
	}
	stockCode, err := parseRequiredStockCode(opts.StockCode)
	if err != nil {
		return "", 0, "", err
	}
	count, err := parseChartCount(opts.Count)
	if err != nil {
		return "", 0, "", err
	}
	to, err := parseChartDateOrDefault(opts.To)
	if err != nil {
		return "", 0, "", err
	}
	return stockCode, count, to, nil
}

func parseMinuteChartOptions(opts minuteChartOptions, unusedArgs []string) (string, int, string, string, int, error) {
	if len(unusedArgs) > 0 {
		return "", 0, "", "", 0, fmt.Errorf("unexpected extra arguments: %v", unusedArgs)
	}
	stockCode, err := parseRequiredStockCode(opts.StockCode)
	if err != nil {
		return "", 0, "", "", 0, err
	}
	count, err := parseChartCount(opts.Count)
	if err != nil {
		return "", 0, "", "", 0, err
	}
	interval, err := parseMinuteInterval(opts.Interval, opts.IntervalSet)
	if err != nil {
		return "", 0, "", "", 0, err
	}
	to, baseDate, err := parseMinuteToOrDefault(opts.To)
	if err != nil {
		return "", 0, "", "", 0, err
	}
	return stockCode, count, to, baseDate, interval, nil
}

func parseChartCount(count int) (int, error) {
	if count < 1 || count > maxChartCount {
		return 0, fmt.Errorf("invalid count %d: expected 1 through 600", count)
	}
	return count, nil
}

func parseChartDateOrDefault(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return chartNow().Format("20060102"), nil
	}
	if !isChartDate(trimmed) {
		return "", fmt.Errorf("invalid to %q: expected YYYYMMDD", value)
	}
	return trimmed, nil
}

func parseMinuteToOrDefault(value string) (string, string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		to := chartNow().Format("20060102")
		return to, to, nil
	}
	switch {
	case isChartDate(trimmed):
		return trimmed, trimmed, nil
	case len(trimmed) == 14 && isDigitsOnly(trimmed):
		return trimmed, trimmed[:8], nil
	default:
		return "", "", fmt.Errorf("invalid to %q: expected YYYYMMDD or YYYYMMDDHHmmss", value)
	}
}

func isChartDate(value string) bool {
	return len(value) == 8 && isDigitsOnly(value)
}

func parseMinuteInterval(interval int, intervalSet bool) (int, error) {
	if !intervalSet {
		return 0, fmt.Errorf("missing interval")
	}
	switch interval {
	case 1, 3, 5, 10, 15, 30, 45, 60:
		return interval, nil
	default:
		return 0, fmt.Errorf("invalid interval %d: expected 1, 3, 5, 10, 15, 30, 45, or 60", interval)
	}
}

func normalizeDayChartOutput(stockCode string, to string, count int, rows []kiwoom.StockDayChartRow) (dateChartOutput, error) {
	candles := make([]dateChartCandle, 0, min(count, len(rows)))
	for i, row := range rows {
		if i >= count {
			break
		}
		candle, err := buildDateChartCandle(row.Date, row.CurrentPrice, row.OpenPrice, row.HighPrice, row.LowPrice, row.TradeAmount)
		if err != nil {
			return dateChartOutput{}, err
		}
		candles = append(candles, candle)
	}
	return dateChartOutput{StockCode: stockCode, Chart: "day", To: to, Count: count, Candles: candles}, nil
}

func normalizeWeekChartOutput(stockCode string, to string, count int, rows []kiwoom.StockWeekChartRow) (dateChartOutput, error) {
	candles := make([]dateChartCandle, 0, min(count, len(rows)))
	for i, row := range rows {
		if i >= count {
			break
		}
		candle, err := buildDateChartCandle(row.Date, row.CurrentPrice, row.OpenPrice, row.HighPrice, row.LowPrice, row.TradeAmount)
		if err != nil {
			return dateChartOutput{}, err
		}
		candles = append(candles, candle)
	}
	return dateChartOutput{StockCode: stockCode, Chart: "week", To: to, Count: count, Candles: candles}, nil
}

func normalizeMinuteChartOutput(stockCode string, to string, count int, interval int, rows []kiwoom.StockMinuteChartRow) (minuteChartOutput, error) {
	candles := make([]minuteChartCandle, 0, min(count, len(rows)))
	for i, row := range rows {
		if i >= count {
			break
		}
		candle, err := buildMinuteChartCandle(row)
		if err != nil {
			return minuteChartOutput{}, err
		}
		candles = append(candles, candle)
	}
	return minuteChartOutput{StockCode: stockCode, Chart: "minute", To: to, Count: count, Candles: candles, Interval: interval}, nil
}

func buildDateChartCandle(date string, closeValue string, openValue string, highValue string, lowValue string, tradeAmountValue string) (dateChartCandle, error) {
	closePrice, err := parseKiwoomInt(closeValue, true)
	if err != nil {
		return dateChartCandle{}, fmt.Errorf("invalid cur_prc for %s: %w", date, err)
	}
	openPrice, err := parseKiwoomInt(openValue, true)
	if err != nil {
		return dateChartCandle{}, fmt.Errorf("invalid open_pric for %s: %w", date, err)
	}
	highPrice, err := parseKiwoomInt(highValue, true)
	if err != nil {
		return dateChartCandle{}, fmt.Errorf("invalid high_pric for %s: %w", date, err)
	}
	lowPrice, err := parseKiwoomInt(lowValue, true)
	if err != nil {
		return dateChartCandle{}, fmt.Errorf("invalid low_pric for %s: %w", date, err)
	}
	tradeAmount, err := parseKiwoomInt(tradeAmountValue, true)
	if err != nil {
		return dateChartCandle{}, fmt.Errorf("invalid trde_prica for %s: %w", date, err)
	}
	return dateChartCandle{Date: date, ClosePrice: closePrice, OpenPrice: openPrice, HighPrice: highPrice, LowPrice: lowPrice, TradeAmount: tradeAmount}, nil
}

func buildMinuteChartCandle(row kiwoom.StockMinuteChartRow) (minuteChartCandle, error) {
	closePrice, err := parseKiwoomInt(row.CurrentPrice, true)
	if err != nil {
		return minuteChartCandle{}, fmt.Errorf("invalid cur_prc for %s: %w", row.Timestamp, err)
	}
	openPrice, err := parseKiwoomInt(row.OpenPrice, true)
	if err != nil {
		return minuteChartCandle{}, fmt.Errorf("invalid open_pric for %s: %w", row.Timestamp, err)
	}
	highPrice, err := parseKiwoomInt(row.HighPrice, true)
	if err != nil {
		return minuteChartCandle{}, fmt.Errorf("invalid high_pric for %s: %w", row.Timestamp, err)
	}
	lowPrice, err := parseKiwoomInt(row.LowPrice, true)
	if err != nil {
		return minuteChartCandle{}, fmt.Errorf("invalid low_pric for %s: %w", row.Timestamp, err)
	}
	return minuteChartCandle{Timestamp: row.Timestamp, ClosePrice: closePrice, OpenPrice: openPrice, HighPrice: highPrice, LowPrice: lowPrice}, nil
}

func encodeChartOutput(output any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}
