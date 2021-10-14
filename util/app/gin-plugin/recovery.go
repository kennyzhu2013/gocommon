package gin_plugin

import (
	"common/util/app"
	"common/util/app/errcode"
	"common/util/process"
	"fmt"
	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				process.DefaultPanicReport.RecoverFromPanic("", "Recovery " + c.Request.URL.RequestURI(), err)
				app.ResponseErr(c, errcode.ErrServerInternal.WithDetail(fmt.Sprint(err)))
				c.Abort()
			}
		}()
		c.Next()
	}
}

