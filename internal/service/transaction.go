package service

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"funds_transaction/internal/model"
	"funds_transaction/internal/request"
	"funds_transaction/pkg/database"
	"funds_transaction/pkg/errors"

	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func Preprocessing(ctx context.Context, data string) *request.PreprocessingResp {
	// 先在每个“已成交”后添加一个特殊分隔符(如"end"), 再按分隔符分割，得到单条记录（最后一条会多一个空字符串，需过滤）
	const separator = "->end"
	withSeparator := strings.ReplaceAll(data, "已成交", fmt.Sprintf("已成交%s", separator))
	rawRecords := strings.Split(withSeparator, separator)
	// 多行匹配正则(使用(?s)单行模式，让.匹配换行符)
	pattern := `(?s)^([^\d]+?)(\d+\.\w+)\s+?(买入|卖出)\s+成交价格:(\d+\.\d+)元\s*成交数量:(\d+)\s+\s*>?\s*成交时间:(\d{4}-\d{2}-\d{2})\s*(\d{2}:\d{2}:\d{2})\s*>?\s*已成交`
	re := regexp.MustCompile(pattern)

	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logrus.Errorf("preprocessing load location err: %v", err)
		return nil
	}

	res := request.PreprocessingResp{}
	// 解析每条记录
	for i, record := range rawRecords {
		// 去除记录中的换行符，统一转为空格（避免换行影响匹配）
		cleanRecord := strings.ReplaceAll(record, "\n", " ")
		cleanRecord = strings.TrimSpace(cleanRecord)
		if cleanRecord == "" {
			continue
		}
		res.RawRecordCount++
		matches := re.FindStringSubmatch(cleanRecord)
		if matches == nil {
			logrus.Errorf("match %dth failed: %s", i+1, cleanRecord)
			continue
		}
		create := fmt.Sprintf("%s %s", matches[6], matches[7])
		createAt, err := time.ParseInLocation("2006-01-02 15:04:05", create, location)
		if err != nil {
			logrus.Errorf("parse %dth in location failed: %s", i+1, create)
			continue
		}
		var transactionType string
		switch matches[3] {
		case model.BuyInChinese:
			transactionType = model.Buy
		case model.SellInChinese:
			transactionType = model.Sell
		}
		unit, err := strconv.ParseFloat(matches[4], 64)
		if err != nil {
			logrus.Errorf("parse %dth unit failed: %s", i+1, matches[4])
			continue
		}
		amount, err := strconv.ParseInt(matches[5], 10, 64)
		if err != nil {
			logrus.Errorf("parse %dth amount failed: %s", i+1, matches[5])
			continue
		}
		preRecord := &request.PreProcessRecord{
			TransactionType: transactionType,
			FundCode:        matches[2],
			Unit:            unit,
			Amount:          amount,
			CreatedAt:       createAt.Unix(),
			Create:          create,
		}
		res.Data = append(res.Data, preRecord)
		res.ParseRecordCount++
	}
	sort.SliceStable(res.Data, func(i, j int) bool {
		return res.Data[i].CreatedAt <= res.Data[j].CreatedAt
	})
	return &res
}

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
		// 检验基金代码是否存在
		_, err := GetFundByCode(ctx, req.FundCode)
		if err != nil {
			logrus.Errorf("get fund by code: %s err: %v", req.FundCode, err)
			return errors.WrapC(errors.CodeBadRequest, err)
		}
		// 处理数据
		req.CalculateLoad()
		if req.CreatedAt == 0 {
			req.CreatedAt = time.Now().Unix()
		}
		// 买入数据
		if req.IsBuy() {
			req.LeftAmount = req.Amount
			err = tx.Create(req).Error
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

var (
	defaultCellStyle = &excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#FFFFFF"},
		},
	}
	grayCellStyle = &excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#D3D3D3"},
		},
	}
	partialCellStyle = &excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#FFEBEB"},
		},
	}
)

