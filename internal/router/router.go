package router

import (
	"funds_transaction/internal/controller"
	"funds_transaction/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	apiV1 := r.Group("/api/v1/transaction")

	apiV1.GET("", controller.Transactions)           // 获取基金交易记录列表
	apiV1.POST("", controller.AddTransaction)        // 新增买入/卖出记录
	apiV1.GET("/:id", controller.GetTransactionByID) // 获取指定ID的详细记录
	apiV1.PUT("/:id", controller.UpdateTransaction)  // 更新某条记录
	apiV1.DELETE("/:id", controller.DelTransaction)  // 删除指定ID记录
	//apiV1.GET("/stats", controller.Funds)         // 计算累计收益、持仓汇总等

	fundV1 := r.Group("api/v1/fund")

	fundV1.GET("", controller.Funds) // 获取基金代码
	fundV1.POST("", controller.Fund) // 新增基金代码
}
