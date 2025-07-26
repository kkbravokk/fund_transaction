package database

import (
	"context"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var DB *gorm.DB

func GetDB(ctx context.Context) *gorm.DB {
	if logrus.GetLevel() == logrus.DebugLevel {
		return DB.WithContext(ctx).Debug()
	}
	return DB.WithContext(ctx)
}

