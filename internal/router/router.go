package router

import (
	"funds_transaction/internal/controller"
	"funds_transaction/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	apiV1 := r.Group("/api/v1/funds")

	apiV1.GET("", controller.Funds)               // 获取基金交易记录列表
	apiV1.POST("", controller.AddFundTransaction) // 新增买入/卖出记录
	apiV1.GET("/:id", controller.GetFundByID)     // 获取指定ID的详细记录
	//apiV1.PUT("/:id", controller.Funds)           // 更新某条记录
	//apiV1.DELETE("/:id", controller.Funds)        // 删除指定ID记录
	//apiV1.GET("/stats", controller.Funds)         // 计算累计收益、持仓汇总等
}
