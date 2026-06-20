#!/bin/bash

# Platform Service 快速启动脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
CONFIG_FILE="$PROJECT_DIR/configs/platform.yaml"

echo "🔧 检查依赖..."

# 检查 MySQL
if ! command -v mysql &> /dev/null; then
    echo "⚠️  MySQL 未安装，请确保运行中且 forge_platform 数据库已创建"
else
    echo "✅ MySQL 就绪"
fi

# 检查 Go
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装"
    exit 1
else
    echo "✅ Go $(go version | awk '{print $3}') 就绪"
fi

# 检查配置文件
if [ ! -f "$CONFIG_FILE" ]; then
    echo "⚠️  配置文件不存在: $CONFIG_FILE"
    echo "   请复制 configs/config.example.yaml 为 configs/platform.yaml 并修改配置"
    exit 1
else
    echo "✅ 配置文件就绪"
fi

# 安装依赖
echo "📦 安装依赖..."
cd "$PROJECT_DIR"
go mod tidy

# 编译
echo "🔨 编译中..."
go build -o bin/platform-service ./cmd/main.go

# 启动
echo "🚀 启动 Platform Service..."
echo "   地址: http://localhost:8880"
echo "   按 Ctrl+C 停止服务"
echo ""
./bin/platform-service -f "$CONFIG_FILE"
