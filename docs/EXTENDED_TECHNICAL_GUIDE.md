# kubelet-wuhrai 扩展技术指南

本文档提供了kubelet-wuhrai二次开发后的完整技术指南，包括新增AI模型配置、HTTP API服务部署、MCP配置等详细说明。

## 目录

1. [新增AI模型配置指南](#新增ai模型配置指南)
2. [HTTP API服务部署](#http-api服务部署)
3. [API接口规范](#api接口规范)
4. [MCP配置详解](#mcp配置详解)
5. [故障排查指南](#故障排查指南)
6. [性能优化建议](#性能优化建议)

## 新增AI模型配置指南

### 1. DeepSeek模型配置

DeepSeek是默认的AI提供商，提供高质量的代码生成和推理能力。

#### 环境变量配置
```bash
export DEEPSEEK_API_KEY="your_deepseek_api_key_here"
```

#### 命令行使用
```bash
# 使用默认DeepSeek模型
kubelet-wuhrai "list all pods in default namespace"

# 指定特定模型
kubelet-wuhrai --llm-provider=deepseek --model=deepseek-chat "describe deployment nginx"
kubelet-wuhrai --llm-provider=deepseek --model=deepseek-coder "generate a kubernetes deployment yaml"
kubelet-wuhrai --llm-provider=deepseek --model=deepseek-reasoner "analyze cluster resource usage"
```

#### 配置文件设置
```yaml
# ~/.config/kubelet-wuhrai/config.yaml
llm-provider: "deepseek"
model: "deepseek-chat"
```

#### 支持的模型
- `deepseek-chat`: 通用对话模型，适合日常Kubernetes管理
- `deepseek-coder`: 代码专用模型，适合生成YAML配置和脚本
- `deepseek-reasoner`: 推理模型，适合复杂问题分析

### 2. 通义千问Qwen模型配置

通义千问通过阿里云DashScope API提供服务。

#### 环境变量配置
```bash
export DASHSCOPE_API_KEY="your_dashscope_api_key_here"
# 或者
export QWEN_API_KEY="your_qwen_api_key_here"
```

#### 命令行使用
```bash
# 使用Qwen模型
kubelet-wuhrai --llm-provider=qwen --model=qwen-plus "check pod status"
kubelet-wuhrai --llm-provider=qwen --model=qwen-turbo "quick cluster overview"
kubelet-wuhrai --llm-provider=qwen --model=qwen2.5-coder-32b "generate helm chart"
```

#### 支持的模型
- `qwen-plus`: 高性能通用模型
- `qwen-turbo`: 快速响应模型
- `qwen-max`: 最强性能模型
- `qwen2.5-72b-instruct`: 大参数量指令模型
- `qwen2.5-coder-32b-instruct`: 代码专用模型
- `qwen2.5-math-72b-instruct`: 数学推理模型

### 3. 字节跳动豆包模型配置

豆包模型通过火山引擎API提供服务。

#### 环境变量配置
```bash
export VOLCES_API_KEY="your_volces_api_key_here"
# 或者
export DOUBAO_API_KEY="your_doubao_api_key_here"
```

#### 命令行使用
```bash
# 使用豆包模型
kubelet-wuhrai --llm-provider=doubao --model=doubao-pro-4k "analyze logs"
kubelet-wuhrai --llm-provider=doubao --model=doubao-lite-32k "long context analysis"
kubelet-wuhrai --llm-provider=doubao --model=doubao-pro-vision "describe cluster topology"
```

#### 支持的模型
- `doubao-pro-4k`: 专业版4K上下文
- `doubao-pro-32k`: 专业版32K上下文
- `doubao-pro-128k`: 专业版128K上下文
- `doubao-lite-4k`: 轻量版4K上下文
- `doubao-pro-vision`: 支持视觉理解
- `doubao-pro-search`: 支持搜索增强

### 4. VLLM和第三方OpenAI兼容服务

支持自部署的VLLM服务和第三方OpenAI兼容API。

#### 环境变量配置
```bash
export OPENAI_API_KEY="your_api_key_here"
export OPENAI_ENDPOINT="http://your-vllm-server:8000/v1"
# 或者
export OPENAI_API_BASE="http://your-vllm-server:8000/v1"
```

#### 命令行使用
```bash
# 使用VLLM部署的模型
kubelet-wuhrai --llm-provider=vllm "vllm://your-server:8000" --model=your-model

# 使用第三方OpenAI兼容服务
kubelet-wuhrai --llm-provider=openai-compatible --model=custom-model
```

## HTTP API服务部署

kubelet-wuhrai可以作为HTTP服务运行，提供RESTful API接口。

### 1. 启动HTTP服务

```bash
# 启动HTML UI模式（包含API服务）
kubelet-wuhrai --user-interface=html --ui-listen-address=0.0.0.0:8888

# 使用配置文件
cat > ~/.config/kubelet-wuhrai/config.yaml << EOF
user-interface: html
ui-listen-address: "0.0.0.0:8888"
llm-provider: "deepseek"
model: "deepseek-chat"
EOF

kubelet-wuhrai
```

### 2. Docker部署

```dockerfile
FROM golang:1.24.3 AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o kubelet-wuhrai ./cmd/

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates curl
COPY --from=builder /src/kubelet-wuhrai /bin/kubelet-wuhrai
EXPOSE 8888
CMD ["kubelet-wuhrai", "--user-interface=html", "--ui-listen-address=0.0.0.0:8888"]
```

```bash
# 构建镜像
docker build -t kubelet-wuhrai:latest .

# 运行容器
docker run -d \
  -p 8888:8888 \
  -e DEEPSEEK_API_KEY="your_api_key" \
  -v ~/.kube:/root/.kube:ro \
  kubelet-wuhrai:latest
```

### 3. Kubernetes部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubelet-wuhrai
  namespace: kubelet-wuhrai
spec:
  replicas: 1
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
        env:
        - name: DEEPSEEK_API_KEY
          valueFrom:
            secretKeyRef:
              name: kubelet-wuhrai-secrets
              key: deepseek-api-key
        volumeMounts:
        - name: kubeconfig
          mountPath: /root/.kube
          readOnly: true
      volumes:
      - name: kubeconfig
        secret:
          secretName: kubelet-wuhrai-kubeconfig
---
apiVersion: v1
kind: Service
metadata:
  name: kubelet-wuhrai-service
  namespace: kubelet-wuhrai
spec:
  selector:
    app: kubelet-wuhrai
  ports:
  - port: 80
    targetPort: 8888
  type: LoadBalancer
```

## API接口规范

### 1. 聊天接口

#### POST /api/v1/chat

发送查询请求到kubelet-wuhrai。

**请求格式:**
```json
{
  "query": "list all pods in default namespace",
  "session_id": "optional_session_id",
  "context": {
    "namespace": "default",
    "cluster": "production"
  }
}
```

**响应格式:**
```json
{
  "response": "Here are the pods in the default namespace...",
  "session_id": "session_1234567890",
  "status": "completed",
  "metadata": {
    "query_length": "35",
    "processed_at": "2025-01-21T10:30:00Z"
  },
  "timestamp": "2025-01-21T10:30:00Z"
}
```

**curl示例:**
```bash
curl -X POST http://localhost:8888/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "query": "show me all deployments",
    "session_id": "my-session-123"
  }'
```

#### POST /api/v1/chat/stream

流式聊天接口，支持实时响应。

**请求格式:**
```json
{
  "query": "analyze cluster resource usage",
  "session_id": "stream_session_123",
  "stream": true
}
```

**响应格式 (Server-Sent Events):**
```
data: {"delta": "Analyzing", "session_id": "stream_session_123", "status": "processing", "done": false}

data: {"delta": " cluster", "session_id": "stream_session_123", "status": "processing", "done": false}

data: {"response": "Complete analysis result...", "session_id": "stream_session_123", "status": "completed", "done": true}
```

**curl示例:**
```bash
curl -X POST http://localhost:8888/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{
    "query": "check pod logs for nginx",
    "stream": true
  }'
```

### 2. 健康检查接口

#### GET /api/v1/health

检查服务健康状态。

**响应格式:**
```json
{
  "status": "healthy",
  "version": "dev",
  "timestamp": "2025-01-21T10:30:00Z"
}
```

**curl示例:**
```bash
curl http://localhost:8888/api/v1/health
```

### 3. 模型信息接口

#### GET /api/v1/models

获取可用模型列表。

**响应格式:**
```json
{
  "models": [
    "deepseek-chat",
    "deepseek-coder",
    "qwen-plus",
    "doubao-pro-4k"
  ],
  "current": "deepseek-chat",
  "provider": "deepseek",
  "metadata": {
    "default_provider": "deepseek",
    "supports_tools": "true"
  },
  "timestamp": "2025-01-21T10:30:00Z"
}
```

### 4. 状态接口

#### GET /api/v1/status

获取服务运行状态。

**响应格式:**
```json
{
  "status": "running",
  "timestamp": "2025-01-21T10:30:00Z",
  "uptime": "2h30m15s",
  "agent": "ready",
  "tools": "available",
  "mcp_client": "enabled"
}
```

## JavaScript SDK示例

```javascript
class KubectlAIClient {
  constructor(baseURL = 'http://localhost:8888') {
    this.baseURL = baseURL;
  }

  async chat(query, sessionId = null, context = {}) {
    const response = await fetch(`${this.baseURL}/api/v1/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        query,
        session_id: sessionId,
        context
      })
    });
    return response.json();
  }

  async streamChat(query, onMessage, sessionId = null) {
    const response = await fetch(`${this.baseURL}/api/v1/chat/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'text/event-stream'
      },
      body: JSON.stringify({
        query,
        session_id: sessionId,
        stream: true
      })
    });

    const reader = response.body.getReader();
    const decoder = new TextDecoder();

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      const chunk = decoder.decode(value);
      const lines = chunk.split('\n');
      
      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = JSON.parse(line.slice(6));
          onMessage(data);
          if (data.done) return;
        }
      }
    }
  }

  async getHealth() {
    const response = await fetch(`${this.baseURL}/api/v1/health`);
    return response.json();
  }

  async getModels() {
    const response = await fetch(`${this.baseURL}/api/v1/models`);
    return response.json();
  }
}

