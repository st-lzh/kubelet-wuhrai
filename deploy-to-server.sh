#!/bin/bash

# kubelet-wuhrai 远程服务器部署脚本
# 用法: ./deploy-to-server.sh user@server:/path/to/install

set -e

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查参数
if [[ $# -lt 1 ]]; then
    echo "用法: $0 user@server [install_path]"
    echo "示例: $0 root@192.168.1.100"
    echo "示例: $0 ubuntu@myserver.com /opt/kubelet-wuhrai"
    exit 1
fi

SERVER="$1"
REMOTE_PATH="${2:-/usr/local/bin}"
BINARY_NAME="kubelet-wuhrai"

# 检查本地二进制文件
if [[ ! -f "./$BINARY_NAME" ]]; then
    log_info "本地未找到二进制文件，尝试编译..."
    if [[ -f "./cmd/main.go" ]]; then
        go build -o "$BINARY_NAME" ./cmd/
        log_success "编译完成"
    else
        log_error "未找到源码，请先编译项目"
        exit 1
    fi
fi

# 检查SSH连接
log_info "测试SSH连接到 $SERVER..."
if ! ssh -o ConnectTimeout=10 "$SERVER" "echo 'SSH连接成功'" >/dev/null 2>&1; then
    log_error "无法连接到服务器 $SERVER"
    log_info "请检查SSH配置和网络连接"
    exit 1
fi

log_success "SSH连接正常"

# 创建远程安装脚本
REMOTE_SCRIPT=$(cat << 'EOF'
#!/bin/bash
set -e

BINARY_NAME="$1"
INSTALL_PATH="$2"
CONFIG_DIR="$3"

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }

# 创建安装目录
log_info "创建安装目录 $INSTALL_PATH..."
sudo mkdir -p "$INSTALL_PATH"
sudo mkdir -p "$CONFIG_DIR"

# 移动二进制文件
log_info "安装 $BINARY_NAME..."
sudo mv "/tmp/$BINARY_NAME" "$INSTALL_PATH/"
sudo chmod +x "$INSTALL_PATH/$BINARY_NAME"

# 创建软链接到 /usr/local/bin (如果安装路径不是 /usr/local/bin)
if [[ "$INSTALL_PATH" != "/usr/local/bin" ]]; then
    sudo ln -sf "$INSTALL_PATH/$BINARY_NAME" "/usr/local/bin/$BINARY_NAME"
    log_info "创建软链接到 /usr/local/bin"
fi

# 设置环境变量
if [[ ":$PATH:" != *":$INSTALL_PATH:"* ]]; then
    echo "export PATH=\"$INSTALL_PATH:\$PATH\"" | sudo tee -a /etc/profile > /dev/null
    log_info "已添加到系统PATH"
fi

# 创建配置文件
CONFIG_FILE="$CONFIG_DIR/config.yaml"
if [[ ! -f "$CONFIG_FILE" ]]; then
    sudo tee "$CONFIG_FILE" > /dev/null << 'CONFIGEOF'
# kubelet-wuhrai 服务器配置

# API 配置 (选择一个)
# deepseek_api_key: "your-api-key"
# openai_api_key: "your-api-key"
# qwen_api_key: "your-api-key"

# 服务器设置
quiet: false
skip_permissions: false
enable_tool_use_shim: false

# 日志配置
log_level: "info"
CONFIGEOF
    log_success "配置文件已创建: $CONFIG_FILE"
fi

# 验证安装
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    VERSION=$("$BINARY_NAME" version 2>/dev/null || echo "unknown")
    log_success "安装成功！版本: $VERSION"
else
    log_info "请重新登录或运行: source /etc/profile"
fi

echo ""
echo "安装完成！"
echo "配置文件: $CONFIG_FILE"
echo "测试命令: $BINARY_NAME version"
EOF
)

# 上传文件和脚本
log_info "上传二进制文件到服务器..."
scp "./$BINARY_NAME" "$SERVER:/tmp/"

log_info "上传安装脚本..."
echo "$REMOTE_SCRIPT" | ssh "$SERVER" "cat > /tmp/install_kubelet.sh && chmod +x /tmp/install_kubelet.sh"

# 执行远程安装
log_info "在服务器上执行安装..."
ssh "$SERVER" "/tmp/install_kubelet.sh '$BINARY_NAME' '$REMOTE_PATH' '/etc/kubelet-wuhrai'"

# 清理临时文件
log_info "清理临时文件..."
ssh "$SERVER" "rm -f /tmp/install_kubelet.sh"

log_success "部署完成！"
echo ""
echo "服务器信息:"
echo "  服务器: $SERVER"
echo "  安装路径: $REMOTE_PATH"
echo "  配置目录: /etc/kubelet-wuhrai"
echo ""
echo "下一步:"
echo "1. SSH登录服务器: ssh $SERVER"
echo "2. 配置API密钥: sudo vi /etc/kubelet-wuhrai/config.yaml"
echo "3. 测试命令: kubelet-wuhrai version"
echo ""
