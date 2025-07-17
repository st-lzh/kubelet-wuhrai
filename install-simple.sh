#!/bin/bash

# kubelet-wuhrai 简化安装脚本
# 专门用于从源码编译安装，适用于没有预编译版本的情况
# 使用方法: curl -fsSL https://raw.githubusercontent.com/st-lzh/kubelet-wuhrai/main/install-simple.sh | bash

set -e

# 配置
REPO="st-lzh/kubelet-wuhrai"
BINARY_NAME="kubelet-wuhrai"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

# 日志函数
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_bold() { echo -e "${BOLD}$1${NC}"; }

# 显示欢迎信息
show_banner() {
    echo ""
    log_bold "=================================================="
    log_bold "        kubelet-wuhrai 源码编译安装"
    log_bold "=================================================="
    echo ""
    log_info "项目地址: https://github.com/$REPO"
    echo ""
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    local missing_deps=()
    
    if ! command -v curl >/dev/null 2>&1; then
        missing_deps+=("curl")
    fi
    
    if ! command -v git >/dev/null 2>&1; then
        missing_deps+=("git")
    fi
    
    if ! command -v go >/dev/null 2>&1; then
        missing_deps+=("go")
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "缺少必要的依赖: ${missing_deps[*]}"
        echo ""
        log_info "安装依赖的方法："
        echo "  Ubuntu/Debian: sudo apt update && sudo apt install -y curl git golang-go"
        echo "  CentOS/RHEL:   sudo yum install -y curl git golang"
        echo "  Alpine:        apk add --no-cache curl git go"
        echo "  macOS:         brew install curl git go"
        echo ""
        exit 1
    fi
    
    local go_version=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
    log_success "依赖检查通过"
    log_info "Go版本: $go_version"
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

# 创建目录
create_directories() {
    log_info "创建必要的目录..."
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    log_success "目录创建完成"
}

# 从源码编译安装
compile_and_install() {
    local target_file="$INSTALL_DIR/$BINARY_NAME"
    
    # 创建临时目录
    local temp_dir="/tmp/kubelet-wuhrai-build-$$"
    log_info "创建临时构建目录: $temp_dir"
    mkdir -p "$temp_dir"
    cd "$temp_dir"
    
    # 克隆源码
    log_info "克隆源码..."
    if git clone "https://github.com/$REPO.git" . >/dev/null 2>&1; then
        log_success "源码克隆完成"
    else
        log_error "克隆源码失败"
        log_info "请检查网络连接或GitHub访问"
        cleanup_and_exit 1
    fi
    
    # 编译
    log_info "编译 $BINARY_NAME..."
    if go build -o "$BINARY_NAME" ./cmd/ >/dev/null 2>&1; then
        log_success "编译完成"
    else
        log_error "编译失败"
        log_info "请检查Go环境和依赖"
        cleanup_and_exit 1
    fi
    
    # 安装
    log_info "安装到 $target_file..."
    mv "$BINARY_NAME" "$target_file"
    chmod +x "$target_file"
    
    # 清理
    cd /
    rm -rf "$temp_dir"
    
    log_success "安装完成"
}

# 清理并退出
cleanup_and_exit() {
    local exit_code=${1:-0}
    if [[ -n "$temp_dir" && -d "$temp_dir" ]]; then
        cd /
        rm -rf "$temp_dir"
    fi
    exit $exit_code
}

# 配置环境变量
setup_environment() {
    log_info "配置环境变量..."
    
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

# ===========================================
# LLM API 配置 (选择一个并取消注释)
# ===========================================

# DeepSeek (推荐，性价比高)
# deepseek_api_key: "your-deepseek-api-key"

# OpenAI
# openai_api_key: "your-openai-api-key"

# 通义千问
# qwen_api_key: "your-qwen-api-key"

# 豆包
# doubao_api_key: "your-doubao-api-key"

# ===========================================
# 应用配置
# ===========================================

quiet: false
skip_permissions: false
enable_tool_use_shim: false
EOF
        log_success "配置文件已创建: $config_file"
    else
        log_info "配置文件已存在: $config_file"
    fi
}

# 验证安装
verify_installation() {
    log_info "验证安装..."
    
    local target_file="$INSTALL_DIR/$BINARY_NAME"
    
    if [[ ! -f "$target_file" ]]; then
        log_error "安装验证失败: 文件不存在"
        exit 1
    fi
    
    if [[ ! -x "$target_file" ]]; then
        log_error "安装验证失败: 文件不可执行"
        exit 1
    fi
    
    if "$target_file" version >/dev/null 2>&1; then
        local version_info=$("$target_file" version 2>/dev/null)
        log_success "安装验证成功！"
        log_info "版本信息: $version_info"
    else
        log_warning "二进制文件已安装，但可能需要配置API密钥才能正常使用"
    fi
}

# 显示完成信息
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
    echo "1. 重新加载环境变量: source $PROFILE_FILE"
    echo "2. 配置API密钥: vi $CONFIG_DIR/config.yaml"
    echo "3. 测试安装: $BINARY_NAME version"
    echo "4. 开始使用: $BINARY_NAME \"获取所有pod\""
    echo ""
    
    log_success "感谢使用 kubelet-wuhrai！"
}

# 主函数
main() {
    show_banner
    check_dependencies
    setup_install_dirs
    create_directories
    compile_and_install
    setup_environment
    create_config
    verify_installation
    show_completion
}

# 错误处理
trap 'cleanup_and_exit 1' ERR

# 执行主函数
main "$@"
