#!/bin/bash

# 登录测试脚本
# 使用方法: ./test_login.sh [username] [password]

# 默认配置
HOST="127.0.0.1"
PORT="8000"
BASE_URL="http://${HOST}:${PORT}"

# 从参数获取用户名和密码，如果没有提供则使用默认值
USERNAME=${1:-"admin"}
PASSWORD=${2:-"123456"}

echo "========================================="
echo "测试登录接口"
echo "========================================="
echo "服务器地址: ${BASE_URL}"
echo "用户名: ${USERNAME}"
echo "密码: ${PASSWORD}"
echo ""

# 测试登录
echo "发送登录请求..."
RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
  -X POST "${BASE_URL}/v1/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${USERNAME}\",
    \"password\": \"${PASSWORD}\"
  }")

# 分离响应体和 HTTP 状态码
HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed '/HTTP_CODE/d')

echo ""
echo "HTTP 状态码: ${HTTP_CODE}"
echo "响应内容:"
echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"

echo ""
echo "========================================="

# 检查是否成功
if [ "$HTTP_CODE" = "200" ]; then
  echo "✅ 登录成功！"
  # 提取 token
  TOKEN=$(echo "$BODY" | jq -r '.data.accessToken' 2>/dev/null)
  if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    echo "Token: ${TOKEN}"
    echo ""
    echo "可以使用以下命令测试需要认证的接口:"
    echo "curl -H \"Authorization: Bearer ${TOKEN}\" ${BASE_URL}/v1/menus"
  fi
else
  echo "❌ 登录失败"
  echo "请检查:"
  echo "1. 服务器是否已启动 (go run ./cmd/server)"
  echo "2. 用户名和密码是否正确"
  echo "3. 数据库中是否存在该用户"
fi