// 使用示例
const client = new KubectlAIClient();

// 普通聊天
const result = await client.chat('list all pods');
console.log(result.response);

// 流式聊天
await client.streamChat('analyze cluster', (data) => {
  if (data.delta) {
    process.stdout.write(data.delta);
  }
  if (data.done) {
    console.log('\nCompleted!');
  }
});
```

## Python SDK示例

```python
import requests
import json
import sseclient

class KubectlAIClient:
    def __init__(self, base_url="http://localhost:8888"):
        self.base_url = base_url
        self.session = requests.Session()

    def chat(self, query, session_id=None, context=None):
        """发送聊天请求"""
        url = f"{self.base_url}/api/v1/chat"
        payload = {
            "query": query,
            "session_id": session_id,
            "context": context or {}
        }
        response = self.session.post(url, json=payload)
        response.raise_for_status()
        return response.json()

    def stream_chat(self, query, session_id=None, context=None):
        """流式聊天请求"""
        url = f"{self.base_url}/api/v1/chat/stream"
        payload = {
            "query": query,
            "session_id": session_id,
            "stream": True,
            "context": context or {}
        }

        response = self.session.post(
            url,
            json=payload,
            headers={"Accept": "text/event-stream"},
            stream=True
        )
        response.raise_for_status()

        client = sseclient.SSEClient(response)
        for event in client.events():
            if event.data:
                yield json.loads(event.data)

    def get_health(self):
        """获取健康状态"""
        url = f"{self.base_url}/api/v1/health"
        response = self.session.get(url)
        response.raise_for_status()
        return response.json()

    def get_models(self):
        """获取可用模型"""
        url = f"{self.base_url}/api/v1/models"
        response = self.session.get(url)
        response.raise_for_status()
        return response.json()

