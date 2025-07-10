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

package kel

import (
	"context"
	"fmt"

	"github.com/google/cel-go/cel"
	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

func NewEnv() (*cel.Env, error) {
	// TODO: Can we / should we do better than AnyType?
	env, err := cel.NewEnv(
		cel.Variable("self", cel.AnyType),
	)

	return env, err
}

type Expression struct {
	CELText string
	Program cel.Program
	AST     *cel.Ast
	Env     *cel.Env
}

func NewExpression(env *cel.Env, celExpression string) (*Expression, error) {
	ast, issues := env.Compile(celExpression)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("invalid expression %q: %w", celExpression, issues.Err())
	}
	prg, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("invalid expression %q: %w", celExpression, err)
	}
	return &Expression{
		CELText: celExpression,
		AST:     ast,
		Program: prg,
		Env:     env,
	}, nil
}

func (x *Expression) Eval(ctx context.Context, self *unstructured.Unstructured) (ref.Val, error) {
	log := klog.FromContext(ctx)
	inputs := x.buildInputs(self)

	out, details, err := x.Program.Eval(inputs)
	if err != nil {
		return nil, fmt.Errorf("evaluating CEL expression: %w", err)
	}
	log.V(2).Info("evaluated CEL expression", "out", out, "details", details)
	return out, nil
}

func (x *Expression) buildInputs(self *unstructured.Unstructured) map[string]any {
	inputs := map[string]any{
		"self": celtypes.NewDynamicMap(&unstructuredToCELAdapter{}, self.Object),
	}
	return inputs
}

type unstructuredToCELAdapter struct {
}

func (a *unstructuredToCELAdapter) NativeToValue(value any) ref.Val {
	switch value := value.(type) {
	case string:
		return celtypes.String(value)
	case int:
		return celtypes.Int(value)
	case int64:
		return celtypes.Int(value)
	case map[string]any:
		return celtypes.NewDynamicMap(a, value)
	default:
		klog.Fatalf("unhandled type %T", value)
		return nil
	}
}
