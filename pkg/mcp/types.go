package mcp

import (
	"github.com/mj37yhyy/gowb/pkg/web"
)

// TransportType MCP传输类型
type TransportType string

const (
	TransportStdio TransportType = "stdio"
	TransportSSE   TransportType = "sse"
)

// ActionDef 定义一个Action的完整信息
type ActionDef struct {
	Handler     web.HandlerFunc // 处理函数
	InputType   interface{}     // 输入参数类型，用于生成Schema
	Description string          // 工具描述
	MCPExpose   bool            // 是否暴露给MCP，默认true
	MCPTags     []string        // 标签，用于分组过滤
}

// AuthConfig 认证配置
type AuthConfig struct {
	// 从环境变量读取默认认证信息
	AccountIDEnv string // 账户ID环境变量名，如 "KCF_ACCOUNT_ID"
	RegionEnv    string // 区域环境变量名，如 "KCF_REGION"
	// Session级别的认证信息（从initialize请求中获取）
	SessionAuth map[string]string
}

// MCPRequest MCP标准请求格式
type MCPRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// MCPResponse MCP标准响应格式
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError MCP错误格式
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Tool MCP工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ServerInfo MCP服务器信息
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerCapabilities MCP服务器能力
type ServerCapabilities struct {
	Tools *ToolsCapability `json:"tools,omitempty"`
}

// ToolsCapability 工具能力
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// InitializeResult 初始化响应结果
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Capabilities    ServerCapabilities `json:"capabilities"`
}

// ListToolsResult 列出工具的响应结果
type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

// CallToolResult 调用工具的响应结果
type CallToolResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ContentItem 内容项
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ToHandlerMap 将ActionDef map转换为HandlerFunc map（向后兼容）
func ToHandlerMap(actions map[string]ActionDef) map[string]web.HandlerFunc {
	handlers := make(map[string]web.HandlerFunc)
	for name, action := range actions {
		handlers[name] = action.Handler
	}
	return handlers
}
