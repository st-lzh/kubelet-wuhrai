# kubelet-wuhrai API调用完整指南

## 🔄 功能对比：二开前后

### 二开前工具 (kubectl-ai)
基于Google的kubectl-ai项目，功能相对简单：
- 基本的kubectl命令转换
- 有限的LLM提供商支持
- 简单的交互模式

### 二开后工具 (kubelet-wuhrai)  
全面增强的Kubernetes AI管理工具：
- ✅ **多种LLM提供商支持**：DeepSeek、OpenAI、通义千问、豆包、Gemini、vLLM、SGLang、OpenRouter等
- ✅ **流式返回支持**：实时显示AI响应，优化用户体验
- ✅ **工具扩展系统**：自定义工具配置和扩展
- ✅ **MCP协议生态**：客户端/服务器模式，外部工具集成
- ✅ **多种交互界面**：终端和Web界面
- ✅ **安全权限控制**：操作确认和权限管理
- ✅ **本地模型支持**：Ollama、vLLM、SGLang、llama.cpp
- ✅ **会话管理**：支持复杂的多轮对话

---

## 🚀 API调用方式（非交互式使用）

### 1. 基础K8s集群管理

#### 1.1 集群状态查询

```bash
# 基本集群状态
kubelet-wuhrai --quiet "获取集群中所有节点状态"

# 详细资源信息
kubelet-wuhrai --quiet "显示所有Pod的资源使用情况，包括CPU和内存"

# 特定命名空间查询
kubelet-wuhrai --quiet "查看kube-system命名空间下的所有资源"
```

**流式返回版本：**
```bash
# 默认支持流式返回，实时显示结果
kubelet-wuhrai --quiet --user-interface terminal "分析集群健康状态并提供优化建议"
```

#### 1.2 应用部署和管理

```bash
# 应用部署
kubelet-wuhrai --quiet "部署一个nginx应用，3个副本，暴露80端口"

# 应用扩缩容
kubelet-wuhrai --quiet "将web应用扩展到5个副本"

# 应用更新
kubelet-wuhrai --quiet "更新nginx应用镜像到最新版本"

# 应用删除（需要确认）
kubelet-wuhrai --quiet --skip-permissions "删除test命名空间下的所有deployment"
```

#### 1.3 故障排查和诊断

```bash
# 自动故障诊断
kubelet-wuhrai --quiet "检查为什么Pod一直处于Pending状态"

# 日志分析
kubelet-wuhrai --quiet "分析nginx应用的错误日志"

# 资源瓶颈分析
kubelet-wuhrai --quiet "找出集群中内存使用率最高的Pod"
```

### 2. 多LLM提供商支持

#### 2.1 云端模型服务

```bash
# DeepSeek (默认，成本低)
export LLM_CLIENT="deepseek"
export DEEPSEEK_API_KEY="your-api-key"
kubelet-wuhrai --llm-provider deepseek --model deepseek-chat --quiet "获取所有Pod"

# OpenAI (功能强大)
export LLM_CLIENT="openai"
export OPENAI_API_KEY="your-api-key"
kubelet-wuhrai --llm-provider openai --model gpt-4o --quiet "检查集群状态"

# OpenRouter (模型最全)
export LLM_CLIENT="openrouter"
export OPENROUTER_API_KEY="your-api-key"
kubelet-wuhrai --llm-provider openrouter --model anthropic/claude-3.5-sonnet --quiet "优化deployment配置"

# 通义千问
export LLM_CLIENT="qwen"
export QWEN_API_KEY="your-api-key"
kubelet-wuhrai --llm-provider qwen --model qwen-plus --quiet "分析集群安全配置"

# 豆包
export LLM_CLIENT="doubao"
export DOUBAO_API_KEY="your-api-key"
kubelet-wuhrai --llm-provider doubao --model doubao-pro-4k --quiet "检查资源配额"

# Google Gemini
export LLM_CLIENT="gemini"
export GEMINI_API_KEY="your-api-key"
kubelet-wuhrai --llm-provider gemini --model gemini-2.5-pro --quiet "生成yaml配置文件"
```

#### 2.2 本地模型服务

```bash
# vLLM 本地部署
export LLM_CLIENT="vllm"
export VLLM_URL="http://localhost:8000"
kubelet-wuhrai --llm-provider vllm --quiet "获取所有服务状态"

# SGLang 本地部署
export LLM_CLIENT="sglang"
export SGLANG_URL="http://localhost:30000"
kubelet-wuhrai --llm-provider sglang --quiet "检查Pod健康状态"

# Ollama 本地部署
export LLM_CLIENT="ollama"
kubelet-wuhrai --llm-provider ollama --model llama3.3:70b --quiet "分析网络策略"

# llama.cpp 本地部署
export LLM_CLIENT="llamacpp"
export LLAMACPP_HOST="http://localhost:8080"
kubelet-wuhrai --llm-provider llamacpp --quiet "检查存储类配置"
```

