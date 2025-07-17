#!/bin/bash

# kubelet-wuhraia 安装脚本
# 使用方法: curl -fsSL https://raw.githubusercontent.com/st-lzh/kubelet-wuhraia/main/install.sh | bash

set -e

BINARY_NAME="kubelet-wuhraia"
INSTALL_DIR="/usr/local/bin"
REPO="st-lzh/kubelet-wuhraia"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# 检测系统
detect_system() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv7l) ARCH="armv7" ;;
        *) 
            log_error "不支持的架构: $ARCH"
            exit 1 
            ;;
    esac

    log_info "检测到系统: $OS/$ARCH"
}

# 获取最新版本
get_latest_version() {
    log_info "获取最新版本信息..."
    VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [[ -z "$VERSION" ]]; then
        log_warn "无法获取最新版本，将下载主分支版本"
        VERSION="latest"
    else
        log_info "最新版本: $VERSION"
    fi
}

# 下载二进制文件
download_binary() {
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$BINARY_NAME-$OS-$ARCH"
    
    log_info "正在下载: $DOWNLOAD_URL"
    
    # 检查是否需要sudo
    if [[ -w "$INSTALL_DIR" ]]; then
        SUDO=""
    else
        SUDO="sudo"
        log_info "需要sudo权限安装到 $INSTALL_DIR"
    fi
    
    # 下载并安装
    if $SUDO curl -fsSL "$DOWNLOAD_URL" -o "$INSTALL_DIR/$BINARY_NAME"; then
        $SUDO chmod +x "$INSTALL_DIR/$BINARY_NAME"
        log_info "下载完成"
    else
        log_error "下载失败，请检查网络连接或版本是否存在"
        exit 1
    fi
}

# 验证安装
verify_installation() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        log_info "✓ $BINARY_NAME 已成功安装"
        log_info "版本: $($BINARY_NAME --version 2>/dev/null || echo '未知')"
    else
        log_warn "$BINARY_NAME 未在PATH中找到"
        log_info "二进制文件位置: $INSTALL_DIR/$BINARY_NAME"
        log_info "请确保 $INSTALL_DIR 在你的PATH中"
    fi
}

# 主函数
main() {
    log_info "开始安装 $BINARY_NAME..."
    
    detect_system
    get_latest_version
    download_binary
    verify_installation
    
    log_info "安装完成! 🎉"
    log_info "使用方法: $BINARY_NAME --help"
}

# 运行主函数
main "$@"
