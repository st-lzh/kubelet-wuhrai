#!/usr/bin/env bash
# kubelet-wuhrai 快速安装脚本
# 简化版一键安装，适用于快速部署

set -e

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# 检查命令是否存在
has_command() {
    command -v "$1" >/dev/null 2>&1
}

# 询问用户
ask() {
    echo -e "${YELLOW}$1 [y/N]:${NC} "
    read -r response
    [[ "$response" =~ ^[Yy]$ ]]
}

echo "🚀 kubelet-wuhrai 快速安装"
echo "=========================="

# 检查Go
if has_command go; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    info "检测到Go版本: $GO_VERSION"
    
    # 简单版本检查
    if [[ "$GO_VERSION" < "1.24" ]]; then
        warn "Go版本可能过低，建议升级到1.24+"
        if ! ask "继续安装？"; then
            error "安装已取消"
        fi
    fi
else
    error "未找到Go环境，请先安装Go 1.24+: https://golang.org/dl/"
fi

# 检查项目文件
if [[ ! -f "go.mod" ]] || [[ ! -d "cmd" ]]; then
    error "请在kubelet-wuhrai项目根目录运行此脚本"
fi

# 快速编译
info "下载依赖..."
go mod download || error "依赖下载失败"

info "编译程序..."
mkdir -p bin
go build -o bin/kubelet-wuhrai ./cmd || error "编译失败"

# 检查编译结果
if [[ ! -f "bin/kubelet-wuhrai" ]]; then
    error "编译失败，未找到二进制文件"
fi

success "编译完成: $(du -h bin/kubelet-wuhrai | cut -f1)"

# 安装到系统 - 优先使用全局目录
INSTALL_DIR=""
USE_GLOBAL=false

# 首先尝试系统全局目录
if [[ -w "/usr/local/bin" ]]; then
    INSTALL_DIR="/usr/local/bin"
    USE_GLOBAL=true
    info "使用系统全局目录: $INSTALL_DIR"
elif sudo -n true 2>/dev/null && [[ -d "/usr/local/bin" ]]; then
    INSTALL_DIR="/usr/local/bin"
    USE_GLOBAL=true
    info "使用sudo安装到系统全局目录: $INSTALL_DIR"
elif [[ -n "${GOPATH:-}" ]] && [[ -d "${GOPATH}/bin" ]]; then
    INSTALL_DIR="${GOPATH}/bin"
elif [[ -d "$HOME/go/bin" ]]; then
    INSTALL_DIR="$HOME/go/bin"
elif [[ -d "$HOME/.local/bin" ]]; then
    INSTALL_DIR="$HOME/.local/bin"
else
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

info "安装到: $INSTALL_DIR"

# 复制二进制文件
if [[ "$USE_GLOBAL" == "true" ]] && [[ ! -w "$INSTALL_DIR" ]]; then
    if [[ -f "$INSTALL_DIR/kubelet-wuhrai" ]]; then
        sudo rm -f "$INSTALL_DIR/kubelet-wuhrai"
    fi
    sudo cp bin/kubelet-wuhrai "$INSTALL_DIR/" || error "安装失败"
    sudo chmod +x "$INSTALL_DIR/kubelet-wuhrai"
    success "已使用sudo安装到系统全局目录"
else
    if [[ -f "$INSTALL_DIR/kubelet-wuhrai" ]]; then
        rm -f "$INSTALL_DIR/kubelet-wuhrai"
    fi
    cp bin/kubelet-wuhrai "$INSTALL_DIR/" || error "安装失败"
    chmod +x "$INSTALL_DIR/kubelet-wuhrai"
fi

# 创建全局符号链接（如果不是已经在全局目录）
if [[ "$INSTALL_DIR" != "/usr/local/bin" ]] && [[ -d "/usr/local/bin" ]]; then
    if sudo -n true 2>/dev/null; then
        info "创建全局符号链接到 /usr/local/bin"
        sudo ln -sf "$INSTALL_DIR/kubelet-wuhrai" "/usr/local/bin/kubelet-wuhrai"
        success "全局符号链接创建成功"
    elif ask "是否创建全局符号链接到 /usr/local/bin (需要sudo权限)？"; then
        sudo ln -sf "$INSTALL_DIR/kubelet-wuhrai" "/usr/local/bin/kubelet-wuhrai"
        success "全局符号链接创建成功"
    fi
