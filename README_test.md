# MCP网关测试指南

这个Python测试程序用于测试MCP网关的三个API接口。

## 安装依赖

```bash
pip install requests
```

## 使用方法

### 1. 查看所有服务列表

```bash
python test_mcp_gateway.py --action list-servers
```

### 2. 查看指定服务的工具列表

```bash
python test_mcp_gateway.py --action list-tools --server "Cursor Rules"
```

### 3. 执行指定服务的工具

```bash
python test_mcp_gateway.py --action execute-tool --server "DuckDuckGo Search" --tool "web_search" --params '{"query": "MCP Protocol"}'
```

## 参数说明

- `--host`: 网关服务地址，默认为 http://localhost:8080
- `--action`: 要执行的操作，必选
  - `list-servers`: 列出所有MCP服务
  - `list-tools`: 列出指定服务的工具
  - `execute-tool`: 执行指定服务的工具
- `--server`: 服务名称，用于`list-tools`和`execute-tool`操作
- `--tool`: 工具名称，用于`execute-tool`操作
- `--params`: 工具参数，JSON格式字符串，用于`execute-tool`操作

## 示例

### 获取所有服务

```bash
python test_mcp_gateway.py --action list-servers
```

输出示例:

```json
{
  "code": 200,
  "message": "成功",
  "data": [
    {
      "name": "Cursor Rules",
      "description": "Provides a bridge to Playbooks Rules API..."
    },
    {
      "name": "Web Content Pick",
      "description": "Extracts structured content from web pages..."
    }
  ]
}
```

### 获取服务的工具列表

```bash
python test_mcp_gateway.py --action list-tools --server "Filesystem"
```

输出示例:

```json
{
  "code": 200,
  "message": "成功",
  "data": [
    {
      "name": "read_file",
      "description": "读取文件内容",
      "parameters": {
        "required": ["path"],
        "properties": {
          "path": {
            "type": "string",
            "description": "文件路径"
          }
        }
      }
    }
  ]
}
```

### 执行工具

```bash
python test_mcp_gateway.py --action execute-tool --server "Time" --tool "get_current_time" --params '{}'
```

输出示例:

```json
{
  "code": 200,
  "message": "成功",
  "data": "2025-04-21T16:25:30+08:00"
}
``` 