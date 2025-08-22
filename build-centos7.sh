#!/usr/bin/env bash
# CentOS 7.9 专用编译脚本
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

log_info "开始为 CentOS 7.9 编译 kubelet-wuhrai..."
log_info "项目根目录: ${REPO_ROOT}"

# 检查Go环境
if ! command -v go &> /dev/null; then
    log_error "Go 未安装或不在PATH中"
    log_error "CentOS 7.9 需要手动安装 Go 1.21+ 版本"
    exit 1
fi

GO_VERSION=$(go version)
log_info "Go版本: ${GO_VERSION}"

# 检查Go版本是否满足要求 (需要 Go 1.21+)
GO_VERSION_NUM=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
MAJOR_VERSION=$(echo "$GO_VERSION_NUM" | cut -d. -f1)
MINOR_VERSION=$(echo "$GO_VERSION_NUM" | cut -d. -f2)

if [ "$MAJOR_VERSION" -lt 1 ] || ([ "$MAJOR_VERSION" -eq 1 ] && [ "$MINOR_VERSION" -lt 21 ]); then
    log_error "Go 版本过低，需要 Go 1.21+ (当前: go${GO_VERSION_NUM})"
    log_error "请升级 Go 版本"
    exit 1
fi

# 创建输出目录
BIN_DIR="${REPO_ROOT}/bin"
DIST_DIR="${REPO_ROOT}/dist"
mkdir -p "${BIN_DIR}"
mkdir -p "${DIST_DIR}"

# 编译信息
VERSION=${VERSION:-"1.0.0"}
COMMIT=${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "none")}
DATE=${DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}

log_info "构建信息:"
log_info "  版本: ${VERSION}"
log_info "  提交: ${COMMIT}"
log_info "  日期: ${DATE}"
log_info "  目标平台: CentOS 7.9 (linux/amd64)"

# 设置 CentOS 7.9 兼容的构建环境
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

# 对于老系统，禁用一些新特性以确保兼容性
export CGO_LDFLAGS="-static"

log_info "构建环境:"
log_info "  CGO_ENABLED: ${CGO_ENABLED}"
log_info "  GOOS: ${GOOS}"
log_info "  GOARCH: ${GOARCH}"

# 构建 LDFLAGS
LDFLAGS="-w -s"
LDFLAGS+=" -X main.version=${VERSION}"
LDFLAGS+=" -X main.commit=${COMMIT}"
LDFLAGS+=" -X main.date=${DATE}"

# 静态链接，确保在 CentOS 7.9 上可以运行
LDFLAGS+=" -extldflags '-static'"

log_info "LDFLAGS: ${LDFLAGS}"

# 清理模块缓存（避免版本冲突）
log_info "清理 Go 模块缓存..."
go clean -modcache || true

# 下载依赖
log_info "下载依赖..."
go mod download
if [ $? -ne 0 ]; then
    log_error "依赖下载失败"
    exit 1
fi

# 编译主程序
log_info "编译主程序 kubelet-wuhrai..."
go build \
    -ldflags "${LDFLAGS}" \
    -tags "netgo osusergo" \
    -installsuffix netgo \
    -o "${BIN_DIR}/kubelet-wuhrai" \
    ./cmd

if [ $? -eq 0 ]; then
    log_success "主程序编译完成: ${BIN_DIR}/kubelet-wuhrai"
else
    log_error "主程序编译失败"
    exit 1
fi

# 检查二进制文件
log_info "检查编译结果..."
if [ -f "${BIN_DIR}/kubelet-wuhrai" ]; then
    file_size=$(du -h "${BIN_DIR}/kubelet-wuhrai" | cut -f1)
    file_info=$(file "${BIN_DIR}/kubelet-wuhrai")
    log_success "主程序编译成功，大小: ${file_size}"
    log_info "文件信息: ${file_info}"
    
    # 检查是否是静态链接的二进制文件
    if ldd "${BIN_DIR}/kubelet-wuhrai" 2>&1 | grep -q "not a dynamic executable"; then
        log_success "二进制文件是静态链接的，适合在 CentOS 7.9 上运行"
    else
        log_warning "二进制文件可能不是静态链接的"
        ldd "${BIN_DIR}/kubelet-wuhrai" 2>/dev/null || true
    fi
    
    # 检查文件权限
    chmod +x "${BIN_DIR}/kubelet-wuhrai"
    
    # 跳过测试程序运行（在macOS上无法运行Linux二进制文件）
    if [ "$(uname)" = "Linux" ]; then
        log_info "测试程序基本功能..."
        "${BIN_DIR}/kubelet-wuhrai" version
        if [ $? -eq 0 ]; then
            log_success "程序基本功能测试通过"
        else
            log_error "程序基本功能测试失败"
            exit 1
        fi
    else
        log_info "跳过运行测试（交叉编译的二进制文件无法在当前平台运行）"
    fi
