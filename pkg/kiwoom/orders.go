package kiwoom

import "context"

const (
	orderListAPIID       = "ka10075"
	cashBuyOrderAPIID    = "kt10000"
	cashSellOrderAPIID   = "kt10001"
	cashCancelAPIID      = "kt10003"
	creditBuyOrderAPIID  = "kt10006"
	creditSellOrderAPIID = "kt10007"
	creditCancelAPIID    = "kt10009"
	orderListMaxPages    = 100
	orderAccountEndpoint = "/api/dostk/acnt"
	orderEndpoint        = "/api/dostk/ordr"
	creditOrderEndpoint  = "/api/dostk/crdordr"
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

type CashOrderRequest struct {
	DomesticExchangeType string `json:"dmst_stex_tp"`
	StockCode            string `json:"stk_cd"`
	OrderQuantity        string `json:"ord_qty"`
	OrderUnitPrice       string `json:"ord_uv"`
	TradeType            string `json:"trde_tp"`
	ConditionPrice       string `json:"cond_uv"`
}

type CashOrderResponse struct {
	OrderID              string `json:"ord_no"`
	DomesticExchangeType string `json:"dmst_stex_tp"`
	ReturnCode           int    `json:"return_code"`
	ReturnMsg            string `json:"return_msg"`
}

type CashCancelRequest struct {
	DomesticExchangeType string `json:"dmst_stex_tp"`
	OriginalOrderID      string `json:"orig_ord_no"`
	StockCode            string `json:"stk_cd"`
	CancelQuantity       string `json:"cncl_qty"`
}

type CashCancelResponse struct {
	OrderID             string `json:"ord_no"`
	BaseOriginalOrderID string `json:"base_orig_ord_no"`
	CancelQuantity      string `json:"cncl_qty"`
	ReturnCode          int    `json:"return_code"`
	ReturnMsg           string `json:"return_msg"`
}

type CreditOrderRequest struct {
	DomesticExchangeType string `json:"dmst_stex_tp"`
	StockCode            string `json:"stk_cd"`
	OrderQuantity        string `json:"ord_qty"`
	OrderUnitPrice       string `json:"ord_uv"`
	TradeType            string `json:"trde_tp"`
	ConditionPrice       string `json:"cond_uv"`
}

type CreditSellOrderRequest struct {
	DomesticExchangeType string `json:"dmst_stex_tp"`
	StockCode            string `json:"stk_cd"`
	OrderQuantity        string `json:"ord_qty"`
	OrderUnitPrice       string `json:"ord_uv"`
	TradeType            string `json:"trde_tp"`
	CreditDealType       string `json:"crd_deal_tp"`
	CreditLoanDate       string `json:"crd_loan_dt"`
	ConditionPrice       string `json:"cond_uv"`
}

type CreditOrderResponse struct {
	OrderID              string `json:"ord_no"`
	DomesticExchangeType string `json:"dmst_stex_tp"`
	ReturnCode           int    `json:"return_code"`
	ReturnMsg            string `json:"return_msg"`
}

type CreditCancelRequest struct {
	DomesticExchangeType string `json:"dmst_stex_tp"`
	OriginalOrderID      string `json:"orig_ord_no"`
	StockCode            string `json:"stk_cd"`
	CancelQuantity       string `json:"cncl_qty"`
}

type CreditCancelResponse struct {
	OrderID             string `json:"ord_no"`
	BaseOriginalOrderID string `json:"base_orig_ord_no"`
	CancelQuantity      string `json:"cncl_qty"`
	ReturnCode          int    `json:"return_code"`
	ReturnMsg           string `json:"return_msg"`
}

// OrderList returns order rows merged from every continuation page.
func (c *Client) OrderList(ctx context.Context, req OrderListRequest) ([]OrderListRow, error) {
	var allRows []OrderListRow

	err := postJSONContinuationPages(ctx, c, orderAccountEndpoint, req, orderListAPIID, "Kiwoom order list request", orderListMaxPages, func(response *orderListPageResponse) error {
		if response.ReturnCode != 0 {
			return kiwoomReturnCodeError("order list request", response.ReturnCode, response.ReturnMsg)
		}
		allRows = append(allRows, response.Rows...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return allRows, nil
}

func (c *Client) CashBuyOrder(ctx context.Context, req CashOrderRequest) (CashOrderResponse, error) {
	return c.cashOrder(ctx, req, cashBuyOrderAPIID, "cash buy order")
}

func (c *Client) CashSellOrder(ctx context.Context, req CashOrderRequest) (CashOrderResponse, error) {
	return c.cashOrder(ctx, req, cashSellOrderAPIID, "cash sell order")
}

func (c *Client) cashOrder(ctx context.Context, req CashOrderRequest, apiID string, operation string) (CashOrderResponse, error) {
	var response CashOrderResponse
	err := c.PostJSON(ctx, orderEndpoint, req, orderHeaders(apiID), &response)
	if err != nil {
		return CashOrderResponse{}, err
	}
	if response.ReturnCode != 0 {
		return CashOrderResponse{}, kiwoomReturnCodeError(operation, response.ReturnCode, response.ReturnMsg)
	}
	return response, nil
}

func (c *Client) CashCancelOrder(ctx context.Context, req CashCancelRequest) (CashCancelResponse, error) {
	var response CashCancelResponse
	err := c.PostJSON(ctx, orderEndpoint, req, orderHeaders(cashCancelAPIID), &response)
	if err != nil {
		return CashCancelResponse{}, err
	}
	if response.ReturnCode != 0 {
		return CashCancelResponse{}, kiwoomReturnCodeError("cash cancel order", response.ReturnCode, response.ReturnMsg)
	}
	return response, nil
}

func (c *Client) CreditBuyOrder(ctx context.Context, req CreditOrderRequest) (CreditOrderResponse, error) {
	var response CreditOrderResponse
	err := c.PostJSON(ctx, creditOrderEndpoint, req, orderHeaders(creditBuyOrderAPIID), &response)
	if err != nil {
		return CreditOrderResponse{}, err
	}
	if response.ReturnCode != 0 {
		return CreditOrderResponse{}, kiwoomReturnCodeError("credit buy order", response.ReturnCode, response.ReturnMsg)
	}
	return response, nil
}

func (c *Client) CreditSellOrder(ctx context.Context, req CreditSellOrderRequest) (CreditOrderResponse, error) {
	var response CreditOrderResponse
	err := c.PostJSON(ctx, creditOrderEndpoint, req, orderHeaders(creditSellOrderAPIID), &response)
	if err != nil {
		return CreditOrderResponse{}, err
	}
	if response.ReturnCode != 0 {
		return CreditOrderResponse{}, kiwoomReturnCodeError("credit sell order", response.ReturnCode, response.ReturnMsg)
	}
	return response, nil
}

func (c *Client) CreditCancelOrder(ctx context.Context, req CreditCancelRequest) (CreditCancelResponse, error) {
	var response CreditCancelResponse
	err := c.PostJSON(ctx, creditOrderEndpoint, req, orderHeaders(creditCancelAPIID), &response)
	if err != nil {
		return CreditCancelResponse{}, err
	}
	if response.ReturnCode != 0 {
		return CreditCancelResponse{}, kiwoomReturnCodeError("credit cancel order", response.ReturnCode, response.ReturnMsg)
	}
	return response, nil
}

func orderHeaders(apiID string) map[string]string {
	return map[string]string{
		"cont-yn":  "N",
		"next-key": "",
		"api-id":   apiID,
	}
}
