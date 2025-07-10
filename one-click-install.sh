#!/usr/bin/env bash
# kubelet-wuhrai 一键安装脚本
# 自动检测Go环境、安装依赖、编译程序并添加到全局变量
# Copyright 2025 kubelet-wuhrai

set -o errexit
set -o nounset
set -o pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

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

log_step() {
    echo -e "${CYAN}${BOLD}[STEP]${NC} $1"
}

# 询问用户确认
ask_user() {
    local question="$1"
    local default="${2:-n}"
    local response
    
    if [[ "$default" == "y" ]]; then
        echo -e "${YELLOW}${question} [Y/n]:${NC} "
    else
        echo -e "${YELLOW}${question} [y/N]:${NC} "
    fi
    
    read -r response
    response=${response:-$default}
    
    if [[ "$response" =~ ^[Yy]([Ee][Ss])?$ ]]; then
        return 0
    else
        return 1
    fi
}

# 检测操作系统
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if command -v apt-get &> /dev/null; then
            OS="ubuntu"
        elif command -v yum &> /dev/null; then
            OS="centos"
        elif command -v dnf &> /dev/null; then
            OS="fedora"
        elif command -v pacman &> /dev/null; then
            OS="arch"
        else
            OS="linux"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        OS="windows"
    else
        OS="unknown"
    fi
    
    # 检测架构
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv7l) ARCH="armv6l" ;;
        *) log_warning "未知架构: $ARCH，将使用 amd64" && ARCH="amd64" ;;
    esac
    
    log_info "检测到系统: $OS ($ARCH)"
}

# 检查Go环境
check_go() {
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        GO_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
        GO_MINOR=$(echo $GO_VERSION | cut -d. -f2)
        
        log_info "检测到Go版本: $GO_VERSION"
        
        # 检查版本是否满足要求 (>= 1.24)
        if [[ $GO_MAJOR -gt 1 ]] || [[ $GO_MAJOR -eq 1 && $GO_MINOR -ge 24 ]]; then
            log_success "Go版本满足要求 (>= 1.24.0)"
            return 0
        else
            log_warning "Go版本过低，需要 >= 1.24.0"
            return 1
        fi
    else
        log_warning "未检测到Go环境"
        return 1
    fi
}

# 安装Go环境
install_go() {
    local go_version="1.24.3"
    local go_url="https://golang.org/dl/go${go_version}.linux-${ARCH}.tar.gz"
    local go_file="go${go_version}.linux-${ARCH}.tar.gz"

    # macOS特殊处理
    if [[ "$OS" == "macos" ]]; then
        go_url="https://golang.org/dl/go${go_version}.darwin-${ARCH}.tar.gz"
        go_file="go${go_version}.darwin-${ARCH}.tar.gz"
    fi

    # Windows特殊处理
    if [[ "$OS" == "windows" ]]; then
        log_error "Windows系统请手动安装Go: https://golang.org/dl/"
        log_info "或者使用WSL2运行此脚本"
        exit 1
    fi
    
    log_step "开始安装Go ${go_version}..."
    
    # 创建临时目录
    local temp_dir=$(mktemp -d)
    cd "$temp_dir"
    
    # 下载Go
    log_info "下载Go安装包..."
    if command -v wget &> /dev/null; then
        wget -q --show-progress "$go_url" -O "$go_file"
    elif command -v curl &> /dev/null; then
        curl -L --progress-bar "$go_url" -o "$go_file"
    else
        log_error "需要wget或curl来下载Go安装包"
        exit 1
    fi
    
    # 安装Go
    log_info "安装Go到 /usr/local/go..."
    
    # 删除旧的Go安装（如果存在）
    if [[ -d "/usr/local/go" ]]; then
        if ask_user "检测到已存在的Go安装，是否删除并重新安装？" "y"; then
            sudo rm -rf /usr/local/go
        else
            log_info "跳过Go安装"
            cd - > /dev/null
            rm -rf "$temp_dir"
            return 0
        fi
    fi
    
    # 解压安装
    sudo tar -C /usr/local -xzf "$go_file"
    
    # 设置环境变量
    local shell_rc=""
    if [[ -n "${BASH_VERSION:-}" ]]; then
        shell_rc="$HOME/.bashrc"
    elif [[ -n "${ZSH_VERSION:-}" ]]; then
        shell_rc="$HOME/.zshrc"
    else
        shell_rc="$HOME/.profile"
    fi
    
    # 检查是否已经添加了Go路径
    if ! grep -q "/usr/local/go/bin" "$shell_rc" 2>/dev/null; then
        log_info "添加Go到PATH..."
        echo "" >> "$shell_rc"
        echo "# Go environment" >> "$shell_rc"
        echo 'export PATH=$PATH:/usr/local/go/bin' >> "$shell_rc"
        echo 'export GOPATH=$HOME/go' >> "$shell_rc"
        echo 'export PATH=$PATH:$GOPATH/bin' >> "$shell_rc"
    fi
    
    # 设置当前会话的环境变量
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=$HOME/go
    export PATH=$PATH:$GOPATH/bin
    
    # 创建GOPATH目录
    mkdir -p "$GOPATH/bin"
    
    # 清理临时文件
    cd - > /dev/null
    rm -rf "$temp_dir"
    
    log_success "Go安装完成！"
    log_info "请运行以下命令重新加载环境变量："
    log_info "source $shell_rc"
    
    # 验证安装
    if /usr/local/go/bin/go version &> /dev/null; then
        local new_version=$(/usr/local/go/bin/go version | awk '{print $3}' | sed 's/go//')
        log_success "Go ${new_version} 安装成功！"
    else
        log_error "Go安装验证失败"
        exit 1
    fi
}