# 使用示例
client = KubectlAIClient()

# 普通聊天
result = client.chat("list all pods in default namespace")
print(result["response"])

# 流式聊天
for chunk in client.stream_chat("analyze cluster resource usage"):
    if chunk.get("delta"):
        print(chunk["delta"], end="")
    if chunk.get("done"):
        print("\nCompleted!")
        break

# 检查健康状态
health = client.get_health()
print(f"Service status: {health['status']}")
```

## MCP配置详解

### 1. MCP客户端模式配置

MCP客户端模式允许kubelet-wuhrai连接到外部MCP服务器获取额外工具。

#### 配置文件位置
```bash
~/.config/kubelet-wuhrai/mcp.yaml
```

#### 基本配置示例
```yaml
servers:
  # 本地stdio服务器
  - name: sequential-thinking
    command: npx
    args:
      - -y
      - "@modelcontextprotocol/server-sequential-thinking"
    env:
      NODE_ENV: production

  # HTTP服务器
  - name: cloudflare-docs
    url: https://docs.mcp.cloudflare.com/mcp

  # 带认证的HTTP服务器
  - name: custom-api
    url: https://api.example.com/mcp
    auth:
      type: bearer
      token: "${MCP_TOKEN}"

  # 基本认证
  - name: secure-server
    url: https://secure.example.com/mcp
    auth:
      type: basic
      username: "${MCP_USERNAME}"
      password: "${MCP_PASSWORD}"
