package database

import "gorm.io/gorm"

const (
	DefaultOrderBy = "id"
	DefaultSort    = "desc"
)

type Sorter struct {
	OrderBy string `json:"order_by" form:"order_by"`
	Sort    string `json:"sort" form:"sort"`
}

func DefaultSorter() *Sorter {
	return &Sorter{
		OrderBy: DefaultOrderBy,
		Sort:    DefaultSort,
	}
}

func (s *Sorter) Load(tx *gorm.DB) *gorm.DB {
	return tx.Order(QuoteSQLOrder(s.OrderBy + " " + s.Sort))
}
