package stocklookup

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock-cli/pkg/kiwoom"
)

func TestLookupUsesSchemaFilteringRankingAndOrder(t *testing.T) {
	rows := []kiwoom.StockInfoRow{
		{Code: "999999", Name: "삼성전자 ETF", MarketName: "거래소"},
		{Code: "005930", Name: "삼성전자", MarketName: "거래소", UpName: "전기전자"},
		{Code: "006400", Name: "삼성SDI", MarketName: "거래소", UpName: "전기전자"},
		{Code: "028260", Name: "삼성물산", MarketName: "거래소"},
		{Code: "000660", Name: "SK하이닉스", MarketName: "거래소"},
		{Code: "111111", Name: "삼성옵션", MarketName: "거래소"},
		{Code: "222222", Name: "삼성비상장", MarketName: "K-OTC"},
	}

	output := Lookup([]string{"삼성전자", "삼성", "없음"}, rows, 2)

	data, err := json.Marshal(output)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, map[string]struct{}{"ok": {}, "queries": {}, "errors": {}}, keySet(decoded))
	assert.Equal(t, true, decoded["ok"])
	assert.Empty(t, decoded["errors"])

	queries := decoded["queries"].([]any)
	require.Len(t, queries, 3)

	exact := queries[0].(map[string]any)
	assert.Equal(t, "삼성전자", exact["query"])
	assert.Equal(t, "exact", exact["status"])
	assert.Equal(t, "exact", exact["match_type"])
	assert.Equal(t, float64(1), exact["total_candidates"])
	assert.Equal(t, false, exact["truncated"])
	exactCandidates := exact["candidates"].([]any)
	require.Len(t, exactCandidates, 1)
	firstExact := exactCandidates[0].(map[string]any)
	assert.Equal(t, map[string]struct{}{
		"code":        {},
		"name":        {},
		"market_name": {},
		"match_type":  {},
		"up_name":     {},
	}, keySet(firstExact))
	assert.Equal(t, "005930", firstExact["code"])
	assert.Equal(t, "전기전자", firstExact["up_name"])

	ambiguous := queries[1].(map[string]any)
	assert.Equal(t, "삼성", ambiguous["query"])
	assert.Equal(t, "ambiguous", ambiguous["status"])
	assert.Equal(t, "partial", ambiguous["match_type"])
	assert.Equal(t, float64(3), ambiguous["total_candidates"])
	assert.Equal(t, true, ambiguous["truncated"])
	ambiguousCandidates := ambiguous["candidates"].([]any)
	require.Len(t, ambiguousCandidates, 2)
	assert.Equal(t, "삼성전자", ambiguousCandidates[0].(map[string]any)["name"])
	assert.Equal(t, "삼성물산", ambiguousCandidates[1].(map[string]any)["name"])

	notFound := queries[2].(map[string]any)
	assert.Equal(t, "없음", notFound["query"])
	assert.Equal(t, "not_found", notFound["status"])
	assert.Nil(t, notFound["match_type"])
	assert.Empty(t, notFound["candidates"])
	assert.NotContains(t, notFound, "total_candidates")
	assert.NotContains(t, notFound, "truncated")
}

func keySet(row map[string]any) map[string]struct{} {
	keys := make(map[string]struct{}, len(row))
	for key := range row {
		keys[key] = struct{}{}
	}
	return keys
}
