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
	if err := RegisterProvider("vllm", newVLLMClientFactory); err != nil {
		klog.Fatalf("Failed to register vllm provider: %v", err)
	}
}

// VLLMClient wraps OpenAIClient to provide vLLM-specific functionality
type VLLMClient struct {
	*OpenAIClient
	baseURL string
}

var _ Client = &VLLMClient{}

// newVLLMClientFactory creates a new vLLM client factory
func newVLLMClientFactory(ctx context.Context, opts ClientOptions) (Client, error) {
	return NewVLLMClient(ctx, opts)
}

// NewVLLMClient creates a new client for vLLM server
func NewVLLMClient(ctx context.Context, opts ClientOptions) (*VLLMClient, error) {
	// Get vLLM server URL from environment or use default
	vllmURL := os.Getenv("VLLM_URL")
	if vllmURL == "" {
		vllmURL = "http://localhost:8000"
	}

	// Override URL if provided in opts
	if opts.URL != nil && opts.URL.String() != "" {
		vllmURL = opts.URL.String()
	}

	klog.Infof("Connecting to vLLM server at: %s", vllmURL)

	// Create OpenAI client with vLLM endpoint
	// vLLM doesn't require an API key, but the OpenAI client does
	// So we use a dummy key
	apiKey := os.Getenv("VLLM_API_KEY")
	if apiKey == "" {
		apiKey = "dummy-key-for-vllm"
	}

	options := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithBaseURL(vllmURL),
	}

	// Support custom HTTP client (e.g., skip SSL verification)
	httpClient := createCustomHTTPClient(opts.SkipVerifySSL)
	options = append(options, option.WithHTTPClient(httpClient))

	openaiClient := &OpenAIClient{
		client: openai.NewClient(options...),
	}

	return &VLLMClient{
		OpenAIClient: openaiClient,
		baseURL:      vllmURL,
	}, nil
}

// StartChat starts a new chat session with vLLM-specific configuration
func (c *VLLMClient) StartChat(systemPrompt, model string) Chat {
	// Use default model if not specified
	if model == "" {
		model = getVLLMDefaultModel()
	}

	klog.V(1).Infof("Starting new vLLM chat session with model: %s", model)
	return c.OpenAIClient.StartChat(systemPrompt, model)
}

// ListModels returns available models from vLLM server
func (c *VLLMClient) ListModels(ctx context.Context) ([]string, error) {
	klog.V(1).Info("Listing models from vLLM server")
	
	models, err := c.OpenAIClient.ListModels(ctx)
	if err != nil {
		klog.Warningf("Failed to list models from vLLM server: %v", err)
		// Return a sensible default for vLLM
		return []string{"default"}, nil
	}
	
	return models, nil
}

// getVLLMDefaultModel returns the default model for vLLM
func getVLLMDefaultModel() string {
	model := os.Getenv("VLLM_MODEL")
	if model == "" {
		// vLLM typically serves one model at a time
		model = "default"
	}
	return model
}

// GenerateCompletion generates a completion using vLLM
func (c *VLLMClient) GenerateCompletion(ctx context.Context, req *CompletionRequest) (CompletionResponse, error) {
	if req.Model == "" {
		req.Model = getVLLMDefaultModel()
	}
	return c.OpenAIClient.GenerateCompletion(ctx, req)
}