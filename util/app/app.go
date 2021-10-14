package app

import (
	"common/util/app/errcode"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	RId   = "router.id"
	Data  = "router.data"
	Code  = "router.code"
	Error = "router.error"
)

func Response(c *gin.Context, data interface{}) {
	Response2(c, http.StatusOK, data)
}

func Response2(c *gin.Context, statusCode int, data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	c.JSON(statusCode, data)
}

func ResponseStream(c *gin.Context, data []byte) {
	c.Data(http.StatusOK, "application/octet-stream", data)
}

func ResponseErr(c *gin.Context, err *errcode.Error) {
	ResponseErr2(c, http.StatusInternalServerError, err)
}

func ResponseErr2(c *gin.Context, statusCode int, err *errcode.Error) {
	if err == nil {
		err = errcode.ErrServerInternal
	}
	c.JSON(statusCode, err)
}