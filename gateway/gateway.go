package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yincongcyincong/mcp-client-go/clients"
	"github.com/yincongcyincong/mcp-client-go/clients/param"

	"mcp-gateway/config"
)

// 工具缓存
type ToolsCache struct {
	Tools map[string][]mcp.Tool
	mu    sync.RWMutex
}

// Gateway 网关结构体
type Gateway struct {
	config     *config.Config
	toolsCache *ToolsCache
}

// 响应结构体
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewGateway 创建新的网关实例
func NewGateway(cfg *config.Config) (*Gateway, error) {
	log.Println("创建网关实例...")
	
	gw := &Gateway{
		config: cfg,
		toolsCache: &ToolsCache{
			Tools: make(map[string][]mcp.Tool),
		},
	}
	
	log.Println("网关实例创建成功")
	return gw, nil
}

// InitializeMCPClients 初始化所有MCP客户端
func (g *Gateway) InitializeMCPClients(ctx context.Context) error {
	log.Println("开始初始化MCP客户端...")
	var clientConfigs []*param.MCPClientConf

	// 初始化所有配置中的服务
	for serverName, serverConfig := range g.config.MCPServers {
		log.Printf("正在配置服务: %s", serverName)
		var conf *param.MCPClientConf

		// 根据类型创建不同的客户端配置
		if serverConfig.Type == "stdio" {
			log.Printf("服务 %s 使用stdio类型", serverName)
			
			// 转换环境变量
			env := make([]string, 0, len(serverConfig.Env))
			for k, v := range serverConfig.Env {
				env = append(env, fmt.Sprintf("%s=%s", k, v))
				log.Printf("添加环境变量: %s=%s", k, v)
			}

			log.Printf("初始化stdio客户端: %s, 命令: %s, 参数: %v", 
				serverName, serverConfig.Command, serverConfig.Args)
			
			conf = clients.InitStdioMCPClient(
				serverName,
				serverConfig.Command,
				env,
				serverConfig.Args,
				mcp.InitializeRequest{},
				nil,
				nil,
			)
		} else {
			// 未来可以支持其他类型的客户端
			log.Printf("不支持的服务器类型: %s，服务名: %s", serverConfig.Type, serverName)
			continue
		}

		clientConfigs = append(clientConfigs, conf)
		log.Printf("已添加客户端配置: %s", serverName)
	}

	if len(clientConfigs) == 0 {
		log.Println("没有可用的MCP客户端配置，跳过注册过程")
		return nil
	}

	// 注册所有客户端（添加超时控制）
	log.Printf("开始注册 %d 个MCP客户端...", len(clientConfigs))
	
	// 创建一个带超时的上下文
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	
	// 使用goroutine执行注册，可以通过通道获取结果
	resultCh := make(chan []error, 1)
	go func() {
		errs := clients.RegisterMCPClient(ctxWithTimeout, clientConfigs)
		resultCh <- errs
	}()
	
	// 等待注册完成或超时
	select {
	case errs := <-resultCh:
		// 处理注册结果
		if len(errs) > 0 {
			hasError := false
			for i, err := range errs {
				if err != nil {
					hasError = true
					log.Printf("注册客户端 %s 时发生错误: %v", clientConfigs[i].Name, err)
				}
			}
			if hasError {
				log.Println("部分客户端注册失败，但将继续处理")
			}
		}
		log.Println("MCP客户端注册过程完成")
	case <-ctxWithTimeout.Done():
		return fmt.Errorf("注册MCP客户端超时（60秒），可能是命令行工具不存在或执行异常")
	}

	// 初始化后立即获取所有服务的工具信息并缓存
	log.Println("开始缓存所有服务的工具信息...")
	for serverName := range g.config.MCPServers {
		log.Printf("正在获取服务 %s 的工具信息...", serverName)
		// 添加超时控制
		ctxWithToolTimeout, cancelTool := context.WithTimeout(ctx, 10*time.Second)
		
		err := g.cacheToolsForServer(ctxWithToolTimeout, serverName)
		if err != nil {
			log.Printf("获取服务 %s 的工具信息失败: %v", serverName, err)
			// 继续处理其他服务，不中断
		}
		
		cancelTool()
	}
	log.Println("工具信息缓存完成")

	return nil
}

// cacheToolsForServer 缓存指定服务的工具信息
func (g *Gateway) cacheToolsForServer(ctx context.Context, serverName string) error {
	log.Printf("开始缓存服务 %s 的工具信息...", serverName)
	
	// 获取客户端
	client, err := clients.GetMCPClient(serverName)
	if err != nil {
		return fmt.Errorf("获取客户端失败: %v", err)
	}
	log.Printf("成功获取服务 %s 的客户端", serverName)

	// 获取工具列表
	log.Printf("正在获取服务 %s 的工具列表...", serverName)
	tools, err := client.GetAllTools(ctx, "")
	if err != nil {
		return fmt.Errorf("获取工具列表失败: %v", err)
	}
	log.Printf("成功获取服务 %s 的工具列表，共 %d 个工具", serverName, len(tools))

	// 缓存工具信息
	g.toolsCache.mu.Lock()
	g.toolsCache.Tools[serverName] = tools
	g.toolsCache.mu.Unlock()
	log.Printf("服务 %s 的工具信息已缓存", serverName)

	return nil
}