else
    log_error "主程序编译失败"
    exit 1
fi

# 创建 CentOS 7.9 专用发布包
log_info "创建 CentOS 7.9 发布包..."
ARCHIVE_NAME="kubelet-wuhrai-${VERSION}-centos7.9-amd64"
ARCHIVE_PATH="${DIST_DIR}/${ARCHIVE_NAME}.tar.gz"

# 创建临时目录用于打包
TEMP_DIR=$(mktemp -d)
PACKAGE_DIR="${TEMP_DIR}/${ARCHIVE_NAME}"
mkdir -p "${PACKAGE_DIR}"

# 复制二进制文件
cp "${BIN_DIR}/kubelet-wuhrai" "${PACKAGE_DIR}/"

# 创建安装脚本
cat > "${PACKAGE_DIR}/install.sh" << 'EOF'
#!/bin/bash
# kubelet-wuhrai CentOS 7.9 安装脚本

set -e

INSTALL_DIR="/usr/local/bin"
BINARY_NAME="kubelet-wuhrai"

echo "安装 kubelet-wuhrai 到 ${INSTALL_DIR}..."

# 检查权限
if [ "$EUID" -ne 0 ]; then
    echo "请使用 sudo 运行此脚本"
    exit 1
fi

# 安装二进制文件
cp "${BINARY_NAME}" "${INSTALL_DIR}/"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

echo "安装完成！"
echo "运行 'kubelet-wuhrai --help' 开始使用"
EOF

chmod +x "${PACKAGE_DIR}/install.sh"

# 创建 README
cat > "${PACKAGE_DIR}/README.md" << EOF
# kubelet-wuhrai for CentOS 7.9

这是专为 CentOS 7.9 系统编译的 kubelet-wuhrai 版本。

## 系统要求

- CentOS 7.9 (x86_64)
- 静态编译的二进制文件，无需额外依赖

## 安装方法

### 方法 1: 使用安装脚本
\`\`\`bash
sudo ./install.sh
\`\`\`

### 方法 2: 手动安装
\`\`\`bash
sudo cp kubelet-wuhrai /usr/local/bin/
sudo chmod +x /usr/local/bin/kubelet-wuhrai
\`\`\`

## 使用方法

\`\`\`bash
kubelet-wuhrai --help
kubelet-wuhrai version
\`\`\`

## 版本信息

- 版本: ${VERSION}
- 提交: ${COMMIT}
- 构建日期: ${DATE}
- 目标平台: CentOS 7.9 (linux/amd64)

## 故障排除

如果遇到权限问题，请确保：
1. 使用 sudo 运行安装脚本
2. /usr/local/bin 在您的 PATH 环境变量中

如果遇到"Permission denied"错误：
\`\`\`bash
chmod +x kubelet-wuhrai
\`\`\`
EOF

# 打包
tar -czf "${ARCHIVE_PATH}" -C "${TEMP_DIR}" "${ARCHIVE_NAME}"
if [ $? -eq 0 ]; then
    archive_size=$(du -h "${ARCHIVE_PATH}" | cut -f1)
    log_success "CentOS 7.9 发布包创建成功: ${ARCHIVE_PATH} (${archive_size})"
else
    log_error "发布包创建失败"
    rm -rf "${TEMP_DIR}"
    exit 1
fi

# 清理临时目录
rm -rf "${TEMP_DIR}"

# 生成校验和
log_info "生成校验和..."
cd "${DIST_DIR}"
sha256sum "${ARCHIVE_NAME}.tar.gz" > "${ARCHIVE_NAME}.tar.gz.sha256"
md5sum "${ARCHIVE_NAME}.tar.gz" > "${ARCHIVE_NAME}.tar.gz.md5"
log_success "校验和文件已生成"

cd "${REPO_ROOT}"

log_success "CentOS 7.9 编译打包完成！"
echo ""
log_info "=========================================="
log_info "输出文件:"
log_info "  二进制文件: ${BIN_DIR}/kubelet-wuhrai"
log_info "  发布包: ${ARCHIVE_PATH}"
log_info "  SHA256: ${ARCHIVE_PATH}.sha256"
log_info "  MD5: ${ARCHIVE_PATH}.md5"
echo ""
log_info "部署到 CentOS 7.9 服务器:"
log_info "  1. 上传发布包到目标服务器"
log_info "  2. 解压: tar -xzf ${ARCHIVE_NAME}.tar.gz"
log_info "  3. 进入目录: cd ${ARCHIVE_NAME}"
log_info "  4. 安装: sudo ./install.sh"
echo ""
log_info "或者直接复制二进制文件:"
log_info "  scp ${BIN_DIR}/kubelet-wuhrai user@centos7-server:/usr/local/bin/"
log_info "=========================================="