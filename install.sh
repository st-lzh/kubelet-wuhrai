#!/bin/bash

# kubelet-wuhrai 一键安装脚本
# 使用方法: curl -fsSL https://raw.githubusercontent.com/st-lzh/kubelet-wuhrai/main/install.sh | bash

set -e

# 版本和仓库信息
REPO="st-lzh/kubelet-wuhrai"
BINARY_NAME="kubelet-wuhrai"
GITHUB_API="https://api.github.com/repos/$REPO"
GITHUB_RELEASES="https://github.com/$REPO/releases"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_bold() {
    echo -e "${BOLD}$1${NC}"
}

# 显示欢迎信息
show_banner() {
    echo ""
    log_bold "=================================================="
    log_bold "        kubelet-wuhrai 一键安装脚本"
    log_bold "=================================================="
    echo ""
    log_info "项目地址: https://github.com/$REPO"
    echo ""
}

# 检测系统架构
detect_arch() {
    local arch=$(uname -m)
    case $arch in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="arm"
            ;;
        i386|i686)
            ARCH="386"
            ;;
        *)
            log_error "不支持的架构: $arch"
            log_info "支持的架构: x86_64, aarch64, armv7l, i386"
            exit 1
            ;;
    esac
    log_info "检测到架构: $ARCH"
}

# 检测操作系统
detect_os() {
    case "$OSTYPE" in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        *)
            log_error "不支持的操作系统: $OSTYPE"
            log_info "支持的系统: Linux, macOS"
            exit 1
            ;;
    esac
    log_info "检测到操作系统: $OS"
}

# 检查依赖
check_dependencies() {
    local missing_deps=()

    if ! command -v curl >/dev/null 2>&1; then
        missing_deps+=("curl")
    fi

    if ! command -v tar >/dev/null 2>&1; then
        missing_deps+=("tar")
    fi

    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "缺少必要的依赖: ${missing_deps[*]}"
        log_info "请先安装这些依赖，然后重新运行脚本"
        exit 1
    fi
}

# 确定安装目录
setup_install_dirs() {
    if [[ $EUID -eq 0 ]]; then
        INSTALL_DIR="/usr/local/bin"
        CONFIG_DIR="/etc/kubelet-wuhrai"
        PROFILE_FILE="/etc/profile"
        IS_ROOT=true
        log_info "检测到root用户，将安装到系统目录"
    else
        INSTALL_DIR="$HOME/.local/bin"
        CONFIG_DIR="$HOME/.config/kubelet-wuhrai"
        # 检测shell类型
        if [[ -n "$ZSH_VERSION" ]]; then
            PROFILE_FILE="$HOME/.zshrc"
        elif [[ -n "$BASH_VERSION" ]]; then
            PROFILE_FILE="$HOME/.bashrc"
        else
            PROFILE_FILE="$HOME/.profile"
        fi
        IS_ROOT=false
        log_info "检测到普通用户，将安装到用户目录"
    fi

    log_info "安装目录: $INSTALL_DIR"
    log_info "配置目录: $CONFIG_DIR"
}