// StartHTTPServer 启动HTTP服务器
func (g *Gateway) StartHTTPServer(port string) error {
	log.Println("开始配置HTTP服务器路由...")
	r := mux.NewRouter()

	// 添加一个根路径处理器，便于调试
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("收到根路径请求: %s", r.URL.Path)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("MCP Gateway is running\n"))
	})

	// API路由
	api := r.PathPrefix("/api").Subrouter()
	
	// 1. 返回服务列表和描述
	api.HandleFunc("/servers", g.listServersHandler).Methods("GET")
	log.Printf("注册路由: GET /api/servers")
	
	// 2. 返回指定服务的工具
	api.HandleFunc("/servers/{server}/tools", g.listToolsHandler).Methods("GET")
	log.Printf("注册路由: GET /api/servers/{server}/tools")
	
	// 3. 执行指定服务的工具
	api.HandleFunc("/servers/{server}/tools/{tool}", g.executeToolHandler).Methods("POST")
	log.Printf("注册路由: POST /api/servers/{server}/tools/{tool}")

	// 设置CORS头和请求日志中间件
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			
			// 设置CORS头
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			// 处理预检请求
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			// 记录请求开始
			log.Printf("[请求开始] %s %s", r.Method, r.URL.Path)
			
			// 调用下一个处理器
			next.ServeHTTP(w, r)
			
			// 记录请求结束和耗时
			log.Printf("[请求结束] %s %s 耗时: %v", r.Method, r.URL.Path, time.Since(startTime))
		})
	})

	// 打印监听地址
	log.Printf("HTTP服务器开始监听: :%s", port)
	
	// 启动HTTP服务器
	return http.ListenAndServe(":"+port, r)
}

// 处理函数1: 列出所有服务
func (g *Gateway) listServersHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("处理请求: 列出所有服务")

	type ServerInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	// 从配置中返回服务列表
	serverList := make([]ServerInfo, 0, len(g.config.MCPServers))
	for name, server := range g.config.MCPServers {
		log.Printf("添加服务到列表: %s", name)
		serverList = append(serverList, ServerInfo{
			Name:        name,
			Description: server.Description,
		})
	}

	log.Printf("返回服务列表，共 %d 个服务", len(serverList))
	g.writeJSON(w, Response{
		Code:    200,
		Message: "成功",
		Data:    serverList,
	})
}

// 处理函数2: 列出指定服务的工具
func (g *Gateway) listToolsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverName := vars["server"]
	log.Printf("处理请求: 列出服务 %s 的工具", serverName)

	// 先从缓存中查找
	g.toolsCache.mu.RLock()
	tools, exists := g.toolsCache.Tools[serverName]
	g.toolsCache.mu.RUnlock()

	if !exists {
		log.Printf("服务 %s 的工具信息未缓存，尝试获取", serverName)
		// 如果缓存中不存在，尝试重新获取
		ctx := r.Context()
		// 添加超时控制
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		
		err := g.cacheToolsForServer(ctxWithTimeout, serverName)
		if err != nil {
			log.Printf("获取服务 %s 的工具失败: %v", serverName, err)
			g.writeError(w, 404, fmt.Sprintf("获取服务 %s 的工具失败: %v", serverName, err))
			return
		}

		g.toolsCache.mu.RLock()
		tools = g.toolsCache.Tools[serverName]
		g.toolsCache.mu.RUnlock()
	}

	if len(tools) == 0 {
		log.Printf("警告: 服务 %s 没有工具", serverName)
		g.writeError(w, 404, fmt.Sprintf("服务 %s 没有可用的工具", serverName))
		return
	}

	log.Printf("返回服务 %s 的工具列表，共 %d 个工具", serverName, len(tools))
	g.writeJSON(w, Response{
		Code:    200,
		Message: "成功",
		Data:    tools,
	})
}

// 处理函数3: 执行指定服务的工具
func (g *Gateway) executeToolHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverName := vars["server"]
	toolName := vars["tool"]
	log.Printf("处理请求: 执行服务 %s 的工具 %s", serverName, toolName)

	// 解析请求参数
	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Printf("解析请求参数失败: %v", err)
		g.writeError(w, 400, "无效的请求参数")
		return
	}
	log.Printf("参数: %v", params)

	// 获取MCP客户端
	log.Printf("获取服务 %s 的客户端", serverName)
	client, err := clients.GetMCPClient(serverName)
	if err != nil {
		log.Printf("服务 %s 不存在: %v", serverName, err)
		g.writeError(w, 404, fmt.Sprintf("服务 %s 不存在", serverName))
		return
	}

	// 执行工具
	log.Printf("开始执行工具 %s", toolName)
	ctx := r.Context()
	// 添加超时控制
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	result, err := client.ExecTools(ctxWithTimeout, toolName, params)
	if err != nil {
		log.Printf("执行工具失败: %v", err)
		g.writeError(w, 500, fmt.Sprintf("执行工具失败: %v", err))
		return
	}

	log.Printf("工具执行成功，返回结果")
	g.writeJSON(w, Response{
		Code:    200,
		Message: "成功",
		Data:    result,
	})
}

// 写入JSON响应
func (g *Gateway) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("写入JSON响应失败: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// 写入错误响应
func (g *Gateway) writeError(w http.ResponseWriter, code int, message string) {
	log.Printf("返回错误: [%d] %s", code, message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := Response{
		Code:    code,
		Message: message,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("写入错误响应失败: %v", err)
	}
} 