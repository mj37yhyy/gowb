package mcp

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mj37yhyy/gowb/pkg/config"
	"github.com/sirupsen/logrus"
)

const MCPProtocolVersion = "2024-11-05"

// Server MCP服务器
type Server struct {
	name        string
	version     string
	description string
	actions     map[string]ActionDef
	tools       []Tool
	authConfig  *AuthConfig
	logger      *logrus.Entry
	config      config.Config
	excludes    map[string]bool
	includes    map[string]bool
}

// NewServer 创建MCP服务器
func NewServer(name, version, description string, actions map[string]ActionDef, authConfig *AuthConfig, conf config.Config) *Server {
	s := &Server{
		name:        name,
		version:     version,
		description: description,
		actions:     actions,
		authConfig:  authConfig,
		config:      conf,
		excludes:    make(map[string]bool),
		includes:    make(map[string]bool),
	}

	// 初始化logger
	s.initLogger()

	// 生成工具列表
	s.generateTools()

	return s
}

// SetExcludes 设置黑名单
func (s *Server) SetExcludes(excludes []string) {
	for _, name := range excludes {
		s.excludes[name] = true
	}
	// 重新生成工具列表
	s.generateTools()
}

// SetIncludes 设置白名单
func (s *Server) SetIncludes(includes []string) {
	for _, name := range includes {
		s.includes[name] = true
	}
	// 重新生成工具列表
	s.generateTools()
}

// initLogger 初始化日志
func (s *Server) initLogger() {
	if s.logger == nil {
		// 使用gowb的日志初始化
		logger := logrus.New()
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetOutput(os.Stderr) // MCP使用stderr输出日志
		logger.SetLevel(logrus.InfoLevel)
		s.logger = logger.WithField("service", "mcp-server")
	}
}

// generateTools 生成工具列表
func (s *Server) generateTools() {
	s.tools = []Tool{}

	for name, action := range s.actions {
		// 检查是否暴露
		if !action.MCPExpose {
			continue
		}

		// 检查黑名单
		if len(s.excludes) > 0 && s.excludes[name] {
			continue
		}

		// 检查白名单
		if len(s.includes) > 0 && !s.includes[name] {
			continue
		}

		// 生成schema
		schema := GenerateSchema(action.InputType)

		// 添加通用的认证字段到schema
		if props, ok := schema["properties"].(map[string]interface{}); ok {
			props["account_id"] = map[string]interface{}{
				"type":        "string",
				"description": "账户ID（可选，如果未设置则使用环境变量）",
			}
			props["region"] = map[string]interface{}{
				"type":        "string",
				"description": "区域（可选，如果未设置则使用环境变量）",
			}
		}

		tool := Tool{
			Name:        name,
			Description: action.Description,
			InputSchema: schema,
		}

		s.tools = append(s.tools, tool)
	}

	s.logger.Infof("Generated %d tools for MCP", len(s.tools))
}

// HandleRequest 处理MCP请求
func (s *Server) HandleRequest(reqData []byte) []byte {
	var req MCPRequest
	if err := json.Unmarshal(reqData, &req); err != nil {
		s.logger.Errorf("Failed to unmarshal request: %v", err)
		return s.errorResponse(nil, -32700, "Parse error", nil)
	}

	s.logger.Infof("Received MCP request: method=%s, id=%v", req.Method, req.ID)

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleListTools(req)
	case "tools/call":
		return s.handleCallTool(req)
	default:
		return s.errorResponse(req.ID, -32601, "Method not found", nil)
	}
}

// handleInitialize 处理初始化请求
func (s *Server) handleInitialize(req MCPRequest) []byte {
	// 提取认证信息（如果有）
	if params := req.Params; params != nil {
		if auth, ok := params["auth"].(map[string]interface{}); ok {
			if s.authConfig == nil {
				s.authConfig = &AuthConfig{
					SessionAuth: make(map[string]string),
				}
			}
			if s.authConfig.SessionAuth == nil {
				s.authConfig.SessionAuth = make(map[string]string)
			}
			for k, v := range auth {
				if str, ok := v.(string); ok {
					s.authConfig.SessionAuth[k] = str
				}
			}
		}
	}

	result := InitializeResult{
		ProtocolVersion: MCPProtocolVersion,
		ServerInfo: ServerInfo{
			Name:    s.name,
			Version: s.version,
		},
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
		},
	}

	return s.successResponse(req.ID, result)
}

// handleListTools 处理列出工具请求
func (s *Server) handleListTools(req MCPRequest) []byte {
	result := ListToolsResult{
		Tools: s.tools,
	}
	return s.successResponse(req.ID, result)
}

// handleCallTool 处理调用工具请求
func (s *Server) handleCallTool(req MCPRequest) []byte {
	params := req.Params
	if params == nil {
		return s.errorResponse(req.ID, -32602, "Invalid params", nil)
	}

	toolName, ok := params["name"].(string)
	if !ok {
		return s.errorResponse(req.ID, -32602, "Missing tool name", nil)
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	s.logger.Infof("Calling tool: %s with arguments: %v", toolName, arguments)

	// 查找action
	action, exists := s.actions[toolName]
	if !exists {
		return s.errorResponse(req.ID, -32602, fmt.Sprintf("Tool not found: %s", toolName), nil)
	}

	// 创建Context
	ctx := CreateContextFromMCP(arguments, s.authConfig, s.logger)

	// 调用Handler
	resp, httpStatus := action.Handler(ctx)

	// 构造MCP响应
	var resultText string
	var isError bool

	if httpStatus >= 400 {
		isError = true
		if resp.Error != nil {
			resultText = fmt.Sprintf("Error: %s - %s", resp.Error.Code, resp.Error.Message)
		} else {
			resultText = fmt.Sprintf("Error: HTTP %d", httpStatus)
		}
	} else {
		// 成功响应，序列化Data
		dataBytes, err := json.MarshalIndent(resp.Data, "", "  ")
		if err != nil {
			resultText = fmt.Sprintf("Success, but failed to serialize result: %v", err)
		} else {
			resultText = string(dataBytes)
		}
	}

	result := CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: resultText,
			},
		},
		IsError: isError,
	}

	return s.successResponse(req.ID, result)
}

// successResponse 构造成功响应
func (s *Server) successResponse(id interface{}, result interface{}) []byte {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	return data
}

// errorResponse 构造错误响应
func (s *Server) errorResponse(id interface{}, code int, message string, data interface{}) []byte {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	respData, _ := json.Marshal(resp)
	return respData
}

// LoadAuthFromEnv 从环境变量加载认证信息
func (s *Server) LoadAuthFromEnv() {
	if s.authConfig == nil {
		s.authConfig = &AuthConfig{}
	}

	if s.authConfig.SessionAuth == nil {
		s.authConfig.SessionAuth = make(map[string]string)
	}

	// 从环境变量读取
	if s.authConfig.AccountIDEnv != "" {
		if accountID := os.Getenv(s.authConfig.AccountIDEnv); accountID != "" {
			s.authConfig.SessionAuth["account_id"] = accountID
			s.logger.Infof("Loaded account_id from env: %s", s.authConfig.AccountIDEnv)
		}
	}

	if s.authConfig.RegionEnv != "" {
		if region := os.Getenv(s.authConfig.RegionEnv); region != "" {
			s.authConfig.SessionAuth["region"] = region
			s.logger.Infof("Loaded region from env: %s", s.authConfig.RegionEnv)
		}
	}
}
