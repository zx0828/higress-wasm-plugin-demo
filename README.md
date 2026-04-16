# Higress WASM Plugin Demo

这是一个基于 Go 语言（WASI）开发的 Higress 自定义插件示例，专为 **Higress AI (All-in-One)** 及 Kubernetes 环境设计。

## 功能特性 (Features)

1.  **身份注入 (Identity Injection)**：自动在请求头中注入 `X-Client-Id` 和 `X-Secret-Id`。
2.  **签名校验 (Signature Calculation)**：
    *   根据 `ClientId` + `RequestBody` + `SecretId` 计算 MD5 签名。
    *   计算结果通过 `X-Sign` 请求头传递给后端，并在响应头中回传以便客户端验证。
3.  **IP 白名单 (IP Whitelist)**：
    *   支持 IP 地址及 CIDR 网段（如 `192.168.0.0/16`）。
    *   匹配白名单的请求将**跳过** Body 签名计算，提升性能。
4.  **动态配置 (Dynamic Config)**：支持在 Higress 控制台通过可视化表单或 JSON 动态修改参数。

## 项目结构 (Project Structure)

*   `main.go`: 插件核心逻辑。
*   `spec.yaml`: 插件元数据，定义了 Higress 控制台的可视化配置表单。
*   `fast-deploy.sh`: **推荐开发模式**。直接编译并拷贝 `.wasm` 到本地 `higress-ai` 容器。
*   `deploy-oci.sh`: **生产部署模式**。构建符合 OCI 规范的 Docker 镜像并推送到仓库。
*   `deploy-http.sh`: **HTTP 分发模式**。编译并移动到 `plugin-server` 托管目录。
*   `wasmplugin.yaml`: Kubernetes 部署参考模板。

## 快速开始 (Quick Start)

### 1. 本地快速部署 (Higress AI All-in-One)

如果您在本地运行了 Higress AI 容器，可以使用此脚本一键部署：

```bash
chmod +x fast-deploy.sh
./fast-deploy.sh
```

脚本会将插件拷贝至容器的 `/tmp/higress-wasm-plugin-demo.wasm`，您只需在控制台配置该文件路径即可。

### 2. HTTP 插件服务器部署

运行以下脚本，将插件托管到私有 HTTP 服务器：

```bash
chmod +x deploy-http.sh
./deploy-http.sh
```

### 3. OCI 镜像部署

修改 `deploy-oci.sh` 中的 `REGISTRY` 地址，然后运行：

```bash
chmod +x deploy-oci.sh
./deploy-oci.sh
```

## 配置示例 (Configuration)

在 Higress 控制台中，您可以使用以下 JSON 配置：

```json
{
  "clientId": "your-client-id",
  "secretId": "your-secret-id",
  "whiteList": [
    "127.0.0.1",
    "192.168.1.0/24"
  ],
  "message": "Processed by Higress Wasm"
}
```

## 开发环境要求 (Requirements)

*   Go 1.24+
*   Docker (用于 OCI 构建)
*   Higress AI 或 Higress 集群

---
Created by Gemini CLI.
