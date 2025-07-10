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

// 在包初始化时注册Qwen提供商工厂函数
func init() {
	if err := RegisterProvider("qwen", newQwenClientFactory); err != nil {
		klog.Fatalf("注册Qwen提供商失败: %v", err)
	}
	// 同时使用dashscope别名注册
	if err := RegisterProvider("dashscope", newQwenClientFactory); err != nil {
		klog.Fatalf("注册DashScope提供商失败: %v", err)
	}
}

// newQwenClientFactory 是创建Qwen客户端的工厂函数
func newQwenClientFactory(ctx context.Context, opts ClientOptions) (Client, error) {
	return NewQwenClient(ctx, opts)
}

// QwenClient 通过DashScope API为Qwen模型实现了gollm.Client接口
type QwenClient struct {
	client openai.Client
}

// 确保QwenClient实现了Client接口
var _ Client = &QwenClient{}

// NewQwenClient 使用DashScope API创建一个新的Qwen客户端
func NewQwenClient(ctx context.Context, opts ClientOptions) (*QwenClient, error) {
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		// 尝试替代的环境变量名称
		apiKey = os.Getenv("QWEN_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("需要DASHSCOPE_API_KEY或QWEN_API_KEY环境变量")
		}
	}

	// DashScope API端点
	baseURL := "https://dashscope.aliyuncs.com/compatible-mode/v1"
	if opts.URL != nil && opts.URL.Host != "" {
		baseURL = opts.URL.String()
	}

	// 通过DashScope为Qwen创建OpenAI兼容的客户端
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	return &QwenClient{
		client: client,
	}, nil
}

// Close 释放客户端使用的资源
func (c *QwenClient) Close() error {
	return nil
}

// ListModels 返回可用的Qwen模型列表
func (c *QwenClient) ListModels(ctx context.Context) ([]string, error) {
	// 通过DashScope提供的Qwen可用模型
	return []string{
		"qwen-plus",
		"qwen-turbo",
		"qwen-max",
		"qwen-max-longcontext",
		"qwen2.5-72b-instruct",
		"qwen2.5-32b-instruct",
		"qwen2.5-14b-instruct",
		"qwen2.5-7b-instruct",
		"qwen2.5-3b-instruct",
		"qwen2.5-1.5b-instruct",
		"qwen2.5-0.5b-instruct",
		"qwen2.5-coder-32b-instruct",
		"qwen2.5-coder-14b-instruct",
		"qwen2.5-coder-7b-instruct",
		"qwen2.5-coder-3b-instruct",
		"qwen2.5-coder-1.5b-instruct",
		"qwen2.5-math-72b-instruct",
		"qwen2.5-math-7b-instruct",
		"qwen2.5-math-1.5b-instruct",
	}, nil
}

// SetResponseSchema 约束LLM响应以匹配提供的模式
func (c *QwenClient) SetResponseSchema(schema *Schema) error {
	// Qwen通过OpenAI兼容API支持结构化输出
	// 这将在聊天会话中处理
	return nil
}

// StartChat 开始一个新的聊天会话
func (c *QwenClient) StartChat(systemPrompt, model string) Chat {
	// 获取此次聊天使用的模型
	selectedModel := getQwenModel(model)

	klog.V(1).Infof("使用模型启动新的Qwen聊天会话: %s", selectedModel)

	// 如果提供了系统提示，则用它初始化历史记录
	history := []openai.ChatCompletionMessageParamUnion{}
	if systemPrompt != "" {
		history = append(history, openai.SystemMessage(systemPrompt))
	}

	return &qwenChatSession{
		client:  c.client,
		history: history,
		model:   selectedModel,
	}
}

