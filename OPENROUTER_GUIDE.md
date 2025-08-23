# OpenRouter集成指南

## 🌟 OpenRouter简介

OpenRouter是一个模型聚合服务平台，提供统一API访问多个AI模型提供商，包括：

- **Anthropic**: Claude-3.5 Sonnet, Claude-3 Opus
- **OpenAI**: GPT-4o, GPT-4o-mini, GPT-3.5-turbo
- **Google**: Gemini 2.5 Flash, Gemini 2.5 Pro
- **Meta**: Llama-3.3-70B, Llama-3.2-90B
- **DeepSeek**: DeepSeek-R1
- **Cohere**: Command-R-Plus
- **Mistral**: Mistral-Large
- **其他数百个模型**

## 🚀 便捷方案：一个API访问所有模型

### 优势
1. **统一接口**：一个API密钥访问所有模型
2. **成本优化**：自动选择最具成本效益的模型
3. **故障切换**：模型不可用时自动切换备用模型
4. **简化计费**：统一账单管理
5. **实时可用性**：实时监控模型状态

### 快速开始

#### 1. 获取API密钥
1. 访问 [OpenRouter](https://openrouter.ai)
2. 注册账户并获取API密钥
3. 可选择按需付费或订阅计划

#### 2. 配置环境

```bash
# 设置OpenRouter API密钥
export OPENROUTER_API_KEY="your-openrouter-api-key"

# 设置默认客户端
export LLM_CLIENT="openrouter"

# 可选：设置默认模型
export OPENROUTER_MODEL="anthropic/claude-3.5-sonnet"
```

#### 3. 基础使用

```bash
# 使用默认模型
kubelet-wuhrai --llm-provider openrouter "获取所有Pod状态"

# 指定特定模型
kubelet-wuhrai --llm-provider openrouter --model "openai/gpt-4o" "分析集群性能"

# 使用成本较低的模型
kubelet-wuhrai --llm-provider openrouter --model "openai/gpt-4o-mini" "获取简单信息"

# 使用最新的Claude模型
kubelet-wuhrai --llm-provider openrouter --model "anthropic/claude-3.5-sonnet" "复杂分析任务"
```

## 📋 推荐模型选择策略

### 按场景选择模型

```bash
# 1. 简单查询（成本优化）
simple_query() {
    kubelet-wuhrai --llm-provider openrouter --model "openai/gpt-4o-mini" --quiet "$1"
}

# 2. 复杂分析（性能优先）
complex_analysis() {
    kubelet-wuhrai --llm-provider openrouter --model "anthropic/claude-3.5-sonnet" --quiet "$1"
}

# 3. 代码生成（专业模型）
code_generation() {
    kubelet-wuhrai --llm-provider openrouter --model "deepseek/deepseek-r1" --quiet "$1"
}

# 4. 快速响应（速度优先）
quick_response() {
    kubelet-wuhrai --llm-provider openrouter --model "google/gemini-2.5-flash" --quiet "$1"
}
```

### 使用示例

```bash
# 日常运维查询
simple_query "获取所有Pod数量"
simple_query "检查服务状态"

# 故障诊断分析
complex_analysis "分析为什么Pod一直重启"
complex_analysis "检查集群安全配置问题"

# YAML生成
code_generation "生成一个nginx deployment的yaml配置"
code_generation "创建一个包含资源限制的Pod定义"

# 快速状态检查
quick_response "集群是否正常运行"
quick_response "有多少个节点在运行"
```

## 💰 成本优化策略

### 模型成本对比（每1M tokens）

| 模型 | 输入成本 | 输出成本 | 适用场景 |
|------|----------|----------|----------|
| GPT-4o-mini | $0.15 | $0.60 | 简单查询 |
| Gemini 2.5 Flash | $0.30 | $1.20 | 快速响应 |
| GPT-4o | $2.50 | $10.00 | 复杂分析 |
| Claude-3.5-Sonnet | $3.00 | $15.00 | 专业分析 |

### 智能成本控制

```bash
#!/bin/bash
# 智能模型选择脚本

choose_model() {
    local query="$1"
    local query_length=${#query}
    
    if [ $query_length -lt 50 ]; then
        echo "openai/gpt-4o-mini"  # 短查询用便宜模型
    elif [[ $query == *"分析"* ]] || [[ $query == *"诊断"* ]]; then
        echo "anthropic/claude-3.5-sonnet"  # 分析任务用高级模型
    else
        echo "google/gemini-2.5-flash"  # 默认用中等模型
    fi
}

# 使用示例
MODEL=$(choose_model "分析集群性能瓶颈")
kubelet-wuhrai --llm-provider openrouter --model "$MODEL" --quiet "分析集群性能瓶颈"
```

## 🔧 高级配置

### 多模型策略配置

```yaml
# ~/.config/kubelet-wuhrai/openrouter.yaml
models:
  # 快速查询模型
  quick:
    - "openai/gpt-4o-mini"
    - "google/gemini-2.5-flash"
  
  # 复杂分析模型
  analysis:
    - "anthropic/claude-3.5-sonnet"
    - "openai/gpt-4o"
  
  # 代码生成模型
  coding:
    - "deepseek/deepseek-r1"
    - "meta-llama/llama-3.3-70b-instruct"

# 故障转移配置
fallback:
  enabled: true
  retry_attempts: 3
  timeout: 30s
```

### 批量处理脚本

```bash
#!/bin/bash
# 批量Kubernetes集群检查

clusters=("prod" "staging" "dev")
queries=(
    "获取Pod数量"
    "检查节点状态"  
    "查看失败的Pod"
    "检查资源使用率"
)

for cluster in "${clusters[@]}"; do
    echo "=== 检查 $cluster 集群 ==="
    
    for query in "${queries[@]}"; do
        echo "查询: $query"
        
        # 根据查询类型选择模型
        if [[ $query == *"资源"* ]]; then
            model="anthropic/claude-3.5-sonnet"
        else
            model="openai/gpt-4o-mini"
        fi
        
        kubelet-wuhrai \
            --kubeconfig ~/.kube/${cluster}-config \
            --llm-provider openrouter \
            --model "$model" \
            --quiet "$query" \
            >> "${cluster}-report.txt"
        
        sleep 1  # 避免API限流
    done
done
```

## 📊 监控和分析

### 使用统计

```bash
# 查看OpenRouter使用统计
curl -H "Authorization: Bearer $OPENROUTER_API_KEY" \
     "https://openrouter.ai/api/v1/auth/key" | jq '.data.usage'

# 查看模型性能统计  
curl -H "Authorization: Bearer $OPENROUTER_API_KEY" \
     "https://openrouter.ai/api/v1/models" | jq '.data[] | {id: .id, context_length: .context_length, pricing: .pricing}'
```

### 自动化报告

```bash
#!/bin/bash
# 生成OpenRouter使用报告

generate_report() {
    local date=$(date +%Y-%m-%d)
    local report_file="openrouter-report-${date}.txt"
    
    echo "OpenRouter 使用报告 - $date" > $report_file
    echo "================================" >> $report_file
    
    # 今日查询次数
    grep "$(date +%Y-%m-%d)" ~/.kubelet-wuhrai.log | wc -l >> $report_file
    
    # 使用的模型统计
    grep "model:" ~/.kubelet-wuhrai.log | sort | uniq -c >> $report_file
    
    # 成本估算
    echo "预估成本: 查看 OpenRouter 控制台" >> $report_file
}

# 设置定期报告
# 添加到 crontab: 0 23 * * * /path/to/generate_report.sh
```

## 🛠️ 故障排除

### 常见问题

1. **API密钥错误**
   ```bash
   # 验证API密钥
   curl -H "Authorization: Bearer $OPENROUTER_API_KEY" \
        "https://openrouter.ai/api/v1/auth/key"
   ```

2. **模型不可用**
   ```bash
   # 检查模型状态
   curl -s "https://openrouter.ai/api/v1/models" | \
        jq '.data[] | select(.id=="anthropic/claude-3.5-sonnet")'
   ```

3. **配额限制**
   ```bash
   # 检查账户余额
   curl -H "Authorization: Bearer $OPENROUTER_API_KEY" \
        "https://openrouter.ai/api/v1/auth/key" | jq '.data.usage'
   ```

### 调试模式

```bash
# 启用详细日志
kubelet-wuhrai -v 2 --llm-provider openrouter --model "openai/gpt-4o" "测试查询"

# 检查网络连接
curl -I "https://openrouter.ai/api/v1/models"

# 验证SSL证书
openssl s_client -connect openrouter.ai:443 -servername openrouter.ai
```

## 🎯 最佳实践

1. **成本控制**: 简单查询使用便宜模型，复杂分析使用高级模型
2. **性能优化**: 批量请求时添加适当延迟避免限流
3. **错误处理**: 配置故障转移模型以提高可靠性
4. **监控**: 定期检查使用量和成本
5. **安全**: 妥善保管API密钥，使用环境变量而非硬编码

OpenRouter为kubelet-wuhrai提供了访问所有主流AI模型的便捷方案，是平衡成本、性能和功能的理想选择。