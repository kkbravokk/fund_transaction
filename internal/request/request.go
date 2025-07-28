package request

import (
	"funds_transaction/internal/model"
	"funds_transaction/pkg/database"
)

type TransactionListReq struct {
	FundCode      string `json:"fund_code" form:"fund_code"`
	StartTime     int64  `json:"start_time" form:"start_time"`
	EndTime       int64  `json:"end_time" form:"end_time"`
	OriginalBuyID int64  `json:"original_buy_id" form:"original_buy_id"`

	*database.Pager
	*database.Sorter
}

type TransactionListResp struct {
	Items []*model.Transaction `json:"items"`
	*database.Pager
}

type AddTransactions struct {
	*model.Transaction
}

type AddTransactionsResp struct {
	ID int64 `json:"id"`
}

type FundListReq struct {
	Code string `json:"code" form:"code"`
	Name string `json:"name" form:"name"`

	*database.Pager
	*database.Sorter
}

type FundListResp struct {
	Items []*model.Fund

	*database.Pager
}
