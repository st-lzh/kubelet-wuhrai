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
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

// RetryConfig defines retry behavior for MCP operations
type RetryConfig struct {
	MaxRetries  int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
	Description string
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig(description string) RetryConfig {
	return RetryConfig{
		MaxRetries:  3,
		BaseDelay:   1 * time.Second,
		MaxDelay:    10 * time.Second,
		Multiplier:  2.0,
		Description: description,
	}
}

// RetryOperation executes an operation with exponential backoff retry
func RetryOperation(ctx context.Context, config RetryConfig, operation func() error) error {
	var lastErr error

	for attempt := 1; attempt <= config.MaxRetries; attempt++ {
		klog.V(3).InfoS("Attempting operation",
			"operation", config.Description,
			"attempt", attempt,
			"maxRetries", config.MaxRetries)

		if err := operation(); err == nil {
			if attempt > 1 {
				klog.V(2).InfoS("Operation succeeded after retry",
					"operation", config.Description,
					"attempt", attempt)
			}
			return nil
		} else {
			lastErr = err

			if attempt < config.MaxRetries {
				delay := calculateBackoffDelay(attempt, config)
				klog.V(3).InfoS("Operation failed, retrying",
					"operation", config.Description,
					"attempt", attempt,
					"error", err,
					"nextRetryIn", delay)

				select {
				case <-ctx.Done():
					return fmt.Errorf("operation cancelled: %w", ctx.Err())
				case <-time.After(delay):
					// Continue to next attempt
				}
			}
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxRetries, lastErr)
}

// calculateBackoffDelay calculates exponential backoff delay with jitter
func calculateBackoffDelay(attempt int, config RetryConfig) time.Duration {
	delay := float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt-1))

	if time.Duration(delay) > config.MaxDelay {
		return config.MaxDelay
	}

	return time.Duration(delay)
}

// expandPath expands the command path, handling ~ and environment variables
// If the path is just a binary name (no path separators), it looks in $PATH
func expandPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Expand environment variables first
	expanded := os.ExpandEnv(path)

	// If the command contains no path separators, look it up in $PATH first
	if !strings.Contains(expanded, string(filepath.Separator)) && !strings.HasPrefix(expanded, "~") {
		klog.V(2).InfoS("Attempting PATH lookup for command", "command", expanded)
		// Try to find the command in $PATH
		if pathResolved, err := exec.LookPath(expanded); err == nil {
			klog.V(2).InfoS("Found command in PATH", "command", expanded, "resolved", pathResolved)
			return pathResolved, nil
		} else {
			klog.V(2).InfoS("Command not found in PATH", "command", expanded, "error", err)
		}
		// If not found in PATH, continue with the original logic below
		klog.V(2).InfoS("Command not found in PATH, trying relative to current directory", "command", expanded)
	} else {
		klog.V(2).InfoS("Skipping PATH lookup", "command", expanded, "hasPathSeparator", strings.Contains(expanded, string(filepath.Separator)), "hasTilde", strings.HasPrefix(expanded, "~"))
	}

	// Handle ~ for home directory
	if strings.HasPrefix(expanded, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting home directory: %w", err)
		}
		expanded = filepath.Join(home, expanded[1:])
	}

	// Clean the path to remove any . or .. elements
	expanded = filepath.Clean(expanded)

	// Make the path absolute if it's not already
	if !filepath.IsAbs(expanded) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getting current working directory: %w", err)
		}
		expanded = filepath.Clean(filepath.Join(cwd, expanded))
	}

	// Verify the file exists and is executable
	info, err := os.Stat(expanded)
	if err != nil {
		return "", fmt.Errorf(ErrPathCheckFmt, expanded, err)
	}

	// Check if it's a regular file and executable
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("path %q is not a regular file", expanded)
	}

	// Check if the file is executable by the current user
	if info.Mode().Perm()&0111 == 0 {
		return "", fmt.Errorf("file %q is not executable", expanded)
	}

	return expanded, nil
}

// =============================================================================
// Helper Functions to Reduce Redundancy
// =============================================================================

// ensureConnected checks if the client is connected and returns an error if not
func (c *Client) ensureConnected() error {
	if c.client == nil {
		return fmt.Errorf("not connected to MCP server")
	}
	return nil
}

// =============================================================================
// MCP Tool Helper Functions
// =============================================================================

// FunctionDefinition is an interface representing generic function schema definitions
// This allows the MCP package to create schemas without directly depending on gollm
type FunctionDefinition interface {
	// Schema returns a representation of the function schema
	Schema() any
}

// SchemaProperty is an interface representing generic schema properties
type SchemaProperty interface {
	// Property returns a representation of the schema property
	Property() any
}

// SchemaBuilder is a function that builds a function definition from a tool
type SchemaBuilder func(tool *Tool) (FunctionDefinition, error)