# 编译项目
build_project() {
    log_step "开始编译kubelet-wuhrai项目..."
    
    local repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    cd "$repo_root"
    
    # 检查项目文件
    if [[ ! -f "go.mod" ]] || [[ ! -d "cmd" ]]; then
        log_error "当前目录不是kubelet-wuhrai项目根目录"
        exit 1
    fi
    
    # 下载依赖
    log_info "下载Go模块依赖..."
    go mod download
    
    # 修复子模块的go.mod文件（如果需要）
    if grep -q "GoogleCloudPlatform" k8s-bench/go.mod 2>/dev/null; then
        log_info "修复k8s-bench模块路径..."
        sed -i 's|github.com/GoogleCloudPlatform/kubectl-ai|github.com/st-lzh/kubelet-wuhrai|g' k8s-bench/go.mod
    fi
    
    if grep -q "GoogleCloudPlatform" kubectl-utils/go.mod 2>/dev/null; then
        log_info "修复kubectl-utils模块路径..."
        sed -i 's|github.com/GoogleCloudPlatform/kubectl-ai|github.com/st-lzh/kubelet-wuhrai|g' kubectl-utils/go.mod
    fi
    
    # 整理依赖
    log_info "整理模块依赖..."
    go mod tidy
    
    # 编译子模块
    for module in gollm k8s-bench kubectl-utils; do
        if [[ -d "$module" ]]; then
            log_info "编译模块: $module"
            cd "$module"
            go mod tidy
            go build ./...
            cd "$repo_root"
        fi
    done
    
    # 编译主程序
    log_info "编译主程序..."
    mkdir -p bin
    
    # 获取版本信息
    local version="dev"
    local commit="none"
    local date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    if command -v git &> /dev/null && git rev-parse --git-dir > /dev/null 2>&1; then
        commit=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
    fi
    
    go build -ldflags "-X main.version=${version} -X main.commit=${commit} -X main.date=${date}" -o bin/kubelet-wuhrai ./cmd
    
    if [[ -f "bin/kubelet-wuhrai" ]]; then
        local file_size=$(du -h bin/kubelet-wuhrai | cut -f1)
        log_success "编译完成！二进制文件大小: ${file_size}"
    else
        log_error "编译失败"
        exit 1
    fi
    
    # 运行测试
    log_info "运行测试..."
    if go test ./... -v > /dev/null 2>&1; then
        log_success "所有测试通过"
    else
        log_warning "部分测试失败，但不影响安装"
    fi
}