fi

# 检查PATH配置
if [[ "$INSTALL_DIR" == "/usr/local/bin" ]] || [[ "$INSTALL_DIR" == "/usr/bin" ]]; then
    info "已安装到系统全局目录，无需配置PATH"
elif [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    warn "$INSTALL_DIR 不在PATH中"

    # 自动添加到shell配置
    SHELL_CONFIGS=()
    if [[ -n "${BASH_VERSION:-}" ]]; then
        SHELL_CONFIGS+=("$HOME/.bashrc")
    fi
    if [[ -n "${ZSH_VERSION:-}" ]]; then
        SHELL_CONFIGS+=("$HOME/.zshrc")
    fi
    SHELL_CONFIGS+=("$HOME/.profile")

    if ask "是否自动添加到PATH？"; then
        for SHELL_RC in "${SHELL_CONFIGS[@]}"; do
            if [[ -f "$SHELL_RC" ]] || [[ "$SHELL_RC" == "$HOME/.profile" ]]; then
                if ! grep -q "kubelet-wuhrai" "$SHELL_RC" 2>/dev/null; then
                    echo "" >> "$SHELL_RC"
                    echo "# kubelet-wuhrai global command" >> "$SHELL_RC"
                    echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$SHELL_RC"
                    success "已添加到 $SHELL_RC"
                fi
            fi
        done
        info "请运行: source ~/.bashrc (或相应的配置文件)"

        # 临时设置PATH
        export PATH="$INSTALL_DIR:$PATH"
    fi
fi

# 创建配置目录
CONFIG_DIR="$HOME/.config/kubelet-wuhrai"
mkdir -p "$CONFIG_DIR"

# 简单配置文件
if [[ ! -f "$CONFIG_DIR/config.yaml" ]]; then
    cat > "$CONFIG_DIR/config.yaml" << 'EOF'
# kubelet-wuhrai 配置文件
llmProvider: "deepseek"
model: "deepseek-chat"
skipPermissions: false
maxIterations: 20
userInterface: "terminal"
EOF
    info "创建配置文件: $CONFIG_DIR/config.yaml"
fi

# 验证安装
info "验证安装..."

# 检查安装位置
FOUND_LOCATIONS=()
for dir in "/usr/local/bin" "$HOME/go/bin" "$HOME/.local/bin" "${GOPATH:-}/bin"; do
    if [[ -f "$dir/kubelet-wuhrai" ]]; then
        FOUND_LOCATIONS+=("$dir/kubelet-wuhrai")
    fi
done

if [[ ${#FOUND_LOCATIONS[@]} -gt 0 ]]; then
    success "找到kubelet-wuhrai安装位置："
    for location in "${FOUND_LOCATIONS[@]}"; do
        info "  - $location"
    done
fi

# 测试命令可用性
if has_command kubelet-wuhrai; then
    VERSION=$(kubelet-wuhrai version 2>/dev/null || echo "版本信息获取失败")
    success "✅ kubelet-wuhrai 全局命令可用！"
    echo "$VERSION"
else
    warn "⚠️ kubelet-wuhrai命令不在当前PATH中"
    info "二进制文件位置: $INSTALL_DIR/kubelet-wuhrai"
    info "请重新加载shell配置或重新打开终端"
fi

echo ""
echo "🎉 安装完成！"
echo "============="
echo "使用方法:"
echo "  kubelet-wuhrai --help     # 查看帮助"
echo "  kubelet-wuhrai version    # 查看版本"
echo "  kubelet-wuhrai           # 交互模式"
echo ""
echo "配置文件: $CONFIG_DIR/config.yaml"
echo "注意: 使用前请配置LLM API密钥环境变量"

# 显示环境变量提示
echo ""
echo "📝 环境变量配置示例:"
echo "export DEEPSEEK_API_KEY='your-api-key'"
echo "export OPENAI_API_KEY='your-api-key'"
echo "export QWEN_API_KEY='your-api-key'"
