#!/bin/bash

# MCP网关测试脚本

# 禁用代理
unset http_proxy
unset https_proxy
unset HTTP_PROXY
unset HTTPS_PROXY

# 默认参数
HOST="http://localhost:8088"
API_PREFIX="/api"
ACTION=""
SERVER=""
TOOL=""
PARAMS="{}"

# 解析命令行参数
while [[ $# -gt 0 ]]; do
  case $1 in
    --host)
      HOST="$2"
      shift 2
      ;;
    --action)
      ACTION="$2"
      shift 2
      ;;
    --server)
      SERVER="$2"
      shift 2
      ;;
    --tool)
      TOOL="$2"
      shift 2
      ;;
    --params)
      PARAMS="$2"
      shift 2
      ;;
    *)
      echo "未知参数: $1"
      exit 1
      ;;
  esac
done

# 检查必需参数
if [ -z "$ACTION" ]; then
  echo "错误: 必须指定--action参数 (list-servers, list-tools, execute-tool)"
  exit 1
fi

# API基础URL
API_URL="${HOST}${API_PREFIX}"

echo "使用API URL: $API_URL"

# 测试与服务器的连接
echo "测试与服务器的连接..."
curl -s --noproxy "*" -o /dev/null -w "连接状态: %{http_code}\n" "${HOST}/"

# 根据操作执行对应的请求
case $ACTION in
  list-servers)
    echo -e "\n获取服务列表..."
    curl -s --noproxy "*" -v "${API_URL}/servers" | jq 2>/dev/null || cat
    ;;
  
  list-tools)
    if [ -z "$SERVER" ]; then
      echo "错误: 使用list-tools操作时必须指定--server参数"
      exit 1
    fi
    
    echo -e "\n获取服务 '$SERVER' 的工具列表..."
    curl -s --noproxy "*" -v "${API_URL}/servers/${SERVER}/tools" | jq 2>/dev/null || cat
    ;;
  
  execute-tool)
    if [ -z "$SERVER" ] || [ -z "$TOOL" ]; then
      echo "错误: 使用execute-tool操作时必须指定--server和--tool参数"
      exit 1
    fi
    
    echo -e "\n执行服务 '$SERVER' 的工具 '$TOOL'..."
    echo "参数: $PARAMS"
    curl -s --noproxy "*" -v -X POST -H "Content-Type: application/json" -d "$PARAMS" "${API_URL}/servers/${SERVER}/tools/${TOOL}" | jq 2>/dev/null || cat
    ;;
  
  *)
    echo "错误: 未知的操作 $ACTION"
    exit 1
    ;;
esac 