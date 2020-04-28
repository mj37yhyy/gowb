package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/mj37yhyy/gowb/pkg/config"
	"github.com/mj37yhyy/gowb/pkg/constant"
	"github.com/mj37yhyy/gowb/pkg/model"

	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
	"github.com/willf/pad"
	"github.com/xiaolin8/lager"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Logging is a middleware function that logs the each request.
func RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now().UTC()
		path := c.Request.URL.Path

		// Skip for the swagger requests.
		reg := regexp.MustCompile("(/swagger/*)")
		if reg.MatchString(path) {
			return
		}

		// Skip for the health check requests.
		if path == "/health" || path == "/favicon.ico" {
			return
		}

		// Read the Body content
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
		}

		// Restore the io.ReadCloser to its original state
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		// The basic information.
		method := c.Request.Method
		ip := c.ClientIP()

		//log.Debugf("New request come in, path: %s, Method: %s, body `%s`", path, method, string(bodyBytes))
		blw := &bodyLogWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = blw

		// Continue.
		c.Next()

		// Calculates the latency.
		end := time.Now().UTC()
		latency := end.Sub(start)

		var code, message string

		// get code and message
		var response = model.NewResponse()
		if err := json.Unmarshal(blw.body.Bytes(), &response); err != nil {
			logger.Println(err, "response body can not unmarshal to model.Response struct, body: `%s`", blw.body.Bytes())
			code = "500"
			message = err.Error()
		} else {
			code = "200"
			message = "ok"
		}

		lager.Infof("%-13s | %-12s | %s %s | {code: %d, message: %s}", latency, ip, pad.Right(method, 5, ""), path, code, message)
	}
}

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取上下文
		c := ctx.Value(constant.ContextKey).(context.Context)
		// 获取配置
		conf := c.Value(constant.ConfigKey).(config.Config)

		// 处理自定义字段
		fieldMap := make(map[string]interface{})
		for _, field := range conf.Log.Fields {
			if field.Ref == "" {
				fieldMap[field.Name] = field.Value
			} else {
				arr := strings.Split(field.Ref, ".")
				htype := arr[2]
				if htype == "header" {
					fieldMap[field.Name] = ctx.Request.Header.Get(arr[3])
				} else if htype == "querystring" {
					fieldMap[field.Name] = ctx.Query(arr[3])
				}
			}
		}
		contextLogger := logger.WithFields(fieldMap)

		// 将logger对象插入上下文
		c = context.WithValue(c, constant.LoggerKey, contextLogger)

		// audit func
		auditFunc := func(field map[string]interface{}, format string, args ...interface{}) {
			auditField := make(map[string]interface{})
			for key, value := range fieldMap {
				auditField[key] = value
			}
			auditField["AuditLog"] = true
			for key, value := range field {
				auditField[key] = value
			}

			auditLogger := logger.WithFields(auditField)
			auditLogger.Infof(format, args...)
		}

		// 将audit function 对象插入上下文
		c = context.WithValue(c, constant.AuditLoggerKey, auditFunc)

		ctx.Set(constant.ContextKey, c)
		// Continue.
		ctx.Next()
	}
}
