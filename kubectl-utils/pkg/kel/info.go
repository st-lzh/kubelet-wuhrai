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
	"strings"

	"github.com/google/cel-go/cel"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

type InfoFunction func(ctx context.Context, self *unstructured.Unstructured) string

// BuildStatusPrinter returns an InfoFunction that attempts to report important values from the evaluation of the CEL expression
func (x *Expression) BuildStatusPrinter(ctx context.Context) (InfoFunction, error) {
	log := klog.FromContext(ctx)

	checkedExpr, err := cel.AstToCheckedExpr(x.AST)
	if err != nil {
		return nil, fmt.Errorf("parsing CEL ast: %w", err)
	}

	v := checkedExpr.Expr.ExprKind
	switch v := v.(type) {
	case *exprpb.Expr_CallExpr:
		printFunction := ""
		switch v.CallExpr.Function {
		case "_==_":
			printFunction = "="
		case "_>=_":
			printFunction = ">="
		case "_<=_":
			printFunction = "<="
		case "_>_":
			printFunction = ">"
		case "_<_":
			printFunction = "<"
		default:
			klog.Warningf("unhandled function %q", v.CallExpr.Function)
			return nil, nil
		}
		log.V(2).Info("recognized function", "function", printFunction)
		return x.buildFunctionPrinterFor(v.CallExpr.Args)

	default:
		klog.Warningf("unhandled expression kind %T", checkedExpr.Expr.ExprKind)
		return nil, nil
	}
}

func (x *Expression) buildFunctionPrinterFor(args []*exprpb.Expr) (InfoFunction, error) {
	checkedExpr, err := cel.AstToCheckedExpr(x.AST)
	if err != nil {
		return nil, fmt.Errorf("parsing CEL ast: %w", err)
	}

	type debugValue struct {
		Key     string
		Program cel.Program
	}
	var debugValues []debugValue

	for _, arg := range args {
		shouldPrint := true

		v := arg.ExprKind
		switch v := v.(type) {
		case *exprpb.Expr_ConstExpr:
			// Don't print constants, 2=2 is not informative
			shouldPrint = false
		case *exprpb.Expr_SelectExpr:
			shouldPrint = true

		default:
			klog.Warningf("unhandled expression kind %T", v)
		}

		if !shouldPrint {
			continue
		}

		checkedArg := proto.Clone(checkedExpr).(*exprpb.CheckedExpr)
		checkedArg.Expr = arg

		ast := cel.CheckedExprToAst(checkedArg)
		celExpression, err := cel.AstToString(ast)
		if err != nil {
			return nil, fmt.Errorf("converting expression to string: %w", err)
		}

		compiled, issues := x.Env.Compile(celExpression)
		if issues != nil && issues.Err() != nil {
			return nil, fmt.Errorf("invalid expression %q: %w", celExpression, issues.Err())
		}
		prg, err := x.Env.Program(compiled)
		if err != nil {
			return nil, fmt.Errorf("invalid expression %q: %w", celExpression, err)
		}

		debugValues = append(debugValues, debugValue{
			Key:     celExpression,
			Program: prg,
		})
	}

	if len(debugValues) == 0 {
		return nil, nil
	}

	return func(ctx context.Context, self *unstructured.Unstructured) string {
		log := klog.FromContext(ctx)

		inputs := x.buildInputs(self)

		var values []string
		for _, debugValue := range debugValues {
			s := ""
			out, details, err := debugValue.Program.Eval(inputs)
			log.V(2).Info("evaluated CEL expression", "out", out, "details", details, "error", err)
			if err == nil {
				s = fmt.Sprintf("%s=%v", debugValue.Key, out.Value())
			} else {
				s = fmt.Sprintf("%s=%v", debugValue.Key, "???")
			}
			values = append(values, s)
		}

		return strings.Join(values, "; ")
	}, nil
}
