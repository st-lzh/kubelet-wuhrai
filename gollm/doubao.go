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

// 在包初始化时注册豆包提供商工厂函数
func init() {
	if err := RegisterProvider("doubao", newDoubaoClientFactory); err != nil {
		klog.Fatalf("注册豆包提供商失败: %v", err)
	}
	// 同时使用volces别名注册
	if err := RegisterProvider("volces", newDoubaoClientFactory); err != nil {
		klog.Fatalf("注册Volces提供商失败: %v", err)
	}
}

// newDoubaoClientFactory 是创建豆包客户端的工厂函数
func newDoubaoClientFactory(ctx context.Context, opts ClientOptions) (Client, error) {
	return NewDoubaoClient(ctx, opts)
}

// DoubaoClient 通过Volces API为豆包模型实现了gollm.Client接口
type DoubaoClient struct {
	client openai.Client
}

// 确保DoubaoClient实现了Client接口
var _ Client = &DoubaoClient{}

// NewDoubaoClient 使用Volces API创建一个新的豆包客户端
func NewDoubaoClient(ctx context.Context, opts ClientOptions) (*DoubaoClient, error) {
	apiKey := os.Getenv("VOLCES_API_KEY")
	if apiKey == "" {
		// 尝试替代的环境变量名称
		apiKey = os.Getenv("DOUBAO_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("需要VOLCES_API_KEY或DOUBAO_API_KEY环境变量")
		}
	}

	// Volces API端点
	baseURL := "https://ark.cn-beijing.volces.com/api/v3"
	if opts.URL != nil && opts.URL.Host != "" {
		baseURL = opts.URL.String()
	}

	// 通过Volces为豆包创建OpenAI兼容的客户端
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	return &DoubaoClient{
		client: client,
	}, nil
}

// Close 释放客户端使用的资源
func (c *DoubaoClient) Close() error {
	return nil
}

// ListModels 返回可用的豆包模型列表
func (c *DoubaoClient) ListModels(ctx context.Context) ([]string, error) {
	// 通过Volces提供的豆包可用模型
	return []string{
		"doubao-pro-4k",
		"doubao-pro-32k",
		"doubao-pro-128k",
		"doubao-lite-4k",
		"doubao-lite-32k",
		"doubao-lite-128k",
		"doubao-pro-vision",
		"doubao-pro-search",
		"doubao-character-4k",
		"doubao-character-32k",
		"doubao-character-128k",
	}, nil
}

// SetResponseSchema constrains LLM responses to match the provided schema.
func (c *DoubaoClient) SetResponseSchema(schema *Schema) error {
	// Doubao supports structured output through OpenAI-compatible API
	// This will be handled in the chat session
	return nil
}

// StartChat starts a new chat session.
func (c *DoubaoClient) StartChat(systemPrompt, model string) Chat {
	// Get the model to use for this chat
	selectedModel := getDoubaoModel(model)

	klog.V(1).Infof("Starting new Doubao chat session with model: %s", selectedModel)

	// Initialize history with system prompt if provided
	history := []openai.ChatCompletionMessageParamUnion{}
	if systemPrompt != "" {
		history = append(history, openai.SystemMessage(systemPrompt))
	}

	return &doubaoChat{
		client:  c.client,
		history: history,
		model:   selectedModel,
	}
}

