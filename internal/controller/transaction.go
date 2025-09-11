package controller

import (
	"net/http"
	"strconv"
	"time"

	"funds_transaction/internal/model"
	"funds_transaction/internal/request"
	"funds_transaction/internal/service"
	"funds_transaction/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Transactions(c *gin.Context) {
	var req request.TransactionListReq
	if err := c.BindQuery(&req); err != nil {
		logrus.Errorf("transactions err: %v", err)
		c.JSON(http.StatusBadRequest, errors.WrapC(errors.CodeBadRequest, err).Error())
		return
	}
	resp := service.Transactions(c, &req)
	c.JSON(http.StatusOK, resp)
}

func AddTransaction(c *gin.Context) {
	var req *model.Transaction
	if err := c.BindJSON(&req); err != nil {
		logrus.Errorf("add transaction err: %v", err)
		c.JSON(http.StatusBadRequest, errors.WrapC(errors.CodeBadRequest, err).Error())
		return
	}
	id, err := service.AddTransaction(c, req)
	if err != nil {
		logrus.Errorf("add transaction err: %v", err)
		code := errors.ParseCoder(err).HTTPStatus()
		c.JSON(code, err.Error())
		return
	}
	c.JSON(http.StatusOK, &request.AddTransactionsResp{ID: id})
}

func GetTransactionByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.WrapC(errors.CodeBadRequest, err).Error())
		return
	}
	fund, err := service.GetTransactionByID(c, id)
	if err != nil {
		logrus.Errorf("get transaction by id err: %v", err)
		code := errors.ParseCoder(err).HTTPStatus()
		c.JSON(code, err.Error())
		return
	}
	c.JSON(http.StatusOK, fund)
}

func UpdateTransaction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.WrapC(errors.CodeBadRequest, err).Error())
		return
	}
	var req *model.Transaction
	if err = c.BindJSON(&req); err != nil {
		logrus.Errorf("update transaction err: %v", err)
		c.JSON(http.StatusBadRequest, errors.WrapC(errors.CodeBadRequest, err).Error())
		return
	}
	req.ID = id
	err = service.UpdateTransaction(c, req)
	if err != nil {
		logrus.Errorf("update transaction by id err: %v", err)
		code := errors.ParseCoder(err).HTTPStatus()
		c.JSON(code, err.Error())
		return
	}
	c.JSON(http.StatusOK, nil)
}

func DelTransaction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.WrapC(errors.CodeBadRequest, err).Error())
		return
	}
	err = service.DelTransaction(c, id)
	if err != nil {
		logrus.Errorf("del transaction by id err: %v", err)
		code := errors.ParseCoder(err).HTTPStatus()
		c.JSON(code, err.Error())
		return
	}
	c.JSON(http.StatusOK, nil)
}

func ExportTransactionToExcel(c *gin.Context) {
	excelFile, err := service.ExportTransactionToExcel(c)
	if err != nil {
		logrus.Errorf("export transaction err: %v", err)
		code := errors.ParseCoder(err).HTTPStatus()
		c.JSON(code, err.Error())
		return
	}
	fileName := "fund_transactions_" + time.Now().Format("20060102") + ".xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+fileName)

	defer func() {
		_ = excelFile.Close()
	}()

	_, err = excelFile.WriteTo(c.Writer)
	if err != nil {
		logrus.Errorf("export transaction err: %v", err)
		code := errors.ParseCoder(err).HTTPStatus()
		c.JSON(code, err.Error())
		return
	}
}
