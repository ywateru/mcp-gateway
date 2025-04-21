# MCP Gateway

MCP Gateway 是一个Golang编写的Model Context Protocol（MCP）网关程序，用于将MCP服务以HTTP API的形式暴露给外部应用使用。

## 功能特点

- 加载并管理MCP服务器配置
- 自动初始化并缓存MCP服务的工具信息
- 提供简洁的HTTP API来访问MCP服务
- 支持多种MCP服务器并发处理请求

## 安装

```bash
# 下载项目
git clone https://github.com/yourusername/mcp-gateway.git
cd mcp-gateway

# 安装依赖
go mod tidy

# 构建项目
go build -o mcp-gateway
```

## 使用方法

### 启动服务

```bash
# 使用默认配置文件和端口启动
./mcp-gateway

# 指定配置文件和端口
./mcp-gateway -config=/path/to/mcp-servers-config.json -port=8088
```

### 配置文件格式

配置文件`mcp-servers-config.json`示例：

```json
{
  "mcpServers": {
    "Service Name": {
      "description": "服务描述",
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "some-mcp-server"],
      "env": {"KEY": "VALUE"}
    }
  }
}
```

## API 接口

### 1. 获取服务列表

- **URL**: `/api/servers`
- **方法**: `GET`
- **响应示例**:

```json
{
  "code": 200,
  "message": "成功",
  "data": [
    {
      "name": "Cursor Rules",
      "description": "Provides a bridge to Playbooks Rules API..."
    }
  ]
}
```

### 2. 获取服务工具列表

- **URL**: `/api/servers/{server}/tools`
- **方法**: `GET`
- **响应示例**:

```json
{
  "code": 200,
  "message": "成功",
  "data": [
    {
      "name": "tool_name",
      "description": "工具描述",
      "parameters": {...}
    }
  ]
}
```

### 3. 执行工具

- **URL**: `/api/servers/{server}/tools/{tool}`
- **方法**: `POST`
- **请求体**: 工具参数（JSON格式）
- **请求示例**:

```json
{
  "param1": "value1",
  "param2": "value2"
}
```

- **响应示例**:

```json
{
  "code": 200,
  "message": "成功",
  "data": {...}  // 工具执行结果
}
```

## 依赖项

- [github.com/gorilla/mux](https://github.com/gorilla/mux) - HTTP路由
- [github.com/yincongcyincong/mcp-client-go](https://github.com/yincongcyincong/mcp-client-go) - MCP客户端库
