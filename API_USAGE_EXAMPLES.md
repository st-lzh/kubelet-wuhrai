# kubelet-wuhrai API 调用示例

本文档提供了使用curl命令调用kubelet-wuhrai API的详细示例，包括自定义API端点配置。

## 🔧 配置信息

### 示例配置
- **Base URL**: `https://your-api-endpoint.com/v1`
- **API Key**: `sk-your-api-key-here`
- **Model**: `gpt-4o`

## 🚀 方式1: 通过kubelet-wuhrai HTTP API调用

### 1.1 环境变量配置

```bash
# 配置您的API信息
export OPENAI_API_KEY="sk-your-api-key-here"
export OPENAI_API_BASE="https://your-api-endpoint.com/v1"
export OPENAI_ENDPOINT="https://your-api-endpoint.com/v1"
```

### 1.2 启动kubelet-wuhrai服务

```bash
# 启动HTTP API服务
kubelet-wuhrai \
  --user-interface=html \
  --ui-listen-address=0.0.0.0:8888 \
  --llm-provider=openai \
  --model=gpt-4o
```

### 1.3 API调用示例

#### 基本聊天请求

```bash
curl -X POST http://localhost:8888/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "query": "获取所有运行中的pod",
    "session_id": "my_session_001"
  }'
```

#### 复杂查询请求

```bash
curl -X POST http://localhost:8888/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "query": "分析集群资源使用情况，找出CPU使用率最高的Pod并提供优化建议",
    "session_id": "analysis_session_001",
    "context": {
      "namespace": "default",
      "analysis_type": "resource_optimization"
    }
  }'
```

#### 流式响应请求

```bash
curl -X POST http://localhost:8888/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{
    "query": "执行集群健康检查并生成详细报告",
    "session_id": "stream_session_001",
    "stream": true
  }'
```

#### 健康检查和状态

```bash
# 健康检查
curl -X GET http://localhost:8888/api/v1/health

# 获取状态
curl -X GET http://localhost:8888/api/v1/status

# 获取模型信息
curl -X GET http://localhost:8888/api/v1/models
```

## 🤖 方式2: 直接调用自定义API端点

### 2.1 基本调用

```bash
curl -X POST https://your-api-endpoint.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-api-key-here" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {
        "role": "user",
        "content": "帮我分析Kubernetes集群中的pod状态"
      }
    ],
    "temperature": 0.7,
    "max_tokens": 1000
  }'
```

### 2.2 带系统提示的调用

```bash
curl -X POST https://your-api-endpoint.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-api-key-here" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {
        "role": "system",
        "content": "你是一个Kubernetes专家，帮助用户管理和分析Kubernetes集群。"
      },
      {
        "role": "user",
        "content": "我的pod一直处于Pending状态，帮我分析可能的原因"
      }
    ],
    "temperature": 0.7,
    "max_tokens": 2000,
    "stream": false
  }'
```

### 2.3 流式调用

```bash
curl -X POST https://your-api-endpoint.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-api-key-here" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {
        "role": "user",
        "content": "生成一个完整的Kubernetes部署配置，包括Deployment、Service和Ingress"
      }
    ],
    "temperature": 0.7,
    "max_tokens": 3000,
    "stream": true
  }'
```

### 2.4 带函数调用的高级用法

```bash
curl -X POST https://your-api-endpoint.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-api-key-here" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {
        "role": "user",
        "content": "检查集群中nginx相关的资源"
      }
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "kubectl_get",
          "description": "执行kubectl get命令获取Kubernetes资源",
          "parameters": {
            "type": "object",
            "properties": {
              "resource": {
                "type": "string",
                "description": "要获取的资源类型，如pods, services, deployments"
              },
              "namespace": {
                "type": "string",
                "description": "命名空间"
              },
              "selector": {
                "type": "string",
                "description": "标签选择器"
              }
            },
            "required": ["resource"]
          }
        }
      }
    ],
    "tool_choice": "auto",
    "temperature": 0.3,
    "max_tokens": 2000
  }'
```

## 📝 配置文件方式

### 创建配置文件

```bash
# 创建配置目录
mkdir -p ~/.config/kubelet-wuhrai

# 创建配置文件
cat > ~/.config/kubelet-wuhrai/config.yaml << EOF
llmProvider: "openai"
model: "gpt-4o"
skipPermissions: false
quiet: false
maxIterations: 20
userInterface: "html"
uiListenAddress: "0.0.0.0:8888"
EOF
```

### 设置环境变量

```bash
# 设置API配置
export OPENAI_API_KEY="sk-your-api-key-here"
export OPENAI_API_BASE="https://your-api-endpoint.com/v1"

# 启动服务
kubelet-wuhrai
```

## 🧪 快速测试脚本

### 创建测试脚本

```bash
cat > test_api.sh << 'EOF'
#!/bin/bash

# API配置
API_KEY="sk-your-api-key-here"
BASE_URL="https://your-api-endpoint.com/v1"
MODEL="gpt-4o"

echo "🚀 测试API连接..."
curl -X POST ${BASE_URL}/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d "{
    \"model\": \"${MODEL}\",
    \"messages\": [
      {
        \"role\": \"user\",
        \"content\": \"Hello, 请回复一个简单的测试消息\"
      }
    ],
    \"temperature\": 0.7,
    \"max_tokens\": 100
  }"

echo -e "\n\n🔧 测试kubelet-wuhrai API..."
curl -X POST http://localhost:8888/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "query": "测试连接",
    "session_id": "test_session"
  }'
EOF

chmod +x test_api.sh
```

### 运行测试

```bash
# 运行测试脚本
./test_api.sh
```

## 📋 参数说明

### 必需参数
- **Base URL**: API服务的基础URL
- **API Key**: 认证密钥
- **Model**: 使用的模型名称

### 可选参数
- **temperature**: 控制回复的随机性 (0.0-2.0)
- **max_tokens**: 最大输出token数
- **stream**: 是否使用流式响应
- **top_p**: 核采样参数 (0.0-1.0)
- **frequency_penalty**: 频率惩罚 (-2.0-2.0)
- **presence_penalty**: 存在惩罚 (-2.0-2.0)

## 🎯 推荐使用方式

**推荐使用方式1（通过kubelet-wuhrai HTTP API）**，因为它提供了：

✅ **Kubernetes工具集成** - 自动执行kubectl命令
✅ **智能上下文管理** - 理解Kubernetes概念
✅ **错误处理** - 友好的错误提示
✅ **会话管理** - 支持多轮对话
✅ **流式响应** - 实时显示处理过程

### 使用示例

```bash
# 直接问Kubernetes问题，系统会自动执行相应命令
curl -X POST http://localhost:8888/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "query": "显示所有pod的状态",
    "session_id": "k8s_session"
  }'

# 系统会自动执行: kubectl get pods --all-namespaces
# 并分析结果返回友好的回复
```

## 🔍 故障排除

### 常见问题

1. **连接失败**
   ```bash
   # 检查API端点是否可访问
   curl -I https://your-api-endpoint.com/v1/models
   ```

2. **认证失败**
   ```bash
   # 验证API密钥格式
   echo "API Key: sk-your-api-key-here"
   ```

3. **模型不存在**
   ```bash
   # 检查可用模型
   curl -X GET https://your-api-endpoint.com/v1/models \
     -H "Authorization: Bearer sk-your-api-key-here"
   ```

---

**注意**: 请确保API密钥的安全性，不要在公共代码库中暴露真实的API密钥。