# 获取最新版本
get_latest_version() {
    log_info "获取最新版本信息..."

    if command -v jq >/dev/null 2>&1; then
        # 使用jq解析JSON
        LATEST_VERSION=$(curl -s "$GITHUB_API/releases/latest" | jq -r '.tag_name')
    else
        # 不使用jq的备用方案
        LATEST_VERSION=$(curl -s "$GITHUB_API/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    fi

    if [[ -z "$LATEST_VERSION" || "$LATEST_VERSION" == "null" ]]; then
        log_warning "无法获取最新版本，使用latest"
        LATEST_VERSION="latest"
        DOWNLOAD_URL="$GITHUB_RELEASES/latest/download/${BINARY_NAME}-${OS}-${ARCH}"
    else
        log_info "最新版本: $LATEST_VERSION"
        DOWNLOAD_URL="$GITHUB_RELEASES/download/${LATEST_VERSION}/${BINARY_NAME}-${OS}-${ARCH}"
    fi
}

# 创建安装目录
create_directories() {
    log_info "创建必要的目录..."

    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"

    log_success "目录创建完成"
}

# 下载并安装二进制文件
download_and_install() {
    local temp_file="/tmp/${BINARY_NAME}-${OS}-${ARCH}"
    local target_file="$INSTALL_DIR/$BINARY_NAME"

    log_info "下载 $BINARY_NAME..."
    log_info "下载地址: $DOWNLOAD_URL"

    # 下载文件
    if curl -fL "$DOWNLOAD_URL" -o "$temp_file"; then
        log_success "下载完成"
    else
        log_error "下载失败"
        log_info "请检查网络连接或手动下载安装"
        log_info "下载地址: $DOWNLOAD_URL"
        exit 1
    fi

    # 安装文件
    log_info "安装到 $target_file..."

    if [[ "$IS_ROOT" == "true" ]]; then
        mv "$temp_file" "$target_file"
    else
        mv "$temp_file" "$target_file"
    fi

    chmod +x "$target_file"
    log_success "安装完成"
}

# 配置环境变量
setup_environment() {
    log_info "配置环境变量..."

    # 检查PATH中是否已包含安装目录
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo "" >> "$PROFILE_FILE"
        echo "# kubelet-wuhrai" >> "$PROFILE_FILE"
        echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$PROFILE_FILE"
        log_success "已添加 $INSTALL_DIR 到 PATH"
        log_info "配置文件: $PROFILE_FILE"
    else
        log_info "$INSTALL_DIR 已在 PATH 中"
    fi
}

# 创建配置文件
create_config() {
    local config_file="$CONFIG_DIR/config.yaml"

    if [[ ! -f "$config_file" ]]; then
        log_info "创建配置文件..."
        cat > "$config_file" << 'EOF'
# kubelet-wuhrai 配置文件
# 更多配置选项请参考: https://github.com/st-lzh/kubelet-wuhrai

# ===========================================
# LLM API 配置 (选择一个并取消注释)
# ===========================================

# DeepSeek (推荐，性价比高)
# deepseek_api_key: "your-deepseek-api-key"

# OpenAI
# openai_api_key: "your-openai-api-key"
# openai_base_url: "https://api.openai.com/v1"

# 通义千问
# qwen_api_key: "your-qwen-api-key"

# 豆包
# doubao_api_key: "your-doubao-api-key"

# ===========================================
# 应用配置
# ===========================================

# 静默模式 (非交互式)
quiet: false

# 跳过权限检查
skip_permissions: false

# 启用工具使用垫片
enable_tool_use_shim: false

# Kubernetes 配置文件路径 (可选)
# kubeconfig: "~/.kube/config"
EOF
        log_success "配置文件已创建: $config_file"
    else
        log_info "配置文件已存在: $config_file"
    fi
}

# 验证安装
verify_installation() {
    log_info "验证安装..."

    # 检查文件是否存在
    if [[ ! -f "$INSTALL_DIR/$BINARY_NAME" ]]; then
        log_error "安装验证失败: 文件不存在"
        exit 1
    fi

    # 检查文件是否可执行
    if [[ ! -x "$INSTALL_DIR/$BINARY_NAME" ]]; then
        log_error "安装验证失败: 文件不可执行"
        exit 1
    fi

    # 尝试运行version命令
    if "$INSTALL_DIR/$BINARY_NAME" version >/dev/null 2>&1; then
        local version_info=$("$INSTALL_DIR/$BINARY_NAME" version 2>/dev/null)
        log_success "安装验证成功！"
        log_info "版本信息: $version_info"
    else
        log_warning "二进制文件已安装，但可能需要配置API密钥才能正常使用"
    fi
}

# 显示安装完成信息
show_completion() {
    echo ""
    log_bold "=================================================="
    log_bold "           🎉 安装完成！"
    log_bold "=================================================="
    echo ""

    log_info "安装位置: $INSTALL_DIR/$BINARY_NAME"
    log_info "配置文件: $CONFIG_DIR/config.yaml"
    echo ""

    log_bold "下一步操作："
    echo ""
    echo "1. 重新加载环境变量："
    echo "   source $PROFILE_FILE"
    echo ""
    echo "2. 配置API密钥："
    echo "   编辑文件: $CONFIG_DIR/config.yaml"
    echo "   取消注释并填入您的API密钥"
    echo ""
    echo "3. 验证安装："
    echo "   $BINARY_NAME version"
    echo "   $BINARY_NAME --help"
    echo ""
    echo "4. 开始使用："
    echo "   $BINARY_NAME \"获取所有pod\""
    echo "   $BINARY_NAME \"查看default命名空间的服务\""
    echo ""

    log_bold "获取帮助："
    echo "   项目文档: https://github.com/$REPO"
    echo "   问题反馈: https://github.com/$REPO/issues"
    echo ""

    log_success "感谢使用 kubelet-wuhrai！"
}

# 主函数
main() {
    show_banner
    check_dependencies
    detect_arch
    detect_os
    setup_install_dirs
    get_latest_version
    create_directories
    download_and_install
    setup_environment
    create_config
    verify_installation
    show_completion
}

# 错误处理
trap 'log_error "安装过程中发生错误，请检查上面的错误信息"' ERR

# 执行主函数
main "$@"
