# Higress MCP Signature Plugin

该插件是一个为大模型 MCP (Model Context Protocol) 架构深度定制的企业级 WASM 插件，专为无法修改代码的旧版后端系统提供安全凭证与签名注入。

## ✨ 核心特性

1. **零性能损耗的路径隔离**
   内置 `pathPrefixes` 拦截机制，仅对 MCP 流量（默认 `/mcp/`）生效。针对网关上的其他普通静态资源或 API 流量直接放行，不消耗任何额外的计算资源。

2. **三重优先级时间戳注入**
   支持从三种渠道获取 `timestamp` 参数，优先级从高到低：
   - 客户端在 HTTP Request Header 中传入的 `timestamp` 
   - 插件 YAML 配置中固化的 `timestamp`
   - 自动调用系统函数生成当前毫秒级时间戳

3. **完美规避 Envoy Envoy Content-Length 截断 Bug**
   基于 Higress 官方 `mcp-server` 的最新架构，本插件**绝不在 WASM 中直接修改 HTTP 报文**。它在 Body 阶段解析原生的 JSON-RPC 请求，将计算好的安全参数（`accessKey`, `accessSecret`, `sign`, `timestamp`）全量精准注入到 `params.arguments` 对象中。
   后续 `mcp-server` 会根据 OpenAPI 规范，无缝、无 Bug 地将它们转换为下游 HTTP Header 和 Body 参数。

## 🚀 工作原理

当 MCP 客户端发起请求时，插件通过截取 URL 中的参数和上下文，利用 HMAC-SHA1 算法计算出数字签名。

**签名算法**: `sign = HMAC-SHA1(accessSecret, accessKey + timestamp)` 
*(返回 Hex 编码的字符串)*

## 🛠️ 如何配置与使用

### 1. 插件配置说明 (YAML)
在 Higress 控制台配置此插件时，可以参考以下结构：

```yaml
# [必选] 您的 API 密钥对
accessKey: "YOUR_ACCESS_KEY"
accessSecret: "YOUR_ACCESS_SECRET"

# [可选] 需要拦截的 MCP 路由前缀（默认是 ["/mcp/"]）
pathPrefixes:
  - "/mcp/"
  - "/api/v1/secure-mcp"

# [可选] 强制写死的时间戳 (通常不需要配置，让其自动生成)
# timestamp: "1777463211205"
```

### 2. 必须配合 OpenAPI 规范使用
因为该插件使用的是“降维注入”策略，后端系统接收这些参数的前提是：**在 Nacos 中导入的 `openapi-spec.json` 必须显式声明这四个参数**。

以 `accessKey` 等头参数为例，您的 OpenAPI 规范的每个接口 `parameters` 中必须包含如下定义：

```json
"parameters": [
  { "name": "accessKey", "in": "header", "required": false, "schema": { "type": "string" } },
  { "name": "accessSecret", "in": "header", "required": false, "schema": { "type": "string" } },
  { "name": "timestamp", "in": "header", "required": false, "schema": { "type": "string" } },
  { "name": "sign", "in": "header", "required": false, "schema": { "type": "string" } }
]
```

*注意：如果后端要求签名放在 Request Body 中，请将它们平铺写入对应接口的 `requestBody.schema.properties` 中，绝不能使用 `allOf`！*

## 📦 部署指南

项目中提供了三个标准部署脚本：

- `./deploy-http.sh` - 使用 Higress Plugin Server 进行本地或轻量级 HTTP 分发。
- `./deploy-oci.sh` - 将插件打包为 OCI 镜像推送到私有 Docker Registry (如 Harbor)。
- `./fast-deploy.sh` - 极速本地测试构建。

> 提示：本项目的初始发布版本为 `1.0.0`。
