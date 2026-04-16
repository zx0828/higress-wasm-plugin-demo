# Higress 高性能国密 SM3 插件

本项目是一个基于 Go 语言（WASI）开发的 Higress 自定义插件，核心功能为身份注入与 **高性能国密 SM3 签名校验**。

## 🌟 核心特性

1.  **极致性能 SM3**：
    *   **手动循环展开**：参考 RustCrypto 优化策略，实施 8 轮循环展开。
    *   **快速位运算**：集成 GG2 函数位运算技巧，减少 CPU 执行周期。
    *   **零分配（Zero-allocation）**：优化内存对齐，减少 WASM 环境下的 GC 压力。
2.  **性能实时监控**：
    *   插件自动记录每次签名计算的**纳秒级耗时**，并输出至网关日志，便于压测与调优。
3.  **身份注入与白名单**：
    *   自动注入 `X-Client-Id` 和 `X-Secret-Id`。
    *   支持基于 **IPTree** 的高效 IP 白名单（支持 CIDR），匹配成功则跳过签名计算。
4.  **工程化部署体系**：
    *   支持 OCI 镜像、HTTP 静态分发、容器热拷贝三种部署模式。

## 📂 项目结构

*   `pkg/sm3/`: 极致优化的内建国密 SM3 算法实现。
*   `main.go`: 插件入口逻辑，包含计时器与 Header 处理。
*   `deploy-http.sh`: **推荐！** 自动编译并托管至插件服务器，生成 SHA256。
*   `fast-deploy.sh`: 本地开发调试快传脚本。
*   `deploy-oci.sh`: 生产环境标准 OCI 镜像构建。

## 🚀 部署指南 (v2.8.0)

### 1. HTTP 模式（适用于 All-in-One 或私有化环境）
确保您已启动 `higress-plugin-server`，然后运行：
```bash
./deploy-http.sh
```
脚本会输出最新的 **SHA256**。在控制台镜像地址填入：
`http://higress-plugin-server:8080/plugins/higress-wasm-plugin-demo/2.8.0/plugin.wasm`

### 2. 单元测试
验证算法一致性（符合 GM/T 0004-2012 标准）：
```bash
go test -v ./pkg/sm3/...
```

## ⚙️ 配置示例

在 Higress 控制台中配置：
```json
{
  "clientId": "test-user",
  "secretId": "secure-key-123",
  "whiteList": ["10.0.0.0/8", "192.168.1.1"],
  "message": "Processed by SM3 Optimized"
}
```

---
Created by Gemini CLI. 基于国密标准与高性能 WASM 实践。
