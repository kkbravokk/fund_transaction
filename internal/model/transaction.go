package model

import "funds_transaction/pkg/utils"

const (
	basicLoadFeePrice = 2000   // 基础手续费的价格，即买入金额少于这个值，按照这个值收取手续费
	InterestRage      = 0.0001 // 利率
)

const (
	Buy  = "buy"
	Sell = "sell"
)

type Transaction struct {
	ID              int64   `json:"id" gorm:"primary"`
	OriginalBuyId   int64   `json:"original_buy_id"`  // 对应的买入
	TransactionType string  `json:"transaction_type"` // 买入或卖出, buy/sell
	FundCode        string  `json:"fund_code"`        // 证券代码
	Unit            float64 `json:"unit"`             // 单价
	Amount          int     `json:"amount"`           // 数量
	Price           float64 `json:"price"`            // 价格
	Load            float64 `json:"load"`             // 手续费
	LeftAmount      int     `json:"left_amount"`      // 剩余数量
	Profit          float64 `json:"profit"`           // 利润
	ProfitMargin    float64 `json:"profit_margin"`    // 利润率
	NetProfit       float64 `json:"net_profit"`       // 利润
	CreatedAt       int64   `json:"created_at"`       // 创建时间
}

func (t *Transaction) TableName() string {
	return "transaction"
}

func (t *Transaction) IsBuy() bool {
	return t.TransactionType == Buy
}

func (t *Transaction) CalculatePrice() {
	t.Price = t.Unit * float64(t.Amount)
}

func (t *Transaction) CalculateLoad() {
	t.CalculatePrice()
	if t.Price >= basicLoadFeePrice {
		t.Load = t.Price * InterestRage
		return
	}
	t.Load = utils.Round(basicLoadFeePrice*InterestRage, 2)
}

func (t *Transaction) CalculateSellProfit(buyUnit float64) {
	t.CalculateLoad()
	// 利润 = (卖出单价 - 买入单价) * 买入数量
	t.Profit = (t.Unit - buyUnit) * float64(t.Amount)
	t.NetProfit = t.Profit - t.Load
	// 利润率 = 利润/ 买入金额
	t.ProfitMargin = t.Profit / (buyUnit * float64(t.Amount))
}
