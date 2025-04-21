package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// ServerConfig 表示单个MCP服务器配置
type ServerConfig struct {
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env"`
}

// Config 代表整个配置文件
type Config struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

// LoadConfig 从文件加载配置
func LoadConfig(filepath string) (*Config, error) {
	log.Printf("正在读取配置文件: %s", filepath)
	
	// 检查文件是否存在
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", filepath)
	}
	
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}
	
	log.Printf("成功读取配置文件，大小: %d 字节", len(data))
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}
	
	log.Printf("配置文件解析成功，包含 %d 个服务", len(config.MCPServers))
	
	// 打印每个服务的基本信息
	for name, server := range config.MCPServers {
		log.Printf("服务: %s, 类型: %s, 命令: %s", name, server.Type, server.Command)
	}
	
	return &config, nil
} 