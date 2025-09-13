package request

import (
	"funds_transaction/internal/model"
	"funds_transaction/pkg/database"
)

type PreprocessingResp struct {
	RawRecordCount   int                 `json:"raw_record_count"`
	ParseRecordCount int                 `json:"parse_record_count"`
	Data             []*PreProcessRecord `json:"data"`
}

type PreProcessRecord struct {
	OriginalBuyId   int     `json:"original_buy_id"`
	TransactionType string  `json:"transaction_type"`
	FundCode        string  `json:"fund_code"`
	Unit            float64 `json:"unit"`
	Amount          int64   `json:"amount"`
	CreatedAt       int64   `json:"created_at"`
	Create          string  `json:"create"`
}

type TransactionListReq struct {
	FundCode        string `json:"fund_code" form:"fund_code"`
	StartTime       int64  `json:"start_time" form:"start_time"`
	EndTime         int64  `json:"end_time" form:"end_time"`
	OriginalBuyID   int64  `json:"original_buy_id" form:"original_buy_id"`
	TransactionType string `json:"transaction_type" form:"transaction_type"`
	Left            bool   `json:"left" form:"left" description:"query left amount great than 0"`

	*database.Pager
	*database.Sorter
}

type TransactionListResp struct {
	Items []*model.Transaction `json:"items"`
	Pager *database.Pager      `json:"pager"`
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
	Items []*model.Fund `json:"items"`

	Pager *database.Pager `json:"pager"`
}
