package service

import (
	"context"
	"fmt"
	"time"

	"funds_transaction/internal/model"
	"funds_transaction/internal/request"
	"funds_transaction/pkg/database"
	"funds_transaction/pkg/errors"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func Funds(ctx context.Context, req *request.FundListReq) *request.FundListResp {
	tx := database.GetDB(ctx)
	return FundsWithTx(tx, req)
}

func FundsWithTx(tx *gorm.DB, req *request.FundListReq) *request.FundListResp {
	t := &model.Transaction{}
	var transactions []*model.Transaction

	query := tx.Table(t.TableName()).Unscoped()
	if req.StartTime > 0 {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime > 0 {
		query = query.Where("created_at <= ?", req.EndTime)
	}
	if req.BuyID > 0 {
		query = query.Where("buy_id = ?", req.BuyID)
	}

	if req.Pager == nil {
		req.Pager = database.DefaultPager()
	}
	req.Pager.Load(query)

	if req.Sorter == nil {
		req.Sorter = database.DefaultSorter()
	}
	req.Sorter.Load(query)

	var total int64
	query.Find(&transactions).Count(&total)
	req.Pager.SetTotalCount(int(total))

	return &request.FundListResp{Items: transactions, Pager: req.Pager}
}

func GetFundByID(ctx context.Context, id int64) (*model.Transaction, error) {
	tx := database.GetDB(ctx)
	return GetFundByIDWithTx(tx, id)
}

func GetFundByIDWithTx(tx *gorm.DB, id int64) (*model.Transaction, error) {
	fund := &model.Transaction{}
	err := tx.First(fund, id).Error
	return fund, err
}

func AddFundTransaction(ctx context.Context, req *model.Transaction) (int64, error) {
	err := database.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		// 处理数据
		req.CalculateLoad()
		if req.CreatedAt.IsZero() {
			req.CreatedAt = time.Now()
		}
		// 买入数据
		if req.IsBuy() {
			req.LeftAmount = req.Amount
			err := tx.Create(req).Error
			return err
		}
		// 卖出数据
		return addSellFundTransaction(tx, req)
	})
	return req.ID, err
}

func addSellFundTransaction(tx *gorm.DB, req *model.Transaction) error {
	buy, err := GetFundByIDWithTx(tx, req.BuyID)
	if err != nil {
		logrus.Errorf("get fund by id err: %v", err)
		return errors.WrapC(errors.CodeUnknownError, err)
	}
	buy.LeftAmount -= req.Amount
	if buy.LeftAmount < 0 {
		return errors.WrapC(errors.CodeBadRequest, fmt.Errorf("buy id %d left amount is not enough", req.BuyID))
	}
	// 计算卖出利润，并保存卖出数据
	req.CalculateSellProfit(buy.Unit)
	if err = tx.Create(req).Error; err != nil {
		logrus.Errorf("create sell fund err: %v", err)
		return errors.WrapC(errors.CodeUnknownError, err)
	}
	// 全部卖出后，计算该买入数据的利润
	if buy.LeftAmount == 0 {
		var profit float64
		sells := FundsWithTx(tx, &request.FundListReq{BuyID: req.BuyID})
		for _, sell := range sells.Items {
			profit += sell.Profit
		}
		buy.Profit = profit
		buy.NetProfit = profit - buy.Load
		buy.ProfitMargin = profit / buy.Price
	}
	if err = tx.Save(buy).Error; err != nil {
		logrus.Errorf("save buy fund err: %v", err)
		return errors.WrapC(errors.CodeUnknownError, err)
	}
	return nil
}
