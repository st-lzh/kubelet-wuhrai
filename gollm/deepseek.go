// Copyright 2024 kubelet-wuhrai Contributors
//
// Licensed under the kubelet-wuhrai Custom License (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://github.com/st-lzh/kubelet-wuhrai/blob/main/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This software is based on kubectl-ai by Google Cloud Platform:
// https://github.com/GoogleCloudPlatform/kubectl-ai

package gollm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"k8s.io/klog/v2"
)

// 在包初始化时注册DeepSeek提供商工厂函数
func init() {
	if err := RegisterProvider("deepseek", newDeepSeekClientFactory); err != nil {
		klog.Fatalf("注册DeepSeek提供商失败: %v", err)
	}
}

// newDeepSeekClientFactory 是创建DeepSeek客户端的工厂函数
func newDeepSeekClientFactory(ctx context.Context, opts ClientOptions) (Client, error) {
	return NewDeepSeekClient(ctx, opts)
}

// DeepSeekClient 为DeepSeek模型实现了gollm.Client接口
type DeepSeekClient struct {
	client openai.Client
}

// 确保DeepSeekClient实现了Client接口
var _ Client = &DeepSeekClient{}

// NewDeepSeekClient 创建一个新的DeepSeek客户端
func NewDeepSeekClient(ctx context.Context, opts ClientOptions) (*DeepSeekClient, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY environment variable is required")
	}

	// DeepSeek API端点
	baseURL := "https://api.deepseek.com"
	if opts.URL != nil && opts.URL.Host != "" {
		baseURL = opts.URL.String()
	}

	// 为DeepSeek创建OpenAI兼容的客户端
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	return &DeepSeekClient{
		client: client,
	}, nil
}

// Close 释放客户端使用的资源
func (c *DeepSeekClient) Close() error {
	return nil
}

// ListModels 返回可用的DeepSeek模型列表
func (c *DeepSeekClient) ListModels(ctx context.Context) ([]string, error) {
	// DeepSeek可用模型
	return []string{
		"deepseek-chat",
		"deepseek-coder",
		"deepseek-reasoner",
	}, nil
}

// SetResponseSchema 约束LLM响应以匹配提供的模式
func (c *DeepSeekClient) SetResponseSchema(schema *Schema) error {
	// DeepSeek通过OpenAI兼容API支持结构化输出
	// 这将在聊天会话中处理
	return nil
}

// StartChat 开始一个新的聊天会话
func (c *DeepSeekClient) StartChat(systemPrompt, model string) Chat {
	// 获取此次聊天使用的模型
	selectedModel := getDeepSeekModel(model)

	klog.V(1).Infof("使用模型启动新的DeepSeek聊天会话: %s", selectedModel)

	// 如果提供了系统提示，则用它初始化历史记录
	history := []openai.ChatCompletionMessageParamUnion{}
	if systemPrompt != "" {
		history = append(history, openai.SystemMessage(systemPrompt))
	}

	return &deepSeekChatSession{
		client:  c.client,
		history: history,
		model:   selectedModel,
	}
}

// getDeepSeekModel 返回适当的DeepSeek模型名称
func getDeepSeekModel(model string) string {
	if model == "" {
		return "deepseek-chat" // 默认模型
	}

	// 将常见模型名称映射到DeepSeek模型
	switch strings.ToLower(model) {
	case "deepseek-chat", "chat":
		return "deepseek-chat"
	case "deepseek-coder", "coder":
		return "deepseek-coder"
	case "deepseek-reasoner", "reasoner":
		return "deepseek-reasoner"
	default:
		// 如果已经是有效的DeepSeek模型名称，则使用它
		if strings.HasPrefix(model, "deepseek-") {
			return model
		}
		// 默认回退
		return "deepseek-chat"
	}
}

// GenerateCompletion 向DeepSeek API发送完成请求
func (c *DeepSeekClient) GenerateCompletion(ctx context.Context, req *CompletionRequest) (CompletionResponse, error) {
	klog.Infof("使用模型调用DeepSeek GenerateCompletion: %s", req.Model)
	klog.V(1).Infof("提示词:\n%s", req.Prompt)

	selectedModel := getDeepSeekModel(req.Model)

	// 使用聊天完成API
	completion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModel(selectedModel),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(req.Prompt),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("生成DeepSeek完成失败: %w", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("DeepSeek无响应")
	}

	response := completion.Choices[0].Message.Content
	klog.V(1).Infof("DeepSeek完成响应: %s", response)

	return &DeepSeekCompletionResponse{
		response: response,
		usage:    completion.Usage,
	}, nil
}

