# kubelet-wuhrai 一键安装指南

## 🚀 一键安装命令

### Linux/macOS 一键安装 (推荐)

```bash
curl -fsSL https://raw.githubusercontent.com/st-lzh/kubelet-wuhrai/main/install.sh | bash
```

这个命令会：
- ✅ 自动检测系统架构 (amd64/arm64/arm/386)
- ✅ 自动检测操作系统 (Linux/macOS)
- ✅ 优先下载预编译二进制文件
- ✅ 如果下载失败，自动从源码编译
- ✅ 智能选择安装目录 (root用户安装到系统目录，普通用户安装到用户目录)
- ✅ 自动配置环境变量
- ✅ 创建示例配置文件
- ✅ 验证安装结果

### 备用安装方法 (源码编译)

如果主安装脚本遇到问题，可以使用专门的源码编译脚本：

```bash
curl -fsSL https://raw.githubusercontent.com/st-lzh/kubelet-wuhrai/main/install-simple.sh | bash
```

这个脚本专门用于从源码编译安装，需要预先安装Go环境。

## 📦 安装位置

### Root用户 (使用sudo)
```bash
sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/st-lzh/kubelet-wuhrai/main/install.sh)"
```
- **二进制文件**: `/usr/local/bin/kubelet-wuhrai`
- **配置文件**: `/etc/kubelet-wuhrai/config.yaml`
- **环境变量**: 添加到 `/etc/profile`

### 普通用户
```bash
curl -fsSL https://raw.githubusercontent.com/st-lzh/kubelet-wuhrai/main/install.sh | bash
```
- **二进制文件**: `~/.local/bin/kubelet-wuhrai`
- **配置文件**: `~/.config/kubelet-wuhrai/config.yaml`
- **环境变量**: 添加到 `~/.bashrc` 或 `~/.zshrc`

## ⚙️ 安装后配置

### 1. 重新加载环境变量
```bash
# 对于普通用户
source ~/.bashrc
# 或者
source ~/.zshrc

# 对于root用户
source /etc/profile
```

### 2. 配置API密钥
编辑配置文件：
```bash
# 普通用户
vi ~/.config/kubelet-wuhrai/config.yaml

# Root用户
sudo vi /etc/kubelet-wuhrai/config.yaml
```

配置示例：
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

### 3. 验证安装
```bash
# 检查版本
kubelet-wuhrai version

# 查看帮助
kubelet-wuhrai --help

# 测试功能 (需要先配置API密钥)
kubelet-wuhrai "获取所有pod"
```

## 🌐 支持的系统

| 操作系统 | 架构 | 支持状态 |
|---------|------|---------|
| Linux | amd64 (x86_64) | ✅ |
| Linux | arm64 (aarch64) | ✅ |
| Linux | arm (armv7l) | ✅ |
| Linux | 386 (i386/i686) | ✅ |
| macOS | amd64 (Intel) | ✅ |
| macOS | arm64 (Apple Silicon) | ✅ |

## 🔧 高级选项

### 指定安装目录
如果您想自定义安装目录，可以下载脚本后修改：
```bash
# 下载脚本
curl -fsSL https://raw.githubusercontent.com/st-lzh/kubelet-wuhrai/main/install.sh -o install.sh

# 编辑脚本，修改INSTALL_DIR变量
vi install.sh

# 运行安装
bash install.sh
```

### 离线安装
如果您的服务器无法访问GitHub，可以：
1. 在有网络的机器上下载二进制文件
2. 使用 `deploy-to-server.sh` 脚本部署到目标服务器

## 🛠️ 故障排除

### 1. 网络连接问题
```bash
# 测试GitHub连接
curl -I https://github.com

# 使用代理
export https_proxy=http://your-proxy:port
curl -fsSL https://raw.githubusercontent.com/st-lzh/kubelet-wuhrai/main/install.sh | bash
```

### 2. 权限问题
```bash
# 如果普通用户安装失败，尝试创建目录
mkdir -p ~/.local/bin

# 或者使用sudo安装到系统目录
sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/st-lzh/kubelet-wuhrai/main/install.sh)"
```

### 3. 命令未找到
```bash
# 检查PATH
echo $PATH

# 手动添加到PATH
export PATH="$HOME/.local/bin:$PATH"

# 或者创建软链接
sudo ln -s ~/.local/bin/kubelet-wuhrai /usr/local/bin/kubelet-wuhrai
```

### 4. 版本不匹配
```bash
# 检查可用版本
curl -s https://api.github.com/repos/st-lzh/kubelet-wuhrai/releases/latest

# 手动下载特定版本
curl -fsSL https://github.com/st-lzh/kubelet-wuhrai/releases/download/v1.0.0/kubelet-wuhrai-linux-amd64 -o kubelet-wuhrai
chmod +x kubelet-wuhrai
sudo mv kubelet-wuhrai /usr/local/bin/
```

## 🗑️ 卸载

### 完全卸载
```bash
# 删除二进制文件
sudo rm -f /usr/local/bin/kubelet-wuhrai
rm -f ~/.local/bin/kubelet-wuhrai

# 删除配置文件
sudo rm -rf /etc/kubelet-wuhrai
rm -rf ~/.config/kubelet-wuhrai

# 从环境变量中移除 (手动编辑)
vi ~/.bashrc  # 删除kubelet-wuhrai相关的export行
vi /etc/profile  # 如果是系统级安装
```

## 📞 获取帮助

- **项目主页**: https://github.com/st-lzh/kubelet-wuhrai
- **问题反馈**: https://github.com/st-lzh/kubelet-wuhrai/issues
- **使用文档**: [README.md](./README.md)
- **安装脚本**: [install.sh](./install.sh)

## 🎯 快速开始

安装完成后，您可以立即开始使用：

```bash
# 1. 一键安装
curl -fsSL https://raw.githubusercontent.com/st-lzh/kubelet-wuhrai/main/install.sh | bash

# 2. 重新加载环境
source ~/.bashrc

# 3. 配置API密钥
vi ~/.config/kubelet-wuhrai/config.yaml

# 4. 开始使用
kubelet-wuhrai "获取所有pod"
kubelet-wuhrai "查看default命名空间的服务"
kubelet-wuhrai "创建一个nginx deployment"
```

享受使用 kubelet-wuhrai！🎉
