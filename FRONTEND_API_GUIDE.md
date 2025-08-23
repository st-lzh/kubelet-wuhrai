# kubelet-wuhrai 前端集成API文档

## 🎯 概述

kubelet-wuhrai提供完整的REST API接口，支持前端应用集成。所有API都支持流式返回，提供实时的AI交互体验。

---

## 🚀 服务启动

### 启动Web API服务

```bash
# 基础启动
kubelet-wuhrai --user-interface html --ui-listen-address 0.0.0.0:8888

# 生产环境启动（推荐配置）
kubelet-wuhrai \
  --user-interface html \
  --ui-listen-address 0.0.0.0:8888 \
  --llm-provider openrouter \
  --model "anthropic/claude-3.5-sonnet" \
  --skip-verify-ssl=false \
  --max-iterations 20 \
  --kubeconfig ~/.kube/readonly-config

# Docker容器化启动
docker run -d \
  --name kubelet-wuhrai-api \
  -p 8888:8888 \
  -v ~/.kube:/root/.kube:ro \
  -e OPENROUTER_API_KEY=your-key \
  -e LLM_CLIENT=openrouter \
  kubelet-wuhrai:latest \
  --user-interface html --ui-listen-address 0.0.0.0:8888
```

---

## 📡 API接口规范

### 基础URL
```
http://localhost:8888
```

### 认证方式
- 当前版本无需认证（内网使用）
- 生产环境建议添加API密钥或JWT认证

---

## 🔗 核心API端点

### 1. 聊天查询接口

#### `POST /api/chat`

**功能**: 发送自然语言查询并获取Kubernetes操作结果

**请求格式**:
```json
{
  "query": "获取所有Pod状态",
  "options": {
    "llm_provider": "openrouter",
    "model": "anthropic/claude-3.5-sonnet",
    "stream": true,
    "max_iterations": 10,
    "namespace": "default",
    "kubeconfig": "/path/to/kubeconfig"
  },
  "context": {
    "session_id": "unique-session-id",
    "user_id": "user-123"
  }
}
```

**响应格式** (非流式):
```json
{
  "status": "success",
  "response": "发现3个Pod正在运行:\n1. nginx-xxx (Running)\n2. redis-xxx (Running)\n3. app-xxx (Pending)",
  "session_id": "unique-session-id",
  "metadata": {
    "execution_time": "2.3s",
    "commands_executed": ["kubectl get pods --all-namespaces"],
    "model_used": "anthropic/claude-3.5-sonnet",
    "token_usage": {
      "input_tokens": 150,
      "output_tokens": 89
    }
  },
  "kubectl_commands": [
    {
      "command": "kubectl get pods --all-namespaces",
      "output": "NAMESPACE   NAME        READY   STATUS    RESTARTS   AGE\ndefault     nginx-xxx   1/1     Running   0          1h",
      "exit_code": 0
    }
  ]
}
```

**响应格式** (流式):
```
Content-Type: text/event-stream

data: {"type": "thinking", "content": "正在分析您的请求..."}

data: {"type": "command", "content": "kubectl get pods --all-namespaces"}

data: {"type": "output", "content": "NAMESPACE   NAME        READY   STATUS\n"}

data: {"type": "text", "content": "发现3个Pod正在运行：\n"}

data: {"type": "text", "content": "1. nginx-xxx (Running)\n"}

data: {"type": "done", "metadata": {"execution_time": "2.3s"}}
```

### 2. 流式聊天接口

#### `POST /api/stream`

**功能**: 服务器发送事件(SSE)流式聊天

**JavaScript示例**:
```javascript
// 流式调用示例
async function streamQuery(query, options = {}) {
  const response = await fetch('/api/stream', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'text/event-stream',
    },
    body: JSON.stringify({
      query: query,
      options: {
        llm_provider: 'openrouter',
        model: 'anthropic/claude-3.5-sonnet',
        stream: true,
        ...options
      }
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
        handleStreamData(data);
      }
    }
  }
}

function handleStreamData(data) {
  switch (data.type) {
    case 'thinking':
      showThinking(data.content);
      break;
    case 'command':
      showCommand(data.content);
      break;
    case 'output':
      showOutput(data.content);
      break;
    case 'text':
      appendText(data.content);
      break;
    case 'done':
      showComplete(data.metadata);
      break;
    case 'error':
      showError(data.content);
      break;
  }
}
```