// DeepSeekCompletionResponse 为DeepSeek实现CompletionResponse接口
type DeepSeekCompletionResponse struct {
	response string
	usage    openai.CompletionUsage
}

func (r *DeepSeekCompletionResponse) Response() string {
	return r.response
}

func (r *DeepSeekCompletionResponse) UsageMetadata() any {
	return r.usage
}

// --- 聊天会话实现 ---

type deepSeekChatSession struct {
	client              openai.Client
	history             []openai.ChatCompletionMessageParamUnion
	model               string
	functionDefinitions []*FunctionDefinition            // 以gollm格式存储
	tools               []openai.ChatCompletionToolParam // 以OpenAI格式存储
}

// 确保deepSeekChatSession实现了Chat接口
var _ Chat = (*deepSeekChatSession)(nil)

// SetFunctionDefinitions 为聊天会话配置可用的函数
func (cs *deepSeekChatSession) SetFunctionDefinitions(functionDefinitions []*FunctionDefinition) error {
	klog.V(1).InfoS("调用deepSeekChatSession.SetFunctionDefinitions", "count", len(functionDefinitions))

	cs.functionDefinitions = functionDefinitions
	cs.tools = nil // 重置工具

	// 将gollm函数定义转换为OpenAI格式
	for _, funcDef := range functionDefinitions {
		if funcDef == nil {
			continue
		}

		// 将模式转换为OpenAI格式
		var params openai.FunctionParameters
		if funcDef.Parameters != nil {
			schemaBytes, err := cs.convertSchemaToBytes(funcDef.Parameters, funcDef.Name)
			if err != nil {
				return fmt.Errorf("转换函数%s的模式失败: %w", funcDef.Name, err)
			}

			if err := json.Unmarshal(schemaBytes, &params); err != nil {
				return fmt.Errorf("解析函数%s的参数失败: %w", funcDef.Name, err)
			}
		}

		tool := openai.ChatCompletionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        funcDef.Name,
				Description: openai.String(funcDef.Description),
				Parameters:  params,
			},
		}

		cs.tools = append(cs.tools, tool)
	}

	klog.V(1).InfoS("deepSeekChatSession函数定义已设置", "tools_count", len(cs.tools))
	return nil
}

// convertSchemaToBytes 使用OpenAI特定的编组将验证的模式转换为JSON字节
func (cs *deepSeekChatSession) convertSchemaToBytes(schema *Schema, functionName string) ([]byte, error) {
	// 使用OpenAI特定的编组行为包装模式
	openAIWrapper := openAISchema{Schema: schema}

	bytes, err := json.Marshal(openAIWrapper)
	if err != nil {
		return nil, fmt.Errorf("转换模式失败: %w", err)
	}

	klog.Infof("函数%s的DeepSeek模式: %s", functionName, string(bytes))

	return bytes, nil
}

// Send 发送用户消息，追加到历史记录，并获取LLM响应
func (cs *deepSeekChatSession) Send(ctx context.Context, contents ...any) (ChatResponse, error) {
	klog.V(1).InfoS("调用deepSeekChatSession.Send", "model", cs.model, "history_len", len(cs.history))

	// 处理并将消息追加到历史记录
	if err := cs.addContentsToHistory(contents); err != nil {
		return nil, err
	}

	// 准备并发送API请求
	chatReq := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(cs.model),
		Messages: cs.history,
	}
	if len(cs.tools) > 0 {
		chatReq.Tools = cs.tools
	}

	// 调用DeepSeek API
	klog.V(1).InfoS("向DeepSeek聊天API发送请求", "model", cs.model, "messages", len(chatReq.Messages), "tools", len(chatReq.Tools))
	completion, err := cs.client.Chat.Completions.New(ctx, chatReq)
	if err != nil {
		klog.Errorf("DeepSeek聊天完成API错误: %v", err)
		return nil, fmt.Errorf("DeepSeek聊天完成失败: %w", err)
	}
	klog.V(1).InfoS("从DeepSeek聊天API收到响应", "id", completion.ID, "choices", len(completion.Choices))

	// 处理响应
	if len(completion.Choices) == 0 {
		klog.Warning("从DeepSeek收到无选择的响应")
		return nil, errors.New("从DeepSeek收到空响应(无选择)")
	}

	// 将助手的响应添加到历史记录
	assistantMessage := completion.Choices[0].Message
	cs.history = append(cs.history, openai.AssistantMessage(assistantMessage.Content))

	// 如果存在工具调用则处理
	if len(assistantMessage.ToolCalls) > 0 {
		// 将工具调用添加到历史记录
		// 用包含工具调用的消息替换最后一条消息
		cs.history[len(cs.history)-1] = assistantMessage.ToParam()
	}

	// 创建并返回响应
	response := &deepSeekChatResponse{
		completion: completion,
	}

	klog.V(1).InfoS("deepSeekChatSession.Send完成", "response_id", completion.ID)
	return response, nil
}

