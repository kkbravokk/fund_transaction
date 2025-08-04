package controller

import (
	"net/http"

	"funds_transaction/internal/model"
	"funds_transaction/internal/request"
	"funds_transaction/internal/service"
	"funds_transaction/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Funds(c *gin.Context) {
	var req request.FundListReq
	if err := c.BindQuery(&req); err != nil {
		logrus.Errorf("funds err: %v", err)
		c.JSON(http.StatusBadRequest, errors.WrapC(errors.CodeBadRequest, err).Error())
		return
	}
	resp := service.Funds(c, &req)
	c.JSON(http.StatusOK, resp)
}

func Fund(c *gin.Context) {
	var req *model.Fund
	if err := c.BindJSON(&req); err != nil {
		logrus.Errorf("add fund err: %v", err)
		c.JSON(http.StatusBadRequest, errors.WrapC(errors.CodeBadRequest, err).Error())
		return
	}
	err := service.Fund(c, req)
	if err != nil {
		logrus.Errorf("add fund err: %v", err)
		code := errors.ParseCoder(err).HTTPStatus()
		c.JSON(code, err.Error())
		return
	}
	c.JSON(http.StatusOK, req)
}
