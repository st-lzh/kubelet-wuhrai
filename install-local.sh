#!/usr/bin/env bash
# 本地安装脚本 - 将编译好的kubelet-wuhrai安装到系统中
# Copyright 2025 kubelet-wuhrai

set -o errexit
set -o nounset
set -o pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# 获取项目根目录
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${REPO_ROOT}"

# 检查二进制文件是否存在
BINARY_PATH="${REPO_ROOT}/bin/kubelet-wuhrai"
if [ ! -f "${BINARY_PATH}" ]; then
    log_error "二进制文件不存在: ${BINARY_PATH}"
    log_info "请先运行 ./build.sh 编译项目"
    exit 1
fi

# 确定安装目录
if [ -n "${GOPATH:-}" ] && [ -d "${GOPATH}/bin" ]; then
    INSTALL_DIR="${GOPATH}/bin"
elif [ -d "${HOME}/go/bin" ]; then
    INSTALL_DIR="${HOME}/go/bin"
elif [ -d "${HOME}/.local/bin" ]; then
    INSTALL_DIR="${HOME}/.local/bin"
elif [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
else
    # 创建用户本地bin目录
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "${INSTALL_DIR}"
fi

log_info "安装 kubelet-wuhrai 到 ${INSTALL_DIR}..."

# 复制二进制文件
cp "${BINARY_PATH}" "${INSTALL_DIR}/"
if [ $? -eq 0 ]; then
    log_success "二进制文件已安装到: ${INSTALL_DIR}/kubelet-wuhrai"
else
    log_error "安装失败"
    exit 1
fi

# 检查PATH
if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
    log_warning "${INSTALL_DIR} 不在您的 PATH 中"
    log_info "请将以下行添加到您的 shell 配置文件中 (~/.bashrc, ~/.zshrc 等):"
    echo "export PATH=\"${INSTALL_DIR}:\$PATH\""
    log_info "然后运行: source ~/.bashrc (或相应的配置文件)"
fi

# 验证安装
log_info "验证安装..."
if command -v kubelet-wuhrai &> /dev/null; then
    VERSION_OUTPUT=$(kubelet-wuhrai version 2>/dev/null || echo "无法获取版本信息")
    log_success "安装成功！"
    log_info "版本信息:"
    echo "${VERSION_OUTPUT}"
else
    log_warning "kubelet-wuhrai 命令不在 PATH 中，但已安装到 ${INSTALL_DIR}"
fi

# 创建配置目录
CONFIG_DIR="${HOME}/.config/kubelet-wuhrai"
if [ ! -d "${CONFIG_DIR}" ]; then
    mkdir -p "${CONFIG_DIR}"
    log_info "创建配置目录: ${CONFIG_DIR}"
fi

# 创建示例配置文件
EXAMPLE_CONFIG="${CONFIG_DIR}/config.yaml.example"
if [ ! -f "${EXAMPLE_CONFIG}" ]; then
    cat > "${EXAMPLE_CONFIG}" << 'EOF'
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

# 文件路径
# kubeConfigPath: ""     # kubeconfig 文件路径
# tracePath: ""          # 跟踪文件路径
# promptTemplateFilePath: ""  # 自定义提示模板文件路径
# extraPromptPaths: []   # 额外提示模板路径
# toolConfigPaths: []    # 自定义工具配置路径
EOF
    log_info "创建示例配置文件: ${EXAMPLE_CONFIG}"
fi

log_success "安装完成！"
log_info ""
log_info "使用方法:"
log_info "  查看帮助: kubelet-wuhrai --help"
log_info "  查看版本: kubelet-wuhrai version"
log_info "  交互模式: kubelet-wuhrai"
log_info "  HTML界面: kubelet-wuhrai --user-interface html"
log_info "  静默模式: kubelet-wuhrai --quiet \"获取所有pod\""
log_info ""
log_info "配置文件:"
log_info "  配置目录: ${CONFIG_DIR}"
log_info "  示例配置: ${EXAMPLE_CONFIG}"
log_info ""
log_info "注意: 使用前请确保已配置相应的 LLM API 密钥环境变量"
