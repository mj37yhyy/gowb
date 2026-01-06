# Gowb MCP Package

这是 gowb 框架的 MCP (Model Context Protocol) 支持包，允许将基于 gowb 的 Web 服务自动暴露为 MCP 服务器。

## 功能特性

- 🚀 **零侵入集成**: 现有 gowb 应用无需修改即可支持 MCP
- 🔄 **自动同步**: Handler 自动转换为 MCP 工具
- 📝 **智能 Schema**: 从 Go 结构体自动生成 JSON Schema
- 🔌 **双传输支持**: stdio（本地）和 SSE（远程）
- 🔐 **灵活认证**: 支持环境变量、Session 和参数级别认证

## 快速开始

### 1. 定义 Actions

```go
// pkg/controller/actions.go
package controller

import (
    "github.com/mj37yhyy/gowb/pkg/mcp"
    "github.com/mj37yhyy/gowb/pkg/web"
)

var Actions = map[string]mcp.ActionDef{
    "CreateUser": {
        Handler:     CreateUserHandler,
        InputType:   CreateUserInput{},
        Description: "创建新用户",
        MCPExpose:   true,
    },
    "GetUser": {
        Handler:     GetUserHandler,
        InputType:   GetUserInput{},
        Description: "获取用户信息",
        MCPExpose:   true,
    },
}

// 向后兼容
var Funcs = mcp.ToHandlerMap(Actions)
```

### 2. 创建 MCP 入口

```go
// cmd/mcp/main.go
package main

import (
    "github.com/mj37yhyy/gowb"
    "github.com/mj37yhyy/gowb/pkg/mcp"
    "myapp/pkg/controller"
)

func main() {
    gowb.BootstrapMCP(gowb.MCPOptions{
        Name:        "my-service",
        Version:     "1.0.0",
        Description: "我的服务",
        ConfigName:  "config",
        ConfigType:  "yaml",
        Actions:     controller.Actions,
        Transport:   mcp.TransportSSE,
        SSEEndpoint: ":8081",
        Auth: mcp.AuthConfig{
            AccountIDEnv: "ACCOUNT_ID",
            RegionEnv:    "REGION",
        },
    })
}
```

### 3. 启动服务

```bash
# SSE 模式
export ACCOUNT_ID="123456"
export REGION="cn-beijing-6"
./mcp-server

# Stdio 模式
export MCP_TRANSPORT="stdio"
./mcp-server
```

## 核心概念

### ActionDef

定义一个可被 MCP 调用的 Action：

```go
type ActionDef struct {
    Handler     web.HandlerFunc // gowb Handler 函数
    InputType   interface{}     // 输入参数类型（用于生成 Schema）
    Description string          // 工具描述
    MCPExpose   bool            // 是否暴露给 MCP
    MCPTags     []string        // 标签（用于分组）
}
```

### 自动 Schema 生成

从 Go 结构体自动生成 JSON Schema：

```go
type CreateUserInput struct {
    Name  string `json:"name" binding:"required" desc:"用户名"`
    Email string `json:"email" binding:"required" desc:"邮箱地址"`
    Age   int    `json:"age" desc:"年龄"`
}
```

生成的 Schema：

```json
{
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "description": "用户名"
    },
    "email": {
      "type": "string",
      "description": "邮箱地址"
    },
    "age": {
      "type": "integer",
      "description": "年龄"
    }
  },
  "required": ["name", "email"]
}
```

### Context 适配

MCP 请求自动转换为 gowb Context：

```go
// MCP 请求
{
  "name": "CreateUser",
  "arguments": {
    "account_id": "123456",
    "name": "Alice",
    "email": "alice@example.com"
  }
}

// 自动转换为 gowb Context
ctx := context.Context{
    HeaderKey: http.Header{
        "account_id": "123456",
    },
    BodyKey: []byte(`{"name":"Alice","email":"alice@example.com"}`),
    ShouldBindKey: func(obj interface{}) error {
        // 自动绑定参数
    },
}
```

## 传输层

### Stdio 传输

适用于本地 AI 工具（Claude Desktop、Cursor 等）：

