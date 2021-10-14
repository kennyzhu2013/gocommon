package gin_plugin

import (
	"bytes"
	log "common/log/newlog"
	"common/util/app"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"strings"
)

type AccessLogWriter struct {
	gin.ResponseWriter
	responseBody *bytes.Buffer
}

func (w *AccessLogWriter) Write(data []byte) (n int, err error) {
	contentType := w.ResponseWriter.Header().Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if n, err = w.responseBody.Write(data); err != nil {
			return n, err
		}
	}
	return w.ResponseWriter.Write(data)
}

func (w *AccessLogWriter) WriteString(s string) (n int, err error) {
	contentType := w.ResponseWriter.Header().Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if n, err = w.responseBody.WriteString(s); err != nil {
			return n, err
		}
	}
	return w.ResponseWriter.WriteString(s)
}

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, _ := c.GetRawData()
		requestUri := c.Request.URL.RequestURI()
		if len(body) > 512 {
			log.Infof("%s %s, requestBody is bigData", c.Request.Method, requestUri)
		} else {
			log.Infof("%s %s, requestBody is %s", c.Request.Method, requestUri, string(body))
		}

		writer := &AccessLogWriter{
			responseBody:   bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = writer
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		c.Next()

		rId, _ := c.Get(app.RId)
		log.InfoK(rId.(string), "%s done with %d, responseBody is %s", requestUri, c.Writer.Status(), writer.responseBody.String())
	}
}