### 3. 模型管理接口

#### `GET /api/models`

**功能**: 获取可用的AI模型列表

**响应格式**:
```json
{
  "status": "success",
  "providers": {
    "openrouter": {
      "models": [
        {
          "id": "anthropic/claude-3.5-sonnet",
          "name": "Claude 3.5 Sonnet",
          "description": "最新的Claude模型，适合复杂推理",
          "pricing": {"input": 3.0, "output": 15.0},
          "context_length": 200000
        },
        {
          "id": "openai/gpt-4o",
          "name": "GPT-4o",
          "description": "OpenAI最新多模态模型",
          "pricing": {"input": 2.5, "output": 10.0},
          "context_length": 128000
        }
      ]
    },
    "vllm": {
      "models": ["default"],
      "base_url": "http://localhost:8000",
      "status": "available"
    }
  }
}
```

#### `GET /api/providers`

**功能**: 获取支持的LLM提供商列表

**响应格式**:
```json
{
  "status": "success",
  "providers": [
    {
      "id": "openrouter",
      "name": "OpenRouter",
      "description": "模型聚合服务，支持数百个模型",
      "type": "cloud",
      "status": "available",
      "default_model": "anthropic/claude-3.5-sonnet"
    },
    {
      "id": "vllm", 
      "name": "vLLM",
      "description": "高性能本地模型服务",
      "type": "local",
      "status": "available",
      "base_url": "http://localhost:8000"
    }
  ]
}
```

### 4. 集群信息接口

#### `GET /api/cluster/info`

**功能**: 获取Kubernetes集群基础信息

**响应格式**:
```json
{
  "status": "success",
  "cluster": {
    "name": "prod-cluster",
    "version": "v1.28.2",
    "nodes": 5,
    "namespaces": 12,
    "pods": 156,
    "services": 34,
    "health": "healthy"
  },
  "kubeconfig": "/path/to/current/kubeconfig"
}
```

### 5. 命令执行接口

#### `POST /api/execute`

**功能**: 直接执行kubectl命令

**请求格式**:
```json
{
  "command": "kubectl get pods --all-namespaces",
  "options": {
    "timeout": 30,
    "stream": true,
    "confirm_dangerous": false
  }
}
```

**响应格式**:
```json
{
  "status": "success",
  "command": "kubectl get pods --all-namespaces",
  "output": "NAMESPACE   NAME        READY   STATUS    RESTARTS   AGE\n...",
  "exit_code": 0,
  "execution_time": "1.2s",
  "is_dangerous": false
}
```

---

## 🎨 前端集成示例

### React.js集成

