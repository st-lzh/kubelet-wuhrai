# 🔧 故障排除指南 | Troubleshooting Guide

本文档包含kubelet-wuhrai常见问题的解决方案。

## 🚨 常见问题

### 1. 安装问题

#### Q: 编译时出现Go版本错误
```
error: go version go1.20.x is not supported
```

**解决方案：**
```bash
# 升级Go到1.24+
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

#### Q: 依赖下载失败
```
error: module not found
```

**解决方案：**
```bash
# 清理模块缓存
go clean -modcache

# 重新下载依赖
go mod download

# 如果仍有问题，尝试代理
export GOPROXY=https://goproxy.cn,direct
go mod download
```

### 2. 连接问题

#### Q: API连接超时
```
Error: context deadline exceeded
```

**解决方案：**
1. 检查网络连接
2. 验证API端点是否正确
3. 检查防火墙设置
4. 使用`--skip-verify-ssl`跳过SSL验证

```bash
kubelet-wuhrai --skip-verify-ssl "your query"
```

#### Q: API密钥无效
```
Error: 401 Unauthorized
```

**解决方案：**
1. 验证API密钥格式
2. 检查密钥是否过期
3. 确认API端点匹配

```bash
# 设置正确的环境变量
export OPENAI_API_KEY="your-valid-key"
export OPENAI_API_BASE="https://your-endpoint.com/v1"
```

### 3. Kubernetes连接问题

#### Q: kubectl命令失败
```
Error: unable to connect to kubernetes cluster
```

**解决方案：**
1. 检查kubeconfig文件
2. 验证集群连接
3. 确认权限设置

```bash
# 测试kubectl连接
kubectl cluster-info

# 指定kubeconfig文件
kubelet-wuhrai --kubeconfig /path/to/kubeconfig "your query"
```

#### Q: 权限不足
```
Error: forbidden: User cannot list pods
```

**解决方案：**
1. 检查RBAC权限
2. 使用有权限的用户
3. 联系集群管理员

### 4. 运行时问题

#### Q: 命令重复执行
```
Running: kubectl get nodes -o wide
Running: kubectl get nodes -o wide
...
```

**解决方案：**
1. 使用更具体的查询
2. 限制迭代次数

```bash
# 使用具体查询
kubelet-wuhrai "获取集群节点列表"

# 限制迭代次数
kubelet-wuhrai --max-iterations 3 "your query"
```

#### Q: 响应不完整
```
AI response was cut off...
```

**解决方案：**
1. 增加超时时间
2. 简化查询内容
3. 分步执行复杂任务

### 5. 自定义工具问题

#### Q: 自定义工具不生效
```
Error: tool not found
```

**解决方案：**
1. 检查配置文件路径
2. 验证YAML格式
3. 确认工具可执行权限

```bash
# 指定自定义工具配置
kubelet-wuhrai --custom-tools-config /path/to/tools.yaml "your query"

# 验证配置文件
yamllint /path/to/tools.yaml
```

## 🔍 调试技巧

### 启用详细日志
```bash
# 启用调试日志
kubelet-wuhrai -v=2 "your query"

# 启用最详细日志
kubelet-wuhrai -v=5 "your query"
```

### 查看跟踪信息
```bash
# 启用跟踪
kubelet-wuhrai --trace-path /tmp/trace.txt "your query"

# 查看跟踪文件
cat /tmp/trace.txt
```

### 测试API连接
```bash
# 测试OpenAI兼容API
curl -X POST "https://your-endpoint.com/v1/chat/completions" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

## 📊 性能优化

### 1. 减少响应时间
- 使用更快的AI模型
- 限制迭代次数
- 使用本地模型

### 2. 降低API成本
- 选择成本较低的模型
- 优化查询内容
- 使用缓存机制

### 3. 提高准确性
- 使用更具体的查询
- 提供更多上下文
- 使用专业模型

## 🆘 获取帮助

如果以上解决方案都无法解决您的问题：

1. **查看日志文件**
   ```bash
   kubelet-wuhrai -v=2 "your query" 2>&1 | tee debug.log
   ```

2. **创建Issue**
   - 访问：https://github.com/st-lzh/kubelet-wuhrai/issues
   - 包含完整的错误信息和环境信息
   - 提供复现步骤

3. **联系支持**
   - 邮箱：lzh094285@gmail.com
   - 包含调试日志和系统信息

## 📋 环境信息收集

创建Issue时，请提供以下信息：

```bash
# 系统信息
uname -a

# Go版本
go version

# kubelet-wuhrai版本
kubelet-wuhrai version

# Kubernetes版本
kubectl version

# 网络连接测试
curl -I https://your-api-endpoint.com
```

## 🔄 常用命令

```bash
# 重新编译
go build -o bin/kubelet-wuhrai ./cmd

# 清理缓存
go clean -cache -modcache

# 运行测试
go test ./...

# 格式化代码
go fmt ./...

# 检查代码
go vet ./...
```
