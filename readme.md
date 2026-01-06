# Gowb

Gowb 是一个基于 Gin 封装的 Go 语言 Web 开发框架，同时内置了对 Model Context Protocol (MCP) 的支持。它旨在简化微服务开发，提供规范的项目结构、统一的响应处理、灵活的配置管理以及开箱即用的 MCP 服务能力。

## ✨ 特性

- **Gin Web 框架封装**:
    - 统一的 `Response` 结构与错误处理。
    - 集成 Prometheus 监控与 Swagger 文档。
    - 内置 GORM 数据库支持 (MySQL)。
    - 结构化日志管理 (Logrus/Zap)。
    - 全局请求追踪 (Trace)。
- **MCP (Model Context Protocol) 支持**:
    - 将 Go 函数轻松暴露为 AI 可调用的 Tools。
    - 支持 `stdio` (本地) 和 `sse` (远程) 传输协议。
    - 自动基于 Struct 生成 JSON Schema。
- **开箱即用**:
    - 优雅停机。
    - 配置文件支持 (YAML/JSON)。
    - 简单的启动引导 `Bootstrap`。

## 📦 安装

```bash
go get github.com/mj37yhyy/gowb
```

## 🚀 快速开始

### 1. Web 服务开发

Gowb 对 Gin 的 Handler 进行了封装，支持直接返回数据对象。

#### 定义 Handler

```go
func HelloHandler(ctx context.Context) (model.Response, web.HttpStatus) {
    resp := model.NewResponse()
    resp.SetData(map[string]string{"message": "Hello World"})
    return *resp, http.StatusOK
}
```

#### 启动 Web 服务

```go
package main

import (
    "context"
    "net/http"
    "github.com/mj37yhyy/gowb"
    "github.com/mj37yhyy/gowb/pkg/model"
    "github.com/mj37yhyy/gowb/pkg/web"
)

func main() {
    g := gowb.Gowb{
        ConfigName: "config", // 配置文件名 (无需后缀)
        ConfigType: "yaml",   // 配置文件类型
        Routers: []web.Router{
            {
                Path:    "/hello",
                Method:  "GET",
                Handler: HelloHandler,
            },
        },
    }

    if err := gowb.Bootstrap(g); err != nil {
        panic(err)
    }
}
```

### 2. MCP 服务开发

Gowb 允许你快速构建 MCP Server，让 AI 模型可以调用你的业务逻辑。

#### 定义 Action

```go
// 需引入 "github.com/mj37yhyy/gowb/pkg/mcp"

var MyActions = map[string]mcp.ActionDef{
    "calculate_sum": {
        Handler: func(ctx context.Context, input struct{ A, B int }) (interface{}, error) {
            return input.A + input.B, nil
        },
        Description: "计算两个数字的和",
        MCPExpose:   true,
    },
}
```

#### 启动 MCP 服务

```go
package main

import (
    "github.com/mj37yhyy/gowb"
    "github.com/mj37yhyy/gowb/pkg/mcp"
)

func main() {
    opts := gowb.MCPOptions{
        Name:        "my-math-tool",
        Version:     "1.0.0",
        Description: "A math assistant MCP server",
        Actions:     MyActions,
        Transport:   mcp.TransportStdio, // 或 mcp.TransportSSE
    }

    if err := gowb.BootstrapMCP(opts); err != nil {
        panic(err)
    }
}
```

## ⚙️ 配置文件

默认支持 `config.yaml`，主要配置项如下：

```yaml
web:
  port: 8080
  runMode: debug
  # logSkipPath: ["/health"]

log:
  level: info    # debug, info, warn, error
  formatter: json # json, text
  printMethod: true

mysql:
  enabled: true
  userName: root
  password: password
  host: 127.0.0.1
  port: 3306
  database: test_db
  maxOpenConns: 100
  maxIdleConns: 10
```

## 📝 统一响应格式

API 默认返回 JSON 格式：

```json
{
  "RequestId": "req-xxx",
  "Error": {
    "Code": "LimitExceeded",
    "Message": "Rate limit exceeded"
  },
  "Data": { ... }
}
```
