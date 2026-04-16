# 使用 alpine 基础镜像，提高 OCI 层解析的稳定性
FROM alpine:latest

# 官方规范：必须将 wasm 文件命名为 plugin.wasm 放在根目录
COPY main.wasm /plugin.wasm

# 官方规范：元数据文件命名为 spec.yaml 放在根目录
COPY spec.yaml /spec.yaml

# 保持镜像尽可能的轻量
RUN chmod +x /plugin.wasm
