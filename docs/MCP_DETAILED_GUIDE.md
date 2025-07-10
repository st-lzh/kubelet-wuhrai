# kubelet-wuhrai MCP (Model Context Protocol) 详细使用指南

本文档提供了kubelet-wuhrai中MCP功能的完整使用指南，包括配置、调用和使用的详细步骤。

## 目录

1. [MCP概述](#mcp概述)
2. [MCP客户端模式详细配置](#mcp客户端模式详细配置)
3. [MCP服务器模式详细配置](#mcp服务器模式详细配置)
4. [完整的配置示例](#完整的配置示例)
5. [实际使用场景](#实际使用场景)
6. [故障排查](#故障排查)

## MCP概述

### 什么是MCP

Model Context Protocol (MCP) 是一个开放标准，用于连接AI助手与外部数据源和工具。kubelet-wuhrai支持两种MCP模式：

- **客户端模式**: kubelet-wuhrai作为MCP客户端，连接到外部MCP服务器获取额外工具
- **服务器模式**: kubelet-wuhrai作为MCP服务器，向其他MCP客户端暴露kubectl工具

### MCP的优势

1. **工具扩展**: 通过MCP可以轻松添加新的工具和功能
2. **标准化**: 使用开放标准，兼容性好
3. **模块化**: 工具可以独立开发和部署
4. **安全性**: 支持多种认证方式

## MCP客户端模式详细配置

### 1. 基本配置文件

创建MCP配置文件：

```bash
mkdir -p ~/.config/kubelet-wuhrai
cat > ~/.config/kubelet-wuhrai/mcp.yaml << 'EOF'
servers:
  # 本地stdio服务器示例
  - name: sequential-thinking
    command: npx
    args:
      - -y
      - "@modelcontextprotocol/server-sequential-thinking"
    env:
      NODE_ENV: production
      DEBUG: "mcp:*"

  # HTTP服务器示例
  - name: weather-api
    url: https://api.weather.com/mcp
    timeout: 30s
    
  # 带认证的HTTP服务器
  - name: github-api
    url: https://api.github.com/mcp
    auth:
      type: bearer
      token: "${GITHUB_TOKEN}"
    headers:
      User-Agent: "kubelet-wuhrai/1.0"
      Accept: "application/vnd.github.v3+json"

  # 基本认证示例
  - name: internal-api
    url: https://internal.company.com/mcp
    auth:
      type: basic
      username: "${MCP_USERNAME}"
      password: "${MCP_PASSWORD}"
    timeout: 60s
    retry_count: 3
EOF
```

### 2. 环境变量配置

设置必要的环境变量：

```bash
# GitHub API认证
export GITHUB_TOKEN="ghp_your_github_token_here"

# 内部API认证
export MCP_USERNAME="your_username"
export MCP_PASSWORD="your_password"

# 其他MCP服务器配置
export MCP_DEBUG="true"
export MCP_TIMEOUT="30s"
```

### 3. 启动MCP客户端模式

```bash
# 基本启动
kubelet-wuhrai --mcp-client "list all pods and analyze their status"

# 带调试信息启动
kubelet-wuhrai --mcp-client -v=2 "check cluster health using external tools"

# 指定配置文件
kubelet-wuhrai --mcp-client --custom-tools-config ~/.config/kubelet-wuhrai/mcp.yaml "your query"
```

### 4. 验证MCP连接

```bash
# 测试MCP连接
kubelet-wuhrai --mcp-client --quiet "test mcp connection"

# 列出可用的MCP工具
kubelet-wuhrai --mcp-client tools

# 查看MCP服务器状态
kubelet-wuhrai --mcp-client status
```

## MCP服务器模式详细配置

### 1. 基本MCP服务器模式

```bash
# 启动基本MCP服务器（只暴露内置工具）
kubelet-wuhrai --mcp-server

# 启动MCP服务器并暴露外部工具
kubelet-wuhrai --mcp-server --external-tools

# 指定自定义工具配置
kubelet-wuhrai --mcp-server --custom-tools-config /path/to/tools.yaml
```

### 2. 自定义工具配置

创建自定义工具配置文件：

```bash
cat > ~/.config/kubelet-wuhrai/tools.yaml << 'EOF'
tools:
  - name: helm
    description: "Helm包管理器，用于Kubernetes应用管理"
    command: "helm"
    command_desc: |
      Helm是Kubernetes的包管理器，用于管理Kubernetes应用程序。
      常用命令：
      - helm install <name> <chart>: 安装应用
      - helm upgrade <name> <chart>: 升级应用
      - helm list: 列出已安装的应用
      - helm uninstall <name>: 卸载应用
      - helm search repo <keyword>: 搜索chart
    parameters:
      - name: action
        type: string
        description: "要执行的Helm操作"
        required: true
        enum: ["install", "upgrade", "list", "uninstall", "search"]
      - name: name
        type: string
        description: "应用名称"
        required: false
      - name: chart
        type: string
        description: "Chart名称或路径"
        required: false

  - name: istio
    description: "Istio服务网格管理工具"
    command: "istioctl"
    command_desc: |
      Istio命令行工具，用于服务网格管理。
      常用命令：
      - istioctl proxy-status: 查看代理状态
      - istioctl analyze: 分析配置
      - istioctl kube-inject: 注入sidecar
      - istioctl proxy-config: 查看代理配置
    parameters:
      - name: subcommand
        type: string
        description: "Istio子命令"
        required: true
        enum: ["proxy-status", "analyze", "kube-inject", "proxy-config"]

  - name: monitoring
    description: "集群监控和指标收集工具"
    command: "kubectl"
    command_desc: |
      使用kubectl收集集群监控数据和指标。
      功能包括：
      - 资源使用情况监控
      - 节点状态检查
      - Pod性能分析
      - 事件监控
    parameters:
      - name: metric_type
        type: string
        description: "要收集的指标类型"
        required: true
        enum: ["cpu", "memory", "disk", "network", "events"]
      - name: namespace
        type: string
        description: "命名空间"
        required: false
        default: "default"
EOF
```

### 3. 与Claude Desktop集成

配置Claude Desktop使用kubelet-wuhrai作为MCP服务器：

```json
{
  "mcpServers": {
    "kubelet-wuhrai": {
      "command": "kubelet-wuhrai",
      "args": ["--mcp-server", "--external-tools"],
      "env": {
        "KUBECONFIG": "/path/to/your/kubeconfig",
        "DEEPSEEK_API_KEY": "your_api_key"
      }
    },
    "kubelet-wuhrai-basic": {
      "command": "kubelet-wuhrai",
      "args": ["--mcp-server"],
      "env": {
        "KUBECONFIG": "/path/to/your/kubeconfig"
      }
    }
  }
}
```

### 4. 与VS Code集成

在VS Code的settings.json中配置：

```json
{
  "mcp.servers": [
    {
      "name": "kubelet-wuhrai",
      "command": "kubelet-wuhrai",
      "args": ["--mcp-server", "--external-tools"],
      "env": {
        "KUBECONFIG": "/Users/username/.kube/config",
        "DEEPSEEK_API_KEY": "your_api_key"
      },
      "description": "Kubernetes AI助手，提供智能集群管理功能"
    }
  ],
  "mcp.autoStart": true,
  "mcp.logLevel": "info"
}
```

## 完整的配置示例

### 1. 生产环境MCP配置

```yaml
# ~/.config/kubelet-wuhrai/mcp-production.yaml
servers:
  # 企业内部工具服务器
  - name: company-tools
    url: https://tools.company.com/mcp
    auth:
      type: bearer
      token: "${COMPANY_MCP_TOKEN}"
    headers:
      X-Environment: "production"
      X-Team: "platform"
    timeout: 60s
    retry_count: 3
    retry_delay: 5s

  # 监控工具服务器
  - name: monitoring-tools
    command: /usr/local/bin/monitoring-mcp-server
    args:
      - --config
      - /etc/monitoring/mcp-config.json
    env:
      PROMETHEUS_URL: "https://prometheus.company.com"
      GRAFANA_URL: "https://grafana.company.com"
      ALERT_MANAGER_URL: "https://alertmanager.company.com"

  # 安全扫描工具
  - name: security-scanner
    url: https://security.company.com/mcp
    auth:
      type: oauth2
      client_id: "${SECURITY_CLIENT_ID}"
      client_secret: "${SECURITY_CLIENT_SECRET}"
      token_url: "https://auth.company.com/oauth/token"
    timeout: 120s

  # 外部API集成
  - name: external-apis
    command: node
    args:
      - /opt/mcp-servers/external-apis/index.js
    env:
      AWS_ACCESS_KEY_ID: "${AWS_ACCESS_KEY_ID}"
      AWS_SECRET_ACCESS_KEY: "${AWS_SECRET_ACCESS_KEY}"
      AZURE_CLIENT_ID: "${AZURE_CLIENT_ID}"
      AZURE_CLIENT_SECRET: "${AZURE_CLIENT_SECRET}"
      GCP_SERVICE_ACCOUNT_KEY: "${GCP_SERVICE_ACCOUNT_KEY}"
```

### 2. 开发环境MCP配置

```yaml
# ~/.config/kubelet-wuhrai/mcp-development.yaml
servers:
  # 本地开发工具
  - name: local-dev-tools
    command: npm
    args:
      - run
      - start:mcp
    cwd: /path/to/local/mcp-server
    env:
      NODE_ENV: development
      DEBUG: "*"

  # 测试环境API
  - name: test-api
    url: http://localhost:3000/mcp
    timeout: 10s
    retry_count: 1

  # 模拟工具
  - name: mock-tools
    command: python3
    args:
      - -m
      - mock_mcp_server
      - --port
      - "8080"
    env:
      MOCK_MODE: "true"
      LOG_LEVEL: "debug"
```

### 3. 启动脚本示例

创建便捷的启动脚本：

```bash
#!/bin/bash
# start-kubelet-wuhrai-mcp.sh

set -e

# 设置环境变量
export KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config}"
export DEEPSEEK_API_KEY="${DEEPSEEK_API_KEY}"

# 检查必要的环境变量
if [ -z "$DEEPSEEK_API_KEY" ]; then
    echo "错误: DEEPSEEK_API_KEY环境变量未设置"
    exit 1
fi

# 选择配置文件
ENVIRONMENT="${1:-development}"
MCP_CONFIG="$HOME/.config/kubelet-wuhrai/mcp-${ENVIRONMENT}.yaml"

if [ ! -f "$MCP_CONFIG" ]; then
    echo "错误: MCP配置文件不存在: $MCP_CONFIG"
    exit 1
fi

echo "使用配置文件: $MCP_CONFIG"
echo "环境: $ENVIRONMENT"

# 启动kubelet-wuhrai MCP客户端
kubelet-wuhrai \
    --mcp-client \
    --custom-tools-config "$MCP_CONFIG" \
    --user-interface html \
    --ui-listen-address "0.0.0.0:8888" \
    -v=1
```

使用脚本：

```bash
# 开发环境
./start-kubelet-wuhrai-mcp.sh development

# 生产环境
./start-kubelet-wuhrai-mcp.sh production
```

## 实际使用场景

### 1. 场景一：集成Helm包管理

#### 配置Helm MCP工具

```bash
# 安装Helm（如果未安装）
curl https://get.helm.sh/helm-v3.12.0-linux-amd64.tar.gz | tar xz
sudo mv linux-amd64/helm /usr/local/bin/

# 添加常用的Helm仓库
helm repo add stable https://charts.helm.sh/stable
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

#### 使用kubelet-wuhrai调用Helm

```bash
# 启动MCP客户端模式
kubelet-wuhrai --mcp-client

# 在交互模式中使用Helm工具
> "使用Helm安装nginx ingress controller"
> "列出所有已安装的Helm应用"
> "升级prometheus监控栈到最新版本"
> "搜索可用的数据库相关的Helm charts"
```

#### 具体命令示例

```bash
# 通过kubelet-wuhrai执行复杂的Helm操作
kubelet-wuhrai --mcp-client "使用Helm在monitoring命名空间安装Prometheus，并配置持久化存储"

# 这将自动执行类似以下的操作：
# 1. kubectl create namespace monitoring
# 2. helm install prometheus prometheus-community/kube-prometheus-stack -n monitoring
# 3. 配置相关的存储类和PVC
```

### 2. 场景二：集成Istio服务网格

#### 配置Istio MCP工具

```bash
# 下载并安装Istio
curl -L https://istio.io/downloadIstio | sh -
export PATH=$PWD/istio-1.18.0/bin:$PATH

# 安装Istio到集群
istioctl install --set values.defaultRevision=default
```

#### 使用kubelet-wuhrai管理Istio

```bash
kubelet-wuhrai --mcp-client "分析Istio服务网格的健康状态"
kubelet-wuhrai --mcp-client "为default命名空间启用Istio sidecar注入"
kubelet-wuhrai --mcp-client "检查所有服务的代理状态"
kubelet-wuhrai --mcp-client "配置流量路由规则，将50%流量路由到v2版本"
```

### 3. 场景三：集成监控和告警

#### 配置监控MCP工具

```yaml
# ~/.config/kubelet-wuhrai/monitoring-mcp.yaml
servers:
  - name: prometheus-tools
    command: python3
    args:
      - /opt/mcp-servers/prometheus/server.py
    env:
      PROMETHEUS_URL: "http://prometheus.monitoring.svc.cluster.local:9090"
      GRAFANA_URL: "http://grafana.monitoring.svc.cluster.local:3000"
      GRAFANA_API_KEY: "${GRAFANA_API_KEY}"

  - name: alertmanager-tools
    command: node
    args:
      - /opt/mcp-servers/alertmanager/index.js
    env:
      ALERTMANAGER_URL: "http://alertmanager.monitoring.svc.cluster.local:9093"
      SLACK_WEBHOOK_URL: "${SLACK_WEBHOOK_URL}"
```

#### 使用kubelet-wuhrai进行监控

```bash
kubelet-wuhrai --mcp-client --custom-tools-config ~/.config/kubelet-wuhrai/monitoring-mcp.yaml

# 监控相关查询
> "检查过去1小时内CPU使用率超过80%的Pod"
> "创建一个告警规则，当内存使用率超过90%时发送Slack通知"
> "生成集群资源使用情况的Grafana仪表板"
> "分析最近的告警趋势并提供优化建议"
```

### 4. 场景四：CI/CD集成

#### 配置CI/CD MCP工具

```yaml
# ~/.config/kubelet-wuhrai/cicd-mcp.yaml
servers:
  - name: jenkins-tools
    url: https://jenkins.company.com/mcp
    auth:
      type: basic
      username: "${JENKINS_USER}"
      password: "${JENKINS_TOKEN}"

  - name: argocd-tools
    command: argocd-mcp-server
    args:
      - --server
      - "https://argocd.company.com"
      - --token
      - "${ARGOCD_TOKEN}"

  - name: gitlab-tools
    url: https://gitlab.company.com/api/v4/mcp
    auth:
      type: bearer
      token: "${GITLAB_TOKEN}"
```

#### 使用kubelet-wuhrai管理CI/CD

```bash
kubelet-wuhrai --mcp-client --custom-tools-config ~/.config/kubelet-wuhrai/cicd-mcp.yaml

# CI/CD相关操作
> "触发应用的部署流水线到staging环境"
> "检查ArgoCD中所有应用的同步状态"
> "回滚production环境的应用到上一个版本"
> "创建一个新的GitLab分支并配置相应的部署流水线"
```

## 故障排查

### 1. 常见问题和解决方案

#### MCP服务器连接失败

```bash
# 问题：无法连接到MCP服务器
# 错误信息：failed to connect to MCP server

# 解决步骤：
# 1. 检查MCP配置文件
cat ~/.config/kubelet-wuhrai/mcp.yaml

# 2. 验证网络连接
curl -v https://your-mcp-server.com/health

# 3. 检查认证信息
echo $MCP_TOKEN

# 4. 测试MCP服务器
kubelet-wuhrai --mcp-client --quiet "test connection"

# 5. 启用调试日志
kubelet-wuhrai --mcp-client -v=2 "your query"
```

#### 工具调用失败

```bash
# 问题：MCP工具调用失败
# 错误信息：tool execution failed

# 解决步骤：
# 1. 检查工具是否可用
kubelet-wuhrai --mcp-client tools

# 2. 验证工具权限
which helm
helm version

# 3. 检查环境变量
env | grep -E "(KUBECONFIG|PATH|HOME)"

# 4. 手动测试工具
helm list
istioctl version
```

#### 认证问题

```bash
# 问题：MCP认证失败
# 错误信息：authentication failed

# 解决步骤：
# 1. 检查API密钥
echo $GITHUB_TOKEN | cut -c1-10

# 2. 验证token有效性
curl -H "Authorization: Bearer $GITHUB_TOKEN" https://api.github.com/user

# 3. 检查权限范围
# 确保token有足够的权限访问所需资源

# 4. 更新认证信息
export GITHUB_TOKEN="new_token_here"
```

### 2. 调试技巧

#### 启用详细日志

```bash
# 启用MCP调试日志
export MCP_DEBUG="true"
export DEBUG="mcp:*"

# 启用kubelet-wuhrai详细日志
kubelet-wuhrai --mcp-client -v=3 "your query"

# 查看日志文件
tail -f /tmp/kubelet-wuhrai.log
```

#### 测试MCP连接

```bash
# 创建测试脚本
cat > test-mcp.sh << 'EOF'
#!/bin/bash
echo "测试MCP连接..."

# 测试基本连接
kubelet-wuhrai --mcp-client --quiet "test mcp connection" || {
    echo "MCP连接失败"
    exit 1
}

# 列出可用工具
echo "可用的MCP工具："
kubelet-wuhrai --mcp-client tools

# 测试工具调用
echo "测试工具调用："
kubelet-wuhrai --mcp-client --quiet "list pods using kubectl tool"

echo "MCP测试完成"
EOF

chmod +x test-mcp.sh
./test-mcp.sh
```

#### 性能监控

```bash
# 监控MCP性能
cat > monitor-mcp.sh << 'EOF'
#!/bin/bash
echo "=== MCP性能监控 ==="

# 测试响应时间
start_time=$(date +%s.%N)
kubelet-wuhrai --mcp-client --quiet "get cluster info" > /dev/null
end_time=$(date +%s.%N)
response_time=$(echo "$end_time - $start_time" | bc)
echo "MCP响应时间: ${response_time}s"

# 检查内存使用
memory_usage=$(ps aux | grep kubelet-wuhrai | grep -v grep | awk '{print $4}')
echo "内存使用: ${memory_usage}%"

# 检查MCP服务器状态
kubelet-wuhrai --mcp-client status
EOF

chmod +x monitor-mcp.sh
./monitor-mcp.sh
```

### 3. 最佳实践

#### 配置管理

```bash
# 使用版本控制管理MCP配置
git init ~/.config/kubelet-wuhrai
cd ~/.config/kubelet-wuhrai
git add mcp.yaml tools.yaml
git commit -m "Initial MCP configuration"

# 创建不同环境的配置
cp mcp.yaml mcp-production.yaml
cp mcp.yaml mcp-staging.yaml
cp mcp.yaml mcp-development.yaml
```

#### 安全考虑

```bash
# 使用环境变量存储敏感信息
export MCP_TOKEN="$(cat /secure/path/to/token)"
export GITHUB_TOKEN="$(kubectl get secret github-token -o jsonpath='{.data.token}' | base64 -d)"

# 设置适当的文件权限
chmod 600 ~/.config/kubelet-wuhrai/mcp.yaml
chmod 700 ~/.config/kubelet-wuhrai/

# 定期轮换API密钥
# 在crontab中添加密钥轮换任务
```

#### 监控和告警

```bash
# 设置MCP健康检查
cat > /etc/cron.d/mcp-health-check << 'EOF'
*/5 * * * * root /usr/local/bin/kubelet-wuhrai --mcp-client --quiet "health check" || echo "MCP health check failed" | logger
EOF

# 配置告警
# 当MCP连接失败时发送通知
```

通过以上详细的配置和使用指南，您可以充分利用kubelet-wuhrai的MCP功能，实现强大的工具集成和自动化管理。
