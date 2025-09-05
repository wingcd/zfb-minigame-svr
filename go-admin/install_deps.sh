#!/bin/bash

echo "========================================"
echo "正在安装Go项目依赖..."
echo "========================================"

# 检查Go版本
if ! command -v go &> /dev/null; then
    echo "错误: 未找到Go环境，请先安装Go 1.16+"
    exit 1
fi

echo "Go版本: $(go version)"
echo

# 安装admin-service依赖
echo "[1/2] 安装admin-service依赖..."
cd admin-service
if ! go mod tidy; then
    echo "错误: admin-service依赖整理失败"
    exit 1
fi

if ! go mod download; then
    echo "错误: admin-service依赖下载失败"
    exit 1
fi
echo "admin-service依赖安装完成!"

# 安装game-service依赖
echo
echo "[2/2] 安装game-service依赖..."
cd ../game-service
if ! go mod tidy; then
    echo "错误: game-service依赖整理失败"
    exit 1
fi

if ! go mod download; then
    echo "错误: game-service依赖下载失败"
    exit 1
fi
echo "game-service依赖安装完成!"

echo
echo "========================================"
echo "所有依赖安装完成!"
echo "========================================"

echo
echo "依赖列表:"
echo "- Beego v2.0.7 (Web框架)"
echo "- JWT Go v3.2.0 (JWT认证)"
echo "- Redis v8.11.5 (缓存数据库)"
echo "- MySQL Driver v1.7.0 (数据库驱动)"
echo "- Crypto (加密库)"

echo
echo "下一步: 运行 ./scripts/build.sh 编译项目"