```jsx
// KubeletWuhraiClient.js
class KubeletWuhraiClient {
  constructor(baseURL = 'http://localhost:8888') {
    this.baseURL = baseURL;
  }

  // 流式查询
  async streamQuery(query, options = {}, onData) {
    const response = await fetch(`${this.baseURL}/api/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'text/event-stream',
      },
      body: JSON.stringify({
        query,
        options: {
          llm_provider: 'openrouter',
          model: 'anthropic/claude-3.5-sonnet',
          stream: true,
          ...options
        }
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
          try {
            const data = JSON.parse(line.slice(6));
            onData(data);
          } catch (e) {
            console.error('Parse error:', e);
          }
        }
      }
    }
  }

  // 非流式查询
  async query(query, options = {}) {
    const response = await fetch(`${this.baseURL}/api/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        query,
        options: {
          llm_provider: 'openrouter',
          model: 'anthropic/claude-3.5-sonnet',
          ...options
        }
      })
    });

    return await response.json();
  }

  // 获取模型列表
  async getModels() {
    const response = await fetch(`${this.baseURL}/api/models`);
    return await response.json();
  }

  // 获取集群信息
  async getClusterInfo() {
    const response = await fetch(`${this.baseURL}/api/cluster/info`);
    return await response.json();
  }
}

// React组件示例
import React, { useState, useEffect } from 'react';

const KubeletChat = () => {
  const [client] = useState(new KubeletWuhraiClient());
  const [query, setQuery] = useState('');
  const [response, setResponse] = useState('');
  const [isStreaming, setIsStreaming] = useState(false);
  const [models, setModels] = useState([]);
  const [selectedModel, setSelectedModel] = useState('anthropic/claude-3.5-sonnet');

  useEffect(() => {
    client.getModels().then(data => {
      if (data.status === 'success') {
        const allModels = Object.values(data.providers).flatMap(p => p.models || []);
        setModels(allModels);
      }
    });
  }, []);

  const handleStreamQuery = async () => {
    setIsStreaming(true);
    setResponse('');

    await client.streamQuery(
      query,
      { model: selectedModel },
      (data) => {
        switch (data.type) {
          case 'thinking':
            setResponse(prev => prev + `🤔 ${data.content}\n`);
            break;
          case 'command':
            setResponse(prev => prev + `💻 执行: ${data.content}\n`);
            break;
          case 'text':
            setResponse(prev => prev + data.content);
            break;
          case 'done':
            setIsStreaming(false);
            break;
          case 'error':
            setResponse(prev => prev + `❌ 错误: ${data.content}\n`);
            setIsStreaming(false);
            break;
        }
      }
    );
  };

  return (
    <div className="kubelet-chat">
      <div className="model-selector">
        <select 
          value={selectedModel} 
          onChange={(e) => setSelectedModel(e.target.value)}
          disabled={isStreaming}
        >
          {models.map(model => (
            <option key={model.id} value={model.id}>
              {model.name} ({model.pricing?.input}$/1M tokens)
            </option>
          ))}
        </select>
      </div>

      <div className="query-input">
        <textarea
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="输入Kubernetes相关查询，例如：获取所有Pod状态"
          disabled={isStreaming}
        />
        <button 
          onClick={handleStreamQuery}
          disabled={isStreaming || !query.trim()}
        >
          {isStreaming ? '处理中...' : '发送查询'}
        </button>
      </div>

      <div className="response-area">
        <pre>{response}</pre>
        {isStreaming && <div className="typing-indicator">AI正在思考...</div>}
      </div>
    </div>
  );
};

export default KubeletChat;
```

### Vue.js集成

```vue
<!-- KubeletChat.vue -->
<template>
  <div class="kubelet-chat">
    <!-- 模型选择器 -->
    <div class="model-selector">
      <select v-model="selectedModel" :disabled="isStreaming">
        <option v-for="model in models" :key="model.id" :value="model.id">
          {{ model.name }} ({{ model.pricing?.input }}$/1M tokens)
        </option>
      </select>
    </div>

    <!-- 查询输入 -->
    <div class="query-input">
      <textarea 
        v-model="query"
        :disabled="isStreaming"
        placeholder="输入Kubernetes查询..."
      ></textarea>
      <button @click="streamQuery" :disabled="isStreaming || !query.trim()">
        {{ isStreaming ? '处理中...' : '发送查询' }}
      </button>
    </div>

    <!-- 响应显示 -->
    <div class="response-area">
      <pre>{{ response }}</pre>
      <div v-if="isStreaming" class="typing-indicator">🤖 AI正在处理...</div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue';

export default {
  name: 'KubeletChat',
  setup() {
    const query = ref('');
    const response = ref('');
    const isStreaming = ref(false);
    const models = ref([]);
    const selectedModel = ref('anthropic/claude-3.5-sonnet');
    const baseURL = 'http://localhost:8888';

    const fetchModels = async () => {
      try {
        const res = await fetch(`${baseURL}/api/models`);
        const data = await res.json();
        if (data.status === 'success') {
          models.value = Object.values(data.providers).flatMap(p => p.models || []);
        }
      } catch (error) {
        console.error('获取模型失败:', error);
      }
    };

    const streamQuery = async () => {
      isStreaming.value = true;
      response.value = '';

      try {
        const res = await fetch(`${baseURL}/api/stream`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'text/event-stream',
          },
          body: JSON.stringify({
            query: query.value,
            options: {
              llm_provider: 'openrouter',
              model: selectedModel.value,
              stream: true
            }
          })
        });

        const reader = res.body.getReader();
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
                handleStreamData(data);
              } catch (e) {
                console.error('解析错误:', e);
              }
            }
          }
        }
      } catch (error) {
        response.value += `❌ 错误: ${error.message}\n`;
      } finally {
        isStreaming.value = false;
      }
    };

    const handleStreamData = (data) => {
      switch (data.type) {
        case 'thinking':
          response.value += `🤔 ${data.content}\n`;
          break;
        case 'command':
          response.value += `💻 执行: ${data.content}\n`;
          break;
        case 'text':
          response.value += data.content;
          break;
        case 'done':
          response.value += `\n✅ 完成 (${data.metadata?.execution_time})\n`;
          break;
        case 'error':
          response.value += `❌ 错误: ${data.content}\n`;
          break;
      }
    };

    onMounted(() => {
      fetchModels();
    });

    return {
      query,
      response,
      isStreaming,
      models,
      selectedModel,
      streamQuery
    };
  }
};
</script>
```

### Angular集成

```typescript
// kubelet-wuhrai.service.ts
import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, Subject } from 'rxjs';

export interface QueryOptions {
  llm_provider?: string;
  model?: string;
  stream?: boolean;
  max_iterations?: number;
  namespace?: string;
}

export interface StreamData {
  type: 'thinking' | 'command' | 'output' | 'text' | 'done' | 'error';
  content: string;
  metadata?: any;
}

@Injectable({
  providedIn: 'root'
})
export class KubeletWuhraiService {
  private baseURL = 'http://localhost:8888';
  
  constructor(private http: HttpClient) {}

  // 非流式查询
  query(query: string, options: QueryOptions = {}): Observable<any> {
    return this.http.post(`${this.baseURL}/api/chat`, {
      query,
      options: {
        llm_provider: 'openrouter',
        model: 'anthropic/claude-3.5-sonnet',
        ...options
      }
    });
  }

  // 流式查询
  streamQuery(query: string, options: QueryOptions = {}): Observable<StreamData> {
    const subject = new Subject<StreamData>();

    fetch(`${this.baseURL}/api/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'text/event-stream',
      },
      body: JSON.stringify({
        query,
        options: {
          llm_provider: 'openrouter',
          model: 'anthropic/claude-3.5-sonnet',
          stream: true,
          ...options
        }
      })
    }).then(response => {
      const reader = response.body!.getReader();
      const decoder = new TextDecoder();

      const readStream = async () => {
        while (true) {
          const { done, value } = await reader.read();
          if (done) {
            subject.complete();
            break;
          }

          const chunk = decoder.decode(value);
          const lines = chunk.split('\n');

          for (const line of lines) {
            if (line.startsWith('data: ')) {
              try {
                const data = JSON.parse(line.slice(6));
                subject.next(data);
              } catch (e) {
                subject.error(e);
              }
            }
          }
        }
      };

      readStream().catch(error => subject.error(error));
    }).catch(error => subject.error(error));

    return subject.asObservable();
  }

  // 获取模型列表
  getModels(): Observable<any> {
    return this.http.get(`${this.baseURL}/api/models`);
  }

  // 获取集群信息
  getClusterInfo(): Observable<any> {
    return this.http.get(`${this.baseURL}/api/cluster/info`);
  }
}
```

### 原生JavaScript集成

```javascript
// kubelet-client.js
class KubeletWuhraiAPI {
  constructor(baseURL = 'http://localhost:8888') {
    this.baseURL = baseURL;
  }

