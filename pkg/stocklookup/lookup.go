package stocklookup

import (
	"sort"
	"strings"
	"unicode/utf8"

	"stock-cli/pkg/kiwoom"
)

type Envelope struct {
	OK      bool         `json:"ok"`
	Queries []Query      `json:"queries"`
	Errors  []ErrorEntry `json:"errors"`
}

type ErrorEntry struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type Query struct {
	Query           string      `json:"query"`
	Status          string      `json:"status"`
	MatchType       *string     `json:"match_type"`
	Candidates      []Candidate `json:"candidates"`
	TotalCandidates *int        `json:"total_candidates,omitempty"`
	Truncated       *bool       `json:"truncated,omitempty"`
}

type Candidate struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	MarketName string `json:"market_name"`
	MatchType  string `json:"match_type"`
	UpName     string `json:"up_name,omitempty"`
}

func Lookup(names []string, rows []kiwoom.StockInfoRow, limit int) Envelope {
	filteredRows := filterAllowedRows(rows)
	queries := make([]Query, 0, len(names))
	for _, name := range names {
		queries = append(queries, buildQuery(name, filteredRows, limit))
	}
	return Envelope{
		OK:      true,
		Queries: queries,
		Errors:  []ErrorEntry{},
	}
}

func buildQuery(name string, rows []kiwoom.StockInfoRow, limit int) Query {
	matches, matchType := matchingRows(rows, name)
	if len(matches) == 0 {
		return Query{
			Query:      name,
			Status:     "not_found",
			MatchType:  nil,
			Candidates: []Candidate{},
		}
	}

	status := "ambiguous"
	if len(matches) == 1 {
		status = "single_partial"
		if matchType == "exact" {
			status = "exact"
		}
	}

	limited := matches
	if len(limited) > limit {
		limited = limited[:limit]
	}
	totalCandidates := len(matches)
	truncated := len(matches) > limit
	return Query{
		Query:           name,
		Status:          status,
		MatchType:       &matchType,
		Candidates:      candidates(limited, matchType),
		TotalCandidates: &totalCandidates,
		Truncated:       &truncated,
	}
}

func matchingRows(rows []kiwoom.StockInfoRow, query string) ([]kiwoom.StockInfoRow, string) {
	queryText := normalizeText(query)
	exactRows := make([]kiwoom.StockInfoRow, 0)
	for _, row := range rows {
		if normalizeText(row.Name) == queryText {
			exactRows = append(exactRows, row)
		}
	}
	if len(exactRows) > 0 {
		sortRows(exactRows, queryText)
		return exactRows, "exact"
	}

	partialRows := make([]kiwoom.StockInfoRow, 0)
	for _, row := range rows {
		if strings.Contains(normalizeText(row.Name), queryText) {
			partialRows = append(partialRows, row)
		}
	}
	if len(partialRows) > 0 {
		sortRows(partialRows, queryText)
		return partialRows, "partial"
	}
	return nil, ""
}

func sortRows(rows []kiwoom.StockInfoRow, queryText string) {
	sort.Slice(rows, func(i, j int) bool {
		left := rowRank(rows[i], queryText)
		right := rowRank(rows[j], queryText)
		if left.matchRank != right.matchRank {
			return left.matchRank < right.matchRank
		}
		if left.nameLength != right.nameLength {
			return left.nameLength < right.nameLength
		}
		if left.codeLength != right.codeLength {
			return left.codeLength < right.codeLength
		}
		return left.code < right.code
	})
}

type rank struct {
	matchRank  int
	nameLength int
	codeLength int
	code       string
}

func rowRank(row kiwoom.StockInfoRow, queryText string) rank {
	normalizedName := normalizeText(row.Name)
	matchRank := 2
	if normalizedName == queryText {
		matchRank = 0
	} else if strings.HasPrefix(normalizedName, queryText) {
		matchRank = 1
	}
	code := strings.TrimSpace(row.Code)
	return rank{
		matchRank:  matchRank,
		nameLength: utf8.RuneCountInString(strings.TrimSpace(row.Name)),
		codeLength: len(code),
		code:       code,
	}
}

func candidates(rows []kiwoom.StockInfoRow, matchType string) []Candidate {
	candidates := make([]Candidate, 0, len(rows))
	for _, row := range rows {
		candidate := Candidate{
			Code:       strings.TrimSpace(row.Code),
			Name:       strings.TrimSpace(row.Name),
			MarketName: strings.TrimSpace(row.MarketName),
			MatchType:  matchType,
		}
		if upName := strings.TrimSpace(row.UpName); upName != "" {
			candidate.UpName = upName
		}
		candidates = append(candidates, candidate)
	}
	return candidates
}

func filterAllowedRows(rows []kiwoom.StockInfoRow) []kiwoom.StockInfoRow {
	filtered := make([]kiwoom.StockInfoRow, 0, len(rows))
	for _, row := range rows {
		if isAllowedRow(row) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func isAllowedRow(row kiwoom.StockInfoRow) bool {
	name := row.Name
	for _, keyword := range []string{"ETF", "ETN", "선물", "옵션"} {
		if strings.Contains(name, keyword) {
			return false
		}
	}
	switch strings.TrimSpace(row.MarketName) {
	case "거래소", "코스닥", "리츠", "인프라투자금융":
		return true
	default:
		return false
	}
}

func normalizeText(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
