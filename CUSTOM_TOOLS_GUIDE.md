# kubelet-wuhrai 自定义工具和MCP工具使用指南

本文档详细介绍如何在kubelet-wuhrai中使用自定义工具和MCP工具，扩展其功能。

## 📋 目录

1. [自定义工具使用](#自定义工具使用)
2. [MCP工具使用](#mcp工具使用)
3. [实际使用示例](#实际使用示例)
4. [故障排除](#故障排除)

## 🛠️ 自定义工具使用

### 1. 自定义工具配置文件

kubelet-wuhrai支持通过YAML配置文件定义自定义工具。

#### 创建工具配置文件

```bash
# 创建配置目录
mkdir -p ~/.config/kubelet-wuhrai

# 创建自定义工具配置文件
cat > ~/.config/kubelet-wuhrai/tools.yaml << 'EOF'
tools:
  # 系统监控工具
  - name: "system_monitor"
    description: "监控系统资源使用情况"
    command: "top"
    args: ["-b", "-n1"]
    timeout: "10s"
    
  # Docker容器管理
  - name: "docker_ps"
    description: "列出Docker容器"
    command: "docker"
    args: ["ps", "-a"]
    timeout: "5s"
    
  # 网络诊断工具
  - name: "network_check"
    description: "检查网络连接"
    command: "ping"
    args: ["-c", "3", "8.8.8.8"]
    timeout: "15s"
    
  # 自定义脚本工具
  - name: "cluster_health"
    description: "检查Kubernetes集群健康状态"
    command: "/usr/local/bin/check-cluster-health.sh"
    args: []
    timeout: "30s"
    working_directory: "/tmp"
    environment:
      KUBECONFIG: "${HOME}/.kube/config"
      
  # 带参数的工具
  - name: "log_analyzer"
    description: "分析日志文件"
    command: "grep"
    args: ["${pattern}", "${file}"]
    timeout: "20s"
    parameters:
      - name: "pattern"
        description: "搜索模式"
        required: true
      - name: "file"
        description: "日志文件路径"
        required: true
        default: "/var/log/syslog"
EOF
```

#### 使用自定义工具配置

```bash
# 指定自定义工具配置文件
kubelet-wuhrai --custom-tools-config ~/.config/kubelet-wuhrai/tools.yaml "检查系统资源使用情况"

# 使用多个配置文件
kubelet-wuhrai --custom-tools-config ~/.config/kubelet-wuhrai/tools.yaml --custom-tools-config /etc/kubelet-wuhrai/extra-tools.yaml "执行网络检查"
```

### 2. 高级自定义工具配置

#### 带条件执行的工具

```yaml
tools:
  # 条件执行工具
  - name: "conditional_restart"
    description: "根据条件重启服务"
    command: "bash"
    args: ["-c", "if systemctl is-failed ${service}; then systemctl restart ${service}; fi"]
    parameters:
      - name: "service"
        description: "服务名称"
        required: true
    timeout: "60s"
    
  # 管道命令工具
  - name: "pod_resource_usage"
    description: "获取Pod资源使用情况"
    command: "bash"
    args: ["-c", "kubectl top pods | grep ${namespace} | sort -k3 -nr"]
    parameters:
      - name: "namespace"
        description: "命名空间"
        default: "default"
    timeout: "30s"
```

#### 带输出处理的工具

```yaml
tools:
  # JSON输出处理
  - name: "parse_pod_status"
    description: "解析Pod状态信息"
    command: "kubectl"
    args: ["get", "pods", "-o", "json"]
    timeout: "20s"
    output_format: "json"
    
  # 表格输出处理
  - name: "format_node_info"
    description: "格式化节点信息"
    command: "kubectl"
    args: ["get", "nodes", "-o", "wide"]
    timeout: "15s"
    output_format: "table"
```

## 🔌 MCP工具使用

### 1. MCP客户端模式

MCP客户端模式允许kubelet-wuhrai连接到外部MCP服务器，获取额外的工具和功能。

#### 配置MCP客户端

```bash
# 创建MCP配置文件
cat > ~/.config/kubelet-wuhrai/mcp.yaml << 'EOF'
servers:
  # 本地stdio MCP服务器
  - name: "sequential-thinking"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-sequential-thinking"]
    env:
      NODE_ENV: "production"
    timeout: "30s"
    
  # HTTP MCP服务器
  - name: "github-tools"
    url: "https://api.github.com/mcp"
    auth:
      type: "bearer"
      token: "${GITHUB_TOKEN}"
    headers:
      User-Agent: "kubelet-wuhrai/1.0"
    timeout: "60s"
    
  # 自定义MCP服务器
  - name: "monitoring-tools"
    url: "http://monitoring.internal.com:8080/mcp"
    auth:
      type: "basic"
      username: "${MCP_USERNAME}"
      password: "${MCP_PASSWORD}"
    timeout: "45s"
EOF
```

#### 启动MCP客户端模式

```bash
# 基本MCP客户端模式
kubelet-wuhrai --mcp-client "使用外部工具分析集群状态"

# 指定MCP配置文件
kubelet-wuhrai --mcp-client --custom-tools-config ~/.config/kubelet-wuhrai/mcp.yaml "执行高级分析"

# 带调试信息
kubelet-wuhrai --mcp-client -v=2 "列出所有可用的MCP工具"
```

### 2. MCP服务器模式

MCP服务器模式将kubelet-wuhrai作为MCP服务器运行，向其他MCP客户端暴露kubectl工具。

#### 启动MCP服务器

```bash
# 基本MCP服务器模式
kubelet-wuhrai --mcp-server

# 暴露外部工具
kubelet-wuhrai --mcp-server --external-tools

# 指定监听地址
kubelet-wuhrai --mcp-server --ui-listen-address=0.0.0.0:9090
```

#### MCP服务器配置

```yaml
# ~/.config/kubelet-wuhrai/mcp-server.yaml
server:
  listen_address: "0.0.0.0:9090"
  timeout: "60s"
  max_connections: 100
  
exposed_tools:
  - "kubectl_get"
  - "kubectl_apply"
  - "kubectl_delete"
  - "kubectl_describe"
  - "bash_execute"
  
security:
  auth_required: true
  allowed_clients:
    - "client1.example.com"
    - "192.168.1.0/24"
```

### 3. 外部工具发现

kubelet-wuhrai可以自动发现并集成外部MCP工具。

```bash
# 启用外部工具发现
kubelet-wuhrai --external-tools --mcp-server "发现并使用所有可用工具"

# 列出发现的工具
kubelet-wuhrai --external-tools tools list

# 测试外部工具
kubelet-wuhrai --external-tools "使用发现的工具执行系统检查"
```

## 🎯 实际使用示例

### 示例1: 集成Helm工具

```yaml
# ~/.config/kubelet-wuhrai/helm-tools.yaml
tools:
  - name: "helm_list"
    description: "列出Helm发布"
    command: "helm"
    args: ["list", "-A"]
    timeout: "30s"
    
  - name: "helm_install"
    description: "安装Helm chart"
    command: "helm"
    args: ["install", "${release_name}", "${chart}", "--namespace", "${namespace}"]
    parameters:
      - name: "release_name"
        description: "发布名称"
        required: true
      - name: "chart"
        description: "Chart名称"
        required: true
      - name: "namespace"
        description: "命名空间"
        default: "default"
    timeout: "300s"
    
  - name: "helm_upgrade"
    description: "升级Helm发布"
    command: "helm"
    args: ["upgrade", "${release_name}", "${chart}", "--namespace", "${namespace}"]
    parameters:
      - name: "release_name"
        description: "发布名称"
        required: true
      - name: "chart"
        description: "Chart名称"
        required: true
      - name: "namespace"
        description: "命名空间"
        default: "default"
    timeout: "300s"
```

使用Helm工具：

```bash
# 使用Helm工具
kubelet-wuhrai --custom-tools-config ~/.config/kubelet-wuhrai/helm-tools.yaml "列出所有Helm发布"

kubelet-wuhrai --custom-tools-config ~/.config/kubelet-wuhrai/helm-tools.yaml "安装nginx ingress controller"

kubelet-wuhrai --custom-tools-config ~/.config/kubelet-wuhrai/helm-tools.yaml "升级prometheus监控栈"
```

### 示例2: 集成监控工具

```yaml
# ~/.config/kubelet-wuhrai/monitoring-tools.yaml
tools:
  - name: "prometheus_query"
    description: "执行Prometheus查询"
    command: "curl"
    args: ["-s", "http://prometheus:9090/api/v1/query?query=${query}"]
    parameters:
      - name: "query"
        description: "PromQL查询语句"
        required: true
    timeout: "30s"
    
  - name: "grafana_dashboard"
    description: "获取Grafana仪表板"
    command: "curl"
    args: ["-s", "-H", "Authorization: Bearer ${GRAFANA_TOKEN}", "http://grafana:3000/api/dashboards/uid/${uid}"]
    parameters:
      - name: "uid"
        description: "仪表板UID"
        required: true
    timeout: "20s"
    environment:
      GRAFANA_TOKEN: "${GRAFANA_API_TOKEN}"
```

### 示例3: MCP客户端集成GitHub

```bash
# 设置GitHub token
export GITHUB_TOKEN="ghp_your_token_here"

# 配置GitHub MCP服务器
cat > ~/.config/kubelet-wuhrai/github-mcp.yaml << 'EOF'
servers:
  - name: "github-api"
    url: "https://api.github.com/mcp"
    auth:
      type: "bearer"
      token: "${GITHUB_TOKEN}"
    headers:
      Accept: "application/vnd.github.v3+json"
    timeout: "60s"
EOF

# 使用GitHub工具
kubelet-wuhrai --mcp-client --custom-tools-config ~/.config/kubelet-wuhrai/github-mcp.yaml "检查仓库的最新提交"
```

## 🔧 故障排除

### 常见问题

1. **自定义工具不可用**
   ```bash
   # 检查工具配置
   kubelet-wuhrai --custom-tools-config ~/.config/kubelet-wuhrai/tools.yaml tools list
   
   # 验证命令路径
   which your-command
   ```

2. **MCP连接失败**
   ```bash
   # 检查MCP服务器状态
   kubelet-wuhrai --mcp-client -v=2 status
   
   # 测试网络连接
   curl -I http://your-mcp-server:port/health
   ```

3. **权限问题**
   ```bash
   # 检查命令权限
   ls -la /path/to/your/command
   
   # 添加执行权限
   chmod +x /path/to/your/command
   ```

4. **环境变量问题**
   ```bash
   # 检查环境变量
   echo $GITHUB_TOKEN
   echo $MCP_USERNAME
   
   # 设置环境变量
   export GITHUB_TOKEN="your-token"
   ```

### 调试技巧

```bash
# 启用详细日志
kubelet-wuhrai --custom-tools-config tools.yaml -v=3 "your query"

# 查看工具执行跟踪
kubelet-wuhrai --trace-path /tmp/trace.log "your query"
cat /tmp/trace.log

# 测试单个工具
kubelet-wuhrai --custom-tools-config tools.yaml tools test tool_name
```

## 📚 更多资源

- [MCP详细使用指南](docs/MCP_DETAILED_GUIDE.md)
- [API调用指南](docs/API_DETAILED_GUIDE.md)
- [扩展技术指南](docs/EXTENDED_TECHNICAL_GUIDE.md)

---

通过自定义工具和MCP工具，您可以大大扩展kubelet-wuhrai的功能，集成各种外部系统和工具，打造强大的Kubernetes管理平台！