func ExportTransactionToExcel(ctx context.Context, exportType string) (*excelize.File, error) {
	f := excelize.NewFile()
	// 定义样式
	grayStyle, _ := f.NewStyle(grayCellStyle)
	defaultStyle, _ := f.NewStyle(defaultCellStyle)
	activeStyle, _ := f.NewStyle(partialCellStyle)

	fundsResp := Funds(ctx, &request.FundListReq{})
	for _, fund := range fundsResp.Items {
		// 获取基金代码对应的交易记录
		transactionReq := &request.TransactionListReq{
			FundCode: fund.Code,
			Pager: &database.Pager{
				Page:     1,
				PageSize: 10000,
			},
		}
		transactions := Transactions(ctx, transactionReq)

		buyTransactions := make(map[string][]*model.Transaction)
		sellTransactions := make(map[int64][]*model.Transaction)
		for _, t := range transactions.Items {
			if t.IsBuy() {
				buyTransactions[t.FundCode] = append(buyTransactions[t.FundCode], t)
			} else {
				sellTransactions[t.OriginalBuyId] = append(sellTransactions[t.OriginalBuyId], t)
			}
		}

		sheetName := fund.Name
		if len(sheetName) > 31 {
			sheetName = sheetName[:31]
		}
		// 创建sheet
		index, err := f.NewSheet(sheetName)
		if err != nil {
			return nil, err
		}
		f.SetActiveSheet(index)

		// 设置表头
		headers := []string{"类型", "单价", "数量", "价格", "时间", "手续费", "剩余数量", "利润", "利润率"}
		for col, header := range headers {
			cell := string(rune(col+'A')) + "1"
			_ = f.SetCellValue(sheetName, cell, header)
		}
		// 按买入单价排序
		buys := buyTransactions[fund.Code]
		sort.Slice(buys, func(i, j int) bool {
			return buys[i].Unit > buys[j].Unit
		})

		row := 1
		for _, buy := range buys {
			if exportType == model.ExportHasLeft && buy.LeftAmount == 0 {
				continue
			}
			row++
			rowName := strconv.Itoa(row)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("A%s", rowName), "买入")
			_ = f.SetCellValue(sheetName, fmt.Sprintf("B%s", rowName), buy.Unit)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("C%s", rowName), buy.Amount)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("D%s", rowName), fmt.Sprintf("%.2f", buy.Price))
			_ = f.SetCellValue(sheetName, fmt.Sprintf("E%s", rowName), time.Unix(buy.CreatedAt, 0).Format("2006-01-02 15:04:05"))
			_ = f.SetCellValue(sheetName, fmt.Sprintf("F%s", rowName), fmt.Sprintf("%.2f", buy.Load))
			_ = f.SetCellValue(sheetName, fmt.Sprintf("G%s", rowName), buy.LeftAmount)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("H%s", rowName), fmt.Sprintf("%.2f", buy.Profit))
			_ = f.SetCellValue(sheetName, fmt.Sprintf("I%s", rowName), fmt.Sprintf("%.2f", buy.ProfitMargin))

			var style int
			switch {
			case buy.LeftAmount == buy.Amount:
				style = defaultStyle
			case buy.LeftAmount == 0:
				style = grayStyle
			default:
				style = activeStyle
			}
			for col := 'A'; col <= 'I'; col++ {
				cell := string(col) + rowName
				_ = f.SetCellStyle(sheetName, cell, cell, style)
			}

			sells := sellTransactions[buy.ID]
			sort.Slice(sells, func(i, j int) bool {
				return sells[i].CreatedAt <= sells[j].CreatedAt
			})

			for _, sell := range sells {
				row++
				rowName = strconv.Itoa(row)
				_ = f.SetCellValue(sheetName, fmt.Sprintf("A%s", rowName), "卖出")
				_ = f.SetCellValue(sheetName, fmt.Sprintf("B%s", rowName), sell.Unit)
				_ = f.SetCellValue(sheetName, fmt.Sprintf("C%s", rowName), sell.Amount)
				_ = f.SetCellValue(sheetName, fmt.Sprintf("D%s", rowName), fmt.Sprintf("%.2f", sell.Price))
				_ = f.SetCellValue(sheetName, fmt.Sprintf("E%s", rowName), time.Unix(sell.CreatedAt, 0).Format("2006-01-02 15:04:05"))
				_ = f.SetCellValue(sheetName, fmt.Sprintf("F%s", rowName), fmt.Sprintf("%.2f", sell.Load))
				//_ = f.SetCellValue(sheetName, fmt.Sprintf("G%s", rowName), sell.LeftAmount)
				_ = f.SetCellValue(sheetName, fmt.Sprintf("H%s", rowName), fmt.Sprintf("%.2f", sell.Profit))
				_ = f.SetCellValue(sheetName, fmt.Sprintf("I%s", rowName), fmt.Sprintf("%.2f", sell.ProfitMargin))

				for col := 'A'; col <= 'I'; col++ {
					cell := string(col) + rowName
					_ = f.SetCellStyle(sheetName, cell, cell, grayStyle)
				}

			}
		}
	}

	defaultSheet := f.GetSheetName(0)
	_ = f.DeleteSheet(defaultSheet)
	return f, nil
}
