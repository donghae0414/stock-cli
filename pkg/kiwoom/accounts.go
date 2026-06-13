package kiwoom

import (
	"context"
)

const accountProfitRateAPIID = "ka10085"

type AccountProfitRateRequest struct {
	ExchangeType string `json:"stex_tp"`
}

type AccountProfitRateResponse struct {
	Rows       []AccountProfitRateRow `json:"acnt_prft_rt"`
	ReturnCode int                    `json:"return_code"`
	ReturnMsg  string                 `json:"return_msg"`
}

type AccountProfitRateRow struct {
	Date              string `json:"dt"`
	StockCode         string `json:"stk_cd"`
	StockName         string `json:"stk_nm"`
	CurrentPrice      string `json:"cur_prc"`
	PurchasePrice     string `json:"pur_pric"`
	PurchaseAmount    string `json:"pur_amt"`
	RemainingQuantity string `json:"rmnd_qty"`
	OrderableQuantity string `json:"clrn_alow_qty"`
	CreditType        string `json:"crd_tp"`
	LoanDate          string `json:"loan_dt"`
}

func (c *Client) AccountProfitRates(ctx context.Context) (AccountProfitRateResponse, error) {
	var response AccountProfitRateResponse
	err := c.PostJSON(
		ctx,
		"/api/dostk/acnt",
		AccountProfitRateRequest{ExchangeType: "0"},
		map[string]string{
			"cont-yn":  "N",
			"next-key": "",
			"api-id":   accountProfitRateAPIID,
		},
		&response,
	)
	if err != nil {
		return AccountProfitRateResponse{}, err
	}
	if response.ReturnCode != 0 {
		return AccountProfitRateResponse{}, kiwoomReturnCodeError("account request", response.ReturnCode)
	}
	return response, nil
}
