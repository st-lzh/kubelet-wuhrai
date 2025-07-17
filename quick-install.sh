#!/bin/bash

# kubelet-wuhrai 快速安装脚本
# 适用于已有二进制文件的情况

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 配置变量
BINARY_NAME="kubelet-wuhrai"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/kubelet-wuhrai"

# 检查权限
if [[ $EUID -ne 0 ]]; then
    log_warn "建议使用sudo运行此脚本以安装到系统目录"
    INSTALL_DIR="$HOME/.local/bin"
    CONFIG_DIR="$HOME/.config/kubelet-wuhrai"
fi

# 检查二进制文件
if [[ ! -f "./$BINARY_NAME" ]]; then
    log_error "未找到 $BINARY_NAME 二进制文件"
    log_info "请先编译项目: go build -o $BINARY_NAME ./cmd/"
    exit 1
fi

# 创建目录
log_info "创建安装目录..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"

# 安装二进制文件
log_info "安装 $BINARY_NAME 到 $INSTALL_DIR..."
cp "./$BINARY_NAME" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# 设置环境变量
SHELL_RC=""
if [[ -n "$ZSH_VERSION" ]]; then
    SHELL_RC="$HOME/.zshrc"
elif [[ -n "$BASH_VERSION" ]]; then
    SHELL_RC="$HOME/.bashrc"
else
    SHELL_RC="$HOME/.profile"
fi

# 检查PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    log_info "添加 $INSTALL_DIR 到 PATH..."
    echo "" >> "$SHELL_RC"
    echo "# kubelet-wuhrai" >> "$SHELL_RC"
    echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$SHELL_RC"
    log_info "已更新 $SHELL_RC"
fi

# 创建配置文件
CONFIG_FILE="$CONFIG_DIR/config.yaml"
if [[ ! -f "$CONFIG_FILE" ]]; then
    log_info "创建配置文件..."
    cat > "$CONFIG_FILE" << 'EOF'
# kubelet-wuhrai 配置文件

# DeepSeek (推荐)
# deepseek_api_key: "your-deepseek-api-key"

# OpenAI
# openai_api_key: "your-openai-api-key"

# 通义千问
# qwen_api_key: "your-qwen-api-key"

# 其他设置
quiet: false
skip_permissions: false
EOF
fi

log_info "安装完成！"
echo ""
echo "下一步："
echo "1. 重新加载环境: source $SHELL_RC"
echo "2. 配置API密钥: $CONFIG_FILE"
echo "3. 测试命令: kubelet-wuhrai version"
echo ""
