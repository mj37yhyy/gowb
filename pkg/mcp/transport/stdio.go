package transport

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/mj37yhyy/gowb/pkg/mcp"
)

// StdioTransport stdio传输层实现
type StdioTransport struct {
	server *mcp.Server
	reader *bufio.Reader
	writer io.Writer
}

// NewStdioTransport 创建stdio传输层
func NewStdioTransport(server *mcp.Server) *StdioTransport {
	return &StdioTransport{
		server: server,
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

// Start 启动stdio传输
func (t *StdioTransport) Start() error {
	fmt.Fprintln(os.Stderr, "[MCP] Starting stdio transport...")

	for {
		// 读取一行JSON-RPC请求
		line, err := t.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Fprintln(os.Stderr, "[MCP] EOF received, shutting down")
				return nil
			}
			fmt.Fprintf(os.Stderr, "[MCP] Error reading from stdin: %v\n", err)
			return err
		}

		// 跳过空行
		if len(line) <= 1 {
			continue
		}

		// 处理请求
		response := t.server.HandleRequest(line)

		// 写入响应
		if _, err := t.writer.Write(response); err != nil {
			fmt.Fprintf(os.Stderr, "[MCP] Error writing to stdout: %v\n", err)
			return err
		}

		// 写入换行符
		if _, err := t.writer.Write([]byte("\n")); err != nil {
			fmt.Fprintf(os.Stderr, "[MCP] Error writing newline: %v\n", err)
			return err
		}

		// 刷新输出
		if flusher, ok := t.writer.(interface{ Flush() error }); ok {
			if err := flusher.Flush(); err != nil {
				fmt.Fprintf(os.Stderr, "[MCP] Error flushing output: %v\n", err)
			}
		}
	}
}

// Stop 停止stdio传输
func (t *StdioTransport) Stop() error {
	// stdio传输没有需要清理的资源
	return nil
}
