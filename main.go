package main

import (
	"context"
	"flag"
	"log"
	"os"

	"mcp-gateway/config"
	"mcp-gateway/gateway"
)

func main() {
	// 配置日志
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Println("MCP网关启动...")

	// 解析命令行参数
	configFile := flag.String("config", "mcp-servers-config.json", "MCP服务器配置文件路径")
	port := flag.String("port", "9090", "HTTP服务器端口")
	flag.Parse()

	log.Printf("使用配置文件: %s", *configFile)
	log.Printf("使用端口: %s", *port)

	// 加载配置
	log.Println("正在加载配置...")
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	log.Printf("成功加载配置，包含 %d 个MCP服务", len(cfg.MCPServers))

	// 初始化网关
	log.Println("正在初始化网关...")
	gw, err := gateway.NewGateway(cfg)
	if err != nil {
		log.Fatalf("初始化网关失败: %v", err)
	}
	log.Println("网关初始化成功")

	// 初始化所有MCP客户端
	log.Println("正在初始化MCP客户端...")
	if err := gw.InitializeMCPClients(context.Background()); err != nil {
		log.Fatalf("初始化MCP客户端失败: %v", err)
	}
	log.Println("所有MCP客户端初始化成功")

	// 启动HTTP服务器
	log.Printf("启动MCP网关服务器，监听端口: %s", *port)
	if err := gw.StartHTTPServer(*port); err != nil {
		log.Fatalf("启动HTTP服务器失败: %v", err)
	}
} 