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
	"testing"
)

func TestCustomTool_AddCommandPrefix(t *testing.T) {
	tests := []struct {
		name           string
		configCommand  string
		inputCommand   string
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "simple command without prefix",
			configCommand:  "gcloud",
			inputCommand:   "compute instances list",
			expectedOutput: "gcloud compute instances list",
			expectError:    false,
		},
		{
			name:           "simple command with prefix",
			configCommand:  "gcloud",
			inputCommand:   "gcloud compute instances list",
			expectedOutput: "gcloud compute instances list",
			expectError:    false,
		},
		{
			name:           "command with pipe",
			configCommand:  "gcloud",
			inputCommand:   "compute instances list | grep test",
			expectedOutput: "compute instances list | grep test",
			expectError:    false,
		},
		{
			name:           "command with redirect",
			configCommand:  "gcloud",
			inputCommand:   "compute instances list > instances.txt",
			expectedOutput: "compute instances list > instances.txt",
			expectError:    false,
		},
		{
			name:           "command with background",
			configCommand:  "gcloud",
			inputCommand:   "compute instances list &",
			expectedOutput: "compute instances list &",
			expectError:    false,
		},
		{
			name:           "command with subshell",
			configCommand:  "gcloud",
			inputCommand:   "(compute instances list)",
			expectedOutput: "(compute instances list)",
			expectError:    false,
		},
		{
			name:           "command with multiple statements",
			configCommand:  "gcloud",
			inputCommand:   "compute instances list; compute disks list",
			expectedOutput: "compute instances list; compute disks list",
			expectError:    false,
		},
		{
			name:           "invalid shell syntax",
			configCommand:  "gcloud",
			inputCommand:   "compute instances list |",
			expectedOutput: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &CustomTool{
				config: CustomToolConfig{
					Command: tt.configCommand,
				},
			}

			output, err := tool.addCommandPrefix(tt.inputCommand)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if output != tt.expectedOutput {
				t.Errorf("expected %q, got %q", tt.expectedOutput, output)
			}
		})
	}
}
