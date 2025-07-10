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

package templates

import (
	"embed"
	"fmt"
	"html/template"
)

//go:embed *.html
var htmlFiles embed.FS

func LoadTemplate(key string) (*template.Template, error) {
	// TODO: Caching
	b, err := htmlFiles.ReadFile(key)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", key, err)
	}

	tmpl, err := template.New(key).Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %w", key, err)
	}
	return tmpl, nil
}
