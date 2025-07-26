package request

import (
	"funds_transaction/internal/model"
	"funds_transaction/pkg/database"
)

type FundListReq struct {
	StartTime int64 `json:"start_time" form:"start_time"`
	EndTime   int64 `json:"end_time" form:"end_time"`
	BuyID     int64 `json:"buy_id" form:"buy_id"`

	*database.Pager
	*database.Sorter
}

type FundListResp struct {
	Items []*model.Transaction `json:"items"`
	*database.Pager
}

type AddFund struct {
	*model.Transaction
}

type AddFundResp struct {
	ID int64 `json:"id"`
}
