#!/usr/bin/env bash
# 编译脚本 - 不使用Docker进行编译打包
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

log_info "开始编译 kubelet-wuhrai 项目..."
log_info "项目根目录: ${REPO_ROOT}"

# 检查Go环境
if ! command -v go &> /dev/null; then
    log_error "Go 未安装或不在PATH中"
    exit 1
fi

GO_VERSION=$(go version)
log_info "Go版本: ${GO_VERSION}"

# 创建输出目录
BIN_DIR="${REPO_ROOT}/bin"
DIST_DIR="${REPO_ROOT}/dist"
mkdir -p "${BIN_DIR}"
mkdir -p "${DIST_DIR}"

# 编译信息
VERSION=${VERSION:-"dev"}
COMMIT=${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "none")}
DATE=${DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}

log_info "构建信息:"
log_info "  版本: ${VERSION}"
log_info "  提交: ${COMMIT}"
log_info "  日期: ${DATE}"

# 构建标志
BUILD_FLAGS="-ldflags=-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

# 编译主程序
log_info "编译主程序 kubelet-wuhrai..."
go build -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" -o "${BIN_DIR}/kubelet-wuhrai" ./cmd
if [ $? -eq 0 ]; then
    log_success "主程序编译完成: ${BIN_DIR}/kubelet-wuhrai"
else
    log_error "主程序编译失败"
    exit 1
fi

# 编译所有子模块
log_info "编译所有Go模块..."

# 查找所有go.mod文件并编译
for go_mod in $(find "${REPO_ROOT}" -name go.mod); do
    module_dir=$(dirname "${go_mod}")
    module_name=$(basename "${module_dir}")
    
    log_info "编译模块: ${module_name} (${module_dir})"
    
    cd "${module_dir}"
    
    # 整理依赖
    go mod tidy
    if [ $? -ne 0 ]; then
        log_warning "模块 ${module_name} 依赖整理失败"
        continue
    fi
    
    # 编译模块
    go build ./...
    if [ $? -eq 0 ]; then
        log_success "模块 ${module_name} 编译成功"
    else
        log_warning "模块 ${module_name} 编译失败"
    fi
    
    cd "${REPO_ROOT}"
done

# 运行测试
log_info "运行测试..."
cd "${REPO_ROOT}"
go test ./... -v
if [ $? -eq 0 ]; then
    log_success "所有测试通过"
else
    log_warning "部分测试失败"
fi

# 检查编译结果
log_info "检查编译结果..."
if [ -f "${BIN_DIR}/kubelet-wuhrai" ]; then
    file_size=$(du -h "${BIN_DIR}/kubelet-wuhrai" | cut -f1)
    log_success "主程序编译成功，大小: ${file_size}"
    
    # 测试程序是否能正常运行
    log_info "测试程序运行..."
    "${BIN_DIR}/kubelet-wuhrai" version
    if [ $? -eq 0 ]; then
        log_success "程序运行测试通过"
    else
        log_error "程序运行测试失败"
        exit 1
    fi
else
    log_error "主程序编译失败"
    exit 1
fi

# 创建发布包
log_info "创建发布包..."
ARCHIVE_NAME="kubelet-wuhrai-${VERSION}-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)"
ARCHIVE_PATH="${DIST_DIR}/${ARCHIVE_NAME}.tar.gz"

tar -czf "${ARCHIVE_PATH}" -C "${BIN_DIR}" kubelet-wuhrai
if [ $? -eq 0 ]; then
    archive_size=$(du -h "${ARCHIVE_PATH}" | cut -f1)
    log_success "发布包创建成功: ${ARCHIVE_PATH} (${archive_size})"
else
    log_error "发布包创建失败"
    exit 1
fi

# 生成校验和
log_info "生成校验和..."
cd "${DIST_DIR}"
sha256sum "${ARCHIVE_NAME}.tar.gz" > "${ARCHIVE_NAME}.tar.gz.sha256"
log_success "校验和文件: ${DIST_DIR}/${ARCHIVE_NAME}.tar.gz.sha256"

cd "${REPO_ROOT}"

log_success "编译打包完成！"
log_info "输出文件:"
log_info "  二进制文件: ${BIN_DIR}/kubelet-wuhrai"
log_info "  发布包: ${ARCHIVE_PATH}"
log_info "  校验和: ${ARCHIVE_PATH}.sha256"

log_info "使用方法:"
log_info "  直接运行: ${BIN_DIR}/kubelet-wuhrai --help"
log_info "  解压发布包: tar -xzf ${ARCHIVE_PATH}"
