# kubelet-wuhrai 安装脚本说明

本项目提供了多种安装脚本，满足不同场景的需求。所有脚本都**不使用Docker**，直接在本地环境编译安装。

## 📦 可用脚本

| 脚本名称 | 用途 | 特点 | 推荐场景 |
|---------|------|------|----------|
| `one-click-install.sh` | 完整一键安装 | 自动安装Go环境 | 首次安装，无Go环境 |
| `quick-install.sh` | 快速安装 | 简洁快速 | 已有Go环境，快速部署 |
| `build.sh` | 编译打包 | 完整构建流程 | 开发者，需要发布包 |
| `install-local.sh` | 本地安装 | 安装已编译程序 | 从编译产物安装 |

## 🚀 推荐安装方式

### 方式1: 一键安装（零配置）

```bash
git clone <repository-url>
cd kubelet-wuhrai
./one-click-install.sh
```

**适用于：**
- 首次安装用户
- 没有Go环境的用户
- 希望完全自动化的用户

**功能：**
- ✅ 自动检测系统环境
- ✅ 询问并安装Go环境
- ✅ 编译所有模块
- ✅ 运行测试验证
- ✅ **智能安装到全局目录** (`/usr/local/bin`)
- ✅ **自动创建全局符号链接**
- ✅ **多shell环境变量配置** (bash/zsh/profile)
- ✅ 创建详细配置文件
- ✅ 支持卸载

### 方式2: 快速安装（推荐）

```bash
git clone <repository-url>
cd kubelet-wuhrai
./quick-install.sh
```

**适用于：**
- 已有Go 1.24+环境
- 希望快速安装的用户
- 开发者和高级用户

**功能：**
- ⚡ 快速编译安装
- ⚡ **优先安装到系统全局目录**
- ⚡ **自动创建全局符号链接**
- ⚡ **智能PATH配置**
- ⚡ 简洁输出界面

## 🔧 详细功能对比

### one-click-install.sh（完整版）

```bash
# 基本安装
./one-click-install.sh

# 卸载
./one-click-install.sh --uninstall

# 查看帮助
./one-click-install.sh --help
```

**特性：**
- 🔍 自动检测操作系统（Linux/macOS）
- 🔍 检查Go版本，支持自动安装Go 1.24.3
- 📦 编译所有子模块（gollm, k8s-bench, kubectl-utils）
- 🧪 运行完整测试套件
- 📁 创建发布包和校验和
- 🛠️ **优先安装到系统全局目录** (`/usr/local/bin`)
- 🔗 **自动创建全局符号链接**
- ⚙️ **多shell环境变量配置** (bash/zsh/profile)
- 📝 创建详细配置文件和环境变量示例
- 🗑️ 支持完整卸载

### quick-install.sh（快速版）

```bash
./quick-install.sh
```

**特性：**
- ⚡ 快速编译（仅主程序）
- ⚡ 简化的依赖检查
- ⚡ 自动PATH配置
- ⚡ 基础配置文件创建
- 📱 友好的用户界面

### build.sh（构建版）

```bash
./build.sh
```

**特性：**
- 🏗️ 完整构建流程
- 📦 生成发布包
- 🔐 生成校验和文件
- 🧪 运行测试
- 📊 详细构建信息

### install-local.sh（安装版）

```bash
./install-local.sh
```

**特性：**
- 📁 从bin/目录安装
- 🛠️ 配置PATH
- ⚙️ 创建配置文件

## 🎯 使用场景

### 场景1: 新用户首次安装

```bash
# 推荐使用完整一键安装
./one-click-install.sh
```

### 场景2: 开发者快速部署

```bash
# 推荐使用快速安装
./quick-install.sh
```

### 场景3: CI/CD自动化

```bash
# 推荐使用构建脚本
./build.sh
# 然后分发bin/kubelet-wuhrai
```

### 场景4: 离线安装

```bash
# 先在有网络的机器上构建
./build.sh

# 然后将bin/kubelet-wuhrai复制到目标机器
# 在目标机器上运行
./install-local.sh
```

## 🌟 全局安装优势

### ✅ 新版本改进

我们的安装脚本现在提供了更好的全局安装体验：

