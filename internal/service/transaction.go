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

func Transactions(ctx context.Context, req *request.TransactionListReq) *request.TransactionListResp {
	tx := database.GetDB(ctx)
	return TransactionsWithTx(tx, req)
}

func TransactionsWithTx(tx *gorm.DB, req *request.TransactionListReq) *request.TransactionListResp {
	t := &model.Transaction{}
	var transactions []*model.Transaction

	query := tx.Table(t.TableName()).Unscoped()
	if req.FundCode != "" {
		query = query.Where("fund_code = ?", req.FundCode)
	}
	if req.StartTime > 0 {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime > 0 {
		query = query.Where("created_at <= ?", req.EndTime)
	}
	if req.OriginalBuyID > 0 {
		query = query.Where("original_buy_id = ?", req.OriginalBuyID)
	}
	if req.TransactionType != "" {
		query = query.Where("transaction_type = ?", req.TransactionType)
	}
	if req.Left {
		query = query.Where("left_amount > 0")
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
	query.Find(&transactions).Limit(-1).Offset(-1).Count(&total)
	req.Pager.SetTotalCount(int(total))

	return &request.TransactionListResp{Items: transactions, Pager: req.Pager}
}

func GetTransactionByID(ctx context.Context, id int64) (*model.Transaction, error) {
	tx := database.GetDB(ctx)
	return GetTransactionByIDWithTx(tx, id)
}

func GetTransactionByIDWithTx(tx *gorm.DB, id int64) (*model.Transaction, error) {
	transaction := &model.Transaction{}
	err := tx.First(transaction, id).Error
	return transaction, err
}

func AddTransaction(ctx context.Context, req *model.Transaction) (int64, error) {
	err := database.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		// 处理数据
		req.CalculateLoad()
		if req.CreatedAt == 0 {
			req.CreatedAt = time.Now().Unix()
		}
		// 买入数据
		if req.IsBuy() {
			req.LeftAmount = req.Amount
			err := tx.Create(req).Error
			return err
		}
		// 卖出数据
		return addSellTransaction(tx, req)
	})
	return req.ID, err
}

func addSellTransaction(tx *gorm.DB, req *model.Transaction) error {
	buy, err := GetTransactionByIDWithTx(tx, req.OriginalBuyId)
	if err != nil {
		logrus.Errorf("get transactions by id err: %v", err)
		return errors.WrapC(errors.CodeUnknownError, err)
	}
	if buy.FundCode != req.FundCode {
		return errors.WrapC(errors.CodeBadRequest, fmt.Errorf("fund code not match"))
	}
	buy.LeftAmount -= req.Amount
	if buy.LeftAmount < 0 {
		return errors.WrapC(errors.CodeBadRequest, fmt.Errorf("original %d left is not enough", req.OriginalBuyId))
	}
	// 计算卖出利润，并保存卖出数据
	req.CalculateSellProfit(buy.Unit)
	if err = tx.Create(req).Error; err != nil {
		logrus.Errorf("create sell transactions err: %v", err)
		return errors.WrapC(errors.CodeUnknownError, err)
	}
	// 全部卖出后，计算该买入数据的利润
	if buy.LeftAmount == 0 {
		var profit float64
		sells := TransactionsWithTx(tx, &request.TransactionListReq{OriginalBuyID: req.OriginalBuyId})
		for _, sell := range sells.Items {
			profit += sell.Profit
		}
		buy.Profit = profit
		buy.NetProfit = profit - buy.Load
		buy.ProfitMargin = profit / buy.Price
	}
	if err = tx.Save(buy).Error; err != nil {
		logrus.Errorf("save buy transactions err: %v", err)
		return errors.WrapC(errors.CodeUnknownError, err)
	}
	return nil
}

func UpdateTransaction(ctx context.Context, req *model.Transaction) error {
	// todo complete
	return nil
}

func DelTransaction(ctx context.Context, id int64) error {
	transaction, err := GetTransactionByID(ctx, id)
	if err != nil {
		return err
	}
	err = database.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		if transaction.IsBuy() {
			deletes := []int64{transaction.ID}
			sells := Transactions(ctx, &request.TransactionListReq{OriginalBuyID: transaction.ID})
			for _, sell := range sells.Items {
				deletes = append(deletes, sell.ID)
			}
			err = tx.Delete(&model.Transaction{}, deletes).Error
			return err
		}
		// 卖出数据，先更新买入，再删除卖出
		var buy *model.Transaction
		buy, err = GetTransactionByIDWithTx(tx, transaction.OriginalBuyId)
		if err != nil {
			logrus.Errorf("get transactions by id err: %v", err)
			return errors.WrapC(errors.CodeUnknownError, err)
		}
		buy.LeftAmount += transaction.Amount
		buy.Profit = 0
		buy.NetProfit = 0
		buy.ProfitMargin = 0
		if err = tx.Save(buy).Error; err != nil {
			logrus.Errorf("save buy transactions err: %v", err)
			return errors.WrapC(errors.CodeUnknownError, err)
		}
		err = tx.Delete(&model.Transaction{}, transaction.ID).Error
		return err
	})
	return err

}
