package cmd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"stock-cli/pkg/kiwoom"
)

func TestNormalizeAccountHoldingsFromFixture(t *testing.T) {
	rows := loadAccountFixture(t)

	holdings, err := normalizeAccountHoldings(rows)
	require.NoError(t, err)
	require.Len(t, holdings, 3)

	for _, holding := range holdings {
		assert.NotEmpty(t, holding.StockCode)
		assert.NotEmpty(t, holding.StockName)
		assert.NotContains(t, holding.StockName, "*")
		assert.NotZero(t, holding.HoldingQuantity)
		if holding.PurchasePrice != 0 {
			require.NotNil(t, holding.ProfitRate)
		}
	}

	first := holdings[0]
	assert.Equal(t, "000001", first.StockCode)
	assert.Equal(t, "Synthetic Alpha", first.StockName)
	assert.Equal(t, int64(1200), first.CurrentPrice)
	assert.Equal(t, int64(1000), first.PurchasePrice)
	assert.Equal(t, percent(20.00), *first.ProfitRate)
	assert.Equal(t, int64(3000), first.PurchaseAmount)
	assert.Equal(t, int64(3), first.HoldingQuantity)
	assert.Equal(t, int64(2), first.OrderableQuantity)
	assert.False(t, first.IsCredit)

	var foundCredit bool
	var foundFreeLot bool
	for _, holding := range holdings {
		if holding.StockCode == "000003" {
			foundCredit = true
			assert.True(t, holding.IsCredit)
		}
		if holding.StockCode == "000004" {
			foundFreeLot = true
			assert.Nil(t, holding.ProfitRate)
		}
	}
	assert.True(t, foundCredit, "expected unstarred credit aggregate row in fixture")
	assert.True(t, foundFreeLot, "expected zero purchase price row in fixture")
}

func TestNormalizeAccountHoldingsProfitRateNullWhenPurchasePriceZero(t *testing.T) {
	holdings, err := normalizeAccountHoldings([]kiwoom.AccountProfitRateRow{
		{
			StockCode:         "000004",
			StockName:         "Synthetic Free Lot",
			CurrentPrice:      "+2945",
			PurchasePrice:     "0",
			PurchaseAmount:    "0",
			RemainingQuantity: "1",
			OrderableQuantity: "1",
			CreditType:        "00",
		},
	})
	require.NoError(t, err)
	require.Len(t, holdings, 1)
	assert.Nil(t, holdings[0].ProfitRate)
}

func TestNormalizeAccountHoldingsRoundsProfitRateToTwoDecimals(t *testing.T) {
	holdings, err := normalizeAccountHoldings([]kiwoom.AccountProfitRateRow{
		{
			StockCode:         "000005",
			StockName:         "Synthetic Rounded Rate",
			CurrentPrice:      "+1167",
			PurchasePrice:     "1000",
			PurchaseAmount:    "1000",
			RemainingQuantity: "1",
			OrderableQuantity: "1",
			CreditType:        "00",
		},
	})
	require.NoError(t, err)
	require.Len(t, holdings, 1)
	require.NotNil(t, holdings[0].ProfitRate)
	assert.Equal(t, percent(16.70), *holdings[0].ProfitRate)

	data, err := json.Marshal(holdings)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"profit_rate":16.70`)
}

func TestNormalizeAccountHoldingsDropsStarredRowsEvenWithQuantity(t *testing.T) {
	holdings, err := normalizeAccountHoldings([]kiwoom.AccountProfitRateRow{
		{StockCode: "000003", StockName: "*Synthetic Credit Detail", CurrentPrice: "-2100", PurchasePrice: "2000", PurchaseAmount: "10000", RemainingQuantity: "5", OrderableQuantity: "5", CreditType: "03"},
		{StockCode: "000003", StockName: "Synthetic Credit Aggregate", CurrentPrice: "-2100", PurchasePrice: "2000", PurchaseAmount: "10000", RemainingQuantity: "5", OrderableQuantity: "4", CreditType: "99"},
	})
	require.NoError(t, err)
	require.Len(t, holdings, 1)
	assert.Equal(t, "Synthetic Credit Aggregate", holdings[0].StockName)
}

func TestNormalizeAccountHoldingsJSONSchema(t *testing.T) {
	holdings, err := normalizeAccountHoldings(loadAccountFixture(t))
	require.NoError(t, err)

	data, err := json.Marshal(holdings)
	require.NoError(t, err)

	var decoded []map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))
	expected := map[string]struct{}{
		"stock_code":         {},
		"stock_name":         {},
		"current_price":      {},
		"purchase_price":     {},
		"profit_rate":        {},
		"purchase_amount":    {},
		"holding_quantity":   {},
		"orderable_quantity": {},
		"is_credit":          {},
	}
	for _, row := range decoded {
		assert.Equal(t, expected, keySet(row))
	}
}

func TestRootCommandContainsAccounts(t *testing.T) {
	var found bool
	for _, child := range Command.Commands {
		if child.Name == "accounts" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestAccountsCommandContainsList(t *testing.T) {
	var found bool
	for _, child := range accountsCmd.Commands {
		if child.Name == "list" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestRunAccountsListMissingCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KIWOOM_APPKEY", "")
	t.Setenv("KIWOOM_SECRETKEY", "")

	err := runAccountsList(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stock config set")
	assert.Contains(t, err.Error(), "KIWOOM_APPKEY")
	assert.Contains(t, err.Error(), "KIWOOM_SECRETKEY")
}

func TestAccountsListRejectsUnexpectedArgs(t *testing.T) {
	cmd := &cli.Command{}
	cmd.Commands = []*cli.Command{&accountsCmd}
	err := cmd.Run(context.Background(), []string{"stock", "accounts", "list", "extra"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected extra arguments")
}

func loadAccountFixture(t *testing.T) []kiwoom.AccountProfitRateRow {
	t.Helper()
	path := filepath.Join("..", "..", "docs", "kiwoom-ka10085-response-raw.json")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var response kiwoom.AccountProfitRateResponse
	require.NoError(t, json.Unmarshal(data, &response))
	return response.Rows
}

func keySet(row map[string]any) map[string]struct{} {
	keys := make(map[string]struct{}, len(row))
	for key := range row {
		keys[key] = struct{}{}
	}
	return keys
}
