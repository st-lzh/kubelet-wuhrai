# kubelet-wuhrai CentOS 7.9 部署指南

这个文档说明如何在 CentOS 7.9 系统上部署 kubelet-wuhrai。

## 系统要求

- **操作系统**: CentOS 7.9 (x86_64)
- **架构**: amd64 (x86_64)
- **依赖**: 无（静态编译的二进制文件）

## 构建信息

- **版本**: 1.0.0
- **编译时间**: 2025-08-22T06:34:28Z
- **提交哈希**: f631049
- **构建选项**: 静态链接，CGO禁用，针对Linux优化

## 文件列表

```
kubelet-wuhrai-1.0.0-centos7.9-amd64.tar.gz     # 主发布包 (10MB)
kubelet-wuhrai-1.0.0-centos7.9-amd64.tar.gz.sha256   # SHA256校验和
kubelet-wuhrai-1.0.0-centos7.9-amd64.tar.gz.md5      # MD5校验和
```

## 部署步骤

### 步骤 1: 下载和验证

```bash
# 下载发布包
wget /path/to/kubelet-wuhrai-1.0.0-centos7.9-amd64.tar.gz

# 验证完整性 (推荐)
sha256sum -c kubelet-wuhrai-1.0.0-centos7.9-amd64.tar.gz.sha256
# 或者
md5sum -c kubelet-wuhrai-1.0.0-centos7.9-amd64.tar.gz.md5
```

### 步骤 2: 解压

```bash
tar -xzf kubelet-wuhrai-1.0.0-centos7.9-amd64.tar.gz
cd kubelet-wuhrai-1.0.0-centos7.9-amd64
```

### 步骤 3: 安装

#### 方法 1: 使用安装脚本 (推荐)

```bash
sudo ./install.sh
```

#### 方法 2: 手动安装

```bash
# 复制到系统目录
sudo cp kubelet-wuhrai /usr/local/bin/

# 设置执行权限
sudo chmod +x /usr/local/bin/kubelet-wuhrai

# 验证安装
kubelet-wuhrai version
```

### 步骤 4: 验证安装

```bash
# 检查版本
kubelet-wuhrai version

# 查看帮助
kubelet-wuhrai --help

# 检查二进制文件信息
file /usr/local/bin/kubelet-wuhrai
```

预期输出：
```
/usr/local/bin/kubelet-wuhrai: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, BuildID[sha1]=..., stripped
```

## 快速部署 (单命令)

如果您有 SSH 访问权限，可以使用以下命令快速部署：

```bash
# 从本地复制到远程服务器
scp kubelet-wuhrai-1.0.0-centos7.9-amd64.tar.gz user@centos7-server:~/

# SSH 到服务器并安装
ssh user@centos7-server << 'EOF'
tar -xzf kubelet-wuhrai-1.0.0-centos7.9-amd64.tar.gz
cd kubelet-wuhrai-1.0.0-centos7.9-amd64
sudo ./install.sh
kubelet-wuhrai version
EOF
```

## 故障排除

### 权限问题

如果遇到 "Permission denied" 错误：

```bash
chmod +x kubelet-wuhrai
sudo chmod +x /usr/local/bin/kubelet-wuhrai
```

### PATH 问题

如果命令找不到，确保 `/usr/local/bin` 在您的 PATH 中：

```bash
echo $PATH
export PATH=$PATH:/usr/local/bin  # 临时添加
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc  # 永久添加
```

### 依赖问题

这个二进制文件是静态编译的，不应该有依赖问题。如果遇到库缺失错误：

```bash
# 检查动态链接（应该显示"不是动态可执行文件"）
ldd /usr/local/bin/kubelet-wuhrai

# 检查系统信息
uname -a
cat /etc/redhat-release
```

### 系统兼容性

确保您的系统是 CentOS 7.9：

```bash
cat /etc/redhat-release
# 应该显示类似: CentOS Linux release 7.9.xxxx (Core)
```

## 卸载

```bash
sudo rm -f /usr/local/bin/kubelet-wuhrai
```

## 校验和

- **SHA256**: `6241b1ff6feef538b2dad115212da6fab72fa71221b8f6dfea5419b7318af8c1`
- **文件大小**: ~10MB (压缩包), ~30MB (二进制文件)

## 技术细节

- **编译器**: Go 1.24.4
- **构建标志**: `-ldflags="-w -s -static"`
- **CGO**: 禁用 (CGO_ENABLED=0)
- **目标平台**: linux/amd64
- **链接方式**: 静态链接
- **优化**: 去除调试信息和符号表

## 支持

如果遇到问题，请检查：

1. 系统版本是否为 CentOS 7.9
2. 架构是否为 x86_64
3. 是否有足够的磁盘空间
4. 是否有执行权限

更多信息请参考项目主文档。