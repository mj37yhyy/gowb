package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/mj37yhyy/gowb/pkg/model"
	logger "github.com/sirupsen/logrus"
	"io/ioutil"
	"time"
)

// 请求进入日志
func RequestInLog(c *gin.Context) {
	c.Set("startExecTime", time.Now())

	bodyBytes, _ := ioutil.ReadAll(c.Request.Body)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes)) // Write body back
	var logFields = logger.Fields{
		"uri":    c.Request.RequestURI,
		"method": c.Request.Method,
		"args":   c.Request.Form,
		"body":   string(bodyBytes),
		"from":   c.ClientIP(),
	}
	action := c.Request.FormValue("Action")
	logger.WithFields(logFields).Infof("%s start", action)
}

// 请求输出日志
func RequestOutLog(c *gin.Context) {
	endExecTime := time.Now()
	//response, _ := c.Get("response")
	st, _ := c.Get("startExecTime")

	startExecTime, _ := st.(time.Time)

	var response = model.NewResponse()
	blw := &bodyLogWriter{
		body:           bytes.NewBufferString(""),
		ResponseWriter: c.Writer,
	}
	c.Writer = blw
	if err := json.Unmarshal(blw.body.Bytes(), &response); err != nil {
		logger.Println(err, "response body can not unmarshal to model.Response struct, body: `%s`", blw.body.Bytes())
	}
	var logFields = logger.Fields{
		"uri":       c.Request.RequestURI,
		"method":    c.Request.Method,
		"from":      c.ClientIP(),
		"response":  response,
		"proc_time": endExecTime.Sub(startExecTime).Seconds(),
	}
	action := c.Request.FormValue("Action")
	logger.WithFields(logFields).Infof("%s end", action)
}

func RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		RequestInLog(c)
		defer RequestOutLog(c)
		c.Next()
	}
}
