package service

import (
	"context"
	"funds_transaction/internal/model"
	"funds_transaction/internal/request"
	"funds_transaction/pkg/database"
)

func Funds(ctx context.Context, req *request.FundListReq) *request.FundListResp {
	t := &model.Fund{}
	var funds []*model.Fund

	query := database.GetDB(ctx).Table(t.TableName()).Unscoped()
	if req.Code != "" {
		query = query.Where("code like ?", req.Code)
	}
	if req.Name != "" {
		query = query.Where("name like ?", req.Name)
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
	query.Find(&funds).Count(&total)
	req.Pager.SetTotalCount(int(total))

	return &request.FundListResp{Items: funds, Pager: req.Pager}
}

func Fund(ctx context.Context, req *model.Fund) error {
	err := database.GetDB(ctx).Create(req).Error
	return err
}