# 安装到系统
install_to_system() {
    log_step "安装kubelet-wuhrai到系统..."

    local binary_path="bin/kubelet-wuhrai"
    if [[ ! -f "$binary_path" ]]; then
        log_error "二进制文件不存在: $binary_path"
        exit 1
    fi

    # 确定安装目录 - 优先使用系统全局目录
    local install_dir=""
    local use_global=false

    # 首先尝试系统全局目录
    if [[ -w "/usr/local/bin" ]]; then
        install_dir="/usr/local/bin"
        use_global=true
        log_info "使用系统全局目录: $install_dir"
    elif sudo -n true 2>/dev/null && [[ -d "/usr/local/bin" ]]; then
        install_dir="/usr/local/bin"
        use_global=true
        log_info "使用sudo安装到系统全局目录: $install_dir"
    elif [[ -n "${GOPATH:-}" ]] && [[ -d "${GOPATH}/bin" ]]; then
        install_dir="${GOPATH}/bin"
    elif [[ -d "${HOME}/go/bin" ]]; then
        install_dir="${HOME}/go/bin"
    elif [[ -d "${HOME}/.local/bin" ]]; then
        install_dir="${HOME}/.local/bin"
    else
        # 创建用户本地bin目录
        install_dir="${HOME}/.local/bin"
        mkdir -p "$install_dir"
    fi

    log_info "安装到: $install_dir"

    # 复制二进制文件
    if [[ "$use_global" == "true" ]] && [[ ! -w "$install_dir" ]]; then
        if [[ -f "$install_dir/kubelet-wuhrai" ]]; then
            sudo rm -f "$install_dir/kubelet-wuhrai"
        fi
        sudo cp "$binary_path" "$install_dir/"
        sudo chmod +x "$install_dir/kubelet-wuhrai"
        log_success "已使用sudo安装到系统全局目录"
    else
        if [[ -f "$install_dir/kubelet-wuhrai" ]]; then
            rm -f "$install_dir/kubelet-wuhrai"
        fi
        cp "$binary_path" "$install_dir/"
        chmod +x "$install_dir/kubelet-wuhrai"
    fi

    # 创建全局符号链接（如果不是已经在全局目录）
    if [[ "$install_dir" != "/usr/local/bin" ]] && [[ -d "/usr/local/bin" ]]; then
        if sudo -n true 2>/dev/null; then
            log_info "创建全局符号链接到 /usr/local/bin"
            sudo ln -sf "$install_dir/kubelet-wuhrai" "/usr/local/bin/kubelet-wuhrai"
            log_success "全局符号链接创建成功"
        elif ask_user "是否创建全局符号链接到 /usr/local/bin (需要sudo权限)？" "y"; then
            sudo ln -sf "$install_dir/kubelet-wuhrai" "/usr/local/bin/kubelet-wuhrai"
            log_success "全局符号链接创建成功"
        fi
    fi

    # 配置shell环境变量
    setup_shell_environment "$install_dir"

    # 创建配置目录和文件
    setup_configuration

    log_success "安装完成！"
}

# 配置shell环境变量
setup_shell_environment() {
    local install_dir="$1"

    # 如果已经在全局PATH中，不需要额外配置
    if [[ "$install_dir" == "/usr/local/bin" ]] || [[ "$install_dir" == "/usr/bin" ]]; then
        log_info "已安装到系统全局目录，无需配置PATH"
        return 0
    fi

    # 检查并添加到PATH
    local shell_configs=()

    # 检测当前shell
    if [[ -n "${BASH_VERSION:-}" ]]; then
        shell_configs+=("$HOME/.bashrc")
    fi
    if [[ -n "${ZSH_VERSION:-}" ]]; then
        shell_configs+=("$HOME/.zshrc")
    fi

    # 添加通用配置文件
    shell_configs+=("$HOME/.profile")

    # 为每个配置文件添加PATH
    for shell_rc in "${shell_configs[@]}"; do
        if [[ -f "$shell_rc" ]] || [[ "$shell_rc" == "$HOME/.profile" ]]; then
            if ! grep -q "kubelet-wuhrai" "$shell_rc" 2>/dev/null; then
                log_info "添加PATH到: $shell_rc"
                echo "" >> "$shell_rc"
                echo "# kubelet-wuhrai global command" >> "$shell_rc"
                echo "export PATH=\"${install_dir}:\$PATH\"" >> "$shell_rc"
            else
                log_info "PATH已存在于: $shell_rc"
            fi
        fi
    done

    # 设置当前会话的PATH
    export PATH="${install_dir}:$PATH"

    log_info "PATH配置完成"
}

