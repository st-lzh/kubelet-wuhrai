# kubelet-wuhrai 扩展版本

这是kubelet-wuhrai的扩展版本，新增了多个AI模型提供商支持和HTTP API服务功能。

## 🚀 新增功能

### 1. 多AI模型支持
- **DeepSeek** (默认): deepseek-chat, deepseek-coder, deepseek-reasoner
- **通义千问Qwen**: qwen-plus, qwen-turbo, qwen-max, qwen2.5系列
- **字节跳动豆包**: doubao-pro-4k, doubao-lite-4k, doubao-pro-vision等
- **VLLM和第三方OpenAI兼容服务**: 支持自定义endpoint

### 2. HTTP API服务
- RESTful API接口
- 流式聊天支持
- 健康检查和状态监控
- JavaScript/Python SDK

### 3. 增强的MCP支持
- 客户端和服务器模式
- 多种认证方式
- 自定义工具集成

## 📦 快速开始

### 1. 构建项目

```bash
# 克隆项目
git clone <your-repo-url>
cd kubelet-wuhrai

# 构建
go mod tidy
go build -o kubelet-wuhrai ./cmd/

# 或使用make（如果有Makefile）
make build
```

### 2. 配置API密钥

```bash
# DeepSeek (默认)
export DEEPSEEK_API_KEY="your_deepseek_api_key"

# 通义千问
export DASHSCOPE_API_KEY="your_dashscope_api_key"

# 豆包
export VOLCES_API_KEY="your_volces_api_key"

# OpenAI兼容服务
export OPENAI_API_KEY="your_api_key"
export OPENAI_ENDPOINT="http://your-server:8000/v1"
```

### 3. 基本使用

```bash
# 使用默认DeepSeek模型
./kubelet-wuhrai "list all pods in default namespace"

# 指定特定模型
./kubelet-wuhrai --llm-provider=qwen --model=qwen-plus "describe deployment nginx"

# 启动HTTP服务
./kubelet-wuhrai --user-interface=html --ui-listen-address=0.0.0.0:8888
```

## 🔧 配置文件

创建配置文件 `~/.config/kubelet-wuhrai/config.yaml`:

```yaml
llm-provider: "deepseek"
model: "deepseek-chat"
user-interface: "html"
ui-listen-address: "0.0.0.0:8888"
max-iterations: 10
skip-permissions: false
```

## 🌐 HTTP API使用

### 启动服务
```bash
./kubelet-wuhrai --user-interface=html --ui-listen-address=0.0.0.0:8888
```

### API调用示例
```bash
# 聊天接口
curl -X POST http://localhost:8888/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "list all pods"}'

# 健康检查
curl http://localhost:8888/api/v1/health

# 获取可用模型
curl http://localhost:8888/api/v1/models
```

## 🐳 Docker部署

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
# 构建和运行
docker build -t kubelet-wuhrai:extended .
docker run -d -p 8888:8888 \
  -e DEEPSEEK_API_KEY="your_api_key" \
  -v ~/.kube:/root/.kube:ro \
  kubelet-wuhrai:extended
```

## 📚 详细文档

完整的技术文档请参考：[docs/EXTENDED_TECHNICAL_GUIDE.md](docs/EXTENDED_TECHNICAL_GUIDE.md)

文档包含：
- 详细的模型配置指南
- API接口规范和示例
- MCP配置详解
- 部署指南
- 故障排查
- 性能优化建议

## 🛠️ 开发和贡献

### 项目结构
```
kubelet-wuhrai/
├── cmd/                    # 主程序入口
├── pkg/                    # 核心包
│   ├── agent/             # AI代理
│   ├── tools/             # 工具系统
│   ├── mcp/               # MCP支持
│   └── ui/                # 用户界面
├── gollm/                 # AI模型抽象层
│   ├── deepseek.go        # DeepSeek提供商
│   ├── qwen.go            # 通义千问提供商
│   ├── doubao.go          # 豆包提供商
│   └── openai.go          # OpenAI兼容提供商
└── docs/                  # 文档
```

### 添加新的AI提供商

1. 在`gollm/`目录创建新的提供商文件
2. 实现`Client`和`Chat`接口
3. 在`init()`函数中注册提供商
4. 更新文档和测试

## 🔍 故障排查

### 常见问题

1. **API密钥错误**
   ```bash
   export DEEPSEEK_API_KEY="your_correct_api_key"
   ```

2. **网络连接问题**
   ```bash
   export LLM_SKIP_VERIFY_SSL=true
   ```

3. **模型不可用**
   ```bash
   ./kubelet-wuhrai models  # 查看可用模型
   ```

### 调试模式
```bash
# 启用详细日志
./kubelet-wuhrai -v=2 "your query"

# 查看trace
./kubelet-wuhrai --trace-path=/tmp/trace.txt "your query"
```

## 📄 许可证

本项目基于Apache 2.0许可证开源。

## 🤝 支持

如有问题或建议，请：
1. 查看[技术文档](docs/EXTENDED_TECHNICAL_GUIDE.md)
2. 检查[故障排查指南](docs/EXTENDED_TECHNICAL_GUIDE.md#故障排查指南)
3. 提交Issue或Pull Request

---

**注意**: 这是kubelet-wuhrai的扩展版本，包含了额外的AI模型支持和HTTP API功能。使用前请确保已正确配置相应的API密钥。