// SendStreaming 是Send的流式版本
func (cs *deepSeekChatSession) SendStreaming(ctx context.Context, contents ...any) (ChatResponseIterator, error) {
	// 目前，我们将其实现为非流式调用
	// TODO: 实现实际的流式支持
	response, err := cs.Send(ctx, contents...)
	if err != nil {
		return nil, err
	}

	// 返回一个产生单个响应的迭代器
	return func(yield func(ChatResponse, error) bool) {
		yield(response, nil)
	}, nil
}

// addContentsToHistory 处理输入内容并将其添加到聊天历史记录
func (cs *deepSeekChatSession) addContentsToHistory(contents []any) error {
	for _, content := range contents {
		switch v := content.(type) {
		case string:
			// 简单文本消息
			cs.history = append(cs.history, openai.UserMessage(v))
		case FunctionCallResult:
			// 函数调用结果
			resultJSON, err := json.Marshal(v.Result)
			if err != nil {
				klog.Errorf("序列化函数调用结果失败: %v", err)
				return fmt.Errorf("序列化函数调用结果%q失败: %w", v.Name, err)
			}
			cs.history = append(cs.history, openai.ToolMessage(string(resultJSON), v.ID))
		default:
			// 尝试转换为字符串
			cs.history = append(cs.history, openai.UserMessage(fmt.Sprintf("%v", v)))
		}
	}
	return nil
}

// --- 响应实现 ---

type deepSeekChatResponse struct {
	completion *openai.ChatCompletion
}

// 确保deepSeekChatResponse实现了ChatResponse接口
var _ ChatResponse = (*deepSeekChatResponse)(nil)

func (r *deepSeekChatResponse) UsageMetadata() any {
	return r.completion.Usage
}

func (r *deepSeekChatResponse) Candidates() []Candidate {
	var candidates []Candidate
	for _, choice := range r.completion.Choices {
		candidates = append(candidates, &deepSeekCandidate{choice: choice})
	}
	return candidates
}

// --- 候选实现 ---

type deepSeekCandidate struct {
	choice openai.ChatCompletionChoice
}

// 确保deepSeekCandidate实现了Candidate接口
var _ Candidate = (*deepSeekCandidate)(nil)

func (c *deepSeekCandidate) String() string {
	return c.choice.Message.Content
}

func (c *deepSeekCandidate) Parts() []Part {
	var parts []Part

	// 如果内容存在则添加文本部分
	if c.choice.Message.Content != "" {
		parts = append(parts, &deepSeekTextPart{text: c.choice.Message.Content})
	}

	// 如果存在函数调用部分则添加
	if len(c.choice.Message.ToolCalls) > 0 {
		var functionCalls []FunctionCall
		for _, toolCall := range c.choice.Message.ToolCalls {
			if toolCall.Function.Name != "" {
				// 解析参数
				var args map[string]any
				if toolCall.Function.Arguments != "" {
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
						klog.Errorf("解析函数参数失败: %v", err)
						args = make(map[string]any)
					}
				}

				functionCalls = append(functionCalls, FunctionCall{
					ID:        toolCall.ID,
					Name:      toolCall.Function.Name,
					Arguments: args,
				})
			}
		}
		if len(functionCalls) > 0 {
			parts = append(parts, &deepSeekFunctionCallPart{functionCalls: functionCalls})
		}
	}

	return parts
}

// --- 部分实现 ---

type deepSeekTextPart struct {
	text string
}

// 确保deepSeekTextPart实现了Part接口
var _ Part = (*deepSeekTextPart)(nil)

func (p *deepSeekTextPart) AsText() (string, bool) {
	return p.text, true
}

func (p *deepSeekTextPart) AsFunctionCalls() ([]FunctionCall, bool) {
	return nil, false
}

type deepSeekFunctionCallPart struct {
	functionCalls []FunctionCall
}

// 确保deepSeekFunctionCallPart实现了Part接口
var _ Part = (*deepSeekFunctionCallPart)(nil)

func (p *deepSeekFunctionCallPart) AsText() (string, bool) {
	return "", false
}

func (p *deepSeekFunctionCallPart) AsFunctionCalls() ([]FunctionCall, bool) {
	return p.functionCalls, true
}

// IsRetryableError 判断错误是否可重试
func (d *deepSeekChatSession) IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否是网络错误或临时错误
	errStr := err.Error()
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "504")
}
