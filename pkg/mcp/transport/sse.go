package transport

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mj37yhyy/gowb/pkg/mcp"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

// SSETransport SSE传输层实现
type SSETransport struct {
	server   *mcp.Server
	endpoint string
	engine   *gin.Engine
	httpSrv  *http.Server
	clients  map[string]*SSEClient
	mu       sync.RWMutex
}

// SSEClient SSE客户端
type SSEClient struct {
	ID       string
	Messages chan []byte
	Done     chan bool
}

// NewSSETransport 创建SSE传输层
func NewSSETransport(server *mcp.Server, endpoint string) *SSETransport {
	return &SSETransport{
		server:   server,
		endpoint: endpoint,
		clients:  make(map[string]*SSEClient),
	}
}

// Start 启动SSE传输
func (t *SSETransport) Start() error {
	gin.SetMode(gin.ReleaseMode)
	t.engine = gin.New()
	t.engine.Use(gin.Recovery())

	// SSE endpoint
	t.engine.GET("/sse", t.handleSSE)

	// Message endpoint (POST)
	t.engine.POST("/message", t.handleMessage)

	// Health check
	t.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	t.httpSrv = &http.Server{
		Addr:    t.endpoint,
		Handler: t.engine,
	}

	log.Printf("[MCP] Starting SSE transport on %s", t.endpoint)

	go func() {
		if err := t.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[MCP] SSE server error: %v", err)
		}
	}()

	return nil
}

// Stop 停止SSE传输
func (t *SSETransport) Stop() error {
	if t.httpSrv != nil {
		return t.httpSrv.Close()
	}
	return nil
}

// handleSSE 处理SSE连接
func (t *SSETransport) handleSSE(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		clientID = fmt.Sprintf("client-%d", time.Now().UnixNano())
	}

	client := &SSEClient{
		ID:       clientID,
		Messages: make(chan []byte, 10),
		Done:     make(chan bool),
	}

	t.mu.Lock()
	t.clients[clientID] = client
	t.mu.Unlock()

	defer func() {
		t.mu.Lock()
		delete(t.clients, clientID)
		t.mu.Unlock()
		close(client.Done)
	}()

	// 设置SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 发送连接成功消息
	c.SSEvent("connected", map[string]string{"client_id": clientID})
	c.Writer.Flush()

	// 保持连接并发送消息
	for {
		select {
		case msg := <-client.Messages:
			c.SSEvent("message", string(msg))
			c.Writer.Flush()
		case <-client.Done:
			return
		case <-c.Request.Context().Done():
			return
		case <-time.After(30 * time.Second):
			// 发送心跳
			c.SSEvent("ping", "")
			c.Writer.Flush()
		}
	}
}

// handleMessage 处理消息请求
func (t *SSETransport) handleMessage(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id required"})
		return
	}

	// 读取请求body
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	// 处理MCP请求
	response := t.server.HandleRequest(body)

	// 如果是SSE客户端，通过SSE发送响应
	t.mu.RLock()
	client, exists := t.clients[clientID]
	t.mu.RUnlock()

	if exists {
		select {
		case client.Messages <- response:
			c.JSON(http.StatusOK, gin.H{"status": "sent"})
		case <-time.After(5 * time.Second):
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "timeout"})
		}
	} else {
		// 如果不是SSE客户端，直接返回响应（兼容普通HTTP调用）
		var respObj map[string]interface{}
		if err := json.Unmarshal(response, &respObj); err == nil {
			c.JSON(http.StatusOK, respObj)
		} else {
			c.Data(http.StatusOK, "application/json", response)
		}
	}
}
