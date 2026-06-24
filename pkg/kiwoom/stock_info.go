package kiwoom

import "context"

const (
	stockInfoEndpoint = "/api/dostk/stkinfo"
	stockInfoAPIID    = "ka10099"
	maxStockInfoPages = 100
)

var defaultStockInfoMarketTypes = []string{"0", "10"}

type StockInfoRequest struct {
	MarketType string `json:"mrkt_tp"`
}

type StockInfoResponse struct {
	Rows       []StockInfoRow `json:"list"`
	ReturnCode int            `json:"return_code"`
	ReturnMsg  string         `json:"return_msg"`
}

type StockInfoRow struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	MarketName string `json:"marketName"`
	UpName     string `json:"upName"`
}

func (c *Client) StockInfoRows(ctx context.Context, marketTypes []string) ([]StockInfoRow, error) {
	if len(marketTypes) == 0 {
		marketTypes = defaultStockInfoMarketTypes
	}

	rows := []StockInfoRow{}
	for _, marketType := range marketTypes {
		req := StockInfoRequest{MarketType: marketType}
		err := postJSONContinuationPages[StockInfoResponse](
			ctx,
			c,
			stockInfoEndpoint,
			req,
			stockInfoAPIID,
			"stock info request",
			maxStockInfoPages,
			func(response *StockInfoResponse) error {
				if response.ReturnCode != 0 {
					return kiwoomReturnCodeError("stock info request", response.ReturnCode, response.ReturnMsg)
				}
				rows = append(rows, response.Rows...)
				return nil
			},
		)
		if err != nil {
			return nil, err
		}
	}
	return rows, nil
}
