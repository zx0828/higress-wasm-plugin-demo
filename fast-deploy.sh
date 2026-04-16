#!/bin/bash
set -e

# 1. 编译 (Go 1.24 原生 Wasm 编译)
echo "==> [1/2] 正在编译 WASM 插件..."
GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o main.wasm ./
echo "✓ 编译成功: main.wasm"

# 2. 拷贝到容器
echo "==> [2/2] 正在拷贝到 higress-ai 容器..."
docker cp main.wasm higress-ai:/tmp/higress-wasm-plugin-demo.wasm
echo "✓ 拷贝完成: /tmp/higress-wasm-plugin-demo.wasm"

# 3. 验证并显示文件信息
echo ""
echo "项目状态:"
docker exec higress-ai ls -lh /tmp/higress-wasm-plugin-demo.wasm
echo ""
echo "部署完成！Higress 会自动热加载该文件。"
