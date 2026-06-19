package kiwoom

import "context"

const (
	chartEndpoint    = "/api/dostk/chart"
	chartDayAPIID    = "ka10081"
	chartWeekAPIID   = "ka10082"
	chartMinuteAPIID = "ka10080"
)

type StockDayChartRequest struct {
	StockCode             string `json:"stk_cd"`
	BaseDate              string `json:"base_dt"`
	UpdatedStockPriceType string `json:"upd_stkpc_tp"`
}

type StockWeekChartRequest struct {
	StockCode             string `json:"stk_cd"`
	BaseDate              string `json:"base_dt"`
	UpdatedStockPriceType string `json:"upd_stkpc_tp"`
}

type StockMinuteChartRequest struct {
	StockCode             string `json:"stk_cd"`
	TickScope             string `json:"tic_scope"`
	UpdatedStockPriceType string `json:"upd_stkpc_tp"`
	BaseDate              string `json:"base_dt"`
}

type StockDayChartResponse struct {
	StockCode  string             `json:"stk_cd"`
	Rows       []StockDayChartRow `json:"stk_dt_pole_chart_qry"`
	ReturnCode int                `json:"return_code"`
	ReturnMsg  string             `json:"return_msg"`
}

type StockWeekChartResponse struct {
	StockCode  string              `json:"stk_cd"`
	Rows       []StockWeekChartRow `json:"stk_stk_pole_chart_qry"`
	ReturnCode int                 `json:"return_code"`
	ReturnMsg  string              `json:"return_msg"`
}

type StockMinuteChartResponse struct {
	StockCode  string                `json:"stk_cd"`
	Rows       []StockMinuteChartRow `json:"stk_min_pole_chart_qry"`
	ReturnCode int                   `json:"return_code"`
	ReturnMsg  string                `json:"return_msg"`
}

type StockDayChartRow struct {
	CurrentPrice string `json:"cur_prc"`
	TradeAmount  string `json:"trde_prica"`
	Date         string `json:"dt"`
	OpenPrice    string `json:"open_pric"`
	HighPrice    string `json:"high_pric"`
	LowPrice     string `json:"low_pric"`
}

type StockWeekChartRow struct {
	CurrentPrice string `json:"cur_prc"`
	TradeAmount  string `json:"trde_prica"`
	Date         string `json:"dt"`
	OpenPrice    string `json:"open_pric"`
	HighPrice    string `json:"high_pric"`
	LowPrice     string `json:"low_pric"`
}

type StockMinuteChartRow struct {
	CurrentPrice string `json:"cur_prc"`
	Timestamp    string `json:"cntr_tm"`
	OpenPrice    string `json:"open_pric"`
	HighPrice    string `json:"high_pric"`
	LowPrice     string `json:"low_pric"`
}

func (c *Client) StockDayChart(ctx context.Context, req StockDayChartRequest) (StockDayChartResponse, error) {
	var response StockDayChartResponse
	err := c.PostJSON(ctx, chartEndpoint, req, chartHeaders(chartDayAPIID), &response)
	if err != nil {
		return StockDayChartResponse{}, err
	}
	if response.ReturnCode != 0 {
		return StockDayChartResponse{}, kiwoomReturnCodeError("day chart request", response.ReturnCode, response.ReturnMsg)
	}
	return response, nil
}

func (c *Client) StockWeekChart(ctx context.Context, req StockWeekChartRequest) (StockWeekChartResponse, error) {
	var response StockWeekChartResponse
	err := c.PostJSON(ctx, chartEndpoint, req, chartHeaders(chartWeekAPIID), &response)
	if err != nil {
		return StockWeekChartResponse{}, err
	}
	if response.ReturnCode != 0 {
		return StockWeekChartResponse{}, kiwoomReturnCodeError("week chart request", response.ReturnCode, response.ReturnMsg)
	}
	return response, nil
}

func (c *Client) StockMinuteChart(ctx context.Context, req StockMinuteChartRequest) (StockMinuteChartResponse, error) {
	var response StockMinuteChartResponse
	err := c.PostJSON(ctx, chartEndpoint, req, chartHeaders(chartMinuteAPIID), &response)
	if err != nil {
		return StockMinuteChartResponse{}, err
	}
	if response.ReturnCode != 0 {
		return StockMinuteChartResponse{}, kiwoomReturnCodeError("minute chart request", response.ReturnCode, response.ReturnMsg)
	}
	return response, nil
}

func chartHeaders(apiID string) map[string]string {
	return map[string]string{
		"cont-yn":  "N",
		"next-key": "",
		"api-id":   apiID,
	}
}