# 创建配置目录和文件
setup_configuration() {
    local config_dir="${HOME}/.config/kubelet-wuhrai"
    mkdir -p "$config_dir"

    # 创建示例配置文件
    if [[ ! -f "$config_dir/config.yaml.example" ]]; then
        cat > "$config_dir/config.yaml.example" << 'EOF'
# kubelet-wuhrai 配置文件示例
# 复制此文件为 config.yaml 并根据需要修改

# LLM 提供商配置
llmProvider: "deepseek"  # 支持: deepseek, openai, qwen, doubao, gemini 等
model: "deepseek-chat"   # 模型名称

# 基本设置
skipPermissions: false   # 是否跳过权限确认
quiet: false            # 是否以静默模式运行
maxIterations: 20       # 最大迭代次数

# UI 设置
userInterface: "terminal"  # terminal 或 html
uiListenAddress: "localhost:8888"  # HTML UI 监听地址

# 高级设置
enableToolUseShim: false  # 是否启用工具使用垫片
skipVerifySSL: false     # 是否跳过SSL验证
removeWorkDir: false     # 是否删除临时工作目录

# 自定义API配置示例
# 如果使用第三方OpenAI兼容API，请设置环境变量：
# export OPENAI_API_KEY="your-api-key"
# export OPENAI_API_BASE="https://your-api-endpoint.com/v1"
EOF
        log_info "创建示例配置文件: $config_dir/config.yaml.example"
    fi

    # 创建环境变量示例文件
    if [[ ! -f "$config_dir/env.example" ]]; then
        cat > "$config_dir/env.example" << 'EOF'
# kubelet-wuhrai 环境变量配置示例
# 复制需要的环境变量到你的 ~/.bashrc 或 ~/.zshrc

# DeepSeek API (默认)
export DEEPSEEK_API_KEY="your-deepseek-api-key"

# OpenAI API
export OPENAI_API_KEY="your-openai-api-key"

# 自定义OpenAI兼容API
export OPENAI_API_KEY="your-custom-api-key"
export OPENAI_API_BASE="https://your-api-endpoint.com/v1"

# 通义千问
export QWEN_API_KEY="your-qwen-api-key"

# 豆包
export DOUBAO_API_KEY="your-doubao-api-key"

# Gemini
export GEMINI_API_KEY="your-gemini-api-key"

# 使用示例：
# source ~/.bashrc
# kubelet-wuhrai --llm-provider openai --model gpt-4o "获取所有pod"
EOF
        log_info "创建环境变量示例文件: $config_dir/env.example"
    fi
}

