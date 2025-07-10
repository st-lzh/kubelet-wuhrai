# 🤝 贡献指南 | Contributing Guide

感谢您对kubelet-wuhrai项目的关注！我们欢迎并感谢所有形式的贡献。

## 🌟 贡献方式

### 📝 报告问题
- 使用GitHub Issues报告bug
- 提供详细的复现步骤
- 包含系统环境信息

### 💡 功能建议
- 通过Issues提出新功能建议
- 详细描述功能需求和使用场景
- 讨论实现方案

### 🔧 代码贡献
- Fork项目到您的GitHub账户
- 创建功能分支进行开发
- 提交Pull Request

## 📋 开发准备

### 环境要求
- Go 1.24+
- Kubernetes集群（用于测试）
- Git

### 本地开发设置
```bash
# 克隆项目
git clone https://github.com/st-lzh/kubelet-wuhrai.git
cd kubelet-wuhrai

# 安装依赖
go mod download

# 编译项目
go build -o bin/kubelet-wuhrai ./cmd
```

## 🔄 贡献流程

### 1. 准备工作
```bash
# Fork项目并克隆
git clone https://github.com/YOUR_USERNAME/kubelet-wuhrai.git
cd kubelet-wuhrai

# 添加上游仓库
git remote add upstream https://github.com/st-lzh/kubelet-wuhrai.git
```

### 2. 创建功能分支
```bash
# 创建并切换到新分支
git checkout -b feature/your-feature-name

# 或者修复bug
git checkout -b fix/your-bug-fix
```

### 3. 开发和测试
```bash
# 编译项目
go build -o bin/kubelet-wuhrai ./cmd

# 运行测试
go test ./...

# 代码格式化
go fmt ./...
```

### 4. 提交代码
```bash
# 添加更改
git add .

# 提交（使用清晰的提交信息）
git commit -m "feat: add new feature description"

# 推送到您的fork
git push origin feature/your-feature-name
```

### 5. 创建Pull Request
- 在GitHub上创建Pull Request
- 填写详细的PR描述
- 等待代码审查

## 📁 项目结构

- `cmd/` - kubelet-wuhrai CLI主程序
- `pkg/` - 核心功能包
  - `agent/` - AI对话和决策逻辑
  - `tools/` - kubectl、bash和自定义工具
  - `mcp/` - MCP协议支持
  - `ui/` - 用户界面（终端和Web）
- `gollm/` - LLM客户端实现
- `examples/` - 使用示例和配置文件
- `docs/` - 项目文档
- `k8s-bench/` - 性能评估工具

## 📝 代码规范

### 提交信息格式
```
type(scope): description

[optional body]

[optional footer]
```

类型说明：
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式化
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

### Go代码规范
- 遵循Go官方代码规范
- 使用`go fmt`格式化代码
- 添加必要的注释
- 编写单元测试

## 🧪 测试指南

### 运行测试
```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./pkg/agent

# 运行测试并显示覆盖率
go test -cover ./...
```

### 集成测试
```bash
# 设置测试环境变量
export OPENAI_API_KEY="your-test-key"
export OPENAI_API_BASE="your-test-endpoint"

# 运行集成测试
./bin/kubelet-wuhrai "获取集群节点信息"
```

## 📄 许可证

通过贡献代码，您同意您的贡献将在与项目相同的许可证下发布。

## 🆘 获取帮助

如果您在贡献过程中遇到问题：

- 查看现有的Issues和Pull Requests
- 创建新的Issue描述您的问题
- 发送邮件至：lzh094285@gmail.com

## 🙏 致谢

感谢所有贡献者的努力！您的贡献让kubelet-wuhrai变得更好。
