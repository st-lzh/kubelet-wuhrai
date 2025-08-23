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

package gollm

import (
	"context"
	"errors"
	"os"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"k8s.io/klog/v2"
)

func init() {
	if err := RegisterProvider("openrouter", newOpenRouterClientFactory); err != nil {
		klog.Fatalf("Failed to register openrouter provider: %v", err)
	}
}

// OpenRouterClient wraps OpenAIClient to provide OpenRouter-specific functionality
type OpenRouterClient struct {
	*OpenAIClient
	baseURL string
}

var _ Client = &OpenRouterClient{}

// newOpenRouterClientFactory creates a new OpenRouter client factory
func newOpenRouterClientFactory(ctx context.Context, opts ClientOptions) (Client, error) {
	return NewOpenRouterClient(ctx, opts)
}

// NewOpenRouterClient creates a new client for OpenRouter API
func NewOpenRouterClient(ctx context.Context, opts ClientOptions) (*OpenRouterClient, error) {
	// Get OpenRouter API key from environment
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OpenRouter API key not found. Set via OPENROUTER_API_KEY env var")
	}

	// OpenRouter base URL
	baseURL := "https://openrouter.ai/api/v1"
	
	// Allow override from environment or opts
	if envURL := os.Getenv("OPENROUTER_URL"); envURL != "" {
		baseURL = envURL
	}
	if opts.URL != nil && opts.URL.String() != "" {
		baseURL = opts.URL.String()
	}

	klog.Infof("Connecting to OpenRouter at: %s", baseURL)

	// Create OpenAI client with OpenRouter configuration
	options := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	}

	// Support custom HTTP client (e.g., skip SSL verification)
	httpClient := createCustomHTTPClient(opts.SkipVerifySSL)
	options = append(options, option.WithHTTPClient(httpClient))

	openaiClient := &OpenAIClient{
		client: openai.NewClient(options...),
	}

	return &OpenRouterClient{
		OpenAIClient: openaiClient,
		baseURL:      baseURL,
	}, nil
}

// StartChat starts a new chat session with OpenRouter-specific configuration
func (c *OpenRouterClient) StartChat(systemPrompt, model string) Chat {
	// Use default model if not specified
	if model == "" {
		model = getOpenRouterDefaultModel()
	}

	klog.V(1).Infof("Starting new OpenRouter chat session with model: %s", model)
	return c.OpenAIClient.StartChat(systemPrompt, model)
}

// ListModels returns available models from OpenRouter
func (c *OpenRouterClient) ListModels(ctx context.Context) ([]string, error) {
	klog.V(1).Info("Listing models from OpenRouter")
	
	models, err := c.OpenAIClient.ListModels(ctx)
	if err != nil {
		klog.Warningf("Failed to list models from OpenRouter: %v", err)
		// Return some popular OpenRouter models as fallback
		return []string{
			"anthropic/claude-3.5-sonnet",
			"openai/gpt-4o",
			"openai/gpt-4o-mini", 
			"google/gemini-2.5-flash",
			"meta-llama/llama-3.3-70b-instruct",
			"deepseek/deepseek-r1",
			"anthropic/claude-3-opus",
			"cohere/command-r-plus",
		}, nil
	}
	
	return models, nil
}

// getOpenRouterDefaultModel returns the default model for OpenRouter
func getOpenRouterDefaultModel() string {
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		// Default to a good general-purpose model
		model = "anthropic/claude-3.5-sonnet"
	}
	return model
}

// GenerateCompletion generates a completion using OpenRouter
func (c *OpenRouterClient) GenerateCompletion(ctx context.Context, req *CompletionRequest) (CompletionResponse, error) {
	if req.Model == "" {
		req.Model = getOpenRouterDefaultModel()
	}
	return c.OpenAIClient.GenerateCompletion(ctx, req)
}