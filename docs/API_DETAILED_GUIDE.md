# kubelet-wuhrai HTTP API 详细调用指南

本文档提供了kubelet-wuhrai HTTP API的完整调用指南，包括详细的请求示例、响应格式和错误处理。

## 目录

1. [API概述](#api概述)
2. [认证和安全](#认证和安全)
3. [详细的API接口](#详细的api接口)
4. [SDK和客户端库](#sdk和客户端库)
5. [错误处理和状态码](#错误处理和状态码)
6. [性能优化](#性能优化)
7. [实际使用场景](#实际使用场景)

## API概述

### 基本信息

- **Base URL**: `http://localhost:8888` (默认)
- **API版本**: v1
- **数据格式**: JSON
- **字符编码**: UTF-8
- **超时设置**: 30秒 (可配置)

### 支持的HTTP方法

- `GET`: 获取资源信息
- `POST`: 创建资源或执行操作
- `OPTIONS`: CORS预检请求

### 启动API服务

```bash
# 基本启动
kubelet-wuhrai --user-interface=html --ui-listen-address=0.0.0.0:8888

# 带完整配置启动
kubelet-wuhrai \
    --user-interface=html \
    --ui-listen-address=0.0.0.0:8888 \
    --llm-provider=deepseek \
    --model=deepseek-chat \
    --mcp-client \
    -v=1
```

## 认证和安全

### CORS配置

API服务器默认启用CORS，支持跨域请求：

```http
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization, X-Requested-With
Access-Control-Max-Age: 86400
```

### 安全头设置

```bash
# 在生产环境中，建议设置安全头
curl -H "X-API-Key: your-api-key" \
     -H "X-Request-ID: $(uuidgen)" \
     -H "User-Agent: MyApp/1.0" \
     http://localhost:8888/api/v1/health
```

## 详细的API接口

### 1. 聊天接口 - POST /api/v1/chat

#### 基本请求

```bash
curl -X POST http://localhost:8888/api/v1/chat \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "query": "list all pods in default namespace",
    "session_id": "session_123",
    "context": {
      "namespace": "default",
      "cluster": "production"
    }
  }'
```

#### 请求参数详解

```json
{
  "query": "string, required - 要执行的查询或命令",
  "session_id": "string, optional - 会话ID，用于保持上下文",
  "context": {
    "namespace": "string, optional - Kubernetes命名空间",
    "cluster": "string, optional - 集群名称",
    "user": "string, optional - 用户标识",
    "environment": "string, optional - 环境标识"
  }
}
```

#### 成功响应示例

```json
{
  "response": "在default命名空间中找到以下Pod:\n\n1. nginx-deployment-7d8b49557f-abc123 (Running)\n2. redis-master-6b8b4f4b4f-def456 (Running)\n3. mysql-0 (Running)\n\n所有Pod状态正常。",
  "session_id": "session_123",
  "status": "completed",
  "metadata": {
    "query_length": "35",
    "processed_at": "2025-01-21T10:30:00Z",
    "execution_time": "2.5s",
    "tools_used": ["kubectl"],
    "model_used": "deepseek-chat"
  },
  "timestamp": "2025-01-21T10:30:00Z"
}
```

#### 错误响应示例

```json
{
  "error": "查询处理失败: kubectl命令执行错误",
  "status": "error",
  "error_code": "KUBECTL_ERROR",
  "error_details": {
    "command": "kubectl get pods -n default",
    "exit_code": 1,
    "stderr": "Error from server (Forbidden): pods is forbidden"
  },
  "timestamp": "2025-01-21T10:30:00Z"
}
```

#### 复杂查询示例

```bash
# 复杂的集群分析查询
curl -X POST http://localhost:8888/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "query": "分析集群资源使用情况，找出CPU和内存使用率最高的Pod，并提供优化建议",
    "session_id": "analysis_session_001",
    "context": {
      "namespace": "production",
      "cluster": "main-cluster",
      "analysis_type": "resource_optimization"
    }
  }'
```

### 2. 流式聊天接口 - POST /api/v1/chat/stream

#### 基本请求

```bash
curl -X POST http://localhost:8888/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -H "Cache-Control: no-cache" \
  -d '{
    "query": "执行集群健康检查并生成详细报告",
    "session_id": "stream_session_001",
    "stream": true
  }'
```

#### 流式响应格式

```
data: {"delta": "开始执行集群健康检查...", "session_id": "stream_session_001", "status": "processing", "done": false, "timestamp": "2025-01-21T10:30:00Z"}

data: {"delta": "\n检查节点状态...", "session_id": "stream_session_001", "status": "processing", "done": false, "timestamp": "2025-01-21T10:30:01Z"}

data: {"delta": "\n- node1: Ready", "session_id": "stream_session_001", "status": "processing", "done": false, "timestamp": "2025-01-21T10:30:02Z"}

data: {"delta": "\n- node2: Ready", "session_id": "stream_session_001", "status": "processing", "done": false, "timestamp": "2025-01-21T10:30:03Z"}

data: {"response": "集群健康检查完成。所有节点状态正常，Pod运行正常，资源使用率在合理范围内。", "session_id": "stream_session_001", "status": "completed", "done": true, "metadata": {"execution_time": "5.2s", "checks_performed": 15}, "timestamp": "2025-01-21T10:30:05Z"}
```

#### JavaScript客户端示例

```javascript
async function streamChat(query, sessionId) {
    const response = await fetch('/api/v1/chat/stream', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Accept': 'text/event-stream'
        },
        body: JSON.stringify({
            query: query,
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
                
                if (data.delta) {
                    // 处理增量数据
                    console.log('Delta:', data.delta);
                    updateUI(data.delta);
                }
                
                if (data.done) {
                    // 处理完成
                    console.log('Completed:', data.response);
                    return data;
                }
            }
        }
    }
}
```

### 3. 健康检查接口 - GET /api/v1/health

#### 基本请求

```bash
curl -X GET http://localhost:8888/api/v1/health \
  -H "Accept: application/json"
```

#### 响应示例

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2025-01-21T10:30:00Z",
  "uptime": "2h30m15s",
  "components": {
    "llm_provider": {
      "status": "healthy",
      "provider": "deepseek",
      "model": "deepseek-chat",
      "last_check": "2025-01-21T10:29:55Z"
    },
    "kubernetes": {
      "status": "healthy",
      "cluster": "main-cluster",
      "version": "v1.28.0",
      "last_check": "2025-01-21T10:29:58Z"
    },
    "mcp_client": {
      "status": "enabled",
      "servers_connected": 3,
      "tools_available": 15,
      "last_check": "2025-01-21T10:29:59Z"
    }
  }
}
```

#### 健康检查脚本

```bash
#!/bin/bash
# health-check.sh

ENDPOINT="http://localhost:8888/api/v1/health"
TIMEOUT=10

response=$(curl -s --max-time $TIMEOUT "$ENDPOINT")
status=$(echo "$response" | jq -r '.status // "unknown"')

if [ "$status" = "healthy" ]; then
    echo "✅ kubelet-wuhrai服务健康"
    exit 0
else
    echo "❌ kubelet-wuhrai服务异常: $status"
    echo "响应: $response"
    exit 1
fi
```

### 4. 模型信息接口 - GET /api/v1/models

#### 基本请求

```bash
curl -X GET http://localhost:8888/api/v1/models \
  -H "Accept: application/json"
```

#### 响应示例

```json
{
  "models": [
    {
      "id": "deepseek-chat",
      "name": "DeepSeek Chat",
      "provider": "deepseek",
      "description": "通用对话模型，适合日常Kubernetes管理",
      "capabilities": ["chat", "function_calling", "reasoning"],
      "context_length": 32768,
      "cost_per_token": 0.0001
    },
    {
      "id": "deepseek-coder",
      "name": "DeepSeek Coder",
      "provider": "deepseek",
      "description": "代码专用模型，适合生成YAML配置和脚本",
      "capabilities": ["chat", "function_calling", "code_generation"],
      "context_length": 16384,
      "cost_per_token": 0.0002
    },
    {
      "id": "qwen-plus",
      "name": "Qwen Plus",
      "provider": "qwen",
      "description": "高性能通用模型",
      "capabilities": ["chat", "function_calling", "analysis"],
      "context_length": 32768,
      "cost_per_token": 0.00015
    }
  ],
  "current": "deepseek-chat",
  "provider": "deepseek",
  "timestamp": "2025-01-21T10:30:00Z",
  "metadata": {
    "total_models": 3,
    "providers": ["deepseek", "qwen", "doubao"],
    "default_provider": "deepseek",
    "supports_function_calling": true
  }
}
```

### 5. 状态接口 - GET /api/v1/status

#### 基本请求

```bash
curl -X GET http://localhost:8888/api/v1/status \
  -H "Accept: application/json"
```

#### 详细响应示例

```json
{
  "status": "running",
  "timestamp": "2025-01-21T10:30:00Z",
  "uptime": "2h30m15s",
  "version": "1.0.0",
  "build_info": {
    "commit": "abc123def456",
    "build_date": "2025-01-20T15:00:00Z",
    "go_version": "go1.21.0"
  },
  "system": {
    "os": "linux",
    "arch": "amd64",
    "cpu_cores": 8,
    "memory_total": "16GB",
    "memory_used": "2.5GB"
  },
  "components": {
    "agent": {
      "status": "ready",
      "active_sessions": 3,
      "total_queries": 1247,
      "avg_response_time": "2.3s"
    },
    "tools": {
      "status": "available",
      "kubectl": "enabled",
      "bash": "enabled",
      "custom_tools": 5,
      "mcp_tools": 10
    },
    "mcp_client": {
      "status": "enabled",
      "servers": [
        {
          "name": "sequential-thinking",
          "status": "connected",
          "tools": 3,
          "last_ping": "2025-01-21T10:29:58Z"
        },
        {
          "name": "github-api",
          "status": "connected",
          "tools": 7,
          "last_ping": "2025-01-21T10:29:59Z"
        }
      ]
    },
    "llm": {
      "provider": "deepseek",
      "model": "deepseek-chat",
      "status": "ready",
      "last_request": "2025-01-21T10:29:45Z",
      "total_tokens_used": 125000,
      "avg_tokens_per_request": 850
    }
  },
  "metrics": {
    "requests_total": 1247,
    "requests_success": 1198,
    "requests_error": 49,
    "success_rate": "96.1%",
    "avg_response_time": "2.3s",
    "p95_response_time": "5.1s",
    "p99_response_time": "8.7s"
  }
}
```

## SDK和客户端库

### 1. Python SDK完整示例

```python
import requests
import json
import time
import sseclient
from typing import Dict, Any, Optional, Iterator, Callable
from dataclasses import dataclass
from datetime import datetime

@dataclass
class ChatResponse:
    response: str
    session_id: str
    status: str
    metadata: Dict[str, Any]
    timestamp: datetime
    error: Optional[str] = None

class KubectlAIClient:
    def __init__(self, base_url: str = "http://localhost:8888", timeout: int = 30):
        self.base_url = base_url.rstrip('/')
        self.timeout = timeout
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': 'kubelet-wuhrai-python-client/1.0',
            'Accept': 'application/json'
        })

    def chat(self, query: str, session_id: Optional[str] = None,
             context: Optional[Dict[str, str]] = None) -> ChatResponse:
        """发送聊天请求"""
        url = f"{self.base_url}/api/v1/chat"
        payload = {
            "query": query,
            "session_id": session_id or f"session_{int(time.time())}",
            "context": context or {}
        }

        try:
            response = self.session.post(url, json=payload, timeout=self.timeout)
            response.raise_for_status()
            data = response.json()

            return ChatResponse(
                response=data.get('response', ''),
                session_id=data.get('session_id', ''),
                status=data.get('status', 'unknown'),
                metadata=data.get('metadata', {}),
                timestamp=datetime.fromisoformat(data.get('timestamp', '').replace('Z', '+00:00')),
                error=data.get('error')
            )
        except requests.exceptions.RequestException as e:
            raise Exception(f"API请求失败: {e}")

    def stream_chat(self, query: str, session_id: Optional[str] = None,
                   context: Optional[Dict[str, str]] = None,
                   on_delta: Optional[Callable[[str], None]] = None) -> Iterator[Dict[str, Any]]:
        """流式聊天请求"""
        url = f"{self.base_url}/api/v1/chat/stream"
        payload = {
            "query": query,
            "session_id": session_id or f"stream_session_{int(time.time())}",
            "stream": True,
            "context": context or {}
        }

        try:
            response = self.session.post(
                url,
                json=payload,
                headers={"Accept": "text/event-stream"},
                stream=True,
                timeout=self.timeout
            )
            response.raise_for_status()

            client = sseclient.SSEClient(response)
            for event in client.events():
                if event.data:
                    data = json.loads(event.data)

                    # 调用delta回调
                    if on_delta and data.get('delta'):
                        on_delta(data['delta'])

                    yield data

                    if data.get('done'):
                        break

        except requests.exceptions.RequestException as e:
            raise Exception(f"流式API请求失败: {e}")

    def get_health(self) -> Dict[str, Any]:
        """获取健康状态"""
        url = f"{self.base_url}/api/v1/health"
        response = self.session.get(url, timeout=self.timeout)
        response.raise_for_status()
        return response.json()

    def wait_for_ready(self, max_wait: int = 60) -> bool:
        """等待服务就绪"""
        start_time = time.time()
        while time.time() - start_time < max_wait:
            try:
                health = self.get_health()
                if health.get('status') == 'healthy':
                    return True
            except:
                pass
            time.sleep(1)
        return False

# 使用示例
client = KubectlAIClient()

# 等待服务就绪
if not client.wait_for_ready():
    print("服务未就绪")
    exit(1)

# 普通聊天
response = client.chat("列出所有命名空间中的Pod")
print(f"响应: {response.response}")

# 流式聊天
def on_delta(delta):
    print(delta, end='', flush=True)

print("\n流式响应:")
for chunk in client.stream_chat("分析集群资源使用情况", on_delta=on_delta):
    if chunk.get('done'):
        print(f"\n完成! 状态: {chunk.get('status')}")
        break
```

### 2. JavaScript SDK完整示例

```javascript
class KubectlAIClient {
    constructor(baseURL = 'http://localhost:8888', options = {}) {
        this.baseURL = baseURL.replace(/\/$/, '');
        this.timeout = options.timeout || 30000;
        this.headers = {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
            'User-Agent': 'kubelet-wuhrai-js-client/1.0',
            ...options.headers
        };
    }

    async chat(query, sessionId = null, context = {}) {
        const payload = {
            query,
            session_id: sessionId || `session_${Date.now()}`,
            context
        };

        const response = await fetch(`${this.baseURL}/api/v1/chat`, {
            method: 'POST',
            headers: this.headers,
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(`API请求失败: ${response.status} - ${errorData.error || response.statusText}`);
        }

        return await response.json();
    }

    async streamChat(query, options = {}) {
        const { sessionId, context, onDelta, onComplete } = options;

        const payload = {
            query,
            session_id: sessionId || `stream_session_${Date.now()}`,
            stream: true,
            context: context || {}
        };

        const response = await fetch(`${this.baseURL}/api/v1/chat/stream`, {
            method: 'POST',
            headers: {
                ...this.headers,
                'Accept': 'text/event-stream'
            },
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            throw new Error(`流式请求失败: ${response.status}`);
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();

        while (true) {
            const { done, value } = await reader.read();
            if (done) break;

            const chunk = decoder.decode(value);
            const lines = chunk.split('\n');

            for (const line of lines) {
                if (line.startsWith('data: ')) {
                    try {
                        const data = JSON.parse(line.slice(6));

                        if (data.delta && onDelta) {
                            onDelta(data.delta);
                        }

                        if (data.done) {
                            if (onComplete) onComplete(data);
                            return data;
                        }
                    } catch (e) {
                        console.warn('解析SSE数据失败:', e);
                    }
                }
            }
        }
    }

    async getHealth() {
        const response = await fetch(`${this.baseURL}/api/v1/health`, {
            headers: { Accept: 'application/json' }
        });
        return await response.json();
    }

    async waitForReady(maxWait = 60000) {
        const startTime = Date.now();

        while (Date.now() - startTime < maxWait) {
            try {
                const health = await this.getHealth();
                if (health.status === 'healthy') {
                    return true;
                }
            } catch (error) {
                // 忽略错误，继续等待
            }

            await new Promise(resolve => setTimeout(resolve, 1000));
        }

        return false;
    }
}

// 使用示例
const client = new KubectlAIClient();

// 等待服务就绪
await client.waitForReady();

// 普通聊天
const response = await client.chat('检查集群状态');
console.log(response.response);

// 流式聊天
await client.streamChat('分析Pod资源使用情况', {
    onDelta: (delta) => process.stdout.write(delta),
    onComplete: (data) => console.log('\n完成!', data.status)
});
```

## 错误处理和状态码

### HTTP状态码说明

| 状态码 | 含义 | 说明 |
|--------|------|------|
| 200 | OK | 请求成功 |
| 400 | Bad Request | 请求格式错误或参数无效 |
| 401 | Unauthorized | 认证失败 |
| 403 | Forbidden | 权限不足 |
| 404 | Not Found | 资源不存在 |
| 429 | Too Many Requests | 请求频率过高 |
| 500 | Internal Server Error | 服务器内部错误 |
| 502 | Bad Gateway | 网关错误 |
| 503 | Service Unavailable | 服务不可用 |
| 504 | Gateway Timeout | 网关超时 |

### 错误响应格式

```json
{
  "error": "错误描述信息",
  "status": "error",
  "error_code": "ERROR_CODE",
  "error_details": {
    "field": "具体错误字段",
    "message": "详细错误信息",
    "suggestion": "解决建议"
  },
  "timestamp": "2025-01-21T10:30:00Z",
  "request_id": "req_123456789"
}
```

### 常见错误处理

#### 1. 查询格式错误

```bash
# 错误请求
curl -X POST http://localhost:8888/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"invalid": "request"}'

# 错误响应
{
  "error": "查询是必需的",
  "status": "error",
  "error_code": "MISSING_QUERY",
  "error_details": {
    "field": "query",
    "message": "请求中缺少必需的query字段",
    "suggestion": "请在请求体中包含query字段"
  },
  "timestamp": "2025-01-21T10:30:00Z"
}
```

#### 2. 服务不可用错误

```bash
# 服务不可用时的响应
{
  "error": "LLM服务暂时不可用",
  "status": "error",
  "error_code": "LLM_UNAVAILABLE",
  "error_details": {
    "provider": "deepseek",
    "message": "无法连接到DeepSeek API",
    "suggestion": "请检查API密钥和网络连接，稍后重试"
  },
  "timestamp": "2025-01-21T10:30:00Z"
}
```

#### 3. 权限错误

```bash
# Kubernetes权限不足
{
  "error": "Kubernetes权限不足",
  "status": "error",
  "error_code": "KUBERNETES_FORBIDDEN",
  "error_details": {
    "resource": "pods",
    "namespace": "kube-system",
    "message": "当前用户无权访问kube-system命名空间的Pod",
    "suggestion": "请联系集群管理员获取相应权限"
  },
  "timestamp": "2025-01-21T10:30:00Z"
}
```

### 错误处理最佳实践

#### Python错误处理示例

```python
import requests
from typing import Optional

class KubectlAIError(Exception):
    def __init__(self, message: str, error_code: str = None, status_code: int = None):
        self.message = message
        self.error_code = error_code
        self.status_code = status_code
        super().__init__(message)

class KubectlAIClient:
    def chat_with_retry(self, query: str, max_retries: int = 3,
                       backoff_factor: float = 1.0) -> dict:
        """带重试机制的聊天请求"""
        for attempt in range(max_retries):
            try:
                return self.chat(query)
            except requests.exceptions.RequestException as e:
                if attempt == max_retries - 1:
                    raise KubectlAIError(f"请求失败，已重试{max_retries}次: {e}")

                wait_time = backoff_factor * (2 ** attempt)
                time.sleep(wait_time)
                continue
            except Exception as e:
                # 解析API错误响应
                if hasattr(e, 'response') and e.response:
                    try:
                        error_data = e.response.json()
                        raise KubectlAIError(
                            error_data.get('error', str(e)),
                            error_data.get('error_code'),
                            e.response.status_code
                        )
                    except:
                        pass
                raise KubectlAIError(str(e))

    def handle_api_error(self, error: KubectlAIError) -> Optional[str]:
        """处理API错误并返回用户友好的消息"""
        if error.error_code == "MISSING_QUERY":
            return "请提供要执行的查询内容"
        elif error.error_code == "LLM_UNAVAILABLE":
            return "AI服务暂时不可用，请稍后重试"
        elif error.error_code == "KUBERNETES_FORBIDDEN":
            return "权限不足，请联系管理员"
        elif error.status_code == 429:
            return "请求过于频繁，请稍后重试"
        else:
            return f"发生错误: {error.message}"

# 使用示例
client = KubectlAIClient()

try:
    response = client.chat_with_retry("列出所有Pod")
    print(response['response'])
except KubectlAIError as e:
    user_message = client.handle_api_error(e)
    print(f"错误: {user_message}")
```

## 性能优化

### 1. 连接池配置

```python
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

class OptimizedKubectlAIClient:
    def __init__(self, base_url: str):
        self.base_url = base_url
        self.session = requests.Session()

        # 配置重试策略
        retry_strategy = Retry(
            total=3,
            backoff_factor=1,
            status_forcelist=[429, 500, 502, 503, 504],
        )

        # 配置连接池
        adapter = HTTPAdapter(
            pool_connections=10,
            pool_maxsize=20,
            max_retries=retry_strategy
        )

        self.session.mount("http://", adapter)
        self.session.mount("https://", adapter)

        # 设置超时
        self.session.timeout = (5, 30)  # (连接超时, 读取超时)
```

### 2. 请求缓存

```python
import hashlib
import json
from functools import lru_cache
from typing import Dict, Any

class CachedKubectlAIClient:
    def __init__(self, base_url: str, cache_size: int = 128):
        self.base_url = base_url
        self.cache_size = cache_size

    def _cache_key(self, query: str, context: Dict[str, str]) -> str:
        """生成缓存键"""
        data = {"query": query, "context": context}
        return hashlib.md5(json.dumps(data, sort_keys=True).encode()).hexdigest()

    @lru_cache(maxsize=128)
    def _cached_chat(self, cache_key: str, query: str, context_str: str) -> dict:
        """缓存的聊天请求"""
        context = json.loads(context_str) if context_str else {}
        return self.chat(query, context=context)

    def chat_cached(self, query: str, context: Dict[str, str] = None) -> dict:
        """带缓存的聊天请求"""
        context = context or {}
        cache_key = self._cache_key(query, context)
        context_str = json.dumps(context, sort_keys=True)

        return self._cached_chat(cache_key, query, context_str)
```

### 3. 批量请求

```python
import asyncio
import aiohttp
from typing import List, Dict

class AsyncKubectlAIClient:
    def __init__(self, base_url: str):
        self.base_url = base_url

    async def batch_chat(self, queries: List[str],
                        concurrency: int = 5) -> List[Dict]:
        """批量处理聊天请求"""
        semaphore = asyncio.Semaphore(concurrency)

        async def single_chat(query: str) -> Dict:
            async with semaphore:
                async with aiohttp.ClientSession() as session:
                    async with session.post(
                        f"{self.base_url}/api/v1/chat",
                        json={"query": query}
                    ) as response:
                        return await response.json()

        tasks = [single_chat(query) for query in queries]
        return await asyncio.gather(*tasks)

# 使用示例
async def main():
    client = AsyncKubectlAIClient("http://localhost:8888")

    queries = [
        "列出所有Pod",
        "检查节点状态",
        "查看集群资源使用情况"
    ]

    results = await client.batch_chat(queries)
    for i, result in enumerate(results):
        print(f"查询 {i+1}: {result['response']}")

asyncio.run(main())
```

## 实际使用场景

### 1. 集群监控仪表板

```javascript
// 集群监控仪表板示例
class ClusterDashboard {
    constructor() {
        this.client = new KubectlAIClient();
        this.updateInterval = 30000; // 30秒更新一次
    }

    async initialize() {
        await this.client.waitForReady();
        this.startMonitoring();
    }

    async getClusterOverview() {
        const response = await this.client.chat(
            "提供集群概览，包括节点状态、Pod数量、资源使用情况",
            null,
            { dashboard: "overview" }
        );
        return this.parseOverviewResponse(response.response);
    }

    async getResourceAlerts() {
        const response = await this.client.chat(
            "检查是否有资源使用率超过80%的Pod或节点，如有请列出",
            null,
            { dashboard: "alerts" }
        );
        return this.parseAlertsResponse(response.response);
    }

    startMonitoring() {
        setInterval(async () => {
            try {
                const overview = await this.getClusterOverview();
                const alerts = await this.getResourceAlerts();

                this.updateDashboard(overview, alerts);
            } catch (error) {
                console.error('监控更新失败:', error);
            }
        }, this.updateInterval);
    }

    updateDashboard(overview, alerts) {
        // 更新仪表板UI
        document.getElementById('cluster-overview').innerHTML = overview;
        document.getElementById('alerts-panel').innerHTML = alerts;
    }
}

// 启动仪表板
const dashboard = new ClusterDashboard();
dashboard.initialize();
```

### 2. 自动化运维脚本

```python
#!/usr/bin/env python3
"""
自动化运维脚本示例
定期检查集群健康状态并执行维护任务
"""

import schedule
import time
from kubectl_ai_client import KubectlAIClient

class AutoOpsManager:
    def __init__(self):
        self.client = KubectlAIClient()
        self.session_id = f"auto_ops_{int(time.time())}"

    def daily_health_check(self):
        """每日健康检查"""
        try:
            response = self.client.chat(
                "执行全面的集群健康检查，包括节点状态、Pod健康、存储使用情况、网络连接性",
                session_id=self.session_id,
                context={"task": "daily_health_check"}
            )

            # 发送报告邮件
            self.send_health_report(response.response)

        except Exception as e:
            self.send_alert(f"健康检查失败: {e}")

    def cleanup_completed_jobs(self):
        """清理已完成的Job"""
        try:
            response = self.client.chat(
                "查找并删除状态为Completed且完成时间超过24小时的Job",
                session_id=self.session_id,
                context={"task": "cleanup_jobs"}
            )

            print(f"Job清理结果: {response.response}")

        except Exception as e:
            self.send_alert(f"Job清理失败: {e}")

    def check_resource_usage(self):
        """检查资源使用情况"""
        try:
            response = self.client.chat(
                "分析集群资源使用情况，如果发现资源使用率超过85%的节点，请提供扩容建议",
                session_id=self.session_id,
                context={"task": "resource_check"}
            )

            if "扩容建议" in response.response:
                self.send_alert(f"需要扩容: {response.response}")

        except Exception as e:
            self.send_alert(f"资源检查失败: {e}")

    def send_health_report(self, report: str):
        """发送健康报告"""
        # 实现邮件发送逻辑
        print(f"健康报告: {report}")

    def send_alert(self, message: str):
        """发送告警"""
        # 实现告警发送逻辑（邮件、Slack等）
        print(f"告警: {message}")

# 设置定时任务
ops_manager = AutoOpsManager()

# 每天早上8点执行健康检查
schedule.every().day.at("08:00").do(ops_manager.daily_health_check)

# 每天凌晨2点清理Job
schedule.every().day.at("02:00").do(ops_manager.cleanup_completed_jobs)

# 每小时检查资源使用情况
schedule.every().hour.do(ops_manager.check_resource_usage)

# 运行调度器
while True:
    schedule.run_pending()
    time.sleep(60)
```

### 3. CI/CD集成

```bash
#!/bin/bash
# ci-cd-integration.sh
# 在CI/CD流水线中集成kubelet-wuhrai

set -e

KUBECTL_AI_URL="${KUBECTL_AI_URL:-http://kubelet-wuhrai.internal:8888}"
ENVIRONMENT="${CI_ENVIRONMENT_NAME:-staging}"
APPLICATION="${CI_PROJECT_NAME}"

# 等待kubelet-wuhrai服务就绪
wait_for_service() {
    echo "等待kubelet-wuhrai服务就绪..."
    for i in {1..30}; do
        if curl -s "${KUBECTL_AI_URL}/api/v1/health" | jq -e '.status == "healthy"' > /dev/null; then
            echo "kubelet-wuhrai服务已就绪"
            return 0
        fi
        echo "等待中... ($i/30)"
        sleep 2
    done
    echo "kubelet-wuhrai服务未就绪"
    exit 1
}

# 部署前检查
pre_deployment_check() {
    echo "执行部署前检查..."

    response=$(curl -s -X POST "${KUBECTL_AI_URL}/api/v1/chat" \
        -H "Content-Type: application/json" \
        -d "{
            \"query\": \"检查${ENVIRONMENT}环境是否准备好部署${APPLICATION}应用，包括资源可用性、依赖服务状态\",
            \"session_id\": \"cicd_${CI_PIPELINE_ID}\",
            \"context\": {
                \"environment\": \"${ENVIRONMENT}\",
                \"application\": \"${APPLICATION}\",
                \"pipeline_id\": \"${CI_PIPELINE_ID}\"
            }
        }")

    status=$(echo "$response" | jq -r '.status')
    if [ "$status" != "completed" ]; then
        echo "部署前检查失败: $(echo "$response" | jq -r '.error')"
        exit 1
    fi

    echo "部署前检查通过"
}

# 执行部署
deploy_application() {
    echo "部署应用..."

    response=$(curl -s -X POST "${KUBECTL_AI_URL}/api/v1/chat" \
        -H "Content-Type: application/json" \
        -d "{
            \"query\": \"部署${APPLICATION}应用到${ENVIRONMENT}环境，使用镜像${CI_REGISTRY_IMAGE}:${CI_COMMIT_SHA}\",
            \"session_id\": \"cicd_${CI_PIPELINE_ID}\",
            \"context\": {
                \"environment\": \"${ENVIRONMENT}\",
                \"application\": \"${APPLICATION}\",
                \"image\": \"${CI_REGISTRY_IMAGE}:${CI_COMMIT_SHA}\",
                \"commit_sha\": \"${CI_COMMIT_SHA}\"
            }
        }")

    status=$(echo "$response" | jq -r '.status')
    if [ "$status" != "completed" ]; then
        echo "部署失败: $(echo "$response" | jq -r '.error')"
        exit 1
    fi

    echo "部署成功"
}

# 部署后验证
post_deployment_verification() {
    echo "执行部署后验证..."

    response=$(curl -s -X POST "${KUBECTL_AI_URL}/api/v1/chat" \
        -H "Content-Type: application/json" \
        -d "{
            \"query\": \"验证${APPLICATION}应用在${ENVIRONMENT}环境的部署状态，检查Pod是否正常运行，服务是否可访问\",
            \"session_id\": \"cicd_${CI_PIPELINE_ID}\",
            \"context\": {
                \"environment\": \"${ENVIRONMENT}\",
                \"application\": \"${APPLICATION}\",
                \"verification\": \"post_deployment\"
            }
        }")

    status=$(echo "$response" | jq -r '.status')
    if [ "$status" != "completed" ]; then
        echo "部署后验证失败: $(echo "$response" | jq -r '.error')"
        exit 1
    fi

    echo "部署后验证通过"
}

# 主流程
main() {
    wait_for_service
    pre_deployment_check
    deploy_application
    post_deployment_verification
    echo "CI/CD流水线执行完成"
}

main "$@"
```

通过以上详细的API调用指南，您可以充分利用kubelet-wuhrai的HTTP API功能，构建强大的Kubernetes管理和自动化系统。
