# kubelet-wuhrai 开发指南

## 🔧 开发环境设置

### 前置要求

- Go 1.21+
- Git
- kubectl
- Docker (可选)

### 环境配置

```bash
# 设置Go代理 (使用阿里源)
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn

# 克隆项目
git clone https://github.com/st-lzh/kubelet-wuhrai.git
cd kubelet-wuhrai

# 安装依赖
go mod tidy
```

## 🚀 开发流程

### 1. 代码开发

```bash
# 创建功能分支
git checkout -b feature/your-feature-name

# 进行开发
# ... 编辑代码 ...

# 测试编译
go build -o kubelet-wuhrai ./cmd/

# 测试功能
./kubelet-wuhrai version
```

### 2. 提交代码

使用项目提供的上传脚本：

```bash
# 基本提交 (使用默认提交消息)
./upload-to-github.sh

# 自定义提交消息
./upload-to-github.sh "feat: 添加新功能

- 实现XXX功能
- 修复XXX问题
- 更新文档"
```

### 3. 脚本功能

`upload-to-github.sh` 脚本会自动：

- ✅ 检查Git仓库状态
- ✅ 清理不必要的测试文件
- ✅ 更新.gitignore文件
- ✅ 添加所有更改到Git
- ✅ 提交更改 (支持自定义消息)
- ✅ 推送到GitHub
- ✅ 提供详细的状态反馈

## 📁 项目结构

```
kubelet-wuhrai/
├── cmd/                    # 主程序入口
│   └── main.go
├── gollm/                  # AI模型提供商
│   ├── deepseek.go        # DeepSeek实现
│   ├── qwen.go            # 通义千问实现
│   ├── doubao.go          # 豆包实现
│   └── openai.go          # OpenAI兼容实现
├── pkg/                    # 核心包
│   ├── agent/             # 智能代理
│   ├── tools/             # 工具系统
│   ├── ui/                # 用户界面
│   └── mcp/               # MCP支持
├── docs/                   # 文档
├── Dockerfile             # Docker配置
├── docker-compose.yml     # Docker Compose配置
├── upload-to-github.sh    # 代码上传脚本
└── README.md              # 项目说明
```

## 🧪 测试

### 单元测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./gollm/

# 运行测试并显示覆盖率
go test -cover ./...
```

### 集成测试

```bash
# 设置测试环境变量
export DEEPSEEK_API_KEY="test_key"

# 编译测试
go build -o kubelet-wuhrai ./cmd/

# 功能测试
./kubelet-wuhrai version
./kubelet-wuhrai --help
```

### Docker测试

```bash
# 构建Docker镜像
docker build -t kubelet-wuhrai:dev .

# 测试运行
docker run --rm kubelet-wuhrai:dev version
```

## 📝 代码规范

### 提交消息格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型 (type):**
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

**示例:**
```
feat(gollm): 添加豆包AI模型支持

- 实现豆包API客户端
- 添加模型配置选项
- 更新文档说明

Closes #123
```

### Go代码规范

- 使用 `gofmt` 格式化代码
- 遵循 Go 官方代码规范
- 添加必要的注释 (中文)
- 错误处理要完整

## 🔄 发布流程

### 版本发布

1. **更新版本号**
   ```bash
   # 在 cmd/main.go 中更新版本
   const version = "v1.1.0"
   ```

2. **更新文档**
   ```bash
   # 更新 CHANGELOG.md
   # 更新 README.md
   ```

3. **提交和推送**
   ```bash
   ./upload-to-github.sh "release: v1.1.0

   - 新增功能列表
   - 修复问题列表
   - 重要变更说明"
   ```

4. **创建GitHub Release**
   - 在GitHub上创建新的Release
   - 添加发布说明
   - 上传编译好的二进制文件

## 🛠️ 常用开发命令

```bash
# 快速编译和测试
go build -o kubelet-wuhrai ./cmd/ && ./kubelet-wuhrai version

# 清理构建文件
go clean
rm -f kubelet-wuhrai

# 更新依赖
go mod tidy
go mod download

# 代码检查
go vet ./...
golint ./...

# 性能分析
go build -o kubelet-wuhrai ./cmd/
./kubelet-wuhrai -cpuprofile=cpu.prof "your query"
go tool pprof cpu.prof
```

## 📞 获取帮助

- **文档**: 查看 `docs/` 目录下的详细文档
- **Issues**: 在GitHub上提交问题
- **讨论**: 使用GitHub Discussions
- **代码审查**: 提交Pull Request

---

**提示**: 使用 `./upload-to-github.sh` 脚本可以大大简化开发流程，每次开发完成后只需要运行一个命令即可完成提交和推送！
