# vLLM和SGLang本地模型支持使用指南

## 概述

kubelet-wuhrai现在支持vLLM和SGLang本地模型部署，提供流式返回功能，让您能够使用自建的大语言模型与Kubernetes集群进行交互。

## vLLM支持

### 1. 启动vLLM服务

```bash
# 使用Docker启动vLLM服务
docker run --gpus all \
    -p 8000:8000 \
    vllm/vllm-openai:latest \
    --model microsoft/DialoGPT-medium \
    --served-model-name default

# 或者直接使用Python
python -m vllm.entrypoints.openai.api_server \
    --model microsoft/DialoGPT-medium \
    --port 8000
```

### 2. 配置环境变量

```bash
# 设置vLLM服务地址（默认为 http://localhost:8000）
export VLLM_URL="http://localhost:8000"

# 设置模型名称（可选，默认为 "default"）
export VLLM_MODEL="microsoft/DialoGPT-medium"

# 设置API密钥（可选，vLLM通常不需要）
export VLLM_API_KEY="dummy-key"

# 设置LLM客户端为vLLM
export LLM_CLIENT="vllm"
```

### 3. 使用vLLM

```bash
# 基本使用
kubelet-wuhrai "获取所有Pod状态"

# 启动Web界面
kubelet-wuhrai --user-interface html

# 指定特定模型
kubelet-wuhrai --model "your-model-name" "检查集群健康状态"
```

## SGLang支持

### 1. 启动SGLang服务

```bash
# 使用Python启动SGLang服务
python -m sglang.launch_server \
    --model-path microsoft/DialoGPT-medium \
    --port 30000 \
    --host 0.0.0.0

# 或使用Docker
docker run --gpus all \
    -p 30000:30000 \
    sglang/sglang:latest \
    --model-path microsoft/DialoGPT-medium \
    --port 30000
```

### 2. 配置环境变量

```bash
# 设置SGLang服务地址（默认为 http://localhost:30000）
export SGLANG_URL="http://localhost:30000"

# 设置模型名称（可选，默认为 "default"）
export SGLANG_MODEL="microsoft/DialoGPT-medium"

# 设置API密钥（可选，SGLang通常不需要）
export SGLANG_API_KEY="dummy-key"

# 设置LLM客户端为SGLang
export LLM_CLIENT="sglang"
```

### 3. 使用SGLang

```bash
# 基本使用
kubelet-wuhrai "显示所有命名空间"

# 查看集群资源
kubelet-wuhrai "检查哪些节点内存使用率最高"

# 应用管理
kubelet-wuhrai "部署一个nginx应用到default命名空间"
```

## 流式返回功能

kubelet-wuhrai默认启用流式返回，提供实时响应体验：

### 特性
- **实时响应**：AI回答会逐步显示，无需等待完整响应
- **即时反馈**：长时间任务可以实时查看进度
- **优化体验**：减少用户等待时间

### 在终端中的表现
```bash
kubelet-wuhrai "分析集群中所有Pod的资源使用情况"

# 输出会逐步显示：
# 正在分析集群中的Pod资源使用情况...
# 
# 发现以下Pod资源使用情况：
# 1. kube-system命名空间：
#    - coredns-xxx: CPU 10m, Memory 64Mi
#    - etcd-xxx: CPU 50m, Memory 128Mi
# ...（内容逐步展示）
```

### Web界面中的表现
- 文本会实时流式显示
- 执行的命令结果会即时更新
- 提供更好的交互体验

## 配置选项

### 通用配置

```yaml
# ~/.config/kubelet-wuhrai/config.yaml

# vLLM配置
vllm_url: "http://localhost:8000"
vllm_model: "your-model-name"
vllm_api_key: "optional-api-key"

# SGLang配置  
sglang_url: "http://localhost:30000"
sglang_model: "your-model-name"
sglang_api_key: "optional-api-key"

# 其他设置
quiet: false
skip_permissions: false
enable_tool_use_shim: false
```

### 环境变量优先级

1. 命令行参数（最高优先级）
2. 环境变量
3. 配置文件
4. 默认值（最低优先级）

## 故障排除

### vLLM连接问题
```bash
# 检查vLLM服务状态
curl http://localhost:8000/v1/models

# 检查vLLM日志
docker logs <vllm-container-id>
```

### SGLang连接问题
```bash
# 检查SGLang服务状态
curl http://localhost:30000/v1/models

# 测试聊天接口
curl -X POST http://localhost:30000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "default", "messages": [{"role": "user", "content": "Hello"}]}'
```

### 常见错误

1. **连接被拒绝**
   - 确保服务正在运行
   - 检查端口是否正确
   - 验证防火墙设置

2. **模型加载失败**
   - 检查模型路径是否正确
   - 确保有足够的GPU内存
   - 查看服务启动日志

3. **API密钥错误**
   - vLLM和SGLang通常不需要API密钥
   - 如果需要，请设置正确的环境变量

## 性能优化

### vLLM优化
- 使用适当的GPU设备
- 调整`--max-model-len`参数
- 启用张量并行：`--tensor-parallel-size 2`

### SGLang优化  
- 使用RadixAttention缓存
- 调整批处理大小
- 启用内存池：`--mem-pool-size 20`

## 与其他模型提供商的比较

| 特性 | vLLM | SGLang | Ollama | OpenAI |
|------|------|--------|---------|---------|
| 本地部署 | ✅ | ✅ | ✅ | ❌ |
| 流式响应 | ✅ | ✅ | ✅ | ✅ |
| GPU加速 | ✅ | ✅ | ✅ | N/A |
| OpenAI兼容 | ✅ | ✅ | ✅ | ✅ |
| 工具调用 | ✅ | ✅ | ✅ | ✅ |
| 批处理优化 | ✅ | ✅ | ❌ | N/A |

kubelet-wuhrai现在支持所有这些提供商，让您可以根据需求选择最适合的解决方案。