1. **智能安装目录选择**
   - 优先安装到 `/usr/local/bin` (系统全局目录)
   - 自动检测sudo权限并询问用户
   - 回退到用户目录 (`~/go/bin`, `~/.local/bin`)

2. **全局符号链接**
   - 即使安装到用户目录，也会创建全局符号链接
   - 确保在任何目录下都能使用 `kubelet-wuhrai` 命令

3. **多Shell支持**
   - 自动配置 bash (`~/.bashrc`)
   - 自动配置 zsh (`~/.zshrc`)
   - 通用配置 (`~/.profile`)

4. **智能验证**
   - 检查多个可能的安装位置
   - 验证全局命令可用性
   - 提供详细的故障排除信息

## 📋 安装后验证

```bash
# 检查安装位置
which kubelet-wuhrai

# 检查版本
kubelet-wuhrai version

# 查看帮助
kubelet-wuhrai --help

# 测试运行（需要API密钥）
kubelet-wuhrai --quiet "获取pod列表"

# 在任意目录下测试
cd /tmp && kubelet-wuhrai version
```

## ⚙️ 环境变量配置

安装完成后，需要配置LLM API密钥：

```bash
# DeepSeek（默认）
export DEEPSEEK_API_KEY="your-api-key"

# 或其他提供商
export OPENAI_API_KEY="your-api-key"
export QWEN_API_KEY="your-api-key"
export DOUBAO_API_KEY="your-api-key"
```

## 🔄 更新和维护

### 更新到最新版本

```bash
git pull
./quick-install.sh  # 快速重新安装
```

### 完全重新安装

```bash
./one-click-install.sh --uninstall  # 卸载
./one-click-install.sh              # 重新安装
```

## 🐛 故障排除

### 常见问题

1. **Go版本问题**
   ```bash
   # 使用一键安装自动处理
   ./one-click-install.sh
   ```

2. **权限问题**
   ```bash
   chmod +x *.sh
   ```

3. **网络问题**
   ```bash
   export GOPROXY=https://goproxy.cn,direct
   ```

4. **PATH问题**
   ```bash
   source ~/.bashrc  # 或 ~/.zshrc
   ```

## 📞 获取帮助

如果遇到问题：

1. 查看脚本帮助：`./one-click-install.sh --help`
2. 检查Go环境：`go version`
3. 查看详细日志输出
4. 提交Issue并附上错误信息

## 🚀 kubelet-wuhrai 完整功能列表

### 🤖 多LLM提供商支持

| 提供商 | 参数值 | 支持的模型 | 环境变量 |
|--------|--------|------------|----------|
| **DeepSeek** (默认) | `deepseek` | `deepseek-chat`, `deepseek-coder`, `deepseek-reasoner` | `DEEPSEEK_API_KEY` |
| **通义千问** | `qwen` | `qwen-plus`, `qwen-turbo`, `qwen-max`, `qwen2.5-*` 系列 | `QWEN_API_KEY` |
| **豆包** | `doubao` | `doubao-pro-4k`, `doubao-lite-4k`, `doubao-pro-vision` 等 | `DOUBAO_API_KEY` |
| **OpenAI** | `openai` | `gpt-4`, `gpt-3.5-turbo` 等 | `OPENAI_API_KEY` |
| **OpenAI兼容** | `openai-compatible` | 自定义模型 | `OPENAI_API_KEY` + `OPENAI_API_BASE` |
| **Gemini** | `gemini` | `gemini-pro`, `gemini-pro-vision` | `GEMINI_API_KEY` |

### 🔧 核心功能

- ✅ **MCP协议支持**: 客户端/服务器模式，外部工具集成
- ✅ **HTTP API服务**: RESTful接口，流式响应
- ✅ **多UI模式**: 终端交互 (`terminal`) / Web界面 (`html`)
- ✅ **工具系统**: kubectl、bash、自定义工具、MCP工具
- ✅ **配置管理**: YAML配置文件，环境变量支持

### 🛠️ 高级功能

