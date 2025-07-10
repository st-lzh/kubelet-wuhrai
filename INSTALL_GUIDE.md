# kubelet-wuhrai 安装指南

本指南提供多种安装方式，从完全自动化到手动安装，满足不同用户的需求。

## 🚀 一键安装（推荐）

### 完整版一键安装

适用于首次安装，会自动检测和安装Go环境：

```bash
# 下载项目
git clone <repository-url>
cd kubelet-wuhrai

# 运行一键安装脚本
./one-click-install.sh
```

**功能特性：**
- ✅ 自动检测操作系统和架构
- ✅ 检查Go环境，询问是否需要安装
- ✅ 自动下载和安装Go 1.24.3
- ✅ 编译所有模块并运行测试
- ✅ 安装到系统PATH
- ✅ 创建配置文件
- ✅ 支持卸载功能

### 快速安装

适用于已有Go环境的用户：

```bash
# 在项目根目录运行
./quick-install.sh
```

**功能特性：**
- ⚡ 快速编译和安装
- ⚡ 自动配置PATH
- ⚡ 创建基础配置文件
- ⚡ 简洁的输出界面

## 📋 安装选项

### 选项1: 完整自动安装

```bash
./one-click-install.sh
```

- 检测Go环境，如果没有会询问是否安装
- 自动下载Go 1.24.3并安装到 `/usr/local/go`
- 编译项目并运行测试
- 安装到系统PATH
- 创建配置文件

### 选项2: 快速安装（需要Go环境）

```bash
./quick-install.sh
```

- 要求已安装Go 1.24+
- 快速编译和安装
- 自动配置PATH

### 选项3: 手动编译安装

```bash
# 编译
./build.sh

# 安装
./install-local.sh
```

### 选项4: 仅编译

```bash
./build.sh
```

## 🔧 安装后配置

### 环境变量

根据使用的LLM提供商配置相应的API密钥：

```bash
# DeepSeek
export DEEPSEEK_API_KEY="your-deepseek-api-key"

# OpenAI
export OPENAI_API_KEY="your-openai-api-key"

# 通义千问
export QWEN_API_KEY="your-qwen-api-key"

# 豆包
export DOUBAO_API_KEY="your-doubao-api-key"

# Gemini
export GEMINI_API_KEY="your-gemini-api-key"
```

将环境变量添加到shell配置文件：

```bash
# Bash用户
echo 'export DEEPSEEK_API_KEY="your-api-key"' >> ~/.bashrc
source ~/.bashrc

# Zsh用户
echo 'export DEEPSEEK_API_KEY="your-api-key"' >> ~/.zshrc
source ~/.zshrc
```

### 配置文件

配置文件位置：`~/.config/kubelet-wuhrai/config.yaml`

```yaml
# LLM配置
llmProvider: "deepseek"  # deepseek, openai, qwen, doubao, gemini
model: "deepseek-chat"   # 具体模型名称

# 基本设置
skipPermissions: false   # 是否跳过危险操作确认
quiet: false            # 静默模式
maxIterations: 20       # 最大迭代次数

# 界面设置
userInterface: "terminal"  # terminal 或 html
uiListenAddress: "localhost:8888"

# 高级设置
enableToolUseShim: false
skipVerifySSL: false
removeWorkDir: false
```

## 🎯 验证安装

```bash
# 检查版本
kubelet-wuhrai version

# 查看帮助
kubelet-wuhrai --help

# 测试运行（需要配置API密钥）
kubelet-wuhrai --quiet "获取所有pod"
```

## 🗑️ 卸载

### 使用脚本卸载

```bash
./one-click-install.sh --uninstall
```

### 手动卸载

```bash
# 删除二进制文件
rm -f ~/go/bin/kubelet-wuhrai
rm -f ~/.local/bin/kubelet-wuhrai
rm -f /usr/local/bin/kubelet-wuhrai

# 删除配置文件（可选）
rm -rf ~/.config/kubelet-wuhrai

# 从shell配置文件中移除PATH设置
# 编辑 ~/.bashrc 或 ~/.zshrc，删除相关行
```

## 🐛 故障排除

### 常见问题

1. **Go版本过低**
   ```bash
   # 检查Go版本
   go version
   
   # 如果版本低于1.24，使用一键安装脚本自动升级
   ./one-click-install.sh
   ```

2. **权限问题**
   ```bash
   # 确保脚本有执行权限
   chmod +x one-click-install.sh quick-install.sh
   
   # 如果安装到/usr/local/bin需要sudo权限
   sudo ./one-click-install.sh
   ```

3. **网络问题**
   ```bash
   # 检查网络连接
   ping golang.org
   
   # 如果在中国大陆，可能需要配置代理
   export GOPROXY=https://goproxy.cn,direct
   ```

4. **PATH问题**
   ```bash
   # 检查PATH
   echo $PATH
   
   # 手动添加到PATH
   export PATH="$HOME/.local/bin:$PATH"
   
   # 永久添加到shell配置
   echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   ```

5. **模块路径问题**
   ```bash
   # 清理模块缓存
   go clean -modcache
   go mod download
   ```

### 获取帮助

如果遇到问题：

1. 检查Go版本：`go version`
2. 检查网络连接：`ping golang.org`
3. 查看详细错误信息
4. 尝试手动编译：`go build ./cmd`
5. 提交Issue并附上错误日志

## 📱 使用示例

### 命令行使用

```bash
# 基本使用
kubelet-wuhrai

# HTML界面
kubelet-wuhrai --user-interface html

# 静默模式
kubelet-wuhrai --quiet "显示所有运行中的pod"

# 指定模型
kubelet-wuhrai --llm-provider openai --model gpt-4

# 跳过权限确认（危险）
kubelet-wuhrai --skip-permissions
```

### API调用使用

```bash
# 启动HTTP API服务
kubelet-wuhrai --user-interface=html --ui-listen-address=0.0.0.0:8888

# 通过curl调用API
curl -X POST http://localhost:8888/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "query": "获取所有运行中的pod",
    "session_id": "my_session"
  }'
```

详细的API调用示例请参考：[API_USAGE_EXAMPLES.md](API_USAGE_EXAMPLES.md)

## 🔄 更新

```bash
# 拉取最新代码
git pull

# 重新安装
./quick-install.sh
```

---

**注意：** 首次使用前请确保已配置相应的LLM API密钥环境变量。
