#!/usr/bin/env bash
# 快速构建 CentOS 7.9 二进制文件
# 简化版本，仅生成二进制文件

set -e

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}快速构建 kubelet-wuhrai for CentOS 7.9...${NC}"

# 设置构建参数
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

# 版本信息
VERSION=${VERSION:-"1.0.0"}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# 构建目录
mkdir -p bin

# 编译
echo "正在编译..."
go build \
    -ldflags "-w -s -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE} -extldflags '-static'" \
    -tags "netgo osusergo" \
    -installsuffix netgo \
    -o bin/kubelet-wuhrai-centos7 \
    ./cmd

if [ $? -eq 0 ]; then
    SIZE=$(du -h bin/kubelet-wuhrai-centos7 | cut -f1)
    echo -e "${GREEN}构建成功！${NC}"
    echo "文件: bin/kubelet-wuhrai-centos7"
    echo "大小: ${SIZE}"
    echo "版本: ${VERSION}"
    echo ""
    echo "部署到 CentOS 7.9:"
    echo "  scp bin/kubelet-wuhrai-centos7 user@server:/usr/local/bin/kubelet-wuhrai"
    echo "  ssh user@server 'chmod +x /usr/local/bin/kubelet-wuhrai'"
else
    echo "构建失败！"
    exit 1
fi