// getDoubaoModel returns the appropriate Doubao model name.
func getDoubaoModel(model string) string {
	if model == "" {
		return "doubao-pro-4k" // Default model
	}

	// Map common model names to Doubao models
	switch strings.ToLower(model) {
	case "doubao-pro-4k", "pro-4k", "pro4k":
		return "doubao-pro-4k"
	case "doubao-pro-32k", "pro-32k", "pro32k":
		return "doubao-pro-32k"
	case "doubao-pro-128k", "pro-128k", "pro128k":
		return "doubao-pro-128k"
	case "doubao-lite-4k", "lite-4k", "lite4k":
		return "doubao-lite-4k"
	case "doubao-lite-32k", "lite-32k", "lite32k":
		return "doubao-lite-32k"
	case "doubao-lite-128k", "lite-128k", "lite128k":
		return "doubao-lite-128k"
	case "doubao-pro-vision", "pro-vision", "vision":
		return "doubao-pro-vision"
	case "doubao-pro-search", "pro-search", "search":
		return "doubao-pro-search"
	case "doubao-character-4k", "character-4k", "character4k":
		return "doubao-character-4k"
	case "doubao-character-32k", "character-32k", "character32k":
		return "doubao-character-32k"
	case "doubao-character-128k", "character-128k", "character128k":
		return "doubao-character-128k"
	default:
		// If it's already a valid Doubao model name, use it
		if strings.HasPrefix(model, "doubao-") {
			return model
		}
		// Default fallback
		return "doubao-pro-4k"
	}
}

// GenerateCompletion sends a completion request to the Doubao API.
func (c *DoubaoClient) GenerateCompletion(ctx context.Context, req *CompletionRequest) (CompletionResponse, error) {
	klog.Infof("Doubao GenerateCompletion called with model: %s", req.Model)
	klog.V(1).Infof("Prompt:\n%s", req.Prompt)

	selectedModel := getDoubaoModel(req.Model)

	// Use the Chat Completions API
	completion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModel(selectedModel),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(req.Prompt),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate Doubao completion: %w", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no response from Doubao")
	}

	response := completion.Choices[0].Message.Content
	klog.V(1).Infof("Doubao completion response: %s", response)

	return &DoubaoCompletionResponse{
		response: response,
		usage:    completion.Usage,
	}, nil
}

// DoubaoCompletionResponse implements CompletionResponse for Doubao.
type DoubaoCompletionResponse struct {
	response string
	usage    openai.CompletionUsage
}

func (r *DoubaoCompletionResponse) Response() string {
	return r.response
}

func (r *DoubaoCompletionResponse) UsageMetadata() any {
	return r.usage
}

// --- Chat Session Implementation ---

type doubaoChat struct {
	client              openai.Client
	history             []openai.ChatCompletionMessageParamUnion
	model               string
	functionDefinitions []*FunctionDefinition            // Stored in gollm format
	tools               []openai.ChatCompletionToolParam // Stored in OpenAI format
}

// Ensure doubaoChat implements the Chat interface.
var _ Chat = (*doubaoChat)(nil)