### 3. Web API服务模式

#### 3.1 启动Web服务

```bash
# 启动Web API服务
kubelet-wuhrai --user-interface html --ui-listen-address 0.0.0.0:8888

# 后台运行
nohup kubelet-wuhrai --user-interface html --ui-listen-address 0.0.0.0:8888 > kubelet-wuhrai.log 2>&1 &

# 指定kubeconfig
kubelet-wuhrai --user-interface html --ui-listen-address 0.0.0.0:8888 --kubeconfig ~/.kube/prod-config
```

#### 3.2 Web API调用

```bash
# HTTP POST请求到Web界面
curl -X POST http://localhost:8888/api/query \
  -H "Content-Type: application/json" \
  -d '{"query": "获取所有Pod状态", "stream": true}'

# 流式API调用
curl -X POST http://localhost:8888/api/stream \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{"query": "分析集群资源使用情况"}'
```

### 4. MCP协议扩展

#### 4.1 MCP客户端模式

```bash
# 连接外部MCP服务
kubelet-wuhrai --mcp-client --quiet "使用监控工具查看Prometheus指标"

# 配置MCP服务器连接
cat > ~/.config/kubelet-wuhrai/mcp.yaml << EOF
servers:
  - name: prometheus
    command: prometheus-mcp-server
    args: ["--url", "http://prometheus:9090"]
  - name: grafana
    command: grafana-mcp-server
    args: ["--url", "http://grafana:3000"]
EOF

kubelet-wuhrai --mcp-client --quiet "查询Grafana中的CPU使用率仪表板"
```

#### 4.2 MCP服务器模式

```bash
# 作为MCP服务器运行
kubelet-wuhrai --mcp-server --external-tools

# 暴露到指定端口
kubelet-wuhrai --mcp-server --external-tools --ui-listen-address 0.0.0.0:8889
```

### 5. 自定义工具扩展

#### 5.1 配置自定义工具

```yaml
# ~/.config/kubelet-wuhrai/tools.yaml
tools:
  - name: "monitoring"
    command: "curl"
    args: ["-s", "http://prometheus:9090/api/v1/query"]
    description: "查询Prometheus监控指标"
    
  - name: "database"
    command: "psql"
    args: ["-h", "postgres", "-U", "admin", "-c"]
    description: "执行数据库查询"
    
  - name: "alert"
    command: "slack-cli"
    args: ["--channel", "#ops", "--message"]
    description: "发送告警到Slack"
```

#### 5.2 使用自定义工具

```bash
# 使用自定义工具配置
kubelet-wuhrai --custom-tools-config ~/.config/kubelet-wuhrai/tools.yaml --quiet "查询CPU使用率超过80%的节点并发送告警"

# 使用多个工具配置文件
kubelet-wuhrai --custom-tools-config ./monitoring-tools.yaml --custom-tools-config ./database-tools.yaml --quiet "从监控系统获取数据并更新CMDB"
```

### 6. 高级安全和权限控制

#### 6.1 权限控制

```bash
# 启用权限确认（默认）
kubelet-wuhrai --quiet "删除所有失败的Pod"  # 会要求确认

# 跳过权限确认（危险）
kubelet-wuhrai --skip-permissions --quiet "强制删除所有PVC"

# 指定kubeconfig
kubelet-wuhrai --kubeconfig ~/.kube/readonly-config --quiet "查看集群状态"
```

#### 6.2 SSL和网络安全

```bash
# 跳过SSL验证（不推荐生产环境）
kubelet-wuhrai --skip-verify-ssl --llm-provider openai --quiet "检查集群"

# 使用自定义CA证书
export SSL_CERT_FILE=/etc/ssl/certs/custom-ca.pem
kubelet-wuhrai --llm-provider custom-endpoint --quiet "安全查询"
```

### 7. 批量操作和自动化

#### 7.1 批量集群管理

```bash
# 多集群切换
for cluster in prod staging dev; do
  kubelet-wuhrai --kubeconfig ~/.kube/${cluster}-config --quiet "获取${cluster}集群Pod数量" >> cluster-report.txt
done

# 并发执行
parallel -j 3 kubelet-wuhrai --kubeconfig ~/.kube/{}-config --quiet "检查{}集群健康状态" ::: prod staging dev
```

#### 7.2 脚本化运维