# 验证安装
verify_installation() {
    log_step "验证安装..."

    # 检查多个可能的位置
    local found_locations=()

    # 检查全局位置
    if [[ -f "/usr/local/bin/kubelet-wuhrai" ]]; then
        found_locations+=("/usr/local/bin/kubelet-wuhrai")
    fi

    # 检查用户位置
    for dir in "$HOME/go/bin" "$HOME/.local/bin" "${GOPATH:-}/bin"; do
        if [[ -f "$dir/kubelet-wuhrai" ]]; then
            found_locations+=("$dir/kubelet-wuhrai")
        fi
    done

    if [[ ${#found_locations[@]} -gt 0 ]]; then
        log_success "找到kubelet-wuhrai安装位置："
        for location in "${found_locations[@]}"; do
            log_info "  - $location"
        done
    fi

    # 测试命令是否可用
    if command -v kubelet-wuhrai &> /dev/null; then
        local version_output=$(kubelet-wuhrai version 2>/dev/null || echo "版本信息获取失败")
        log_success "✅ kubelet-wuhrai 全局命令可用！"
        echo "$version_output"

        # 测试基本功能
        log_info "测试基本功能..."
        local test_result=$(kubelet-wuhrai --help 2>/dev/null | head -1 || echo "帮助信息获取失败")
        if [[ "$test_result" != "帮助信息获取失败" ]]; then
            log_success "✅ 基本功能测试通过"
        else
            log_warning "⚠️ 基本功能测试失败"
        fi
    else
        log_warning "⚠️ kubelet-wuhrai 命令不在当前PATH中"
        log_info "请运行以下命令重新加载环境变量："

        # 提供具体的重新加载命令
        if [[ -f "$HOME/.bashrc" ]]; then
            log_info "  source ~/.bashrc"
        fi
        if [[ -f "$HOME/.zshrc" ]]; then
            log_info "  source ~/.zshrc"
        fi
        if [[ -f "$HOME/.profile" ]]; then
            log_info "  source ~/.profile"
        fi

        log_info "或者重新打开终端"

        # 提供直接路径使用方式
        if [[ ${#found_locations[@]} -gt 0 ]]; then
            log_info "或者直接使用完整路径："
            log_info "  ${found_locations[0]} --help"
        fi
    fi
}

# 检查网络连接
check_network() {
    log_info "检查网络连接..."
    if command -v ping &> /dev/null; then
        if ping -c 1 8.8.8.8 &> /dev/null || ping -c 1 golang.org &> /dev/null; then
            log_success "网络连接正常"
            return 0
        else
            log_warning "网络连接可能有问题"
            return 1
        fi
    else
        log_warning "无法检查网络连接"
        return 0
    fi
}

# 卸载功能
uninstall() {
    log_step "卸载kubelet-wuhrai..."

    # 查找并删除二进制文件
    local found=false
    for dir in "$HOME/go/bin" "$HOME/.local/bin" "/usr/local/bin" "${GOPATH:-}/bin"; do
        if [[ -f "$dir/kubelet-wuhrai" ]]; then
            log_info "删除: $dir/kubelet-wuhrai"
            rm -f "$dir/kubelet-wuhrai"
            found=true
        fi
    done

    if [[ "$found" == "true" ]]; then
        log_success "kubelet-wuhrai 已卸载"
    else
        log_warning "未找到kubelet-wuhrai安装"
    fi

    # 询问是否删除配置文件
    if [[ -d "$HOME/.config/kubelet-wuhrai" ]]; then
        if ask_user "是否删除配置文件？" "n"; then
            rm -rf "$HOME/.config/kubelet-wuhrai"
            log_info "配置文件已删除"
        fi
    fi
}

# 显示帮助信息
show_help() {
    echo "kubelet-wuhrai 一键安装脚本"
    echo ""
    echo "用法:"
    echo "  $0                安装kubelet-wuhrai (完整安装)"
    echo "  $0 --quick        快速安装 (需要已有Go环境)"
    echo "  $0 --uninstall    卸载kubelet-wuhrai"
    echo "  $0 --help         显示此帮助信息"
    echo ""
    echo "功能:"
    echo "  - 自动检测并安装Go环境 (完整模式)"
    echo "  - 编译kubelet-wuhrai项目"
    echo "  - 安装到系统PATH"
    echo "  - 创建配置文件"
    echo "  - 验证安装结果"
    echo ""
    echo "模式说明:"
    echo "  完整模式: 自动检测并安装Go环境，适合首次安装"
    echo "  快速模式: 跳过Go安装，适合已有Go环境的快速部署"
    echo ""
}

# 快速安装模式 (融合quick-install.sh的功能)
quick_install() {
    log_step "快速安装模式..."

    # 检查Go环境 (必须存在)
    if ! command -v go &> /dev/null; then
        log_error "快速模式需要已安装的Go环境"
        log_info "请使用完整模式: $0 (不带--quick参数)"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "检测到Go版本: $GO_VERSION"

    # 简单版本检查
    if [[ "$GO_VERSION" < "1.24" ]]; then
        log_warning "Go版本可能过低，建议升级到1.24+"
        if ! ask_user "继续安装？" "n"; then
            log_error "安装已取消"
            exit 1
        fi
    fi

    # 检查项目文件
    local repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    cd "$repo_root"

    if [[ ! -f "go.mod" ]] || [[ ! -d "cmd" ]]; then
        log_error "请在kubelet-wuhrai项目根目录运行此脚本"
        exit 1
    fi

    # 快速编译
    log_info "下载依赖..."
    go mod download || { log_error "依赖下载失败"; exit 1; }

    log_info "编译程序..."
    mkdir -p bin

    # 获取版本信息
    local version="dev"
    local commit="none"
    local date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    if command -v git &> /dev/null && git rev-parse --git-dir > /dev/null 2>&1; then
        commit=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
    fi

    go build -ldflags "-X main.version=${version} -X main.commit=${commit} -X main.date=${date}" -o bin/kubelet-wuhrai ./cmd || { log_error "编译失败"; exit 1; }

    # 检查编译结果
    if [[ ! -f "bin/kubelet-wuhrai" ]]; then
        log_error "编译失败，未找到二进制文件"
        exit 1
    fi

    local file_size=$(du -h bin/kubelet-wuhrai | cut -f1)
    log_success "编译完成: ${file_size}"

    # 安装到系统
    install_to_system

    # 验证安装
    verify_installation
}

# 主函数
main() {
    # 处理命令行参数
    case "${1:-}" in
        --quick)
            # 快速安装模式
            echo -e "${BOLD}${CYAN}"
            echo "=================================================="
            echo "    kubelet-wuhrai 快速安装模式"
            echo "=================================================="
            echo -e "${NC}"

            detect_os
            quick_install

            echo -e "${BOLD}${GREEN}"
            echo "=================================================="
            echo "           🎉 快速安装完成！"
            echo "=================================================="
            echo -e "${NC}"

            log_info "🚀 使用方法:"
            log_info "  查看帮助: kubelet-wuhrai --help"
            log_info "  查看版本: kubelet-wuhrai version"
            log_info "  交互模式: kubelet-wuhrai"
            log_info ""
            log_info "🔑 API密钥配置示例:"
            log_info "  export DEEPSEEK_API_KEY=\"your-key\""
            log_info "  export OPENAI_API_KEY=\"your-key\""
            log_info "  export OPENAI_API_BASE=\"https://your-api.com/v1\""

            exit 0
            ;;
        --uninstall)
            uninstall
            exit 0
            ;;
        --help|-h)
            show_help
            exit 0
            ;;
        "")
            # 继续正常安装流程 (完整模式)
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac

    echo -e "${BOLD}${CYAN}"
    echo "=================================================="
    echo "    kubelet-wuhrai 完整安装模式"
    echo "=================================================="
    echo -e "${NC}"

    # 检查网络连接
    check_network

    # 检测操作系统
    detect_os

    # 检查Go环境
    if ! check_go; then
        if ask_user "是否需要安装Go环境？" "y"; then
            install_go
        else
            log_error "Go环境是必需的，安装已取消"
            exit 1
        fi
    fi

    # 编译项目
    build_project

    # 安装到系统
    install_to_system

    # 验证安装
    verify_installation
    
    echo -e "${BOLD}${GREEN}"
    echo "=================================================="
    echo "           🎉 安装完成！"
    echo "=================================================="
    echo -e "${NC}"

    log_info "🚀 使用方法:"
    log_info "  查看帮助: kubelet-wuhrai --help"
    log_info "  查看版本: kubelet-wuhrai version"
    log_info "  交互模式: kubelet-wuhrai"
    log_info "  HTML界面: kubelet-wuhrai --user-interface html"
    log_info "  静默模式: kubelet-wuhrai --quiet \"获取所有pod\""
    log_info ""
    log_info "📁 配置文件:"
    log_info "  示例配置: ~/.config/kubelet-wuhrai/config.yaml.example"
    log_info "  环境变量: ~/.config/kubelet-wuhrai/env.example"
    log_info ""
    log_info "🔑 API密钥配置 (选择一个):"
    log_info "  DeepSeek: export DEEPSEEK_API_KEY=\"your-key\""
    log_info "  OpenAI:   export OPENAI_API_KEY=\"your-key\""
    log_info "  自定义:   export OPENAI_API_KEY=\"your-key\" OPENAI_API_BASE=\"https://your-api.com/v1\""
    log_info ""
    log_info "🌟 使用示例:"
    log_info "  # 使用DeepSeek (默认)"
    log_info "  export DEEPSEEK_API_KEY=\"your-key\""
    log_info "  kubelet-wuhrai \"获取所有运行中的pod\""
    log_info ""
    log_info "  # 使用自定义API"
    log_info "  export OPENAI_API_KEY=\"your-key\""
    log_info "  export OPENAI_API_BASE=\"https://ai.wuhrai.com/v1\""
    log_info "  kubelet-wuhrai --llm-provider openai --model gpt-4o --skip-permissions \"创建nginx deployment\""
    log_info ""
    log_info "📖 更多信息请查看: API_USAGE_EXAMPLES.md"
}

# 运行主函数
main "$@"
