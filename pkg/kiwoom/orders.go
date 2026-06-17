package kiwoom

import "context"

const (
	orderListAPIID    = "ka10075"
	orderListMaxPages = 100
)

type OrderListRequest struct {
	AllStockType string `json:"all_stk_tp"`
	TradeType    string `json:"trde_tp"`
	StockCode    string `json:"stk_cd"`
	ExchangeType string `json:"stex_tp"`
}

type OrderListRow struct {
	OrderID          string `json:"ord_no"`
	OriginalOrderID  string `json:"orig_ord_no"`
	StockCode        string `json:"stk_cd"`
	StockName        string `json:"stk_nm"`
	OrderedQuantity  string `json:"ord_qty"`
	OrderedPrice     string `json:"ord_pric"`
	UnfilledQuantity string `json:"oso_qty"`
	OrderKind        string `json:"io_tp_nm"`
	FilledQuantity   string `json:"cntr_qty"`
	CurrentPrice     string `json:"cur_prc"`
	ExchangeType     string `json:"stex_tp"`
}

type orderListPageResponse struct {
	Rows       []OrderListRow `json:"oso"`
	ReturnCode int            `json:"return_code"`
	ReturnMsg  string         `json:"return_msg"`
}

// OrderList returns order rows merged from every continuation page.
func (c *Client) OrderList(ctx context.Context, req OrderListRequest) ([]OrderListRow, error) {
	var allRows []OrderListRow

	err := postJSONContinuationPages(ctx, c, "/api/dostk/acnt", req, orderListAPIID, "Kiwoom order list request", orderListMaxPages, func(response *orderListPageResponse) error {
		if response.ReturnCode != 0 {
			return kiwoomReturnCodeError("order list request", response.ReturnCode)
		}
		allRows = append(allRows, response.Rows...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return allRows, nil
}
