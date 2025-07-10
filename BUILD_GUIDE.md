# kubelet-wuhrai 编译打包指南

本指南介绍如何在不使用Docker的情况下编译和打包 kubelet-wuhrai 项目。

## 🚀 快速开始

### 前置要求

- Go 1.24.0 或更高版本
- Git
- Linux/macOS/Windows 环境

### 一键编译打包

```bash
# 克隆项目（如果还没有）
git clone <repository-url>
cd kubelet-wuhrai

# 运行编译脚本
./build.sh
```

## 📁 输出文件

编译完成后，您将得到以下文件：

```
bin/
├── kubelet-wuhrai              # 可执行二进制文件 (~43MB)

dist/
├── kubelet-wuhrai-dev-linux-x86_64.tar.gz        # 发布包 (~20MB)
└── kubelet-wuhrai-dev-linux-x86_64.tar.gz.sha256 # 校验和文件
```

## 🔧 手动编译步骤

如果您想手动编译，可以按照以下步骤：

### 1. 下载依赖

```bash
go mod download
```

### 2. 编译主程序

```bash
mkdir -p bin
go build -ldflags "-X main.version=dev -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/kubelet-wuhrai ./cmd
```

### 3. 编译子模块

```bash
# 编译 gollm 模块
cd gollm && go build ./... && cd ..

# 编译 k8s-bench 模块
cd k8s-bench && go mod tidy && go build ./... && cd ..

# 编译 kubectl-utils 模块
cd kubectl-utils && go mod tidy && go build ./... && cd ..
```

### 4. 运行测试

```bash
go test ./... -v
```

## 📦 安装到系统

使用提供的安装脚本将编译好的程序安装到系统中：

```bash
./install-local.sh
```

安装脚本会：
- 将二进制文件复制到适当的目录（如 `$GOPATH/bin` 或 `~/.local/bin`）
- 创建配置目录 `~/.config/kubelet-wuhrai`
- 生成示例配置文件
- 检查 PATH 设置

## 🎯 验证安装

```bash
# 检查版本
kubelet-wuhrai version

# 查看帮助
kubelet-wuhrai --help

# 测试运行（需要配置 LLM API 密钥）
kubelet-wuhrai --quiet "获取所有pod"
```

## 🔍 构建脚本功能

`build.sh` 脚本提供以下功能：

- ✅ 自动检测 Go 环境
- ✅ 编译主程序和所有子模块
- ✅ 运行完整测试套件
- ✅ 生成版本信息
- ✅ 创建发布包和校验和
- ✅ 彩色输出和详细日志
- ✅ 错误处理和验证

## 📋 支持的平台

当前构建脚本支持：
- Linux (x86_64, arm64)
- macOS (Intel, Apple Silicon)
- Windows (通过 WSL 或 Git Bash)

## 🛠️ 自定义构建

您可以通过环境变量自定义构建：

```bash
# 自定义版本信息
VERSION=v1.0.0 COMMIT=abc123 DATE=2025-01-01T00:00:00Z ./build.sh

# 仅编译不运行测试
SKIP_TESTS=1 ./build.sh
```

## 🐛 故障排除

### 常见问题

1. **Go 版本过低**
   ```
   解决方案: 升级到 Go 1.24.0 或更高版本
   ```

2. **依赖下载失败**
   ```bash
   # 清理模块缓存
   go clean -modcache
   go mod download
   ```

3. **权限问题**
   ```bash
   # 确保脚本有执行权限
   chmod +x build.sh install-local.sh
   ```

4. **PATH 问题**
   ```bash
   # 添加到 PATH
   echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   ```

### 获取帮助

如果遇到问题，请：
1. 检查 Go 版本：`go version`
2. 检查依赖：`go mod verify`
3. 查看详细日志：运行构建脚本时的输出
4. 提交 Issue 并附上错误信息

## 📝 配置文件

安装后，您可以在 `~/.config/kubelet-wuhrai/config.yaml` 中配置：

```yaml
# LLM 提供商配置
llmProvider: "deepseek"
model: "deepseek-chat"

# 基本设置
skipPermissions: false
quiet: false
maxIterations: 20

# UI 设置
userInterface: "terminal"
uiListenAddress: "localhost:8888"
```

## 🎉 完成

现在您已经成功编译并安装了 kubelet-wuhrai！

使用 `kubelet-wuhrai --help` 查看所有可用选项，开始您的 Kubernetes 自然语言交互之旅。
