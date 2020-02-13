package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/mj37yhyy/gowb/pkg/utils"
	"github.com/mj37yhyy/gowb/pkg/web"
	"github.com/mj37yhyy/gowb/pkg/web/model"
	logger "github.com/sirupsen/logrus"
	"github.com/xiaolin8/lager"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/willf/pad"
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

type Fields struct {
	name  string
	value string
	ref   string
}

type Log struct {
	level       string
	formatter   string
	printMethod bool
	fields      []Fields
}

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取上下文
		c := ctx.Value(web.ContextKey).(context.Context)
		// 获取配置
		conf := c.Value(web.ConfigKey).(*utils.Config)

		// 解析并处理yaml
		var _log = Log{}
		if err := conf.Unmarshal(&_log); err != nil {
			// 日志json格式
			if _log.formatter == "json" {
				// Log as JSON instead of the default ASCII formatter.
				logger.SetFormatter(&logger.JSONFormatter{})
			}

			// Output to stdout instead of the default stderr
			// Can be any io.Writer, see below for File example
			logger.SetOutput(os.Stdout)

			// 日志级别
			// Only log the warning severity or above.
			level, err := logger.ParseLevel(_log.level)
			if err != nil {
				panic(err)
			}
			logger.SetLevel(level)

			// 打印函数与文件
			logger.SetReportCaller(_log.printMethod)

			// 处理自定义字段
			fieldMap := make(map[string]interface{})
			for _, field := range _log.fields {
				if field.ref == "" {
					fieldMap[field.name] = field.value
				} else {
					arr := strings.Split(field.ref, ".")
					htype := arr[2]
					if htype == "header" {
						fieldMap[field.name] = ctx.Request.Header.Get(arr[3])
					} else if htype == "querystring" {
						fieldMap[field.name] = ctx.Param(arr[3])
					}
				}
			}
			contextLogger := logger.WithFields(fieldMap)

			// 将logger对象插入上下文
			c = context.WithValue(c, web.LoggerKey, contextLogger)
			ctx.Set(web.ContextKey, c)
		} else {
			panic(err)
		}
	}
}