```

#### 启用MCP客户端
```bash
# 命令行启用
kubelet-wuhrai --mcp-client

# 配置文件启用
echo "mcp-client: true" >> ~/.config/kubelet-wuhrai/config.yaml
```

#### 环境变量配置
```bash
# MCP服务器认证
export MCP_TOKEN="your_mcp_token"
export MCP_USERNAME="your_username"
export MCP_PASSWORD="your_password"

# 特定服务器配置
export MCP_SEQUENTIAL_THINKING_COMMAND="npx"
export MCP_CLOUDFLARE_DOCS_URL="https://docs.mcp.cloudflare.com/mcp"
```

### 2. MCP服务器模式配置

kubelet-wuhrai可以作为MCP服务器运行，向其他MCP客户端暴露kubectl工具。

#### 基本MCP服务器模式
```bash
# 只暴露内置工具
kubelet-wuhrai --mcp-server

# 暴露内置工具和外部MCP工具
kubelet-wuhrai --mcp-server --external-tools
```

#### 与Claude Desktop集成
```json
// Claude Desktop配置文件
{
  "mcpServers": {
    "kubelet-wuhrai": {
      "command": "kubelet-wuhrai",
      "args": ["--mcp-server"]
    }
  }
}
```

#### 与VS Code集成
```json
// VS Code settings.json
{
  "mcp.servers": [
    {
      "name": "kubelet-wuhrai",
      "command": "kubelet-wuhrai",
      "args": ["--mcp-server", "--external-tools"]
    }
  ]
}
```

### 3. 高级MCP配置

#### 自定义工具配置
```yaml
# ~/.config/kubelet-wuhrai/tools.yaml
- name: helm
  description: "Helm package manager for Kubernetes"
  command: "helm"
  command_desc: |
    Helm command-line interface for managing Kubernetes applications.
    Common commands:
    - helm install <name> <chart>
    - helm upgrade <name> <chart>
    - helm list
    - helm uninstall <name>

- name: istio
  description: "Istio service mesh management"
  command: "istioctl"
  command_desc: |
    Istio command-line tool for service mesh management.
    Common commands:
    - istioctl proxy-status
    - istioctl analyze
    - istioctl kube-inject
```

#### MCP工具发现和注册
```bash
# 查看可用工具
kubelet-wuhrai --mcp-client tools

# 测试MCP连接
kubelet-wuhrai --mcp-client --quiet "test mcp connection"
```

## 故障排查指南

### 1. 常见问题和解决方案

#### API密钥问题
```bash
# 问题：API key not found
# 解决方案：
export DEEPSEEK_API_KEY="your_api_key"
# 或检查配置文件中的密钥设置

# 验证密钥是否正确
kubelet-wuhrai --quiet "test connection"
```

#### 网络连接问题
```bash
# 问题：connection timeout
# 解决方案：
export LLM_SKIP_VERIFY_SSL=true  # 跳过SSL验证
# 或设置代理
export HTTP_PROXY="http://proxy:8080"
export HTTPS_PROXY="http://proxy:8080"
```

#### 模型不可用
```bash
# 问题：model not found
# 解决方案：
kubelet-wuhrai models  # 查看可用模型
kubelet-wuhrai --llm-provider=deepseek --model=deepseek-chat "test"
```

#### MCP连接失败
```bash
# 问题：MCP server connection failed
# 解决方案：
# 1. 检查MCP配置
cat ~/.config/kubelet-wuhrai/mcp.yaml

# 2. 测试MCP服务器
npx -y @modelcontextprotocol/server-sequential-thinking

# 3. 检查环境变量
env | grep MCP
```

### 2. 日志和调试

#### 启用详细日志
```bash
# 启用调试日志
kubelet-wuhrai -v=2 "your query"

# 查看日志文件
tail -f /tmp/kubelet-wuhrai.log

# 启用trace
kubelet-wuhrai --trace-path=/tmp/kubelet-wuhrai-trace.txt "your query"
```

#### HTTP API调试
```bash
# 测试API连接
curl -v http://localhost:8888/api/v1/health

