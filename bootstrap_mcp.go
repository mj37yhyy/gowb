package gowb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/mj37yhyy/gowb/pkg/config"
	"github.com/mj37yhyy/gowb/pkg/db"
	gowbLog "github.com/mj37yhyy/gowb/pkg/log"
	"github.com/mj37yhyy/gowb/pkg/mcp"
	"github.com/mj37yhyy/gowb/pkg/mcp/transport"
	"github.com/mj37yhyy/gowb/pkg/utils"
)

const mcpLogo = `
 _____   _____   _          __  _____       __  __   _____  ____  
/  ___| /  _  \ | |        / / |  _  \     |  \/  | /  ___||  _ \ 
| |     | | | | | |  __   / /  | |_| |     | |\/| | | |    | |_) |
| |  _  | | | | | | /  | / /   |  _  {     | |  | | | |    |  __/ 
| |_| | | |_| | | |/   |/ /    | |_| |     | |  | | | |___ | |    
\_____/ \_____/ |___/|___/     |_____/     |__|  |__| \____||_|    
`

// MCPOptions MCP服务器配置选项
type MCPOptions struct {
	Name             string                   // 服务名称
	Version          string                   // 服务版本
	Description      string                   // 服务描述
	ConfigName       string                   // 配置文件名（可选）
	ConfigType       string                   // 配置文件类型（可选）
	Config           config.Config            // 配置对象（可选）
	Actions          map[string]mcp.ActionDef // Action定义
	Transport        mcp.TransportType        // 传输类型：stdio或sse
	SSEEndpoint      string                   // SSE模式下的监听地址，如":8081"
	Auth             mcp.AuthConfig           // 认证配置
	AutoCreateTables []interface{}            // 自动创建的数据库表
	ExcludeActions   []string                 // 黑名单：不暴露的Action
	IncludeActions   []string                 // 白名单：只暴露这些Action（如果设置）
}

// BootstrapMCP 启动MCP服务器
func BootstrapMCP(opts MCPOptions) error {
	fmt.Println(mcpLogo)

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	// 加载配置
	var conf config.Config
	if opts.ConfigName != "" && opts.ConfigType != "" {
		cu, err := utils.NewConfig(opts.ConfigName, opts.ConfigType)
		if err != nil {
			return err
		}
		if err := cu.Unmarshal(&conf); err != nil {
			return err
		}
	} else if unsafe.Sizeof(opts.Config) > 0 {
		conf = opts.Config
	} else {
		return errors.New("ConfigName and ConfigType is empty!")
	}

	// 初始化上下文
	ctx := context.Background()
	ctx = context.WithValue(ctx, "config", conf)

	// 初始化MySQL（如果配置了）
	if conf.Mysql.Enabled {
		if err := db.InitMysql(ctx); err != nil {
			return err
		}
		// 自动建表
		for _, t := range opts.AutoCreateTables {
			db.DB.AutoMigrate(t)
		}
	}

	// 初始化日志
	if err := gowbLog.InitLogger(ctx); err != nil {
		return err
	}

	// 验证Actions
	if opts.Actions == nil || len(opts.Actions) == 0 {
		return errors.New("Actions cannot be empty")
	}

	// 设置默认值
	if opts.Name == "" {
		opts.Name = "gowb-mcp-server"
	}
	if opts.Version == "" {
		opts.Version = "1.0.0"
	}
	if opts.Transport == "" {
		opts.Transport = mcp.TransportSSE
	}
	if opts.Transport == mcp.TransportSSE && opts.SSEEndpoint == "" {
		opts.SSEEndpoint = ":8081"
	}

	// 创建MCP服务器
	server := mcp.NewServer(opts.Name, opts.Version, opts.Description, opts.Actions, &opts.Auth, conf)

	// 设置过滤规则
	if len(opts.ExcludeActions) > 0 {
		server.SetExcludes(opts.ExcludeActions)
	}
	if len(opts.IncludeActions) > 0 {
		server.SetIncludes(opts.IncludeActions)
	}

	// 从环境变量加载认证信息
	server.LoadAuthFromEnv()

	// 根据传输类型启动
	switch opts.Transport {
	case mcp.TransportStdio:
		return startStdioTransport(server)
	case mcp.TransportSSE:
		return startSSETransport(server, opts.SSEEndpoint)
	default:
		return fmt.Errorf("unsupported transport type: %s", opts.Transport)
	}
}

// startStdioTransport 启动stdio传输
func startStdioTransport(server *mcp.Server) error {
	t := transport.NewStdioTransport(server)
	return t.Start()
}

// startSSETransport 启动SSE传输
func startSSETransport(server *mcp.Server, endpoint string) error {
	t := transport.NewSSETransport(server, endpoint)
	if err := t.Start(); err != nil {
		return err
	}

	// 阻塞等待信号（复用web包的信号处理）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutdown MCP Server...")

	// 优雅关闭
	return t.Stop()
}
