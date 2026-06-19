package cmd

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"stock-cli/pkg/kiwoom"
)

func TestRootCommandContainsChart(t *testing.T) {
	var found bool
	for _, child := range Command.Commands {
		if child.Name == "chart" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestChartCommandContainsSubcommands(t *testing.T) {
	found := map[string]bool{}
	for _, child := range chartCmd.Commands {
		found[child.Name] = true
	}
	assert.True(t, found["day"])
	assert.True(t, found["week"])
	assert.True(t, found["minute"])
}

func TestParseDateChartOptions(t *testing.T) {
	oldNow := chartNow
	defer func() { chartNow = oldNow }()
	chartNow = func() time.Time { return time.Date(2026, 6, 18, 13, 20, 0, 0, time.Local) }

	stockCode, count, to, err := parseDateChartOptions(chartOptions{StockCode: "005930", Count: defaultChartCount}, nil)
	require.NoError(t, err)
	assert.Equal(t, "005930", stockCode)
	assert.Equal(t, 120, count)
	assert.Equal(t, "20260618", to)

	_, _, _, err = parseDateChartOptions(chartOptions{StockCode: "005930", Count: 601, To: "20260618"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid count")

	_, _, _, err = parseDateChartOptions(chartOptions{StockCode: "005930", Count: 1, To: "20260618132000"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected YYYYMMDD")
}

func TestParseMinuteChartOptions(t *testing.T) {
	stockCode, count, to, baseDate, interval, err := parseMinuteChartOptions(minuteChartOptions{
		chartOptions: chartOptions{StockCode: "005930", Count: 2, To: "20260618132000"},
		Interval:     1,
		IntervalSet:  true,
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, "005930", stockCode)
	assert.Equal(t, 2, count)
	assert.Equal(t, "20260618132000", to)
	assert.Equal(t, "20260618", baseDate)
	assert.Equal(t, 1, interval)

	_, _, _, _, _, err = parseMinuteChartOptions(minuteChartOptions{
		chartOptions: chartOptions{StockCode: "005930", Count: 2, To: "20260618"},
	}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing interval")

	_, _, _, _, _, err = parseMinuteChartOptions(minuteChartOptions{
		chartOptions: chartOptions{StockCode: "005930", Count: 2, To: "20260618"},
		Interval:     2,
		IntervalSet:  true,
	}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid interval")
}

func TestNormalizeDayChartOutputUsesSchemaAndAbsoluteNumbers(t *testing.T) {
	output, err := normalizeDayChartOutput("005930", "20260618", 1, []kiwoom.StockDayChartRow{
		{Date: "20260618", CurrentPrice: "-70100", OpenPrice: "+69800", HighPrice: "-70500", LowPrice: "-69600", TradeAmount: "-648525"},
		{Date: "20260617", CurrentPrice: "69000", OpenPrice: "68000", HighPrice: "70000", LowPrice: "67000", TradeAmount: "100"},
	})
	require.NoError(t, err)

	data, err := json.Marshal(output)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, map[string]struct{}{
		"stock_code": {},
		"chart":      {},
		"to":         {},
		"count":      {},
		"candles":    {},
	}, keySet(decoded))
	assert.Equal(t, "005930", decoded["stock_code"])
	assert.Equal(t, "day", decoded["chart"])
	candles := decoded["candles"].([]any)
	require.Len(t, candles, 1)
	first := candles[0].(map[string]any)
	assert.Equal(t, map[string]struct{}{
		"date":         {},
		"close_price":  {},
		"open_price":   {},
		"high_price":   {},
		"low_price":    {},
		"trade_amount": {},
	}, keySet(first))
	assert.Equal(t, float64(70100), first["close_price"])
	assert.Equal(t, float64(648525), first["trade_amount"])
}

func TestNormalizeMinuteChartOutputUsesSchemaAndPreservesTimestampTo(t *testing.T) {
	output, err := normalizeMinuteChartOutput("005930", "20260618132000", 1, 1, []kiwoom.StockMinuteChartRow{
		{Timestamp: "20260618132000", CurrentPrice: "-78800", OpenPrice: "-78850", HighPrice: "-78900", LowPrice: "-78800"},
	})
	require.NoError(t, err)

	data, err := json.Marshal(output)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, map[string]struct{}{
		"stock_code": {},
		"chart":      {},
		"to":         {},
		"count":      {},
		"candles":    {},
		"interval":   {},
	}, keySet(decoded))
	assert.Equal(t, "20260618132000", decoded["to"])
	assert.Equal(t, float64(1), decoded["interval"])

	candles := decoded["candles"].([]any)
	require.Len(t, candles, 1)
	first := candles[0].(map[string]any)
	assert.Equal(t, map[string]struct{}{
		"timestamp":   {},
		"close_price": {},
		"open_price":  {},
		"high_price":  {},
		"low_price":   {},
	}, keySet(first))
	assert.Equal(t, float64(78800), first["close_price"])
}

func TestChartValidationRejectsBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KIWOOM_APPKEY", "")
	t.Setenv("KIWOOM_SECRETKEY", "")

	err := runChartMinute(context.Background(), minuteChartOptions{
		chartOptions: chartOptions{StockCode: "005930", Count: 1, To: "20260618"},
	}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing interval")
	assert.NotContains(t, err.Error(), "stock config set")
}

func TestChartCommandsRejectUnexpectedArgs(t *testing.T) {
	cmd := &cli.Command{}
	cmd.Commands = []*cli.Command{&chartCmd}
	err := cmd.Run(context.Background(), []string{"stock", "chart", "day", "--stock-code", "005930", "extra"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected extra arguments")
}