// SetFunctionDefinitions configures the available functions for the chat session.
func (cs *doubaoChat) SetFunctionDefinitions(functionDefinitions []*FunctionDefinition) error {
	klog.V(1).InfoS("doubaoChat.SetFunctionDefinitions called", "count", len(functionDefinitions))

	cs.functionDefinitions = functionDefinitions
	cs.tools = nil // Reset tools

	// Convert gollm function definitions to OpenAI format
	for _, funcDef := range functionDefinitions {
		if funcDef == nil {
			continue
		}

		// Convert schema to OpenAI format
		var params openai.FunctionParameters
		if funcDef.Parameters != nil {
			schemaBytes, err := cs.convertSchemaToBytes(funcDef.Parameters, funcDef.Name)
			if err != nil {
				return fmt.Errorf("failed to convert schema for function %s: %w", funcDef.Name, err)
			}

			if err := json.Unmarshal(schemaBytes, &params); err != nil {
				return fmt.Errorf("failed to parse parameters for function %s: %w", funcDef.Name, err)
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

	klog.V(1).InfoS("doubaoChat function definitions set", "tools_count", len(cs.tools))
	return nil
}

// convertSchemaToBytes converts a validated schema to JSON bytes using OpenAI-specific marshaling
func (cs *doubaoChat) convertSchemaToBytes(schema *Schema, functionName string) ([]byte, error) {
	// Wrap the schema with OpenAI-specific marshaling behavior
	openAIWrapper := openAISchema{Schema: schema}

	bytes, err := json.Marshal(openAIWrapper)
	if err != nil {
		return nil, fmt.Errorf("failed to convert schema: %w", err)
	}

	klog.Infof("Doubao schema for function %s: %s", functionName, string(bytes))

	return bytes, nil
}

// Send sends the user message(s), appends to history, and gets the LLM response.
func (cs *doubaoChat) Send(ctx context.Context, contents ...any) (ChatResponse, error) {
	klog.V(1).InfoS("doubaoChat.Send called", "model", cs.model, "history_len", len(cs.history))

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

	// Call the Doubao API via Volces
	klog.V(1).InfoS("Sending request to Doubao Chat API", "model", cs.model, "messages", len(chatReq.Messages), "tools", len(chatReq.Tools))
	completion, err := cs.client.Chat.Completions.New(ctx, chatReq)
	if err != nil {
		klog.Errorf("Doubao ChatCompletion API error: %v", err)
		return nil, fmt.Errorf("Doubao chat completion failed: %w", err)
	}
	klog.V(1).InfoS("Received response from Doubao Chat API", "id", completion.ID, "choices", len(completion.Choices))

	// Process the response
	if len(completion.Choices) == 0 {
		klog.Warning("Received response with no choices from Doubao")
		return nil, errors.New("received empty response from Doubao (no choices)")
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
	response := &doubaoResponse{
		completion: completion,
	}

	klog.V(1).InfoS("doubaoChat.Send completed", "response_id", completion.ID)
	return response, nil
}

// SendStreaming is the streaming version of Send.
func (cs *doubaoChat) SendStreaming(ctx context.Context, contents ...any) (ChatResponseIterator, error) {
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
func (cs *doubaoChat) addContentsToHistory(contents []any) error {
	for _, content := range contents {
		switch v := content.(type) {
		case string:
			// Simple text message
			cs.history = append(cs.history, openai.UserMessage(v))
		case FunctionCallResult:
			// Function call result
			resultJSON, err := json.Marshal(v.Result)
			if err != nil {
				klog.Errorf("Failed to serialize function call result: %v", err)
				return fmt.Errorf("failed to serialize function call result %q: %w", v.Name, err)
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

type doubaoResponse struct {
	completion *openai.ChatCompletion
}

// Ensure doubaoResponse implements ChatResponse.
var _ ChatResponse = (*doubaoResponse)(nil)

func (r *doubaoResponse) UsageMetadata() any {
	return r.completion.Usage
}

func (r *doubaoResponse) Candidates() []Candidate {
	var candidates []Candidate
	for _, choice := range r.completion.Choices {
		candidates = append(candidates, &doubaoCandidate{choice: choice})
	}
	return candidates
}

// --- Candidate Implementation ---

type doubaoCandidate struct {
	choice openai.ChatCompletionChoice
}

// Ensure doubaoCandidate implements Candidate.
var _ Candidate = (*doubaoCandidate)(nil)

func (c *doubaoCandidate) String() string {
	return c.choice.Message.Content
}

func (c *doubaoCandidate) Parts() []Part {
	var parts []Part

	// Add text part if content exists
	if c.choice.Message.Content != "" {
		parts = append(parts, &doubaoTextPart{text: c.choice.Message.Content})
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
			parts = append(parts, &doubaoFunctionCallPart{functionCalls: functionCalls})
		}
	}

	return parts
}

// --- Part Implementations ---

type doubaoTextPart struct {
	text string
}

// Ensure doubaoTextPart implements Part.
var _ Part = (*doubaoTextPart)(nil)

func (p *doubaoTextPart) AsText() (string, bool) {
	return p.text, true
}

func (p *doubaoTextPart) AsFunctionCalls() ([]FunctionCall, bool) {
	return nil, false
}

type doubaoFunctionCallPart struct {
	functionCalls []FunctionCall
}

// Ensure doubaoFunctionCallPart implements Part.
var _ Part = (*doubaoFunctionCallPart)(nil)

func (p *doubaoFunctionCallPart) AsText() (string, bool) {
	return "", false
}

func (p *doubaoFunctionCallPart) AsFunctionCalls() ([]FunctionCall, bool) {
	return p.functionCalls, true
}

// IsRetryableError 判断错误是否可重试
func (d *doubaoChat) IsRetryableError(err error) bool {
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
