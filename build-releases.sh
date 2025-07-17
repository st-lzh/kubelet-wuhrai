#!/bin/bash

# 构建多架构预编译版本脚本
# 用于创建GitHub releases的二进制文件

set -e

# 配置
BINARY_NAME="kubelet-wuhrai"
VERSION=${1:-"v1.0.0"}
RELEASE_DIR="releases"

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

echo "========================================"
echo "  构建 kubelet-wuhrai 预编译版本"
echo "========================================"
echo ""

log_info "版本: $VERSION"
log_info "输出目录: $RELEASE_DIR"
echo ""

# 清理并创建发布目录
rm -rf "$RELEASE_DIR"
mkdir -p "$RELEASE_DIR"

# 支持的平台和架构
declare -a platforms=(
    "linux/amd64"
    "linux/arm64" 
    "linux/arm"
    "linux/386"
    "darwin/amd64"
    "darwin/arm64"
)

log_info "开始构建多架构版本..."
echo ""

for platform in "${platforms[@]}"; do
    IFS='/' read -r os arch <<< "$platform"
    
    output_name="${BINARY_NAME}-${os}-${arch}"
    if [[ "$os" == "windows" ]]; then
        output_name="${output_name}.exe"
    fi
    
    log_info "构建 $platform -> $output_name"
    
    # 设置环境变量并构建
    env GOOS="$os" GOARCH="$arch" go build \
        -ldflags="-w -s -X main.version=$VERSION" \
        -o "$RELEASE_DIR/$output_name" \
        ./cmd/
    
    # 检查文件大小
    size=$(ls -lh "$RELEASE_DIR/$output_name" | awk '{print $5}')
    log_success "✓ $output_name ($size)"
done

echo ""
log_success "所有版本构建完成！"
echo ""

# 显示构建结果
log_info "构建文件列表:"
ls -la "$RELEASE_DIR/"

echo ""
log_info "文件大小统计:"
du -h "$RELEASE_DIR"/*

echo ""
log_info "SHA256校验和:"
cd "$RELEASE_DIR"
for file in *; do
    if [[ -f "$file" ]]; then
        sha256sum "$file" || shasum -a 256 "$file"
    fi
done
cd ..

echo ""
log_warn "下一步操作:"
echo "1. 测试二进制文件: ./$RELEASE_DIR/$BINARY_NAME-linux-amd64 version"
echo "2. 创建GitHub release: gh release create $VERSION $RELEASE_DIR/*"
echo "3. 或手动上传到: https://github.com/st-lzh/kubelet-wuhrai/releases"
echo ""
