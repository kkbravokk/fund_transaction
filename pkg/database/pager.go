package database

import "gorm.io/gorm"

const (
	DefaultPage     = 1
	DefaultPageSize = 20
)

type Pager struct {
	Page       int `json:"page" form:"page"`
	PageSize   int `json:"page_size" form:"page_size"`
	TotalCount int `json:"total_item"`
	TotalPage  int `json:"total_page"`
}

func DefaultPager() *Pager {
	return &Pager{
		Page:     DefaultPage,
		PageSize: DefaultPageSize,
	}
}

func (p *Pager) SetTotalCount(n int) {
	p.TotalCount = n
	if p.PageSize <= 0 {
		p.TotalPage = 0
		return
	}
	p.TotalPage = p.TotalCount / p.PageSize
	if p.TotalCount%p.PageSize > 0 {
		p.TotalPage++
	}
}

func (p *Pager) Load(tx *gorm.DB) *gorm.DB {
	return tx.Offset(p.Offset()).Limit(p.PageSize)
}

func (p *Pager) Offset() int {
	return (p.Page - 1) * p.PageSize
}
