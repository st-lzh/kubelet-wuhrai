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
	"reflect"
	"strings"

	"k8s.io/klog/v2"
)

// BuildSchemaFor will build a schema for the given golang type.
// Because this does not have description populated, it is more useful for the response schema than tools/functions.
func BuildSchemaFor(t reflect.Type) *Schema {
	out := &Schema{}

	switch t.Kind() {
	case reflect.String:
		out.Type = TypeString
	case reflect.Bool:
		out.Type = TypeBoolean
	case reflect.Int:
		out.Type = TypeInteger
	case reflect.Struct:
		out.Type = TypeObject
		out.Properties = make(map[string]*Schema)
		numFields := t.NumField()
		required := []string{}
		for i := 0; i < numFields; i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" {
				continue
			}
			if strings.HasSuffix(jsonTag, ",omitempty") {
				jsonTag = strings.TrimSuffix(jsonTag, ",omitempty")
			} else {
				required = append(required, jsonTag)
			}

			fieldType := field.Type

			fieldSchema := BuildSchemaFor(fieldType)
			out.Properties[jsonTag] = fieldSchema
		}

		if len(required) != 0 {
			out.Required = required
		}
	case reflect.Slice:
		out.Type = TypeArray
		out.Items = BuildSchemaFor(t.Elem())
	default:
		klog.Fatalf("unhandled kind %v", t.Kind())
	}

	return out
}
