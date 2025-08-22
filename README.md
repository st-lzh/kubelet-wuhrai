<div align="center">

# 🚀 kubelet-wuhrai

**智能Kubernetes管理工具 | AI-Powered Kubernetes Management Tool**

[![License](https://img.shields.io/badge/License-Custom-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-Compatible-326CE5?logo=kubernetes)](https://kubernetes.io/)
[![AI Powered](https://img.shields.io/badge/AI-Powered-FF6B6B?logo=openai)](https://openai.com/)

*基于自然语言与Kubernetes集群交互的智能命令行工具*

[🚀 快速开始](#-快速开始) • [🛠️ 功能特性](#️-功能特性) • [📦 构建部署](#-构建部署) • [🤝 贡献](#-贡献)

</div>

---

## 📖 项目简介

kubelet-wuhrai 是一个革命性的Kubernetes管理工具，让您能够使用自然语言与Kubernetes集群进行交互。通过集成先进的大语言模型，它能理解您的意图并自动转换为相应的kubectl命令。

### 🌟 基于开源项目

本项目基于Google的 [kubectl-ai](https://github.com/GoogleCloudPlatform/kubectl-ai) 进行二次开发，在原有功能基础上增加了更多AI模型支持、自定义工具集成和MCP协议支持。

## 🛠️ 功能特性

### 🤖 多AI模型支持
- **DeepSeek** (默认) - 高性能代码生成模型
- **OpenAI** - GPT-4, GPT-3.5-turbo等
- **通义千问** - 阿里云Qwen系列模型  
- **豆包** - 字节跳动Doubao系列
- **Gemini** - Google Gemini模型
- **自定义API** - 支持OpenAI兼容的第三方API

### 🎯 智能交互
- 🗣️ **自然语言查询** - 用中文或英文描述需求
- 🖥️ **多种界面模式** - 终端交互 / Web界面
- ⚡ **实时响应** - 流式输出，即时反馈
- 🔒 **安全确认** - 危险操作前自动询问确认

### 🔧 扩展功能
- 🛠️ **自定义工具** - 集成您的专用脚本和命令
- 🔌 **MCP协议支持** - 连接外部工具和服务
- 🌐 **HTTP API** - RESTful接口，支持集成
- 📊 **Web仪表板** - 直观的图形化界面

## 🚀 快速开始

### 📦 安装

#### 方法1: 从源码编译

```bash
# 克隆项目
git clone https://github.com/st-lzh/kubelet-wuhrai.git
cd kubelet-wuhrai

# 编译安装
go build -o kubelet-wuhrai ./cmd/
sudo mv kubelet-wuhrai /usr/local/bin/

# 或使用提供的安装脚本
./install.sh
```

#### 方法2: CentOS 7.9 专用版本

对于CentOS 7.9系统，可以使用专门优化的构建脚本：

```bash
# 构建 CentOS 7.9 版本
./build-centos7.sh

# 或快速构建
./quick-build-centos7.sh
```

详细部署指南请参考: [CENTOS7_DEPLOYMENT.md](CENTOS7_DEPLOYMENT.md)

### 🔑 配置API密钥

安装完成后，需要配置AI模型的API密钥：

```bash
# 编辑配置文件
vi ~/.config/kubelet-wuhrai/config.yaml
```

配置文件示例：
```yaml
# 选择一个AI提供商并取消注释
deepseek_api_key: "your-deepseek-api-key"
# openai_api_key: "your-openai-api-key"
# qwen_api_key: "your-qwen-api-key"

# 其他设置
quiet: false
skip_permissions: false
enable_tool_use_shim: false
```

或使用环境变量：
```bash
# DeepSeek (推荐)
export DEEPSEEK_API_KEY="your-deepseek-api-key"

# OpenAI
export OPENAI_API_KEY="your-openai-api-key"
```

### 🎯 开始使用

```bash
# 检查版本
kubelet-wuhrai version

# 基础查询
kubelet-wuhrai "获取所有pod"

# 集群状态检查
kubelet-wuhrai "检查集群健康状态"

# 应用部署
kubelet-wuhrai "部署一个nginx应用"

# 启动Web界面
kubelet-wuhrai --user-interface html
```

## 💡 使用示例

### 🔍 集群管理
```bash
# 查看集群状态
kubelet-wuhrai "显示集群中所有节点的状态"

# 检查资源使用
kubelet-wuhrai "哪些pod使用的内存最多？"

# 故障排查
kubelet-wuhrai "找出所有失败的pod并显示错误信息"
```

### 🚀 应用部署
```bash
# 部署应用
kubelet-wuhrai "创建一个nginx deployment，3个副本"

# 扩缩容
kubelet-wuhrai "将nginx应用扩展到5个副本"

# 更新应用
kubelet-wuhrai "更新nginx镜像到最新版本"
```

## 📦 构建部署

### 🛠️ 本地构建
```bash
# 基本构建
go build -o kubelet-wuhrai ./cmd/

# 使用 Makefile 构建
make build

# CentOS 7.9 专用构建
./build-centos7.sh
```

### 🐳 Docker 部署
```bash
# 构建 Docker 镜像
docker build -t kubelet-wuhrai .

# 运行容器
docker run -d -p 8888:8888 kubelet-wuhrai
```

## 📖 文档

### 📚 核心文档
- [🔧 构建指南](BUILD_GUIDE.md) - 编译和构建
- [🎯 使用指南](USAGE.md) - 基本使用方法
- [🔧 自定义工具](CUSTOM_TOOLS_GUIDE.md) - 自定义工具和MCP工具
- [🐠 CentOS 7.9 部署](CENTOS7_DEPLOYMENT.md) - CentOS 7.9 专用部署指南
- [🔧 故障排除](TROUBLESHOOTING.md) - 常见问题解决

### 📦 配置示例
- [🛠️ 自定义工具配置](examples/custom-tools.yaml)
- [🔌 MCP配置示例](examples/mcp-config.yaml)

## 🤝 贡献

我们欢迎所有形式的贡献！

### 🔧 开发贡献
1. Fork 本项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 📝 文档贡献
- 改进文档和示例
- 翻译文档到其他语言
- 报告和修复文档错误

### 🐛 问题报告
- 使用 [Issues](https://github.com/st-lzh/kubelet-wuhrai/issues) 报告bug
- 提供详细的复现步骤
- 包含系统环境信息

## 📄 许可证

本项目采用自定义许可证：

- ✅ **个人使用** - 完全免费
- ✅ **学习研究** - 完全免费  
- ✅ **开源项目** - 完全免费
- ⚠️ **商业使用** - 需要联系作者获得授权

详细信息请查看 [LICENSE](LICENSE) 文件。

**商业使用授权请联系**: lzh094285@gmail.com

## 🙏 致谢

- 感谢 [Google kubectl-ai](https://github.com/GoogleCloudPlatform/kubectl-ai) 项目提供的基础框架
- 感谢所有贡献者和社区成员的支持
- 感谢各大AI模型提供商的技术支持

---

<div align="center">

**⭐ 如果这个项目对您有帮助，请给我们一个星标！**

[🌟 Star](https://github.com/st-lzh/kubelet-wuhrai) • [🐛 Report Bug](https://github.com/st-lzh/kubelet-wuhrai/issues) • [💡 Request Feature](https://github.com/st-lzh/kubelet-wuhrai/issues)

</div>
