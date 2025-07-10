# kubelet-wuhrai 详细部署和配置指南

本文档提供了kubelet-wuhrai在各种环境中的详细部署指南和配置说明，包括开发、测试、生产环境的完整部署步骤。

## 目录

1. [环境准备](#环境准备)
2. [本地开发环境部署](#本地开发环境部署)
3. [Docker容器化部署](#docker容器化部署)
4. [Kubernetes集群部署](#kubernetes集群部署)
5. [生产环境部署](#生产环境部署)
6. [配置管理](#配置管理)
7. [监控和日志](#监控和日志)
8. [安全配置](#安全配置)
9. [故障排查](#故障排查)

## 环境准备

### 系统要求

- **操作系统**: Linux (推荐 Ubuntu 20.04+, CentOS 8+), macOS 10.15+, Windows 10+
- **CPU**: 最低 2 核，推荐 4 核以上
- **内存**: 最低 4GB，推荐 8GB 以上
- **存储**: 最低 10GB 可用空间
- **网络**: 能够访问互联网和Kubernetes API服务器

### 依赖软件

#### 必需组件

```bash
# Go语言环境 (1.21+)
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# kubectl (1.25+)
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Git
sudo apt-get update && sudo apt-get install -y git

# 基本工具
sudo apt-get install -y curl wget jq bc
```

#### 可选组件

```bash
# Docker (用于容器化部署)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Helm (用于Kubernetes包管理)
curl https://get.helm.sh/helm-v3.12.0-linux-amd64.tar.gz | tar xz
sudo mv linux-amd64/helm /usr/local/bin/

# Node.js (用于MCP服务器)
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
```

### API密钥准备

在开始部署前，请准备以下API密钥：

```bash
# DeepSeek API密钥 (必需)
export DEEPSEEK_API_KEY="your_deepseek_api_key_here"

# 通义千问API密钥 (可选)
export DASHSCOPE_API_KEY="your_dashscope_api_key_here"

# 豆包API密钥 (可选)
export VOLCES_API_KEY="your_volces_api_key_here"

# 其他可选API密钥
export OPENAI_API_KEY="your_openai_api_key_here"
export GITHUB_TOKEN="your_github_token_here"
```

## 本地开发环境部署

### 1. 源码编译部署

#### 克隆和编译

```bash
# 克隆项目
git clone https://github.com/your-org/kubelet-wuhrai.git
cd kubelet-wuhrai

# 安装依赖
go mod tidy

# 编译项目
go build -o kubelet-wuhrai ./cmd/

# 验证编译
./kubelet-wuhrai --version
```

#### 基本配置

```bash
# 创建配置目录
mkdir -p ~/.config/kubelet-wuhrai

# 创建基本配置文件
cat > ~/.config/kubelet-wuhrai/config.yaml << 'EOF'
llm-provider: "deepseek"
model: "deepseek-chat"
user-interface: "terminal"
max-iterations: 10
skip-permissions: false
quiet: false
remove-workdir: true
EOF

# 设置环境变量
echo 'export DEEPSEEK_API_KEY="your_api_key_here"' >> ~/.bashrc
source ~/.bashrc
```

#### 启动和测试

```bash
# 基本启动测试
./kubelet-wuhrai --quiet "get pods"

# 启动HTTP服务模式
./kubelet-wuhrai --user-interface=html --ui-listen-address=localhost:8888

# 在另一个终端测试API
curl http://localhost:8888/api/v1/health
```

### 2. 开发环境配置

#### 开发配置文件

```yaml
# ~/.config/kubelet-wuhrai/dev-config.yaml
llm-provider: "deepseek"
model: "deepseek-chat"
user-interface: "html"
ui-listen-address: "localhost:8888"
max-iterations: 5
skip-permissions: true
quiet: false
remove-workdir: true
mcp-client: true
custom-tools-config:
  - ~/.config/kubelet-wuhrai/dev-tools.yaml
trace-path: "/tmp/kubelet-wuhrai-dev-trace.txt"
```

#### 开发工具配置

```yaml
# ~/.config/kubelet-wuhrai/dev-tools.yaml
tools:
  - name: dev-kubectl
    description: "开发环境kubectl工具"
    command: "kubectl"
    command_desc: "Kubernetes命令行工具，用于开发环境"
    
  - name: dev-docker
    description: "Docker开发工具"
    command: "docker"
    command_desc: "Docker容器管理工具"
    
  - name: dev-helm
    description: "Helm开发工具"
    command: "helm"
    command_desc: "Kubernetes包管理器"
```

#### 启动开发环境

```bash
# 使用开发配置启动
./kubelet-wuhrai --config ~/.config/kubelet-wuhrai/dev-config.yaml

# 或使用环境变量
export KUBECTL_AI_CONFIG=~/.config/kubelet-wuhrai/dev-config.yaml
./kubelet-wuhrai
```

## Docker容器化部署

### 1. 构建Docker镜像

#### 创建Dockerfile

```dockerfile
# Dockerfile
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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kubelet-wuhrai ./cmd/

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
RUN mkdir -p /app/config /app/data /app/logs && \
    chown -R kubelet-wuhrai:kubelet-wuhrai /app

# 复制编译好的二进制文件
COPY --from=builder /src/kubelet-wuhrai /usr/local/bin/kubelet-wuhrai
RUN chmod +x /usr/local/bin/kubelet-wuhrai

# 切换到非root用户
USER kubelet-wuhrai

# 设置工作目录
WORKDIR /app

# 暴露端口
EXPOSE 8888

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8888/api/v1/health || exit 1

# 启动命令
CMD ["kubelet-wuhrai", "--user-interface=html", "--ui-listen-address=0.0.0.0:8888"]
```

#### 构建和测试镜像

```bash
# 构建镜像
docker build -t kubelet-wuhrai:latest .

# 查看镜像信息
docker images kubelet-wuhrai

# 测试运行
docker run --rm -it \
    -e DEEPSEEK_API_KEY="your_api_key" \
    -v ~/.kube:/app/.kube:ro \
    -p 8888:8888 \
    kubelet-wuhrai:latest

# 测试API
curl http://localhost:8888/api/v1/health
```

### 2. Docker Compose部署

#### 创建docker-compose.yml

```yaml
# docker-compose.yml
version: '3.8'

services:
  kubelet-wuhrai:
    build: .
    image: kubelet-wuhrai:latest
    container_name: kubelet-wuhrai
    restart: unless-stopped
    ports:
      - "8888:8888"
    environment:
      - DEEPSEEK_API_KEY=${DEEPSEEK_API_KEY}
      - DASHSCOPE_API_KEY=${DASHSCOPE_API_KEY}
      - VOLCES_API_KEY=${VOLCES_API_KEY}
      - KUBECONFIG=/app/.kube/config
    volumes:
      - ~/.kube:/app/.kube:ro
      - ./config:/app/config:ro
      - ./logs:/app/logs
      - ./data:/app/data
    networks:
      - kubelet-wuhrai-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8888/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # 可选：添加Prometheus监控
  prometheus:
    image: prom/prometheus:latest
    container_name: kubelet-wuhrai-prometheus
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--web.enable-lifecycle'
    networks:
      - kubelet-wuhrai-network

  # 可选：添加Grafana仪表板
  grafana:
    image: grafana/grafana:latest
    container_name: kubelet-wuhrai-grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin123
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards:ro
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources:ro
    networks:
      - kubelet-wuhrai-network

networks:
  kubelet-wuhrai-network:
    driver: bridge

volumes:
  prometheus_data:
  grafana_data:
```

#### 环境变量配置

```bash
# .env文件
DEEPSEEK_API_KEY=your_deepseek_api_key_here
DASHSCOPE_API_KEY=your_dashscope_api_key_here
VOLCES_API_KEY=your_volces_api_key_here

# 可选配置
KUBECTL_AI_LOG_LEVEL=info
KUBECTL_AI_MAX_ITERATIONS=10
KUBECTL_AI_TIMEOUT=30s
```

#### 启动Docker Compose

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f kubelet-wuhrai

# 停止服务
docker-compose down

# 完全清理（包括数据卷）
docker-compose down -v
```

### 3. 多环境Docker配置

#### 开发环境配置

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  kubelet-wuhrai:
    build:
      context: .
      dockerfile: Dockerfile.dev
    image: kubelet-wuhrai:dev
    environment:
      - KUBECTL_AI_ENV=development
      - KUBECTL_AI_LOG_LEVEL=debug
      - KUBECTL_AI_SKIP_PERMISSIONS=true
    volumes:
      - .:/app/src:ro  # 挂载源代码用于开发
      - ~/.kube:/app/.kube:ro
    command: ["go", "run", "./cmd/", "--user-interface=html", "--ui-listen-address=0.0.0.0:8888", "-v=2"]
```

#### 生产环境配置

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  kubelet-wuhrai:
    image: kubelet-wuhrai:latest
    restart: always
    environment:
      - KUBECTL_AI_ENV=production
      - KUBECTL_AI_LOG_LEVEL=info
      - KUBECTL_AI_SKIP_PERMISSIONS=false
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '1.0'
          memory: 2G
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "5"
```

#### 使用不同环境配置

```bash
# 开发环境
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# 生产环境
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## Kubernetes集群部署

### 1. 基本Kubernetes部署

#### 创建命名空间和配置

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: kubelet-wuhrai
  labels:
    name: kubelet-wuhrai
    app.kubernetes.io/name: kubelet-wuhrai
    app.kubernetes.io/version: "1.0.0"
---
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubelet-wuhrai-config
  namespace: kubelet-wuhrai
data:
  config.yaml: |
    llm-provider: "deepseek"
    model: "deepseek-chat"
    user-interface: "html"
    ui-listen-address: "0.0.0.0:8888"
    max-iterations: 10
    skip-permissions: false
    quiet: false
    remove-workdir: true
    mcp-client: true

  tools.yaml: |
    tools:
      - name: kubectl
        description: "Kubernetes命令行工具"
        command: "kubectl"
        command_desc: "用于管理Kubernetes集群的命令行工具"

      - name: helm
        description: "Helm包管理器"
        command: "helm"
        command_desc: "Kubernetes应用程序包管理器"
```

#### 创建Secret存储API密钥

```yaml
# secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: kubelet-wuhrai-secrets
  namespace: kubelet-wuhrai
type: Opaque
data:
  deepseek-api-key: <base64-encoded-deepseek-api-key>
  dashscope-api-key: <base64-encoded-dashscope-api-key>
  volces-api-key: <base64-encoded-volces-api-key>
```

```bash
# 创建Secret的命令
kubectl create secret generic kubelet-wuhrai-secrets \
  --from-literal=deepseek-api-key="your_deepseek_api_key" \
  --from-literal=dashscope-api-key="your_dashscope_api_key" \
  --from-literal=volces-api-key="your_volces_api_key" \
  -n kubelet-wuhrai
```

#### 创建ServiceAccount和RBAC

```yaml
# rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubelet-wuhrai
  namespace: kubelet-wuhrai
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubelet-wuhrai-cluster-role
rules:
- apiGroups: [""]
  resources: ["pods", "services", "endpoints", "persistentvolumeclaims", "events", "configmaps", "secrets", "nodes", "namespaces"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "daemonsets", "replicasets", "statefulsets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["extensions"]
  resources: ["deployments", "daemonsets", "replicasets", "ingresses"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses", "networkpolicies"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["batch"]
  resources: ["jobs", "cronjobs"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["autoscaling"]
  resources: ["horizontalpodautoscalers"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["policy"]
  resources: ["poddisruptionbudgets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["metrics.k8s.io"]
  resources: ["pods", "nodes"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kubelet-wuhrai-cluster-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kubelet-wuhrai-cluster-role
subjects:
- kind: ServiceAccount
  name: kubelet-wuhrai
  namespace: kubelet-wuhrai
```

#### 创建Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubelet-wuhrai
  namespace: kubelet-wuhrai
  labels:
    app: kubelet-wuhrai
    version: "1.0.0"
spec:
  replicas: 2
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: kubelet-wuhrai
  template:
    metadata:
      labels:
        app: kubelet-wuhrai
        version: "1.0.0"
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8888"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: kubelet-wuhrai
      securityContext:
        runAsNonRoot: true
        runAsUser: 1001
        fsGroup: 1001
      containers:
      - name: kubelet-wuhrai
        image: kubelet-wuhrai:latest
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8888
          name: http
          protocol: TCP
        env:
        - name: DEEPSEEK_API_KEY
          valueFrom:
            secretKeyRef:
              name: kubelet-wuhrai-secrets
              key: deepseek-api-key
        - name: DASHSCOPE_API_KEY
          valueFrom:
            secretKeyRef:
              name: kubelet-wuhrai-secrets
              key: dashscope-api-key
              optional: true
        - name: VOLCES_API_KEY
          valueFrom:
            secretKeyRef:
              name: kubelet-wuhrai-secrets
              key: volces-api-key
              optional: true
        - name: KUBECTL_AI_CONFIG
          value: "/app/config/config.yaml"
        - name: KUBECTL_AI_TOOLS_CONFIG
          value: "/app/config/tools.yaml"
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
        - name: tmp
          mountPath: /tmp
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            cpu: 2000m
            memory: 4Gi
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8888
          initialDelaySeconds: 30
          periodSeconds: 30
          timeoutSeconds: 10
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /api/v1/health
            port: 8888
          initialDelaySeconds: 5
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        startupProbe:
          httpGet:
            path: /api/v1/health
            port: 8888
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 5
          failureThreshold: 12
      volumes:
      - name: config
        configMap:
          name: kubelet-wuhrai-config
      - name: tmp
        emptyDir: {}
      nodeSelector:
        kubernetes.io/os: linux
      tolerations:
      - key: "node.kubernetes.io/not-ready"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 300
      - key: "node.kubernetes.io/unreachable"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 300
```

#### 创建Service和Ingress

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: kubelet-wuhrai-service
  namespace: kubelet-wuhrai
  labels:
    app: kubelet-wuhrai
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8888"
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8888
    protocol: TCP
    name: http
  selector:
    app: kubelet-wuhrai
---
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kubelet-wuhrai-ingress
  namespace: kubelet-wuhrai
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
spec:
  tls:
  - hosts:
    - kubelet-wuhrai.yourdomain.com
    secretName: kubelet-wuhrai-tls
  rules:
  - host: kubelet-wuhrai.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: kubelet-wuhrai-service
            port:
              number: 80
```

### 2. 使用Helm部署

#### 创建Helm Chart

```bash
# 创建Helm Chart
helm create kubelet-wuhrai-chart
cd kubelet-wuhrai-chart

# 清理默认文件
rm -rf templates/*
rm values.yaml
```

#### Helm Values文件

```yaml
# values.yaml
replicaCount: 2

image:
  repository: kubelet-wuhrai
  pullPolicy: IfNotPresent
  tag: "latest"

nameOverride: ""
fullnameOverride: ""

serviceAccount:
  create: true
  annotations: {}
  name: ""

podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8888"
  prometheus.io/path: "/metrics"

podSecurityContext:
  runAsNonRoot: true
  runAsUser: 1001
  fsGroup: 1001

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: false
  runAsNonRoot: true
  runAsUser: 1001

service:
  type: ClusterIP
  port: 80
  targetPort: 8888

ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
  hosts:
    - host: kubelet-wuhrai.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: kubelet-wuhrai-tls
      hosts:
        - kubelet-wuhrai.yourdomain.com

resources:
  limits:
    cpu: 2000m
    memory: 4Gi
  requests:
    cpu: 500m
    memory: 1Gi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

nodeSelector: {}

tolerations: []

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app.kubernetes.io/name
            operator: In
            values:
            - kubelet-wuhrai
        topologyKey: kubernetes.io/hostname

config:
  llmProvider: "deepseek"
  model: "deepseek-chat"
  userInterface: "html"
  uiListenAddress: "0.0.0.0:8888"
  maxIterations: 10
  skipPermissions: false
  quiet: false
  removeWorkdir: true
  mcpClient: true

secrets:
  deepseekApiKey: ""
  dashscopeApiKey: ""
  volcesApiKey: ""

monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 30s
    path: /metrics

# 生产环境配置
production:
  enabled: false
  replicaCount: 3
  resources:
    limits:
      cpu: 4000m
      memory: 8Gi
    requests:
      cpu: 1000m
      memory: 2Gi
```

#### 部署Helm Chart

```bash
# 添加必要的Helm仓库
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo add cert-manager https://charts.jetstack.io
helm repo update

# 安装依赖
helm install ingress-nginx ingress-nginx/ingress-nginx -n ingress-nginx --create-namespace
helm install cert-manager cert-manager/cert-manager -n cert-manager --create-namespace --set installCRDs=true

# 安装kubelet-wuhrai
helm install kubelet-wuhrai ./kubelet-wuhrai-chart \
  --namespace kubelet-wuhrai \
  --create-namespace \
  --set secrets.deepseekApiKey="your_deepseek_api_key" \
  --set ingress.hosts[0].host="kubelet-wuhrai.yourdomain.com"

# 升级部署
helm upgrade kubelet-wuhrai ./kubelet-wuhrai-chart \
  --namespace kubelet-wuhrai \
  --set image.tag="v1.1.0"

# 查看部署状态
helm status kubelet-wuhrai -n kubelet-wuhrai

# 卸载
helm uninstall kubelet-wuhrai -n kubelet-wuhrai
```

## 生产环境部署

### 1. 生产环境架构设计

#### 高可用架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   Ingress       │    │   kubelet-wuhrai    │
│   (External)    │───▶│   Controller    │───▶│   Pods (3+)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   Monitoring    │
                       │   & Logging     │
                       └─────────────────┘
```

#### 生产环境清单

```bash
# 生产环境部署清单
cat > production-checklist.md << 'EOF'
## 生产环境部署清单

### 基础设施
- [ ] Kubernetes集群版本 >= 1.25
- [ ] 节点资源充足 (CPU: 8核+, 内存: 16GB+)
- [ ] 存储类配置正确
- [ ] 网络策略配置
- [ ] 备份策略制定

### 安全配置
- [ ] RBAC权限最小化
- [ ] Pod安全策略配置
- [ ] 网络策略限制
- [ ] Secret加密存储
- [ ] 镜像安全扫描

### 监控和日志
- [ ] Prometheus监控配置
- [ ] Grafana仪表板部署
- [ ] 日志聚合配置
- [ ] 告警规则设置
- [ ] 健康检查配置

### 高可用性
- [ ] 多副本部署 (3+)
- [ ] 反亲和性配置
- [ ] 自动扩缩容配置
- [ ] 滚动更新策略
- [ ] 故障转移测试

### 性能优化
- [ ] 资源限制配置
- [ ] JVM参数优化
- [ ] 连接池配置
- [ ] 缓存策略配置
- [ ] 负载测试完成
EOF
```

### 2. 生产环境配置

#### 生产环境Values文件

```yaml
# values-production.yaml
replicaCount: 3

image:
  repository: your-registry.com/kubelet-wuhrai
  pullPolicy: Always
  tag: "v1.0.0"

imagePullSecrets:
  - name: registry-secret

serviceAccount:
  create: true
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT:role/kubelet-wuhrai-role

podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8888"
  prometheus.io/path: "/metrics"
  cluster-autoscaler.kubernetes.io/safe-to-evict: "true"

podSecurityContext:
  runAsNonRoot: true
  runAsUser: 1001
  fsGroup: 1001
  seccompProfile:
    type: RuntimeDefault

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 1001

service:
  type: ClusterIP
  port: 80
  targetPort: 8888
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
    service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled: "true"

ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "1000"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "300"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "300"
  hosts:
    - host: kubelet-wuhrai.production.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: kubelet-wuhrai-production-tls
      hosts:
        - kubelet-wuhrai.production.com

resources:
  limits:
    cpu: 4000m
    memory: 8Gi
  requests:
    cpu: 1000m
    memory: 2Gi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 60

nodeSelector:
  node-type: compute-optimized

tolerations:
- key: "compute-optimized"
  operator: "Equal"
  value: "true"
  effect: "NoSchedule"

affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchExpressions:
        - key: app.kubernetes.io/name
          operator: In
          values:
          - kubelet-wuhrai
      topologyKey: kubernetes.io/hostname
  nodeAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      preference:
        matchExpressions:
        - key: node-type
          operator: In
          values:
          - compute-optimized

config:
  llmProvider: "deepseek"
  model: "deepseek-chat"
  userInterface: "html"
  uiListenAddress: "0.0.0.0:8888"
  maxIterations: 15
  skipPermissions: false
  quiet: false
  removeWorkdir: true
  mcpClient: true
  logLevel: "info"
  timeout: "60s"

secrets:
  deepseekApiKey: ""  # 从外部Secret管理系统获取
  dashscopeApiKey: ""
  volcesApiKey: ""

monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 15s
    path: /metrics
    scrapeTimeout: 10s

networkPolicy:
  enabled: true
  ingress:
    - from:
      - namespaceSelector:
          matchLabels:
            name: ingress-nginx
      ports:
      - protocol: TCP
        port: 8888
    - from:
      - namespaceSelector:
          matchLabels:
            name: monitoring
      ports:
      - protocol: TCP
        port: 8888

podDisruptionBudget:
  enabled: true
  minAvailable: 2

persistence:
  enabled: true
  storageClass: "fast-ssd"
  size: 10Gi
  accessMode: ReadWriteOnce
```

#### 生产环境部署脚本

```bash
#!/bin/bash
# deploy-production.sh

set -e

NAMESPACE="kubelet-wuhrai"
CHART_PATH="./kubelet-wuhrai-chart"
VALUES_FILE="values-production.yaml"
RELEASE_NAME="kubelet-wuhrai"

echo "=== 生产环境部署脚本 ==="

# 检查必要的工具
command -v kubectl >/dev/null 2>&1 || { echo "kubectl未安装"; exit 1; }
command -v helm >/dev/null 2>&1 || { echo "helm未安装"; exit 1; }

# 检查集群连接
kubectl cluster-info >/dev/null 2>&1 || { echo "无法连接到Kubernetes集群"; exit 1; }

# 创建命名空间
echo "创建命名空间..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# 创建Secret (从环境变量或外部系统获取)
echo "创建API密钥Secret..."
kubectl create secret generic kubelet-wuhrai-secrets \
  --from-literal=deepseek-api-key="${DEEPSEEK_API_KEY}" \
  --from-literal=dashscope-api-key="${DASHSCOPE_API_KEY:-}" \
  --from-literal=volces-api-key="${VOLCES_API_KEY:-}" \
  -n $NAMESPACE \
  --dry-run=client -o yaml | kubectl apply -f -

# 部署或升级
if helm list -n $NAMESPACE | grep -q $RELEASE_NAME; then
    echo "升级现有部署..."
    helm upgrade $RELEASE_NAME $CHART_PATH \
        --namespace $NAMESPACE \
        --values $VALUES_FILE \
        --wait \
        --timeout 10m
else
    echo "首次部署..."
    helm install $RELEASE_NAME $CHART_PATH \
        --namespace $NAMESPACE \
        --values $VALUES_FILE \
        --wait \
        --timeout 10m
fi

# 验证部署
echo "验证部署状态..."
kubectl rollout status deployment/$RELEASE_NAME -n $NAMESPACE --timeout=300s

# 检查Pod状态
echo "检查Pod状态..."
kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=kubelet-wuhrai

# 检查服务状态
echo "检查服务状态..."
kubectl get svc -n $NAMESPACE

# 运行健康检查
echo "运行健康检查..."
kubectl run health-check --rm -i --restart=Never --image=curlimages/curl -- \
    curl -f http://kubelet-wuhrai-service.$NAMESPACE.svc.cluster.local/api/v1/health

echo "=== 部署完成 ==="
echo "访问地址: https://kubelet-wuhrai.production.com"
```

### 3. 蓝绿部署策略

#### 蓝绿部署配置

```yaml
# blue-green-deployment.yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: kubelet-wuhrai-rollout
  namespace: kubelet-wuhrai
spec:
  replicas: 3
  strategy:
    blueGreen:
      activeService: kubelet-wuhrai-active
      previewService: kubelet-wuhrai-preview
      autoPromotionEnabled: false
      scaleDownDelaySeconds: 30
      prePromotionAnalysis:
        templates:
        - templateName: success-rate
        args:
        - name: service-name
          value: kubelet-wuhrai-preview
      postPromotionAnalysis:
        templates:
        - templateName: success-rate
        args:
        - name: service-name
          value: kubelet-wuhrai-active
  selector:
    matchLabels:
      app: kubelet-wuhrai
  template:
    metadata:
      labels:
        app: kubelet-wuhrai
    spec:
      containers:
      - name: kubelet-wuhrai
        image: kubelet-wuhrai:latest
        ports:
        - containerPort: 8888
        resources:
          requests:
            cpu: 1000m
            memory: 2Gi
          limits:
            cpu: 4000m
            memory: 8Gi
---
apiVersion: v1
kind: Service
metadata:
  name: kubelet-wuhrai-active
  namespace: kubelet-wuhrai
spec:
  selector:
    app: kubelet-wuhrai
  ports:
  - port: 80
    targetPort: 8888
---
apiVersion: v1
kind: Service
metadata:
  name: kubelet-wuhrai-preview
  namespace: kubelet-wuhrai
spec:
  selector:
    app: kubelet-wuhrai
  ports:
  - port: 80
    targetPort: 8888
```

#### 分析模板

```yaml
# analysis-template.yaml
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: success-rate
  namespace: kubelet-wuhrai
spec:
  args:
  - name: service-name
  metrics:
  - name: success-rate
    interval: 30s
    count: 5
    successCondition: result[0] >= 0.95
    failureLimit: 3
    provider:
      prometheus:
        address: http://prometheus.monitoring.svc.cluster.local:9090
        query: |
          sum(rate(http_requests_total{service="{{args.service-name}}",code!~"5.."}[2m])) /
          sum(rate(http_requests_total{service="{{args.service-name}}"}[2m]))
  - name: avg-response-time
    interval: 30s
    count: 5
    successCondition: result[0] <= 1000
    failureLimit: 3
    provider:
      prometheus:
        address: http://prometheus.monitoring.svc.cluster.local:9090
        query: |
          histogram_quantile(0.95,
            sum(rate(http_request_duration_seconds_bucket{service="{{args.service-name}}"}[2m])) by (le)
          ) * 1000
```

## 配置管理

### 1. 环境配置分离

#### 配置文件结构

```
config/
├── base/
│   ├── config.yaml
│   ├── tools.yaml
│   └── mcp.yaml
├── development/
│   ├── config.yaml
│   ├── secrets.yaml
│   └── overrides.yaml
├── staging/
│   ├── config.yaml
│   ├── secrets.yaml
│   └── overrides.yaml
└── production/
    ├── config.yaml
    ├── secrets.yaml
    └── overrides.yaml
```

#### 基础配置

```yaml
# config/base/config.yaml
llm-provider: "deepseek"
model: "deepseek-chat"
user-interface: "html"
ui-listen-address: "0.0.0.0:8888"
max-iterations: 10
skip-permissions: false
quiet: false
remove-workdir: true
mcp-client: true
trace-path: "/tmp/kubelet-wuhrai-trace.txt"
```

#### 环境特定配置

```yaml
# config/development/config.yaml
llm-provider: "deepseek"
model: "deepseek-chat"
max-iterations: 5
skip-permissions: true
quiet: false
log-level: "debug"
timeout: "30s"

# config/staging/config.yaml
llm-provider: "deepseek"
model: "deepseek-chat"
max-iterations: 10
skip-permissions: false
quiet: false
log-level: "info"
timeout: "60s"

# config/production/config.yaml
llm-provider: "deepseek"
model: "deepseek-chat"
max-iterations: 15
skip-permissions: false
quiet: true
log-level: "warn"
timeout: "120s"
```

### 2. 配置热重载

#### 配置监控脚本

```bash
#!/bin/bash
# config-watcher.sh

CONFIG_DIR="/app/config"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
PID_FILE="/tmp/kubelet-wuhrai.pid"

inotifywait -m -e modify "$CONFIG_FILE" |
while read path action file; do
    echo "配置文件已更改: $file"

    # 验证配置文件
    if kubelet-wuhrai --config "$CONFIG_FILE" --validate-config; then
        echo "配置验证通过，重新加载..."

        # 发送SIGHUP信号重新加载配置
        if [ -f "$PID_FILE" ]; then
            kill -HUP $(cat "$PID_FILE")
            echo "配置重新加载完成"
        else
            echo "未找到PID文件，需要重启服务"
        fi
    else
        echo "配置验证失败，忽略更改"
    fi
done
```

## 监控和日志

### 1. Prometheus监控配置

#### ServiceMonitor配置

```yaml
# monitoring/servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kubelet-wuhrai
  namespace: kubelet-wuhrai
  labels:
    app: kubelet-wuhrai
spec:
  selector:
    matchLabels:
      app: kubelet-wuhrai
  endpoints:
  - port: http
    path: /metrics
    interval: 15s
    scrapeTimeout: 10s
    honorLabels: true
```

#### Prometheus规则

```yaml
# monitoring/prometheus-rules.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: kubelet-wuhrai-rules
  namespace: kubelet-wuhrai
spec:
  groups:
  - name: kubelet-wuhrai.rules
    rules:
    - alert: KubectlAIHighErrorRate
      expr: rate(http_requests_total{job="kubelet-wuhrai",code=~"5.."}[5m]) > 0.1
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "kubelet-wuhrai高错误率"
        description: "kubelet-wuhrai在过去5分钟内错误率超过10%"

    - alert: KubectlAIHighResponseTime
      expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job="kubelet-wuhrai"}[5m])) > 5
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "kubelet-wuhrai响应时间过长"
        description: "kubelet-wuhrai 95%分位响应时间超过5秒"

    - alert: KubectlAIDown
      expr: up{job="kubelet-wuhrai"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "kubelet-wuhrai服务不可用"
        description: "kubelet-wuhrai服务已停止响应超过1分钟"

    - alert: KubectlAIHighMemoryUsage
      expr: container_memory_usage_bytes{pod=~"kubelet-wuhrai-.*"} / container_spec_memory_limit_bytes > 0.9
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "kubelet-wuhrai内存使用率过高"
        description: "kubelet-wuhrai内存使用率超过90%"
```

#### Grafana仪表板

```json
{
  "dashboard": {
    "id": null,
    "title": "kubelet-wuhrai监控仪表板",
    "tags": ["kubelet-wuhrai", "kubernetes"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "请求率",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{job=\"kubelet-wuhrai\"}[5m])",
            "legendFormat": "{{method}} {{code}}"
          }
        ],
        "yAxes": [
          {
            "label": "请求/秒"
          }
        ]
      },
      {
        "id": 2,
        "title": "响应时间",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.50, rate(http_request_duration_seconds_bucket{job=\"kubelet-wuhrai\"}[5m]))",
            "legendFormat": "50%分位"
          },
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job=\"kubelet-wuhrai\"}[5m]))",
            "legendFormat": "95%分位"
          },
          {
            "expr": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job=\"kubelet-wuhrai\"}[5m]))",
            "legendFormat": "99%分位"
          }
        ],
        "yAxes": [
          {
            "label": "秒"
          }
        ]
      },
      {
        "id": 3,
        "title": "错误率",
        "type": "singlestat",
        "targets": [
          {
            "expr": "rate(http_requests_total{job=\"kubelet-wuhrai\",code=~\"5..\"}[5m]) / rate(http_requests_total{job=\"kubelet-wuhrai\"}[5m]) * 100",
            "legendFormat": "错误率"
          }
        ],
        "valueName": "current",
        "format": "percent",
        "thresholds": "5,10"
      },
      {
        "id": 4,
        "title": "内存使用",
        "type": "graph",
        "targets": [
          {
            "expr": "container_memory_usage_bytes{pod=~\"kubelet-wuhrai-.*\"}",
            "legendFormat": "{{pod}}"
          }
        ],
        "yAxes": [
          {
            "label": "字节"
          }
        ]
      },
      {
        "id": 5,
        "title": "CPU使用",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(container_cpu_usage_seconds_total{pod=~\"kubelet-wuhrai-.*\"}[5m])",
            "legendFormat": "{{pod}}"
          }
        ],
        "yAxes": [
          {
            "label": "CPU核心"
          }
        ]
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}
```

### 2. 日志配置

#### Fluentd配置

```yaml
# logging/fluentd-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
  namespace: kubelet-wuhrai
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/containers/kubelet-wuhrai-*.log
      pos_file /var/log/fluentd-kubelet-wuhrai.log.pos
      tag kubernetes.kubelet-wuhrai
      format json
      time_key time
      time_format %Y-%m-%dT%H:%M:%S.%NZ
    </source>

    <filter kubernetes.kubelet-wuhrai>
      @type kubernetes_metadata
      @id filter_kube_metadata
    </filter>

    <filter kubernetes.kubelet-wuhrai>
      @type parser
      key_name log
      reserve_data true
      <parse>
        @type json
      </parse>
    </filter>

    <match kubernetes.kubelet-wuhrai>
      @type elasticsearch
      host elasticsearch.logging.svc.cluster.local
      port 9200
      index_name kubelet-wuhrai
      type_name _doc
      include_tag_key true
      tag_key @log_name
      <buffer>
        @type file
        path /var/log/fluentd-buffers/kubelet-wuhrai.buffer
        flush_mode interval
        retry_type exponential_backoff
        flush_thread_count 2
        flush_interval 5s
        retry_forever
        retry_max_interval 30
        chunk_limit_size 2M
        queue_limit_length 8
        overflow_action block
      </buffer>
    </match>
```

#### 日志聚合部署

```yaml
# logging/elasticsearch.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
  namespace: logging
spec:
  serviceName: elasticsearch
  replicas: 3
  selector:
    matchLabels:
      app: elasticsearch
  template:
    metadata:
      labels:
        app: elasticsearch
    spec:
      containers:
      - name: elasticsearch
        image: docker.elastic.co/elasticsearch/elasticsearch:7.17.0
        env:
        - name: cluster.name
          value: kubelet-wuhrai-logs
        - name: node.name
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: discovery.seed_hosts
          value: "elasticsearch-0.elasticsearch,elasticsearch-1.elasticsearch,elasticsearch-2.elasticsearch"
        - name: cluster.initial_master_nodes
          value: "elasticsearch-0,elasticsearch-1,elasticsearch-2"
        - name: ES_JAVA_OPTS
          value: "-Xms2g -Xmx2g"
        ports:
        - containerPort: 9200
          name: http
        - containerPort: 9300
          name: transport
        volumeMounts:
        - name: data
          mountPath: /usr/share/elasticsearch/data
        resources:
          requests:
            cpu: 1000m
            memory: 4Gi
          limits:
            cpu: 2000m
            memory: 4Gi
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: fast-ssd
      resources:
        requests:
          storage: 100Gi
```

## 安全配置

### 1. Pod安全策略

```yaml
# security/pod-security-policy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: kubelet-wuhrai-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
  readOnlyRootFilesystem: true
```

### 2. 网络策略

```yaml
# security/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kubelet-wuhrai-network-policy
  namespace: kubelet-wuhrai
spec:
  podSelector:
    matchLabels:
      app: kubelet-wuhrai
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8888
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 8888
  egress:
  - to: []
    ports:
    - protocol: TCP
      port: 443  # HTTPS
    - protocol: TCP
      port: 53   # DNS
    - protocol: UDP
      port: 53   # DNS
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 6443  # Kubernetes API
```

### 3. Secret管理

#### 使用External Secrets Operator

```yaml
# security/external-secret.yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: kubelet-wuhrai-secrets
  namespace: kubelet-wuhrai
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-secret-store
    kind: SecretStore
  target:
    name: kubelet-wuhrai-secrets
    creationPolicy: Owner
  data:
  - secretKey: deepseek-api-key
    remoteRef:
      key: kubelet-wuhrai/deepseek
      property: api-key
  - secretKey: dashscope-api-key
    remoteRef:
      key: kubelet-wuhrai/dashscope
      property: api-key
  - secretKey: volces-api-key
    remoteRef:
      key: kubelet-wuhrai/volces
      property: api-key
```

## 故障排查

### 1. 常见问题诊断

#### 诊断脚本

```bash
#!/bin/bash
# troubleshoot.sh

NAMESPACE="kubelet-wuhrai"
APP_LABEL="app=kubelet-wuhrai"

echo "=== kubelet-wuhrai故障诊断工具 ==="

# 检查Pod状态
echo "1. 检查Pod状态..."
kubectl get pods -n $NAMESPACE -l $APP_LABEL -o wide

# 检查Pod事件
echo "2. 检查Pod事件..."
kubectl get events -n $NAMESPACE --field-selector involvedObject.kind=Pod

# 检查服务状态
echo "3. 检查服务状态..."
kubectl get svc -n $NAMESPACE

# 检查Ingress状态
echo "4. 检查Ingress状态..."
kubectl get ingress -n $NAMESPACE

# 检查ConfigMap
echo "5. 检查ConfigMap..."
kubectl get configmap -n $NAMESPACE

# 检查Secret
echo "6. 检查Secret..."
kubectl get secret -n $NAMESPACE

# 检查资源使用
echo "7. 检查资源使用..."
kubectl top pods -n $NAMESPACE

# 检查日志
echo "8. 检查最近日志..."
kubectl logs -n $NAMESPACE -l $APP_LABEL --tail=50

# 网络连接测试
echo "9. 网络连接测试..."
kubectl run network-test --rm -i --restart=Never --image=nicolaka/netshoot -- \
    nslookup kubelet-wuhrai-service.$NAMESPACE.svc.cluster.local

# API健康检查
echo "10. API健康检查..."
kubectl run api-test --rm -i --restart=Never --image=curlimages/curl -- \
    curl -f http://kubelet-wuhrai-service.$NAMESPACE.svc.cluster.local/api/v1/health

echo "=== 诊断完成 ==="
```

### 2. 性能调优

#### 资源优化建议

```yaml
# performance/resource-optimization.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: performance-tuning
  namespace: kubelet-wuhrai
data:
  jvm-options: |
    -Xms2g
    -Xmx4g
    -XX:+UseG1GC
    -XX:MaxGCPauseMillis=200
    -XX:+UnlockExperimentalVMOptions
    -XX:+UseCGroupMemoryLimitForHeap
    -XX:+UseStringDeduplication

  go-options: |
    GOGC=100
    GOMEMLIMIT=4GiB
    GOMAXPROCS=4

  optimization-tips: |
    1. 根据实际负载调整副本数
    2. 使用节点亲和性将Pod调度到高性能节点
    3. 配置适当的资源请求和限制
    4. 启用水平Pod自动扩缩容
    5. 使用快速存储类
    6. 优化网络策略减少延迟
    7. 定期清理不必要的资源
```

### 3. 备份和恢复

#### 备份脚本

```bash
#!/bin/bash
# backup.sh

NAMESPACE="kubelet-wuhrai"
BACKUP_DIR="/backup/kubelet-wuhrai/$(date +%Y%m%d-%H%M%S)"

mkdir -p $BACKUP_DIR

echo "开始备份kubelet-wuhrai..."

# 备份配置
kubectl get configmap -n $NAMESPACE -o yaml > $BACKUP_DIR/configmaps.yaml
kubectl get secret -n $NAMESPACE -o yaml > $BACKUP_DIR/secrets.yaml

# 备份部署配置
kubectl get deployment -n $NAMESPACE -o yaml > $BACKUP_DIR/deployments.yaml
kubectl get service -n $NAMESPACE -o yaml > $BACKUP_DIR/services.yaml
kubectl get ingress -n $NAMESPACE -o yaml > $BACKUP_DIR/ingress.yaml

# 备份RBAC
kubectl get serviceaccount -n $NAMESPACE -o yaml > $BACKUP_DIR/serviceaccounts.yaml
kubectl get clusterrole kubelet-wuhrai-cluster-role -o yaml > $BACKUP_DIR/clusterrole.yaml
kubectl get clusterrolebinding kubelet-wuhrai-cluster-role-binding -o yaml > $BACKUP_DIR/clusterrolebinding.yaml

# 备份监控配置
kubectl get servicemonitor -n $NAMESPACE -o yaml > $BACKUP_DIR/servicemonitor.yaml 2>/dev/null || true
kubectl get prometheusrule -n $NAMESPACE -o yaml > $BACKUP_DIR/prometheusrule.yaml 2>/dev/null || true

echo "备份完成: $BACKUP_DIR"
```

#### 恢复脚本

```bash
#!/bin/bash
# restore.sh

BACKUP_DIR="$1"
NAMESPACE="kubelet-wuhrai"

if [ -z "$BACKUP_DIR" ]; then
    echo "用法: $0 <备份目录>"
    exit 1
fi

echo "从 $BACKUP_DIR 恢复kubelet-wuhrai..."

# 创建命名空间
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# 恢复配置
kubectl apply -f $BACKUP_DIR/configmaps.yaml
kubectl apply -f $BACKUP_DIR/secrets.yaml

# 恢复RBAC
kubectl apply -f $BACKUP_DIR/serviceaccounts.yaml
kubectl apply -f $BACKUP_DIR/clusterrole.yaml
kubectl apply -f $BACKUP_DIR/clusterrolebinding.yaml

# 恢复部署
kubectl apply -f $BACKUP_DIR/deployments.yaml
kubectl apply -f $BACKUP_DIR/services.yaml
kubectl apply -f $BACKUP_DIR/ingress.yaml

# 恢复监控配置
kubectl apply -f $BACKUP_DIR/servicemonitor.yaml 2>/dev/null || true
kubectl apply -f $BACKUP_DIR/prometheusrule.yaml 2>/dev/null || true

echo "恢复完成"
```

通过以上详细的部署和配置指南，您可以在各种环境中成功部署kubelet-wuhrai，并确保其稳定、安全、高效地运行。