// getQwenModel 返回适当的Qwen模型名称
func getQwenModel(model string) string {
	if model == "" {
		return "qwen-plus" // 默认模型
	}

	// 将常见模型名称映射到Qwen模型
	switch strings.ToLower(model) {
	case "qwen-plus", "plus":
		return "qwen-plus"
	case "qwen-turbo", "turbo":
		return "qwen-turbo"
	case "qwen-max", "max":
		return "qwen-max"
	case "qwen-max-longcontext", "max-longcontext", "longcontext":
		return "qwen-max-longcontext"
	case "qwen2.5-72b-instruct", "qwen2.5-72b", "72b":
		return "qwen2.5-72b-instruct"
	case "qwen2.5-32b-instruct", "qwen2.5-32b", "32b":
		return "qwen2.5-32b-instruct"
	case "qwen2.5-14b-instruct", "qwen2.5-14b", "14b":
		return "qwen2.5-14b-instruct"
	case "qwen2.5-7b-instruct", "qwen2.5-7b", "7b":
		return "qwen2.5-7b-instruct"
	case "qwen2.5-3b-instruct", "qwen2.5-3b", "3b":
		return "qwen2.5-3b-instruct"
	case "qwen2.5-1.5b-instruct", "qwen2.5-1.5b", "1.5b":
		return "qwen2.5-1.5b-instruct"
	case "qwen2.5-0.5b-instruct", "qwen2.5-0.5b", "0.5b":
		return "qwen2.5-0.5b-instruct"
	case "qwen2.5-coder-32b-instruct", "coder-32b", "coder32b":
		return "qwen2.5-coder-32b-instruct"
	case "qwen2.5-coder-14b-instruct", "coder-14b", "coder14b":
		return "qwen2.5-coder-14b-instruct"
	case "qwen2.5-coder-7b-instruct", "coder-7b", "coder7b":
		return "qwen2.5-coder-7b-instruct"
	case "qwen2.5-coder-3b-instruct", "coder-3b", "coder3b":
		return "qwen2.5-coder-3b-instruct"
	case "qwen2.5-coder-1.5b-instruct", "coder-1.5b", "coder1.5b":
		return "qwen2.5-coder-1.5b-instruct"
	case "qwen2.5-math-72b-instruct", "math-72b", "math72b":
		return "qwen2.5-math-72b-instruct"
	case "qwen2.5-math-7b-instruct", "math-7b", "math7b":
		return "qwen2.5-math-7b-instruct"
	case "qwen2.5-math-1.5b-instruct", "math-1.5b", "math1.5b":
		return "qwen2.5-math-1.5b-instruct"
	default:
		// 如果已经是有效的Qwen模型名称，则使用它
		if strings.HasPrefix(model, "qwen") {
			return model
		}
		// 默认回退
		return "qwen-plus"
	}
}

// GenerateCompletion 向Qwen API发送完成请求
func (c *QwenClient) GenerateCompletion(ctx context.Context, req *CompletionRequest) (CompletionResponse, error) {
	klog.Infof("使用模型调用Qwen GenerateCompletion: %s", req.Model)
	klog.V(1).Infof("提示词:\n%s", req.Prompt)

	selectedModel := getQwenModel(req.Model)

	// 使用聊天完成API
	completion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModel(selectedModel),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(req.Prompt),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("生成Qwen完成失败: %w", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("Qwen无响应")
	}

	response := completion.Choices[0].Message.Content
	klog.V(1).Infof("Qwen完成响应: %s", response)

	return &QwenCompletionResponse{
		response: response,
		usage:    completion.Usage,
	}, nil
}

// QwenCompletionResponse 为Qwen实现CompletionResponse接口
type QwenCompletionResponse struct {
	response string
	usage    openai.CompletionUsage
}

func (r *QwenCompletionResponse) Response() string {
	return r.response
}

func (r *QwenCompletionResponse) UsageMetadata() any {
	return r.usage
}

// --- 聊天会话实现 ---

type qwenChatSession struct {
	client              openai.Client
	history             []openai.ChatCompletionMessageParamUnion
	model               string
	functionDefinitions []*FunctionDefinition            // 以gollm格式存储
	tools               []openai.ChatCompletionToolParam // 以OpenAI格式存储
}

// 确保qwenChatSession实现了Chat接口
var _ Chat = (*qwenChatSession)(nil)

// SetFunctionDefinitions 为聊天会话配置可用的函数
func (cs *qwenChatSession) SetFunctionDefinitions(functionDefinitions []*FunctionDefinition) error {
	klog.V(1).InfoS("调用qwenChatSession.SetFunctionDefinitions", "count", len(functionDefinitions))

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

	klog.V(1).InfoS("qwenChatSession函数定义已设置", "tools_count", len(cs.tools))
	return nil
}

// convertSchemaToBytes 使用OpenAI特定的编组将验证的模式转换为JSON字节
func (cs *qwenChatSession) convertSchemaToBytes(schema *Schema, functionName string) ([]byte, error) {
	// 使用OpenAI特定的编组行为包装模式
	openAIWrapper := openAISchema{Schema: schema}

	bytes, err := json.Marshal(openAIWrapper)
	if err != nil {
		return nil, fmt.Errorf("转换模式失败: %w", err)
	}

	klog.Infof("函数%s的Qwen模式: %s", functionName, string(bytes))

	return bytes, nil
}

