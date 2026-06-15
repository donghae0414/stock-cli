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
	assert.Equal(t, FundingTypeCash, first.FundingType)

	var foundCredit bool
	var foundFreeLot bool
	for _, holding := range holdings {
		if holding.StockCode == "000003" {
			foundCredit = true
			assert.Equal(t, FundingTypeCredit, holding.FundingType)
		}
		if holding.StockCode == "000004" {
			foundFreeLot = true
			assert.Nil(t, holding.ProfitRate)
		}
	}
	assert.True(t, foundCredit, "expected unstarred credit aggregate row in fixture")
	assert.True(t, foundFreeLot, "expected zero purchase price row in fixture")
}

func TestFundingTypeFromKiwoomCreditType(t *testing.T) {
	tests := []struct {
		name       string
		creditType string
		want       FundingType
	}{
		{name: "cash", creditType: "00", want: FundingTypeCash},
		{name: "credit buy", creditType: "03", want: FundingTypeCredit},
		{name: "credit aggregate", creditType: "99", want: FundingTypeCredit},
		{name: "empty unknown", creditType: "", want: FundingTypeCredit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, fundingTypeFromKiwoomCreditType(tt.creditType))
		})
	}
}

func TestAccountProfitRateRowUnmarshalsLoanDate(t *testing.T) {
	rows := loadAccountFixture(t)

	var found bool
	for _, row := range rows {
		if row.StockCode == "000003" && row.StockName == "*Synthetic Credit Detail" {
			found = true
			assert.Equal(t, "20260601", row.LoanDate)
		}
	}
	assert.True(t, found, "expected credit detail row in fixture")
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

func TestNormalizeCreditDetailAccountHoldingsFromFixture(t *testing.T) {
	holdings, err := normalizeCreditDetailAccountHoldings(loadAccountFixture(t))
	require.NoError(t, err)
	require.Len(t, holdings, 3)

	var foundCash bool
	var foundCreditDetail bool
	var foundFreeLot bool
	for _, holding := range holdings {
		assert.NotContains(t, holding.StockName, "*")
		assert.NotZero(t, holding.HoldingQuantity)
		if holding.StockCode == "000001" {
			foundCash = true
			assert.Equal(t, FundingTypeCash, holding.FundingType)
			assert.Equal(t, "", holding.LoanDate)
		}
		if holding.StockCode == "000003" {
			foundCreditDetail = true
			assert.Equal(t, FundingTypeCredit, holding.FundingType)
			assert.Equal(t, "Synthetic Credit Detail", holding.StockName)
			assert.Equal(t, "20260601", holding.LoanDate)
		}
		if holding.StockCode == "000004" {
			foundFreeLot = true
			assert.Equal(t, FundingTypeCash, holding.FundingType)
			assert.Equal(t, "", holding.LoanDate)
			assert.Nil(t, holding.ProfitRate)
		}
	}
	assert.True(t, foundCash)
	assert.True(t, foundCreditDetail)
	assert.True(t, foundFreeLot)
}

func TestNormalizeCreditDetailAccountHoldingsFiltersCreditAggregates(t *testing.T) {
	rows := []kiwoom.AccountProfitRateRow{
		{StockCode: "000010", StockName: "Synthetic Cash", CurrentPrice: "+1000", PurchasePrice: "900", PurchaseAmount: "900", RemainingQuantity: "1", OrderableQuantity: "1", CreditType: "00"},
		{StockCode: "000011", StockName: "*Synthetic Starred Cash", CurrentPrice: "+1000", PurchasePrice: "900", PurchaseAmount: "900", RemainingQuantity: "1", OrderableQuantity: "1", CreditType: "00"},
		{StockCode: "000020", StockName: "*Synthetic Credit Detail", CurrentPrice: "-2100", PurchasePrice: "2000", PurchaseAmount: "10000", RemainingQuantity: "5", OrderableQuantity: "5", CreditType: "03", LoanDate: "20260601"},
		{StockCode: "000020", StockName: "*Synthetic Credit Detail", CurrentPrice: "-2100", PurchasePrice: "2050", PurchaseAmount: "8200", RemainingQuantity: "4", OrderableQuantity: "4", CreditType: "03", LoanDate: "20260602"},
		{StockCode: "000020", StockName: "Synthetic Credit Aggregate", CurrentPrice: "-2100", PurchasePrice: "2022", PurchaseAmount: "18200", RemainingQuantity: "9", OrderableQuantity: "9", CreditType: "03"},
		{StockCode: "000021", StockName: "Synthetic Unmatched Credit Aggregate", CurrentPrice: "-1800", PurchasePrice: "1700", PurchaseAmount: "1700", RemainingQuantity: "1", OrderableQuantity: "1", CreditType: "03"},
		{StockCode: "000022", StockName: "*Synthetic Zero Credit Detail", CurrentPrice: "-1000", PurchasePrice: "900", PurchaseAmount: "0", RemainingQuantity: "0", OrderableQuantity: "0", CreditType: "03", LoanDate: "20260603"},
	}

	holdings, err := normalizeCreditDetailAccountHoldings(rows)
	require.NoError(t, err)
	require.Len(t, holdings, 4)

	namesByCode := map[string][]string{}
	loanDatesByCode := map[string][]string{}
	for _, holding := range holdings {
		assert.NotContains(t, holding.StockName, "*")
		namesByCode[holding.StockCode] = append(namesByCode[holding.StockCode], holding.StockName)
		loanDatesByCode[holding.StockCode] = append(loanDatesByCode[holding.StockCode], holding.LoanDate)
	}

	assert.Equal(t, []string{"Synthetic Cash"}, namesByCode["000010"])
	assert.Equal(t, []string{""}, loanDatesByCode["000010"])
	assert.Equal(t, []string{"Synthetic Starred Cash"}, namesByCode["000011"])
	assert.Equal(t, []string{""}, loanDatesByCode["000011"])
	assert.Equal(t, []string{"Synthetic Credit Detail", "Synthetic Credit Detail"}, namesByCode["000020"])
	assert.ElementsMatch(t, []string{"20260601", "20260602"}, loanDatesByCode["000020"])
	assert.Empty(t, namesByCode["000021"], "unstarred credit aggregate rows must be hidden in detail mode")
	assert.Empty(t, namesByCode["000022"], "zero-quantity detail rows must be hidden")
}

func TestNormalizeCreditDetailAccountHoldingsJSONSchema(t *testing.T) {
	holdings, err := normalizeCreditDetailAccountHoldings(loadAccountFixture(t))
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
		"funding_type":       {},
		"loan_date":          {},
	}
	var foundCash bool
	var foundCredit bool
	for _, row := range decoded {
		assert.Equal(t, expected, keySet(row))
		assert.IsType(t, "", row["loan_date"])
		assert.IsType(t, "", row["funding_type"])
		switch row["funding_type"] {
		case string(FundingTypeCash):
			foundCash = true
		case string(FundingTypeCredit):
			foundCredit = true
		default:
			t.Fatalf("unexpected funding_type: %v", row["funding_type"])
		}
	}
	assert.True(t, foundCash)
	assert.True(t, foundCredit)
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
		"funding_type":       {},
	}
	var foundCash bool
	var foundCredit bool
	for _, row := range decoded {
		assert.Equal(t, expected, keySet(row))
		assert.IsType(t, "", row["funding_type"])
		switch row["funding_type"] {
		case string(FundingTypeCash):
			foundCash = true
		case string(FundingTypeCredit):
			foundCredit = true
		default:
			t.Fatalf("unexpected funding_type: %v", row["funding_type"])
		}
	}
	assert.True(t, foundCash)
	assert.True(t, foundCredit)
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

func TestAccountsListCommandContainsCreditDetailFlag(t *testing.T) {
	var found bool
	for _, flag := range accountsListCmd.Flags {
		for _, name := range flag.Names() {
			if name == "credit-detail" {
				found = true
			}
		}
	}
	assert.True(t, found)
}

func TestRunAccountsListMissingCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KIWOOM_APPKEY", "")
	t.Setenv("KIWOOM_SECRETKEY", "")

	err := runAccountsList(context.Background(), accountListOptions{}, nil)
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
