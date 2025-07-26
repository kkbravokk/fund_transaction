package controller

import (
	"net/http"
	"strconv"

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

func AddFundTransaction(c *gin.Context) {
	var req *model.Transaction
	if err := c.BindJSON(&req); err != nil {
		logrus.Errorf("add fund transaction err: %v", err)
		c.JSON(http.StatusBadRequest, errors.WrapC(errors.CodeBadRequest, err).Error())
		return
	}
	id, err := service.AddFundTransaction(c, req)
	if err != nil {
		logrus.Errorf("add fund transaction err: %v", err)
		code := errors.ParseCoder(err).HTTPStatus()
		c.JSON(code, err.Error())
		return
	}
	c.JSON(http.StatusOK, &request.AddFundResp{ID: id})
}

func GetFundByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.WrapC(errors.CodeBadRequest, err).Error())
		return
	}
	fund, err := service.GetFundByID(c, id)
	if err != nil {
		logrus.Errorf("get fund by id err: %v", err)
		code := errors.ParseCoder(err).HTTPStatus()
		c.JSON(code, err.Error())
		return
	}
	c.JSON(http.StatusOK, fund)
}
