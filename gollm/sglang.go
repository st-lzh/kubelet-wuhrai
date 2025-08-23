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
	"os"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"k8s.io/klog/v2"
)

func init() {
	if err := RegisterProvider("sglang", newSGLangClientFactory); err != nil {
		klog.Fatalf("Failed to register sglang provider: %v", err)
	}
}

// SGLangClient wraps OpenAIClient to provide SGLang-specific functionality
type SGLangClient struct {
	*OpenAIClient
	baseURL string
}

var _ Client = &SGLangClient{}

// newSGLangClientFactory creates a new SGLang client factory
func newSGLangClientFactory(ctx context.Context, opts ClientOptions) (Client, error) {
	return NewSGLangClient(ctx, opts)
}

// NewSGLangClient creates a new client for SGLang server
func NewSGLangClient(ctx context.Context, opts ClientOptions) (*SGLangClient, error) {
	// Get SGLang server URL from environment or use default
	sglangURL := os.Getenv("SGLANG_URL")
	if sglangURL == "" {
		sglangURL = "http://localhost:30000"
	}

	// Override URL if provided in opts
	if opts.URL != nil && opts.URL.String() != "" {
		sglangURL = opts.URL.String()
	}

	klog.Infof("Connecting to SGLang server at: %s", sglangURL)

	// Create OpenAI client with SGLang endpoint
	// SGLang uses OpenAI compatible API but may not require an API key
	apiKey := os.Getenv("SGLANG_API_KEY")
	if apiKey == "" {
		apiKey = "dummy-key-for-sglang"
	}

	options := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithBaseURL(sglangURL),
	}

	// Support custom HTTP client (e.g., skip SSL verification)
	httpClient := createCustomHTTPClient(opts.SkipVerifySSL)
	options = append(options, option.WithHTTPClient(httpClient))

	openaiClient := &OpenAIClient{
		client: openai.NewClient(options...),
	}

	return &SGLangClient{
		OpenAIClient: openaiClient,
		baseURL:      sglangURL,
	}, nil
}

// StartChat starts a new chat session with SGLang-specific configuration
func (c *SGLangClient) StartChat(systemPrompt, model string) Chat {
	// Use default model if not specified
	if model == "" {
		model = getSGLangDefaultModel()
	}

	klog.V(1).Infof("Starting new SGLang chat session with model: %s", model)
	return c.OpenAIClient.StartChat(systemPrompt, model)
}

// ListModels returns available models from SGLang server
func (c *SGLangClient) ListModels(ctx context.Context) ([]string, error) {
	klog.V(1).Info("Listing models from SGLang server")
	
	models, err := c.OpenAIClient.ListModels(ctx)
	if err != nil {
		klog.Warningf("Failed to list models from SGLang server: %v", err)
		// Return a sensible default for SGLang
		return []string{"default"}, nil
	}
	
	return models, nil
}

// getSGLangDefaultModel returns the default model for SGLang
func getSGLangDefaultModel() string {
	model := os.Getenv("SGLANG_MODEL")
	if model == "" {
		// SGLang typically serves one model at a time
		model = "default"
	}
	return model
}

// GenerateCompletion generates a completion using SGLang
func (c *SGLangClient) GenerateCompletion(ctx context.Context, req *CompletionRequest) (CompletionResponse, error) {
	if req.Model == "" {
		req.Model = getSGLangDefaultModel()
	}
	return c.OpenAIClient.GenerateCompletion(ctx, req)
}