  // 通用请求方法
  async request(endpoint, options = {}) {
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers
      },
      ...options
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return response.json();
  }

  // 流式请求
  async streamRequest(endpoint, data, onData) {
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'text/event-stream',
      },
      body: JSON.stringify(data)
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder();

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      const chunk = decoder.decode(value);
      chunk.split('\n').forEach(line => {
        if (line.startsWith('data: ')) {
          try {
            const data = JSON.parse(line.slice(6));
            onData(data);
          } catch (e) {
            console.error('Parse error:', e);
          }
        }
      });
    }
  }

  // API方法
  async queryCluster(query, options = {}) {
    return this.request('/api/chat', {
      method: 'POST',
      body: JSON.stringify({
        query,
        options: {
          llm_provider: 'openrouter',
          model: 'anthropic/claude-3.5-sonnet',
          ...options
        }
      })
    });
  }

  async streamQueryCluster(query, options = {}, onData) {
    return this.streamRequest('/api/stream', {
      query,
      options: {
        llm_provider: 'openrouter',
        model: 'anthropic/claude-3.5-sonnet',
        stream: true,
        ...options
      }
    }, onData);
  }

  async getModels() {
    return this.request('/api/models');
  }

  async getProviders() {
    return this.request('/api/providers');
  }

  async getClusterInfo() {
    return this.request('/api/cluster/info');
  }

  async executeCommand(command, options = {}) {
    return this.request('/api/execute', {
      method: 'POST',
      body: JSON.stringify({
        command,
        options: {
          timeout: 30,
          stream: false,
          ...options
        }
      })
    });
  }
}