```go
gowb.BootstrapMCP(gowb.MCPOptions{
    Transport: mcp.TransportStdio,
    // ...
})
```

**特点**:
- 通过 stdin/stdout 通信
- 无需网络配置
- 适合本地开发和调试

### SSE 传输

适用于远程访问和 Web 集成：

```go
gowb.BootstrapMCP(gowb.MCPOptions{
    Transport:   mcp.TransportSSE,
    SSEEndpoint: ":8081",
    // ...
})
```

**特点**:
- HTTP 服务器，支持远程访问
- 支持多客户端连接
- 兼容普通 HTTP 调用

**Endpoints**:
- `GET /sse?client_id=xxx` - SSE 连接
- `POST /message?client_id=xxx` - 发送消息
- `GET /health` - 健康检查

## 认证配置

### 环境变量

```go
Auth: mcp.AuthConfig{
    AccountIDEnv: "ACCOUNT_ID",
    RegionEnv:    "REGION",
}
```

服务启动时自动从环境变量读取。

### Session 级别

客户端在 initialize 时传递：

```json
{
  "method": "initialize",
  "params": {
    "auth": {
      "account_id": "123456",
      "region": "cn-beijing-6"
    }
  }
}
```

### 参数级别

每次调用时覆盖：

```json
{
  "name": "CreateUser",
  "arguments": {
    "account_id": "another-account",
    "name": "Bob"
  }
}
```

## 高级功能

### 过滤工具

**黑名单**（不暴露某些 Action）：

```go
gowb.BootstrapMCP(gowb.MCPOptions{
    ExcludeActions: []string{"InternalAPI", "DebugAction"},
    // ...
})
```

**白名单**（只暴露指定 Action）：

```go
gowb.BootstrapMCP(gowb.MCPOptions{
    IncludeActions: []string{"CreateUser", "GetUser"},
    // ...
})
```

### 自定义 Schema

如果自动生成的 Schema 不满足需求，可以手动指定：

```go
"CustomAction": {
    Handler: CustomHandler,
    InputType: nil,  // 不自动生成
    Description: "自定义 Action",
    MCPExpose: true,
}
```

## 架构设计

```
gowb/pkg/mcp/
├── types.go          # 类型定义
├── server.go         # MCP 服务器核心
├── schema.go         # JSON Schema 生成
├── context.go        # Context 适配器
└── transport/
    ├── stdio.go      # Stdio 传输实现
    └── sse.go        # SSE 传输实现
```

## 最佳实践

### 1. 结构体设计

```go
type MyInput struct {
    // 使用 json tag 定义字段名
    Name string `json:"name" binding:"required"`
    
    // 使用 desc tag 添加描述
    Email string `json:"email" binding:"required" desc:"用户邮箱"`
    
    // 可选字段不加 required
    Age int `json:"age,omitempty" desc:"年龄（可选）"`
}
```

### 2. Handler 实现

```go
func CreateUser(ctx context.Context) (model.Response, web.HttpStatus) {
    // 1. 解析参数
    var input CreateUserInput
    bindFunc := ctx.Value(constant.ShouldBindKey).(func(interface{}) error)
    if err := bindFunc(&input); err != nil {
        return errorResponse(err), http.StatusBadRequest
    }
    
    // 2. 业务逻辑
    user, err := service.CreateUser(input)
    if err != nil {
        return errorResponse(err), http.StatusInternalServerError
    }
    
    // 3. 返回结果
    resp := model.NewResponse()
    resp.SetData(user)
    return *resp, http.StatusOK
}
```

### 3. 错误处理

```go
// 使用标准的 gowb Response 格式
resp := model.NewResponse()
resp.SetError(model.ErrorInfo{
    Code:    "InvalidParameter",
    Message: "用户名不能为空",
})
return *resp, http.StatusBadRequest
```

## 兼容性

- ✅ 与现有 gowb HTTP 服务完全兼容
- ✅ 支持事务（OpenFlatTransaction）
- ✅ 支持中间件
- ✅ 支持数据库自动迁移
- ✅ 支持日志和监控

## License

MIT