```bash
#!/bin/bash
# 自动化运维脚本

# 检查集群状态
STATUS=$(kubelet-wuhrai --quiet "检查集群是否有异常Pod，返回简短状态")

# 根据状态执行操作
if [[ $STATUS == *"异常"* ]]; then
    # 详细诊断
    kubelet-wuhrai --quiet "详细分析异常Pod并提供修复建议" > diagnostic.log
    
    # 发送告警
    kubelet-wuhrai --custom-tools-config ./alert-tools.yaml --quiet "发送集群异常告警"
fi

# 定期健康检查
*/15 * * * * /usr/local/bin/kubelet-wuhrai --quiet "检查集群CPU和内存使用率，如果超过80%则记录详情" >> /var/log/k8s-health.log
```

### 8. 流式返回和实时监控

#### 8.1 实时日志分析

```bash
# 流式日志分析
kubelet-wuhrai --quiet "持续监控nginx应用日志，发现错误时报告" &

# 实时资源监控
kubelet-wuhrai --quiet "每30秒检查一次Pod资源使用情况，显示趋势"
```

#### 8.2 长时间任务执行

```bash
# 复杂的集群分析任务
kubelet-wuhrai --max-iterations 50 --quiet "全面分析集群配置，包括安全性、性能、资源利用率，并生成优化报告"

# 指定跟踪文件
kubelet-wuhrai --trace-path ./cluster-analysis.trace --quiet "执行集群深度分析"
```

### 9. 集成和API开发

#### 9.1 作为微服务集成

```bash
# 容器化运行
docker run -d \
  -p 8888:8888 \
  -v ~/.kube:/root/.kube:ro \
  -e LLM_CLIENT=deepseek \
  -e DEEPSEEK_API_KEY=your-key \
  kubelet-wuhrai:latest \
  --user-interface html --ui-listen-address 0.0.0.0:8888

# Kubernetes部署
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubelet-wuhrai
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
        - name: LLM_CLIENT
          value: "deepseek"
        - name: DEEPSEEK_API_KEY
          valueFrom:
            secretKeyRef:
              name: ai-secrets
              key: deepseek-key
        args:
        - --user-interface=html
        - --ui-listen-address=0.0.0.0:8888
EOF
```

#### 9.2 REST API调用

```bash
# 查询API
curl -X POST "http://localhost:8888/api/query" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "获取所有Pod状态",
    "options": {
      "llm_provider": "deepseek",
      "model": "deepseek-chat",
      "stream": true,
      "max_iterations": 10
    }
  }'

# 流式API
curl -X POST "http://localhost:8888/api/stream" \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{
    "query": "分析集群性能瓶颈",
    "stream": true
  }' \
  --no-buffer
```

### 10. 性能优化和成本控制

#### 10.1 模型选择策略

```bash
# 简单查询使用快速模型
simple_query() {
    kubelet-wuhrai --llm-provider deepseek --model deepseek-chat --quiet "$1"
}

# 复杂分析使用高级模型
complex_analysis() {
    kubelet-wuhrai --llm-provider openai --model gpt-4o --quiet "$1"
}

# 使用示例
simple_query "获取Pod数量"
complex_analysis "全面分析集群安全配置并提供改进建议"
```

#### 10.2 本地模型优化

```bash
# vLLM高性能配置
export VLLM_URL="http://localhost:8000"
export VLLM_MODEL="meta-llama/llama-3.3-70b-instruct"

# 批量处理降低成本
queries_file="queries.txt"
while IFS= read -r query; do
    kubelet-wuhrai --llm-provider vllm --quiet "$query" >> results.txt
    sleep 1  # 避免API限流
done < "$queries_file"
```

## 📊 性能对比

| 功能 | kubectl-ai | kubelet-wuhrai |
|------|-----------|----------------|
| LLM提供商 | 1-2个 | 11+个 |
| 流式返回 | ❌ | ✅ |
| 本地模型 | ❌ | ✅ |
| Web界面 | 简单 | 完整HTML UI |
| MCP协议 | ❌ | ✅ |
| 自定义工具 | ❌ | ✅ |
| 权限控制 | 基础 | 高级 |
| 多集群支持 | ❌ | ✅ |
| API服务模式 | ❌ | ✅ |

## 🎯 最佳实践建议

1. **生产环境使用**：启用权限确认，使用只读kubeconfig进行查询操作
2. **成本优化**：简单查询使用DeepSeek，复杂分析使用GPT-4
3. **安全第一**：敏感操作使用`--skip-permissions=false`（默认）
4. **性能优化**：使用本地模型（vLLM/SGLang）减少网络延迟
5. **集成开发**：使用Web API模式集成到运维平台
6. **监控告警**：结合MCP协议连接监控系统

kubelet-wuhrai提供了全面的Kubernetes AI管理能力，支持从简单查询到复杂自动化的各种使用场景。