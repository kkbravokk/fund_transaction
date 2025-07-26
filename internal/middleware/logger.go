package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const LogIDHeader = "X-LOGID"

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		rid := c.GetHeader(LogIDHeader)
		if rid == "" {
			rid = uuid.NewString()
		}

		blw := &bodyLogWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = blw

		reqBody, err := ginRequestBody(c)
		if err != nil {
			logrus.Errorf("%s: %s, request URI: %s, error: %+v", LogIDHeader, rid, c.Request.RequestURI, err)
		} else {
			logrus.Infof("%s: %s, request URI: %s, request body: %s", LogIDHeader, rid, c.Request.RequestURI, reqBody)
		}

		c.Next()

		costTime := time.Since(t)
		respBody := blw.body.String()
		logrus.Infof("%s: %s, response body: %s, cost: %d ms", LogIDHeader, rid, respBody, costTime)
	}
}

func ginRequestBody(c *gin.Context) (body []byte, err error) {
	body, err = io.ReadAll(c.Request.Body)
	if err != nil {
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	return
}
