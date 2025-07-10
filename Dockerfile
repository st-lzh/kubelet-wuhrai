# kubectl-ai 扩展版本 Dockerfile
FROM golang:1.21.5-alpine AS builder

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /src

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o kubelet-wuhrai ./cmd/

# 运行时镜像
FROM alpine:3.18

# 安装运行时依赖
RUN apk add --no-cache ca-certificates curl jq bash

# 安装kubectl
RUN curl -LO "https://dl.k8s.io/release/v1.28.0/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# 创建非root用户
RUN addgroup -g 1001 kubelet-wuhrai && \
    adduser -D -u 1001 -G kubelet-wuhrai kubelet-wuhrai

# 创建必要的目录
RUN mkdir -p /app/config /app/data /app/logs /app/.kube && \
    chown -R kubelet-wuhrai:kubelet-wuhrai /app

# 复制编译好的二进制文件
COPY --from=builder /src/kubelet-wuhrai /usr/local/bin/kubelet-wuhrai
RUN chmod +x /usr/local/bin/kubelet-wuhrai

# 复制配置文件模板
COPY --chown=kubectl-ai:kubectl-ai docs/config-templates/ /app/config/

# 切换到非root用户
USER kubelet-wuhrai

# 设置工作目录
WORKDIR /app

# 设置环境变量
ENV KUBELET_WUHRAI_CONFIG=/app/config/config.yaml
ENV KUBECONFIG=/app/.kube/config

# 暴露端口
EXPOSE 8888

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8888/api/v1/health || exit 1

# 启动命令
CMD ["kubelet-wuhrai", "--user-interface=html", "--ui-listen-address=0.0.0.0:8888"]
