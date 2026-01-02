package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// CalculatorTool provides basic math calculation capability
type CalculatorTool struct {
	toolInfo *schema.ToolInfo
}

func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{
		toolInfo: &schema.ToolInfo{
			Name: "calculator",
			Desc: "Perform mathematical calculations. Supports +, -, *, /, and parentheses.",
			ParamsOneOf: schema.NewParamsOneOfByParams(
				map[string]*schema.ParameterInfo{
					"expression": {
						Type:     schema.String,
						Desc:     "The mathematical expression to evaluate (e.g., '2 + 3 * 4')",
						Required: true,
					},
				},
			),
		},
	}
}

func (t *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return t.toolInfo, nil
}

type calcInput struct {
	Expression string `json:"expression"`
}

func (t *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input calcInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("parse arguments: %w", err)
	}

	result, err := evalExpr(input.Expression)
	if err != nil {
		return "", fmt.Errorf("evaluate expression: %w", err)
	}

	return fmt.Sprintf("Result: %v", result), nil
}

// evalExpr evaluates a simple math expression
func evalExpr(expr string) (float64, error) {
	// Parse the expression
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return 0, fmt.Errorf("invalid expression: %w", err)
	}

	return eval(node)
}

func eval(node ast.Expr) (float64, error) {
	switch n := node.(type) {
	case *ast.BasicLit:
		if n.Kind == token.INT || n.Kind == token.FLOAT {
			return strconv.ParseFloat(n.Value, 64)
		}
		return 0, fmt.Errorf("unsupported literal: %s", n.Value)

	case *ast.BinaryExpr:
		left, err := eval(n.X)
		if err != nil {
			return 0, err
		}
		right, err := eval(n.Y)
		if err != nil {
			return 0, err
		}

		switch n.Op {
		case token.ADD:
			return left + right, nil
		case token.SUB:
			return left - right, nil
		case token.MUL:
			return left * right, nil
		case token.QUO:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		default:
			return 0, fmt.Errorf("unsupported operator: %s", n.Op)
		}

	case *ast.ParenExpr:
		return eval(n.X)

	case *ast.UnaryExpr:
		val, err := eval(n.X)
		if err != nil {
			return 0, err
		}
		if n.Op == token.SUB {
			return -val, nil
		}
		return val, nil

	default:
		return 0, fmt.Errorf("unsupported expression type")
	}
}