// 使用示例
const api = new KubeletWuhraiAPI();

// 1. 基础查询
api.queryCluster('获取所有Pod状态')
  .then(result => console.log(result))
  .catch(error => console.error(error));

// 2. 流式查询
api.streamQueryCluster(
  '分析集群性能瓶颈并提供优化建议',
  { model: 'anthropic/claude-3.5-sonnet' },
  (data) => {
    console.log('Stream data:', data);
    // 处理流式数据
  }
);

// 3. 获取可用模型
api.getModels().then(models => {
  console.log('Available models:', models);
});

// 4. 直接执行命令
api.executeCommand('kubectl get pods').then(result => {
  console.log('Command result:', result);
});
```

---

## 🛡️ 错误处理和重试机制

### 错误类型定义

```javascript
class KubeletError extends Error {
  constructor(type, message, details = {}) {
    super(message);
    this.type = type;
    this.details = details;
  }
}

// 错误处理示例
async function safeQuery(query, options = {}) {
  const maxRetries = 3;
  let lastError;

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      return await api.queryCluster(query, options);
    } catch (error) {
      lastError = error;
      
      if (error.status === 429) { // 限流
        await sleep(Math.pow(2, attempt) * 1000); // 指数退避
        continue;
      }
      
      if (error.status >= 500) { // 服务器错误
        if (attempt < maxRetries) {
          await sleep(1000 * attempt);
          continue;
        }
      }
      
      throw error; // 不可重试的错误
    }
  }
  
  throw lastError;
}

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}
```

---

## 📊 实时监控和状态管理

### WebSocket连接（如果后续支持）

```javascript
class KubeletWebSocket {
  constructor(url = 'ws://localhost:8888/ws') {
    this.url = url;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
  }

  connect() {
    this.ws = new WebSocket(this.url);
    
    this.ws.onopen = () => {
      console.log('WebSocket连接已建立');
      this.reconnectAttempts = 0;
    };

    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      this.handleMessage(data);
    };

