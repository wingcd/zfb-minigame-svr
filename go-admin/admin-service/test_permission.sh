#!/bin/bash

echo "=== 权限控制测试 ==="
echo ""

# 启动服务器（后台运行）
echo "1. 启动服务器..."
./admin-service &
SERVER_PID=$!
sleep 3

# 测试1: 登录获取token
echo "2. 测试登录获取token..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/admin/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }')

echo "登录响应: $LOGIN_RESPONSE"
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "❌ 登录失败，无法获取token"
    kill $SERVER_PID
    exit 1
fi

echo "✅ 登录成功，获取到token: ${TOKEN:0:20}..."
echo ""

# 测试2: 使用token访问需要权限的接口
echo "3. 测试访问应用管理接口（需要app_manage权限）..."
APP_RESPONSE=$(curl -s -X POST http://localhost:8080/app/getAll \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}')

echo "应用管理接口响应: $APP_RESPONSE"

# 检查是否返回权限错误
if echo "$APP_RESPONSE" | grep -q "权限不足"; then
    echo "✅ 权限控制正常工作 - 返回了权限不足错误"
else
    echo "❓ 权限控制状态: $(echo $APP_RESPONSE | grep -o '"code":[^,]*' | head -1)"
fi

echo ""

# 测试3: 访问统计接口（需要stats_view权限）
echo "4. 测试访问统计接口（需要stats_view权限）..."
STATS_RESPONSE=$(curl -s -X POST http://localhost:8080/stat/dashboard \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}')

echo "统计接口响应: $STATS_RESPONSE"

# 检查是否返回权限错误
if echo "$STATS_RESPONSE" | grep -q "权限不足"; then
    echo "✅ 权限控制正常工作 - 返回了权限不足错误"
else
    echo "❓ 权限控制状态: $(echo $STATS_RESPONSE | grep -o '"code":[^,]*' | head -1)"
fi

echo ""

# 测试4: 访问不需要特定权限的接口
echo "5. 测试访问认证相关接口（不需要特定权限）..."
PROFILE_RESPONSE=$(curl -s -X GET http://localhost:8080/api/auth/profile \
  -H "Authorization: Bearer $TOKEN")

echo "个人资料接口响应: $PROFILE_RESPONSE"
echo ""

# 清理
echo "6. 清理资源..."
kill $SERVER_PID
echo "✅ 测试完成"