# 检查API响应
curl -X POST http://localhost:8888/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "test"}' | jq .
```

### 3. 性能问题诊断

#### 响应时间慢
```bash
# 使用更快的模型
kubelet-wuhrai --model=qwen-turbo "quick query"

# 减少上下文长度
kubelet-wuhrai --max-iterations=5 "simple query"

# 使用本地模型
kubelet-wuhrai --llm-provider=ollama --model=gemma3:latest "query"
```

#### 内存使用过高
```bash
# 限制工作目录大小
kubelet-wuhrai --remove-workdir "query"

# 使用轻量级模型
kubelet-wuhrai --model=doubao-lite-4k "query"
```

## 性能优化建议

### 1. 模型选择策略

#### 按用途选择模型
```bash
# 日常查询 - 使用快速模型
kubelet-wuhrai --model=qwen-turbo "list pods"

# 代码生成 - 使用代码专用模型
kubelet-wuhrai --model=deepseek-coder "generate deployment yaml"

# 复杂分析 - 使用强力模型
kubelet-wuhrai --model=qwen-max "analyze cluster performance"

# 长文本处理 - 使用大上下文模型
kubelet-wuhrai --model=doubao-pro-128k "analyze large log file"
```

#### 模型性能对比
| 模型 | 响应速度 | 质量 | 成本 | 适用场景 |
|------|----------|------|------|----------|
| qwen-turbo | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 快速查询 |
| deepseek-chat | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 通用对话 |
| deepseek-coder | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | 代码生成 |
| qwen-max | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ | 复杂分析 |
| doubao-pro-128k | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | 长文本 |

### 2. 配置优化

#### 生产环境配置
```yaml
# ~/.config/kubelet-wuhrai/config.yaml
llm-provider: "deepseek"
model: "deepseek-chat"
max-iterations: 10
skip-permissions: false
quiet: false
remove-workdir: true
user-interface: "html"
ui-listen-address: "0.0.0.0:8888"
```

#### 开发环境配置
```yaml
llm-provider: "qwen"
model: "qwen-turbo"
max-iterations: 5
skip-permissions: true
quiet: true
remove-workdir: true
```

### 3. 缓存和优化

#### 启用工具缓存
```bash
# 缓存kubectl输出
kubelet-wuhrai --quiet "get pods" > /tmp/pods.cache

# 使用缓存数据
cat /tmp/pods.cache | kubelet-wuhrai "analyze these pods"
```

#### 批量操作优化
```bash
# 批量查询
kubelet-wuhrai --quiet "get all resources in namespace production" | \
kubelet-wuhrai "summarize resource usage"

# 管道操作
kubectl get pods -o json | \
kubelet-wuhrai "find pods with high memory usage"
```

### 4. 监控和指标

#### 性能监控脚本
```bash
#!/bin/bash
# monitor-kubelet-wuhrai.sh

echo "=== kubelet-wuhrai Performance Monitor ==="
echo "Timestamp: $(date)"

# 测试响应时间
start_time=$(date +%s.%N)
kubelet-wuhrai --quiet "get pods" > /dev/null
end_time=$(date +%s.%N)
response_time=$(echo "$end_time - $start_time" | bc)

echo "Response time: ${response_time}s"

# 检查内存使用
memory_usage=$(ps aux | grep kubelet-wuhrai | grep -v grep | awk '{print $4}')
echo "Memory usage: ${memory_usage}%"

# 检查API健康状态
api_status=$(curl -s http://localhost:8888/api/v1/health | jq -r .status)
echo "API status: $api_status"
```

#### Prometheus监控配置
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'kubelet-wuhrai'
    static_configs:
      - targets: ['localhost:8888']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

## 总结

本技术指南涵盖了kubelet-wuhrai二次开发后的完整使用方法，包括：

1. **多AI模型支持**: DeepSeek、Qwen、豆包等主流模型
2. **HTTP API服务**: 完整的RESTful接口和流式API
3. **MCP集成**: 客户端和服务器模式的完整配置
4. **部署方案**: Docker、Kubernetes等生产环境部署
5. **故障排查**: 常见问题的诊断和解决方案
6. **性能优化**: 针对不同场景的最佳实践

通过本指南，您可以充分利用kubelet-wuhrai的扩展功能，构建强大的Kubernetes智能管理系统。
