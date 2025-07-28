package model

type Fund struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (f *Fund) TableName() string {
	return "fund"
}
