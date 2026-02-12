#!/bin/bash

# 菜单数据导出脚本
# 使用方法: ./export.sh

cd "$(dirname "$0")/../.."

go run tools/menu-exporter/main.go \
  -host localhost \
  -port 3306 \
  -user root \
  -password qynfqepwq \
  -database nunu_test \
  -output internal/server/initializer/menu.go

echo ""
echo "✅ 导出完成！请检查生成的文件: internal/server/initializer/menu.go"
