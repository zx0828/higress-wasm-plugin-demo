#!/bin/bash
# Higress WASM Plugin - HTTP 模式部署脚本
set -e

# --- 配置区 ---
PLUGIN_NAME="higress-wasm-plugin-demo"
VERSION="2.8.0"
BASE_DIR="/app/plugin-server"
TARGET_DIR="${BASE_DIR}/${PLUGIN_NAME}/${VERSION}"
SERVER_PORT="8085" # 宿主机访问端口
# --------------

echo "============================================"
echo "Higress WASM Plugin - HTTP 模式部署"
echo "============================================"

# 第 1 步：编译
echo "[步骤 1/3] 正在编译 WASM 插件..."
GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o main.wasm ./
echo "✓ 编译成功: main.wasm"

# 第 2 步：创建目录并移动文件
echo ""
echo "[步骤 2/3] 准备服务器目录结构..."
sudo mkdir -p "${TARGET_DIR}"
sudo cp main.wasm "${TARGET_DIR}/plugin.wasm"
sudo chmod -R 755 "${BASE_DIR}"
echo "✓ 文件已就绪: ${TARGET_DIR}/plugin.wasm"

# 第 3 步：生成配置信息
echo ""
echo "[步骤 3/3] 生成 Higress 配置参数..."
SHA256=$(sha256sum main.wasm | cut -d' ' -f1)
# 获取宿主机 IP (尝试获取常用的内网 IP)
HOST_IP=$(hostname -I | awk '{print $1}')

echo "--------------------------------------------"
echo "部署完成！请在 Higress 控制台填入以下信息："
echo ""
echo "镜像地址 (URL):"
echo "http://${HOST_IP}:${SERVER_PORT}/plugins/${PLUGIN_NAME}/${VERSION}/plugin.wasm"
echo ""
echo "SHA256 校验和:"
echo "${SHA256}"
echo ""
echo "容器内部访问地址 (若已加入同一网络):"
echo "http://higress-plugin-server:8080/plugins/${PLUGIN_NAME}/${VERSION}/plugin.wasm"
echo "============================================"
