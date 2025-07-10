# kubelet-wuhrai 使用指南

## 🚀 快速开始

### 1. 设置API密钥

```bash
# DeepSeek (推荐，默认)
export DEEPSEEK_API_KEY="your_deepseek_api_key"

# 通义千问 (可选)
export DASHSCOPE_API_KEY="your_dashscope_api_key"

# 豆包 (可选)
export VOLCES_API_KEY="your_volces_api_key"
```

### 2. 构建和运行

```bash
# 构建
go mod tidy
go build -o kubelet-wuhrai ./cmd/

# 基本使用
./kubelet-wuhrai "list all pods"

# 启动Web界面
./kubelet-wuhrai --user-interface=html --ui-listen-address=localhost:8888
```

### 3. Docker部署

```bash
# 设置环境变量
cp .env.example .env
# 编辑.env文件，填入API密钥

# 启动服务
docker-compose up -d

# 访问Web界面
open http://localhost:8888
```

## 📋 常用命令

```bash
# 查看版本
./kubelet-wuhrai version

# 查看帮助
./kubelet-wuhrai --help

# 指定模型
./kubelet-wuhrai --model=deepseek-coder "generate deployment yaml"

# 指定提供商
./kubelet-wuhrai --llm-provider=qwen --model=qwen-plus "analyze cluster"

# 启用MCP客户端
./kubelet-wuhrai --mcp-client "your query"

# 启动MCP服务器
./kubelet-wuhrai --mcp-server
```

## 🔧 配置文件

配置文件位置：`~/.config/kubelet-wuhrai/config.yaml`

```yaml
llm-provider: "deepseek"
model: "deepseek-chat"
user-interface: "terminal"
max-iterations: 10
skip-permissions: false
mcp-client: false
```

## 📚 更多文档

- [详细技术指南](docs/EXTENDED_TECHNICAL_GUIDE.md)
- [MCP使用指南](docs/MCP_DETAILED_GUIDE.md)
- [API调用指南](docs/API_DETAILED_GUIDE.md)
- [部署指南](docs/DEPLOYMENT_DETAILED_GUIDE.md)

## 🆘 故障排查

### 常见问题

1. **API密钥错误**
   ```bash
   export DEEPSEEK_API_KEY="正确的密钥"
   ```

2. **网络连接问题**
   ```bash
   # 设置代理
   export HTTP_PROXY=http://proxy:8080
   export HTTPS_PROXY=http://proxy:8080
   ```

3. **权限问题**
   ```bash
   # 检查kubeconfig
   kubectl cluster-info
   ```

### 获取帮助

- 查看日志：`tail -f /tmp/kubelet-wuhrai.log`
- 启用调试：`./kubelet-wuhrai -v=2 "your query"`
- 提交Issue：[GitHub Issues](https://github.com/st-lzh/kubelet-wuhrai/issues)