    this.ws.onclose = () => {
      console.log('WebSocket连接已关闭');
      this.attemptReconnect();
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket错误:', error);
    };
  }

  attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      setTimeout(() => {
        console.log(`尝试重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
        this.connect();
      }, Math.pow(2, this.reconnectAttempts) * 1000);
    }
  }

  send(data) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  handleMessage(data) {
    // 处理实时消息
    console.log('收到消息:', data);
  }
}
```

---

## 🎯 生产环境部署建议

### 1. 安全配置

```yaml
# docker-compose.yml
version: '3.8'
services:
  kubelet-wuhrai:
    image: kubelet-wuhrai:latest
    ports:
      - "127.0.0.1:8888:8888"  # 只监听本地
    environment:
      - OPENROUTER_API_KEY=${OPENROUTER_API_KEY}
      - LLM_CLIENT=openrouter
    volumes:
      - ./kubeconfig:/root/.kube:ro
      - ./logs:/var/log/kubelet-wuhrai
    restart: unless-stopped
    command: >
      --user-interface html
      --ui-listen-address 0.0.0.0:8888
      --skip-permissions=false
      --max-iterations 20
```

### 2. 负载均衡配置

```nginx
# nginx.conf
upstream kubelet_wuhrai {
    server 127.0.0.1:8888;
    server 127.0.0.1:8889;  # 多实例
}

server {
    listen 80;
    server_name k8s-ai.company.com;

    location /api/ {
        proxy_pass http://kubelet_wuhrai;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # 支持SSE流式响应
        proxy_buffering off;
        proxy_cache off;
    }
}
```

### 3. 监控和日志

```javascript
// 监控客户端状态
class HealthMonitor {
  constructor(apiClient) {
    this.api = apiClient;
    this.isHealthy = true;
    this.lastCheck = null;
  }

  async startMonitoring(interval = 30000) {
    setInterval(async () => {
      try {
        await this.api.getClusterInfo();
        this.isHealthy = true;
        this.lastCheck = new Date();
      } catch (error) {
        this.isHealthy = false;
        console.error('健康检查失败:', error);
      }
    }, interval);
  }

  getStatus() {
    return {
      healthy: this.isHealthy,
      lastCheck: this.lastCheck,
      uptime: Date.now() - this.startTime
    };
  }
}
```

---

## 🎨 UI组件建议

### 推荐的前端组件

1. **查询输入组件**
   - 智能提示（常用K8s查询）
   - 历史记录
   - 快捷操作按钮

2. **响应显示组件**
   - 代码高亮
   - 可复制的kubectl命令
   - 执行结果表格化显示

3. **模型选择组件**
   - 成本显示
   - 性能指标
   - 推荐标签

4. **状态监控组件**
   - 连接状态
   - 响应时间
   - API使用统计

### CSS样式建议

```css
.kubelet-chat {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.model-selector {
  margin-bottom: 20px;
}

.query-input {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.query-input textarea {
  flex: 1;
  min-height: 100px;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 10px;
}

.response-area {
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 20px;
  background: #f8f9fa;
  min-height: 300px;
}

.typing-indicator {
  color: #007bff;
  font-style: italic;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
```

---

## 🚀 快速开始模板

### HTML + JavaScript 简单示例

```html
<!DOCTYPE html>
<html>
<head>
    <title>Kubelet Wuhrai Web Client</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; }
        .container { max-width: 800px; margin: 0 auto; padding: 20px; }
        textarea { width: 100%; height: 100px; margin: 10px 0; }
        button { padding: 10px 20px; background: #007bff; color: white; border: none; border-radius: 4px; }
        button:disabled { background: #ccc; }
        .response { border: 1px solid #ddd; padding: 20px; background: #f8f9fa; white-space: pre-wrap; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Kubelet Wuhrai AI Assistant</h1>
        
        <div>
            <select id="modelSelect">
                <option value="anthropic/claude-3.5-sonnet">Claude 3.5 Sonnet</option>
                <option value="openai/gpt-4o">GPT-4o</option>
                <option value="openai/gpt-4o-mini">GPT-4o Mini (便宜)</option>
            </select>
        </div>
        
        <textarea id="queryInput" placeholder="输入Kubernetes查询，例如：获取所有Pod状态"></textarea>
        <button onclick="executeQuery()" id="queryBtn">发送查询</button>
        
        <div id="response" class="response"></div>
    </div>

    <script>
        const api = new KubeletWuhraiAPI();
        
        async function executeQuery() {
            const query = document.getElementById('queryInput').value;
            const model = document.getElementById('modelSelect').value;
            const responseDiv = document.getElementById('response');
            const btn = document.getElementById('queryBtn');
            
            if (!query.trim()) return;
            
            btn.disabled = true;
            btn.textContent = '处理中...';
            responseDiv.textContent = '';
            
            try {
                await api.streamQueryCluster(
                    query,
                    { model: model },
                    (data) => {
                        switch (data.type) {
                            case 'thinking':
                                responseDiv.textContent += `🤔 ${data.content}\n`;
                                break;
                            case 'command':
                                responseDiv.textContent += `💻 执行: ${data.content}\n`;
                                break;
                            case 'text':
                                responseDiv.textContent += data.content;
                                break;
                            case 'done':
                                responseDiv.textContent += `\n✅ 完成\n`;
                                break;
                            case 'error':
                                responseDiv.textContent += `❌ 错误: ${data.content}\n`;
                                break;
                        }
                        responseDiv.scrollTop = responseDiv.scrollHeight;
                    }
                );
            } catch (error) {
                responseDiv.textContent += `❌ 连接错误: ${error.message}`;
            } finally {
                btn.disabled = false;
                btn.textContent = '发送查询';
            }
        }

        // KubeletWuhraiAPI类定义（如前所述）
        class KubeletWuhraiAPI {
            constructor(baseURL = 'http://localhost:8888') {
                this.baseURL = baseURL;
            }

            async streamQueryCluster(query, options = {}, onData) {
                const response = await fetch(`${this.baseURL}/api/stream`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Accept': 'text/event-stream',
                    },
                    body: JSON.stringify({
                        query,
                        options: {
                            llm_provider: 'openrouter',
                            model: 'anthropic/claude-3.5-sonnet',
                            stream: true,
                            ...options
                        }
                    })
                });

                const reader = response.body.getReader();
                const decoder = new TextDecoder();

                while (true) {
                    const { done, value } = await reader.read();
                    if (done) break;

                    const chunk = decoder.decode(value);
                    chunk.split('\n').forEach(line => {
                        if (line.startsWith('data: ')) {
                            try {
                                const data = JSON.parse(line.slice(6));
                                onData(data);
                            } catch (e) {
                                console.error('Parse error:', e);
                            }
                        }
                    });
                }
            }
        }
    </script>
</body>
</html>
```

---

## 📋 API测试清单

### 使用curl测试

```bash
# 1. 测试基础查询
curl -X POST http://localhost:8888/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "query": "获取所有Pod状态",
    "options": {
      "llm_provider": "openrouter",
      "model": "anthropic/claude-3.5-sonnet"
    }
  }'

# 2. 测试流式查询
curl -X POST http://localhost:8888/api/stream \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{
    "query": "分析集群健康状态",
    "options": {
      "stream": true,
      "model": "openai/gpt-4o-mini"
    }
  }' \
  --no-buffer

# 3. 测试模型列表
curl http://localhost:8888/api/models

# 4. 测试集群信息
curl http://localhost:8888/api/cluster/info

# 5. 测试命令执行
curl -X POST http://localhost:8888/api/execute \
  -H "Content-Type: application/json" \
  -d '{
    "command": "kubectl get pods",
    "options": {"timeout": 30}
  }'
```

### 前端开发checklist

- [ ] 实现流式响应处理
- [ ] 添加错误重试机制
- [ ] 实现模型选择器
- [ ] 添加查询历史记录
- [ ] 实现响应内容高亮
- [ ] 添加加载状态指示器
- [ ] 实现健康监控
- [ ] 添加使用统计
- [ ] 实现批量操作
- [ ] 添加导出功能

---

## 🎯 总结

kubelet-wuhrai现在提供了完整的前端集成方案：

1. **REST API**: 完整的HTTP接口
2. **流式SSE**: 实时响应体验
3. **多框架支持**: React/Vue/Angular示例
4. **错误处理**: 完善的重试机制
5. **生产部署**: Docker和负载均衡配置

所有API都支持OpenRouter的数百个模型，为前端提供最大的灵活性和最佳的用户体验！