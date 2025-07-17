# kubelet-wuhrai Linux 安装脚本使用指南

本项目提供了三个不同的安装脚本，适用于不同的部署场景。

## 📦 脚本说明

### 1. `install-linux.sh` - 完整安装脚本
**适用场景**: 全新安装，支持多种安装方式
**功能特点**:
- 自动检测系统架构 (amd64/arm64/arm)
- 支持root和普通用户安装
- 优先使用本地二进制文件
- 支持本地编译
- 可从GitHub下载预编译版本
- 自动配置环境变量
- 创建示例配置文件

### 2. `quick-install.sh` - 快速安装脚本
**适用场景**: 已有编译好的二进制文件，快速本地安装
**功能特点**:
- 轻量级，执行速度快
- 自动检测shell类型 (bash/zsh)
- 智能选择安装目录
- 简化的配置文件

### 3. `deploy-to-server.sh` - 远程部署脚本
**适用场景**: 部署到远程Linux服务器
**功能特点**:
- SSH远程部署
- 自动上传二进制文件
- 远程执行安装
- 支持自定义安装路径

## 🚀 使用方法

### 方法一：完整安装 (推荐)

```bash
# 下载并运行完整安装脚本
./install-linux.sh
```

**安装位置**:
- Root用户: `/usr/local/bin` (系统级)
- 普通用户: `~/.local/bin` (用户级)

### 方法二：快速安装

```bash
# 先编译项目
go build -o kubelet-wuhrai ./cmd/

# 运行快速安装
./quick-install.sh
```

### 方法三：远程服务器部署

```bash
# 部署到远程服务器 (默认路径)
./deploy-to-server.sh user@server

# 部署到指定路径
./deploy-to-server.sh user@192.168.1.100 /opt/kubelet-wuhrai

# 示例
./deploy-to-server.sh root@myserver.com
./deploy-to-server.sh ubuntu@192.168.1.100 /usr/local/bin
```

## ⚙️ 配置说明

### 配置文件位置
- **系统级**: `/etc/kubelet-wuhrai/config.yaml`
- **用户级**: `~/.config/kubelet-wuhrai/config.yaml`

### 配置示例

```yaml
# 选择一个AI提供商
deepseek_api_key: "your-deepseek-api-key"
# openai_api_key: "your-openai-api-key"
# qwen_api_key: "your-qwen-api-key"

# 其他设置
quiet: false
skip_permissions: false
enable_tool_use_shim: false
```

## 🔧 环境变量设置

脚本会自动将安装目录添加到以下文件：
- **系统级**: `/etc/profile`
- **用户级**: `~/.bashrc` 或 `~/.zshrc`

手动设置环境变量：
```bash
# 添加到PATH
export PATH="/usr/local/bin:$PATH"

# 或者用户级安装
export PATH="$HOME/.local/bin:$PATH"
```

## 📋 安装后验证

```bash
# 重新加载环境变量
source ~/.bashrc  # 或 source /etc/profile

# 检查版本
kubelet-wuhrai version

# 查看帮助
kubelet-wuhrai --help

# 测试功能 (需要配置API密钥)
kubelet-wuhrai "获取所有pod"
```

## 🛠️ 故障排除

### 1. 命令未找到
```bash
# 检查PATH
echo $PATH

# 手动重新加载
source ~/.bashrc
# 或
source /etc/profile
```

### 2. 权限问题
```bash
# 检查文件权限
ls -la /usr/local/bin/kubelet-wuhrai

# 修复权限
sudo chmod +x /usr/local/bin/kubelet-wuhrai
```

### 3. SSH部署失败
```bash
# 检查SSH连接
ssh user@server "echo 'test'"

# 检查SSH密钥
ssh-add -l
```

### 4. 配置文件问题
```bash
# 检查配置文件
cat ~/.config/kubelet-wuhrai/config.yaml

# 重新创建配置文件
rm ~/.config/kubelet-wuhrai/config.yaml
./quick-install.sh
```

## 📝 卸载方法

```bash
# 删除二进制文件
sudo rm -f /usr/local/bin/kubelet-wuhrai
# 或用户级
rm -f ~/.local/bin/kubelet-wuhrai

# 删除配置文件
sudo rm -rf /etc/kubelet-wuhrai
# 或用户级
rm -rf ~/.config/kubelet-wuhrai

# 从环境变量中移除 (手动编辑)
vi ~/.bashrc  # 删除相关的export行
```

## 🔗 相关链接

- [项目主页](https://github.com/st-lzh/kubelet-wuhrai)
- [使用文档](./README.md)
- [故障排除](./TROUBLESHOOTING.md)
