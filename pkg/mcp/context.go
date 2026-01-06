package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/mj37yhyy/gowb/pkg/constant"
	"github.com/sirupsen/logrus"
	"net/http"
)

// CreateContextFromMCP 从MCP请求创建gowb标准Context
func CreateContextFromMCP(args map[string]interface{}, authConfig *AuthConfig, logger *logrus.Entry) context.Context {
	ctx := context.Background()

	// 添加logger
	ctx = context.WithValue(ctx, constant.LoggerKey, logger)

	// 构造Header
	header := http.Header{}

	// 从参数中提取header信息（如果有）
	if accountID, ok := args["account_id"].(string); ok {
		header.Set("account_id", accountID)
		delete(args, "account_id") // 从参数中移除，避免重复
	} else if authConfig != nil && authConfig.SessionAuth != nil {
		// 从Session认证信息中获取
		if accountID, ok := authConfig.SessionAuth["account_id"]; ok {
			header.Set("account_id", accountID)
		}
	}

	if region, ok := args["region"].(string); ok {
		header.Set("region", region)
		delete(args, "region")
	} else if authConfig != nil && authConfig.SessionAuth != nil {
		if region, ok := authConfig.SessionAuth["region"]; ok {
			header.Set("region", region)
		}
	}

	if requestID, ok := args["request_id"].(string); ok {
		header.Set("X-REQUEST-ID", requestID)
		delete(args, "request_id")
	}

	ctx = context.WithValue(ctx, constant.HeaderKey, header)

	// 构造Body（剩余的参数作为JSON body）
	bodyBytes, _ := json.Marshal(args)
	ctx = context.WithValue(ctx, constant.BodyKey, bodyBytes)

	// 构造Params（用于兼容某些Handler）
	params := make(map[string][]string)
	for k, v := range args {
		if str, ok := v.(string); ok {
			params[k] = []string{str}
		}
	}
	ctx = context.WithValue(ctx, constant.ParamsKey, params)

	// 添加ShouldBind函数（用于参数绑定）
	ctx = context.WithValue(ctx, constant.ShouldBindKey, func(obj interface{}) error {
		return json.Unmarshal(bodyBytes, obj)
	})

	ctx = context.WithValue(ctx, constant.ShouldBindWithKey,
		func(obj interface{}, bt constant.BindingType) error {
			if bt == constant.BindingJson {
				return json.Unmarshal(bodyBytes, obj)
			}
			// 其他类型暂不支持
			return json.Unmarshal(bodyBytes, obj)
		})

	// 构造Request（某些Handler可能需要）
	req, _ := http.NewRequest("POST", "/mcp", bytes.NewReader(bodyBytes))
	req.Header = header
	ctx = context.WithValue(ctx, constant.RequestKey, req)

	return ctx
}
