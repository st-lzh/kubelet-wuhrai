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
	"encoding/json"
	"testing"
)

func TestConvertSchemaForOpenAI(t *testing.T) {
	tests := []struct {
		name           string
		inputSchema    *Schema
		expectedType   SchemaType
		expectedError  bool
		validateResult func(t *testing.T, result *Schema)
	}{
		// Core logic tests
		{
			name:          "nil schema",
			inputSchema:   nil,
			expectedType:  TypeObject,
			expectedError: false,
			validateResult: func(t *testing.T, result *Schema) {
				if result.Properties == nil {
					t.Error("expected properties map to be initialized")
				}
				if len(result.Properties) != 0 {
					t.Error("expected empty properties map")
				}
			},
		},
		{
			name: "simple string schema",
			inputSchema: &Schema{
				Type:        TypeString,
				Description: "A simple string",
			},
			expectedType:  TypeString,
			expectedError: false,
			validateResult: func(t *testing.T, result *Schema) {
				if result.Description != "A simple string" {
					t.Errorf("expected description 'A simple string', got %q", result.Description)
				}
			},
		},
		{
			name: "simple number schema",
			inputSchema: &Schema{
				Type: TypeNumber,
			},
			expectedType:  TypeNumber,
			expectedError: false,
		},
		{
			name: "integer schema converted to number",
			inputSchema: &Schema{
				Type:        TypeInteger,
				Description: "An integer value",
			},
			expectedType:  TypeNumber, // OpenAI prefers number for integers
			expectedError: false,
			validateResult: func(t *testing.T, result *Schema) {
				if result.Description != "An integer value" {
					t.Errorf("expected description preserved")
				}
			},
		},
		{
			name: "boolean schema",
			inputSchema: &Schema{
				Type: TypeBoolean,
			},
			expectedType:  TypeBoolean,
			expectedError: false,
		},
		{
			name: "empty type defaults to object",
			inputSchema: &Schema{
				Description: "No type specified",
			},
			expectedType:  TypeObject,
			expectedError: false,
			validateResult: func(t *testing.T, result *Schema) {
				if result.Properties == nil {
					t.Error("expected properties map to be initialized")
				}
			},
		},
		{
			name: "unknown type defaults to object",
			inputSchema: &Schema{
				Type: "unknown",
			},
			expectedType:  TypeObject,
			expectedError: false,
			validateResult: func(t *testing.T, result *Schema) {
				if result.Properties == nil {
					t.Error("expected properties map to be initialized")
				}
			},
		},
		{
			name: "object schema with properties",
			inputSchema: &Schema{
				Type: TypeObject,
				Properties: map[string]*Schema{
					"name": {Type: TypeString, Description: "User name"},
					"age":  {Type: TypeInteger, Description: "User age"},
				},
				Required: []string{"name"},
			},
			expectedType:  TypeObject,
			expectedError: false,
			validateResult: func(t *testing.T, result *Schema) {
				if len(result.Properties) != 2 {
					t.Errorf("expected 2 properties, got %d", len(result.Properties))
				}
				if result.Properties["name"].Type != TypeString {
					t.Error("expected name property to be string")
				}
				// Age should be converted from integer to number
				if result.Properties["age"].Type != TypeNumber {
					t.Error("expected age property to be converted to number")
				}
				if len(result.Required) != 1 || result.Required[0] != "name" {
					t.Error("expected required fields to be preserved")
				}
			},
		},
		{
			name: "object schema without properties",
			inputSchema: &Schema{
				Type: TypeObject,
			},
			expectedType:  TypeObject,
			expectedError: false,
			validateResult: func(t *testing.T, result *Schema) {
				if result.Properties == nil {
					t.Error("expected properties map to be initialized")
				}
				if len(result.Properties) != 0 {
					t.Error("expected empty properties map")
				}
			},
		},
		{
			name: "array schema with string items",
			inputSchema: &Schema{
				Type:  TypeArray,
				Items: &Schema{Type: TypeString},
			},
			expectedType:  TypeArray,
			expectedError: false,
			validateResult: func(t *testing.T, result *Schema) {
				if result.Items == nil {
					t.Error("expected items schema to be present")
				}
				if result.Items.Type != TypeString {
					t.Error("expected items to be string type")
				}
			},
		},
		{
			name: "array schema with integer items (converted to number)",
			inputSchema: &Schema{
				Type:  TypeArray,
				Items: &Schema{Type: TypeInteger},
			},
			expectedType:  TypeArray,
			expectedError: false,
			validateResult: func(t *testing.T, result *Schema) {
				if result.Items == nil {
					t.Error("expected items schema to be present")
				}
				if result.Items.Type != TypeNumber {
					t.Error("expected items to be converted to number type")
				}
			},
		},
		{
			name: "array schema without items (defaults to string)",
			inputSchema: &Schema{
				Type: TypeArray,
			},
			expectedType:  TypeArray,
			expectedError: false,
			validateResult: func(t *testing.T, result *Schema) {
				if result.Items == nil {
					t.Error("expected items schema to be defaulted")
				}
				if result.Items.Type != TypeString {
					t.Error("expected default items to be string type")
				}
			},
		},
		{
			name: "nested object in array",
			inputSchema: &Schema{
				Type: TypeArray,
				Items: &Schema{
					Type: TypeObject,
					Properties: map[string]*Schema{
						"id":   {Type: TypeInteger},
						"name": {Type: TypeString},
					},
				},
			},
			expectedType:  TypeArray,
			expectedError: false,
			validateResult: func(t *testing.T, result *Schema) {
				if result.Items == nil {
					t.Error("expected items schema to be present")
				}
				if result.Items.Type != TypeObject {
					t.Error("expected items to be object type")
				}
				if result.Items.Properties["id"].Type != TypeNumber {
					t.Error("expected nested integer to be converted to number")
				}
				if result.Items.Properties["name"].Type != TypeString {
					t.Error("expected nested string to remain string")
				}
			},
		},

		// Built-in tool schema tests
		{
			name: "kubectl tool schema",
			inputSchema: &Schema{
				Type: TypeObject,
				Properties: map[string]*Schema{
					"command": {
						Type:        TypeString,
						Description: "The complete kubectl command to execute",
					},
					"modifies_resource": {
						Type:        TypeString,
						Description: "Whether the command modifies a kubernetes resource",
					},
				},
			},
			expectedType:  TypeObject,
			expectedError: false,
			validateResult: func(t *testing.T, result *Schema) {
				if len(result.Properties) != 2 {
					t.Errorf("expected 2 properties, got %d", len(result.Properties))
				}
				if result.Properties["command"].Type != TypeString {
					t.Error("expected command property to be string")
				}
				if result.Properties["modifies_resource"].Type != TypeString {
					t.Error("expected modifies_resource property to be string")
				}
				// Properties should be initialized
				if result.Properties == nil {
					t.Error("expected properties to be initialized")
				}
			},
		},
		{
			name: "bash tool schema",
			inputSchema: &Schema{
				Type: TypeObject,
				Properties: map[string]*Schema{
					"command": {
						Type:        TypeString,
						Description: "The bash command to execute",
					},
					"modifies_resource": {
						Type:        TypeString,
						Description: "Whether the command modifies a kubernetes resource",
					},
				},
			},
			expectedType:  TypeObject,
			expectedError: false,
			validateResult: func(t *testing.T, result *Schema) {
				if len(result.Properties) != 2 {
					t.Errorf("expected 2 properties, got %d", len(result.Properties))
				}
				// All string properties should remain strings
				if result.Properties["command"].Type != TypeString {
					t.Error("expected command property to remain string")
				}
				if result.Properties["modifies_resource"].Type != TypeString {
					t.Error("expected modifies_resource property to remain string")
				}
			},
		},
		{
			name: "mcp tool schema with complex nested structure",
			inputSchema: &Schema{
				Type: TypeObject,
				Properties: map[string]*Schema{
					"server_name": {
						Type:        TypeString,
						Description: "Name of the MCP server",
					},
					"method": {
						Type:        TypeString,
						Description: "MCP method name",
					},
					"params": {
						Type: TypeObject,
						Properties: map[string]*Schema{
							"query": {Type: TypeString},
							"limit": {Type: TypeInteger}, // Should convert to number
						},
					},
				},
				Required: []string{"server_name", "method"},
			},
			expectedType:  TypeObject,
			expectedError: false,
			validateResult: func(t *testing.T, result *Schema) {
				// Check top-level properties
				if len(result.Properties) != 3 {
					t.Errorf("expected 3 properties, got %d", len(result.Properties))
				}
				// Check nested object conversion
				params := result.Properties["params"]
				if params.Type != TypeObject {
					t.Error("expected params to be object type")
				}
				if params.Properties == nil {
					t.Error("expected params properties to be initialized")
				}
				// Check nested integer conversion
				if params.Properties["limit"].Type != TypeNumber {
					t.Error("expected nested limit property to be converted to number")
				}
				// Check required fields preservation
				if len(result.Required) != 2 {
					t.Error("expected required fields to be preserved")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertSchemaForOpenAI(tt.inputSchema)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected non-nil result")
				return
			}

			if result.Type != tt.expectedType {
				t.Errorf("expected type %q, got %q", tt.expectedType, result.Type)
			}

			// Run custom validation if provided
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

// TestConvertSchemaToBytes tests the JSON-level fix for the omitempty issue
func TestConvertSchemaToBytes(t *testing.T) {
	session := &openAIChatSession{}

	// Test case: Object schema with empty properties map (which gets omitted by omitempty)
	schema := &Schema{
		Type:       TypeObject,
		Properties: make(map[string]*Schema), // Empty map gets omitted by omitempty
	}

	bytes, err := session.convertSchemaToBytes(schema, "test_function")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// Parse the JSON to verify it has properties field
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(bytes, &schemaMap); err != nil {
		t.Errorf("failed to unmarshal schema: %v", err)
		return
	}

	// Verify the schema has type: object
	if schemaType, ok := schemaMap["type"].(string); !ok || schemaType != "object" {
		t.Errorf("expected type 'object', got %v", schemaMap["type"])
	}

	// Verify the schema has properties field (even if empty)
	if _, hasProperties := schemaMap["properties"]; !hasProperties {
		t.Error("expected properties field to be present in JSON, but it was missing")
	}

	// Verify properties is an empty object
	if props, ok := schemaMap["properties"].(map[string]interface{}); !ok {
		t.Error("expected properties to be an object")
	} else if len(props) != 0 {
		t.Errorf("expected empty properties object, got %v", props)
	}
}
