#!/bin/bash

# 上传预编译文件到GitHub Release
# 需要设置 GITHUB_TOKEN 环境变量

set -e

# 配置
REPO="st-lzh/kubelet-wuhrai"
RELEASE_ID="233175959"
RELEASE_DIR="releases"

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

echo "========================================"
echo "  上传预编译文件到GitHub Release"
echo "========================================"
echo ""

# 检查GitHub token
if [[ -z "$GITHUB_TOKEN" ]]; then
    log_error "请设置 GITHUB_TOKEN 环境变量"
    echo "获取token: https://github.com/settings/tokens"
    exit 1
fi

# 检查releases目录
if [[ ! -d "$RELEASE_DIR" ]]; then
    log_error "releases目录不存在，请先运行 ./build-releases.sh"
    exit 1
fi

log_info "Release ID: $RELEASE_ID"
log_info "上传目录: $RELEASE_DIR"
echo ""

# 上传每个文件
for file in "$RELEASE_DIR"/*; do
    if [[ -f "$file" ]]; then
        filename=$(basename "$file")
        log_info "上传 $filename..."
        
        # 上传文件到GitHub Release
        response=$(curl -s -w "%{http_code}" \
            -H "Authorization: token $GITHUB_TOKEN" \
            -H "Content-Type: application/octet-stream" \
            --data-binary @"$file" \
            "https://uploads.github.com/repos/$REPO/releases/$RELEASE_ID/assets?name=$filename")
        
        # 检查响应状态码
        http_code="${response: -3}"
        if [[ "$http_code" == "201" ]]; then
            log_success "✓ $filename 上传成功"
        else
            log_error "✗ $filename 上传失败 (HTTP $http_code)"
            echo "响应: ${response%???}"
        fi
    fi
done

echo ""
log_success "上传完成！"
log_info "查看release: https://github.com/$REPO/releases/tag/v1.0.0"
echo ""
