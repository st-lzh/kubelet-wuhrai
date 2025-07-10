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

package tools

import (
	"context"
	"fmt"

	"github.com/st-lzh/kubelet-wuhrai/gollm"
	"github.com/st-lzh/kubelet-wuhrai/pkg/mcp"
	"k8s.io/klog/v2"
)

// =============================================================================
// Schema Conversion Functions (kubectl-ai specific)
// =============================================================================

// ConvertToolToGollm converts an MCP tool to gollm.FunctionDefinition with a simple schema
func ConvertToolToGollm(mcpTool *mcp.Tool) (*gollm.FunctionDefinition, error) {
	def := &gollm.FunctionDefinition{
		Name:        mcpTool.Name,
		Description: mcpTool.Description,
		Parameters:  mcpTool.InputSchema,
	}
	return def, nil
}

// =============================================================================
// MCP Tool Implementation
// =============================================================================

// MCPTool wraps an MCP server tool to implement the Tool interface.
// It serves as an adapter between MCP-based tools and kubectl-ai's tool system.
type MCPTool struct {
	serverName  string
	toolName    string
	description string
	schema      *gollm.FunctionDefinition
	manager     *mcp.Manager
}

// NewMCPTool creates a new MCP tool wrapper.
func NewMCPTool(serverName, toolName, description string, schema *gollm.FunctionDefinition, manager *mcp.Manager) *MCPTool {
	return &MCPTool{
		serverName:  serverName,
		toolName:    toolName,
		description: description,
		schema:      schema,
		manager:     manager,
	}
}

// Name returns the tool name.
func (t *MCPTool) Name() string {
	return t.toolName
}

// ServerName returns the MCP server name.
func (t *MCPTool) ServerName() string {
	return t.serverName
}

// Description returns the tool description.
func (t *MCPTool) Description() string {
	return t.description
}

// FunctionDefinition returns the tool's function definition.
func (t *MCPTool) FunctionDefinition() *gollm.FunctionDefinition {
	return t.schema
}

// TODO(tuannvm): This is a placeholder implementation. Need to implement detection of interactive MCP tools.
// IsInteractive checks if the tool requires interactive input.
func (t *MCPTool) IsInteractive(args map[string]any) (bool, error) {
	return false, nil
}

// CheckModifiesResource determines if the command modifies kubernetes resources
// For MCP tools, we'll conservatively assume they might modify resources
// since we can't easily determine this for arbitrary external tools
// Returns "yes", "no", or "unknown"
func (t *MCPTool) CheckModifiesResource(args map[string]any) string {
	// Since MCP tools can be arbitrary external tools and we don't have a way to know
	// if they modify resources, we'll conservatively return "unknown"
	return "unknown"
}

// Run executes the MCP tool by calling the appropriate MCP server.
func (t *MCPTool) Run(ctx context.Context, args map[string]any) (any, error) {
	log := klog.FromContext(ctx)

	// Get MCP client for the server
	client, exists := t.manager.GetClient(t.serverName)
	if !exists {
		return nil, fmt.Errorf("MCP server %q not connected", t.serverName)
	}

	// // Convert arguments to proper types for MCP server using the MCP package's functions
	// args = mcp.ConvertArgs(args)

	// Execute tool on MCP server
	result, err := client.CallTool(ctx, t.toolName, args)
	if err != nil {
		log.Info("tool info", "name", t.toolName, "schema", t.schema)
		log.Info("call info", "args", args)
		return nil, fmt.Errorf("calling MCP tool %q on server %q: %w", t.toolName, t.serverName, err)
	}

	return result, nil
}
