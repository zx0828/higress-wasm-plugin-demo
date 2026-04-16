#!/bin/bash

# Higress WASM Plugin - 符合官方规范的 OCI 镜像部署脚本
set -e

# 配置信息
REGISTRY="172.30.187.161:5000"
IMAGE_NAME="higress-wasm-plugin-demo"
VERSION="2.8.0"

echo "============================================"
echo "Higress WASM Plugin - 标准化 OCI 部署"
echo "============================================"

# 第 1 步：编译 WASM 文件
echo "[步骤 1/3] 编译 WASM 二进制文件..."
GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o main.wasm ./
echo "✓ 编译成功: main.wasm"

# 第 2 步：构建 Docker 镜像 (带上官方规范要求的 Label)
echo ""
echo "[步骤 2/3] 构建符合规范的镜像..."
docker build \
  --label "org.opencontainers.image.title=${IMAGE_NAME}" \
  --label "org.opencontainers.image.description=A demo wasm plugin with IPTree whitelist and signing" \
  --label "org.opencontainers.image.version=${VERSION}" \
  -t ${REGISTRY}/${IMAGE_NAME}:${VERSION} .

echo "✓ 镜像构建完成: ${REGISTRY}/${IMAGE_NAME}:${VERSION}"

# 第 3 步：推送到本地 Registry
echo ""
echo "[步骤 3/3] 推送到本地仓库..."
docker push ${REGISTRY}/${IMAGE_NAME}:${VERSION}
echo "✓ 推送成功！"

echo ""
echo "============================================"
echo "部署提示 (重要)："
echo "1. 请在控制台修改镜像地址为："
echo "   oci://local-registry:5000/${IMAGE_NAME}:${VERSION}"
echo "2. 点击保存并启用。"
echo "3. 重新进入编辑页面，即可看到表单视图。"
echo "============================================"