// Send sends the user message(s), appends to history, and gets the LLM response.
func (cs *qwenChatSession) Send(ctx context.Context, contents ...any) (ChatResponse, error) {
	klog.V(1).InfoS("qwenChatSession.Send called", "model", cs.model, "history_len", len(cs.history))

	// Process and append messages to history
	if err := cs.addContentsToHistory(contents); err != nil {
		return nil, err
	}

	// Prepare and send API request
	chatReq := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(cs.model),
		Messages: cs.history,
	}
	if len(cs.tools) > 0 {
		chatReq.Tools = cs.tools
	}

	// Call the Qwen API via DashScope
	klog.V(1).InfoS("Sending request to Qwen Chat API", "model", cs.model, "messages", len(chatReq.Messages), "tools", len(chatReq.Tools))
	completion, err := cs.client.Chat.Completions.New(ctx, chatReq)
	if err != nil {
		klog.Errorf("Qwen ChatCompletion API error: %v", err)
		return nil, fmt.Errorf("Qwen chat completion failed: %w", err)
	}
	klog.V(1).InfoS("Received response from Qwen Chat API", "id", completion.ID, "choices", len(completion.Choices))

	// Process the response
	if len(completion.Choices) == 0 {
		klog.Warning("Received response with no choices from Qwen")
		return nil, errors.New("received empty response from Qwen (no choices)")
	}

	// Add assistant's response to history
	assistantMessage := completion.Choices[0].Message
	cs.history = append(cs.history, openai.AssistantMessage(assistantMessage.Content))

	// Handle tool calls if present
	if len(assistantMessage.ToolCalls) > 0 {
		// Replace the last message with the one containing tool calls
		cs.history[len(cs.history)-1] = assistantMessage.ToParam()
	}

	// Create and return response
	response := &qwenChatResponse{
		completion: completion,
	}

	klog.V(1).InfoS("qwenChatSession.Send completed", "response_id", completion.ID)
	return response, nil
}

// SendStreaming is the streaming version of Send.
func (cs *qwenChatSession) SendStreaming(ctx context.Context, contents ...any) (ChatResponseIterator, error) {
	// For now, we'll implement this as a non-streaming call
	// TODO: Implement actual streaming support
	response, err := cs.Send(ctx, contents...)
	if err != nil {
		return nil, err
	}

	// Return an iterator that yields the single response
	return func(yield func(ChatResponse, error) bool) {
		yield(response, nil)
	}, nil
}

// addContentsToHistory processes the input contents and adds them to the chat history.
func (cs *qwenChatSession) addContentsToHistory(contents []any) error {
	for _, content := range contents {
		switch v := content.(type) {
		case string:
			// Simple text message
			cs.history = append(cs.history, openai.UserMessage(v))
		case FunctionCallResult:
			// Function call result
			resultJSON, err := json.Marshal(v.Result)
			if err != nil {
				klog.Errorf("序列化函数调用结果失败: %v", err)
				return fmt.Errorf("序列化函数调用结果%q失败: %w", v.Name, err)
			}
			cs.history = append(cs.history, openai.ToolMessage(string(resultJSON), v.ID))
		default:
			// Try to convert to string
			cs.history = append(cs.history, openai.UserMessage(fmt.Sprintf("%v", v)))
		}
	}
	return nil
}

// --- Response Implementation ---

type qwenChatResponse struct {
	completion *openai.ChatCompletion
}

// Ensure qwenChatResponse implements ChatResponse.
var _ ChatResponse = (*qwenChatResponse)(nil)

func (r *qwenChatResponse) UsageMetadata() any {
	return r.completion.Usage
}

func (r *qwenChatResponse) Candidates() []Candidate {
	var candidates []Candidate
	for _, choice := range r.completion.Choices {
		candidates = append(candidates, &qwenCandidate{choice: choice})
	}
	return candidates
}

// --- Candidate Implementation ---

type qwenCandidate struct {
	choice openai.ChatCompletionChoice
}

// Ensure qwenCandidate implements Candidate.
var _ Candidate = (*qwenCandidate)(nil)

func (c *qwenCandidate) String() string {
	return c.choice.Message.Content
}

func (c *qwenCandidate) Parts() []Part {
	var parts []Part

	// Add text part if content exists
	if c.choice.Message.Content != "" {
		parts = append(parts, &qwenTextPart{text: c.choice.Message.Content})
	}

	// Add function call parts if they exist
	if len(c.choice.Message.ToolCalls) > 0 {
		var functionCalls []FunctionCall
		for _, toolCall := range c.choice.Message.ToolCalls {
			if toolCall.Function.Name != "" {
				// Parse arguments
				var args map[string]any
				if toolCall.Function.Arguments != "" {
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
						klog.Errorf("Failed to parse function arguments: %v", err)
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
			parts = append(parts, &qwenFunctionCallPart{functionCalls: functionCalls})
		}
	}

	return parts
}

// --- Part Implementations ---

type qwenTextPart struct {
	text string
}

// Ensure qwenTextPart implements Part.
var _ Part = (*qwenTextPart)(nil)

func (p *qwenTextPart) AsText() (string, bool) {
	return p.text, true
}

func (p *qwenTextPart) AsFunctionCalls() ([]FunctionCall, bool) {
	return nil, false
}

type qwenFunctionCallPart struct {
	functionCalls []FunctionCall
}

// Ensure qwenFunctionCallPart implements Part.
var _ Part = (*qwenFunctionCallPart)(nil)

func (p *qwenFunctionCallPart) AsText() (string, bool) {
	return "", false
}

func (p *qwenFunctionCallPart) AsFunctionCalls() ([]FunctionCall, bool) {
	return p.functionCalls, true
}

// IsRetryableError 判断错误是否可重试
func (q *qwenChatSession) IsRetryableError(err error) bool {
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
