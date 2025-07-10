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

package mcp

import (
	"context"
	"fmt"

	mcpclient "github.com/mark3labs/mcp-go/client"
)

// MCPClient defines the common interface for all MCP client implementations
type MCPClient interface {
	// Name returns the name of this client
	Name() string

	// Connect establishes a connection to the MCP server
	Connect(ctx context.Context) error

	// Close closes the connection to the MCP server
	Close() error

	// ListTools lists all available tools from the MCP server
	ListTools(ctx context.Context) ([]Tool, error)

	// CallTool calls a tool on the MCP server and returns the result as a string
	CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error)

	// ensureConnected makes sure the client is connected
	ensureConnected() error

	// getUnderlyingClient returns the underlying mcpclient.Client
	getUnderlyingClient() *mcpclient.Client
}

// ClientConfig contains all configuration options for MCP clients
type ClientConfig struct {
	// Common fields
	Name string

	// For stdio-based clients
	Command string
	Args    []string
	Env     []string

	// For HTTP-based clients
	URL          string
	Auth         *AuthConfig
	OAuthConfig  *OAuthConfig
	Timeout      int
	UseStreaming bool // Whether to use streaming HTTP for better performance

	// No LLM configuration needed - MCP doesn't need to know about LLM models
}

// AuthConfig represents authentication options for HTTP MCP servers
type AuthConfig struct {
	Type       string // "none", "basic", "bearer", "api-key"
	Username   string // For basic auth
	Password   string // For basic auth
	Token      string // For bearer auth
	ApiKey     string // For API key auth
	HeaderName string // Custom header name for API key
}

// OAuthConfig represents OAuth configuration for HTTP MCP servers
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
	AuthURL      string
	Scopes       []string
	RedirectURL  string
}

// NewMCPClient creates a new MCP client with the appropriate implementation based on the config
func NewMCPClient(config ClientConfig) (MCPClient, error) {
	// Validate common configuration
	if config.Name == "" {
		return nil, fmt.Errorf("client name is required")
	}

	// Choose the appropriate client implementation
	if config.URL != "" {
		// Use HTTP client
		return NewHTTPClient(config), nil
	}

	// Default to stdio client
	if config.Command == "" {
		return nil, fmt.Errorf("either URL or Command must be specified")
	}
	return NewStdioClient(config), nil
}
