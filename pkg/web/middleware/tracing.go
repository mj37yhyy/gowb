package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/mj37yhyy/gowb/pkg/config"
	"github.com/mj37yhyy/gowb/pkg/constant"
)

func Tracing() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取上下文
		c := ctx.Value(constant.ContextKey).(context.Context)
		// 获取配置
		conf := c.Value(constant.ConfigKey).(config.Config)
		openTracingHeaderNames := conf.Trace.Fields
		headers := make(map[string]string)
		for _, headerName := range openTracingHeaderNames {
			headers[headerName] = ctx.Request.Header.Get(headerName)
		}
		c = context.WithValue(c, constant.TraceKey, headers)
		ctx.Set(constant.ContextKey, c)
		// Continue.
		ctx.Next()
	}
}