- ✅ **自定义工具配置**: `--custom-tools-config`
- ✅ **MCP客户端模式**: `--mcp-client`
- ✅ **MCP服务器模式**: `--mcp-server`
- ✅ **外部工具发现**: `--external-tools`
- ✅ **提示模板自定义**: `--prompt-template-file-path`
- ✅ **工具使用垫片**: `--enable-tool-use-shim`
- ✅ **调试追踪**: `--trace-path`

## 🌐 跨主机部署

### ✅ 二进制文件可移植性

kubelet-wuhrai编译为**静态链接**的二进制文件，可以直接复制到其他主机使用：

- ✅ 包含所有依赖，无需Go环境
- ✅ 支持Linux x86_64架构
- ✅ 单文件部署，简单可靠

### 📦 部署方式

#### 方式1: 手动复制
```bash
# 复制到远程主机
scp bin/kubelet-wuhrai user@target-host:/tmp/
ssh user@target-host "sudo mv /tmp/kubelet-wuhrai /usr/local/bin/ && sudo chmod +x /usr/local/bin/kubelet-wuhrai"
```

#### 方式2: 使用部署脚本
```bash
# 一键部署到远程主机
./deploy-to-remote.sh user@target-host --install-kubectl --copy-kubeconfig --setup-env
```

#### 方式3: 发布包部署
```bash
# 创建发布包
./build.sh

# 复制并解压
scp dist/kubelet-wuhrai-*.tar.gz user@target-host:/tmp/
ssh user@target-host "cd /tmp && tar -xzf kubelet-wuhrai-*.tar.gz && sudo mv kubelet-wuhrai /usr/local/bin/"
```

### ⚠️ 依赖要求

目标主机需要：
- **kubectl**: 执行Kubernetes命令
- **kubeconfig**: 集群访问配置
- **API密钥**: LLM服务认证

---

**选择建议：**
- 🆕 新用户 → `one-click-install.sh`
- ⚡ 快速安装 → `quick-install.sh`
- 🏗️ 开发构建 → `build.sh`
- 📦 离线安装 → `build.sh` + `install-local.sh`
- 🌐 远程部署 → `deploy-to-remote.sh`

## 📚 完整文档索引

### 🔧 核心文档
- [安装指南](INSTALL_GUIDE.md) - 详细安装步骤和故障排除
- [构建指南](BUILD_GUIDE.md) - 编译打包完整流程
- [使用指南](USAGE.md) - 基本使用方法和命令

### 🛠️ 高级功能文档
- **[自定义工具和MCP指南](CUSTOM_TOOLS_GUIDE.md)** - 自定义工具和MCP工具详细使用
- [API调用示例](API_USAGE_EXAMPLES.md) - curl命令调用示例
- [扩展技术指南](docs/EXTENDED_TECHNICAL_GUIDE.md) - 完整技术文档
- [MCP详细指南](docs/MCP_DETAILED_GUIDE.md) - MCP协议深度使用
- [API详细指南](docs/API_DETAILED_GUIDE.md) - HTTP API完整文档

### 📦 示例和模板
- [自定义工具配置](examples/custom-tools.yaml) - 工具配置模板
- [MCP配置示例](examples/mcp-config.yaml) - MCP服务器配置
- [测试脚本](examples/test-custom-tools.sh) - 功能测试脚本

### 🚀 快速开始

#### 使用自定义工具
```bash
# 复制示例配置
cp examples/custom-tools.yaml ~/.config/kubelet-wuhrai/tools.yaml

# 使用自定义工具
kubelet-wuhrai --custom-tools-config ~/.config/kubelet-wuhrai/tools.yaml "检查系统资源使用情况"
```

#### 使用MCP工具
```bash
# 复制MCP配置
cp examples/mcp-config.yaml ~/.config/kubelet-wuhrai/mcp.yaml

# 设置环境变量
export GITHUB_TOKEN="your-token"

# 启动MCP客户端
kubelet-wuhrai --mcp-client --custom-tools-config ~/.config/kubelet-wuhrai/mcp.yaml "使用外部工具分析集群"
```

#### 启动HTTP API服务
```bash
# 启动Web界面和API服务
kubelet-wuhrai --user-interface=html --ui-listen-address=0.0.0.0:8888

# 通过API调用
curl -X POST http://localhost:8888/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "获取所有pod", "session_id": "test"}'
```
