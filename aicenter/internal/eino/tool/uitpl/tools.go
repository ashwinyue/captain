package uitpl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// GetUITemplateTool returns the get_ui_template tool
type GetUITemplateTool struct{}

func (t *GetUITemplateTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_ui_template",
		Desc: "获取指定 UI 模板的详细 schema 格式和使用示例。当需要展示订单、产品、物流等结构化数据时，必须先调用此工具获取格式要求。",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"template_name": {
					Type:     schema.String,
					Desc:     "模板名称，可选值: order, product, product_list, logistics, price_comparison",
					Required: true,
				},
			},
		),
	}, nil
}

type getTemplateInput struct {
	TemplateName string `json:"template_name"`
}

func (t *GetUITemplateTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input getTemplateInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("parse arguments: %w", err)
	}

	return GenerateTemplateDetail(input.TemplateName), nil
}

// RenderUITool returns the render_ui tool
type RenderUITool struct{}

func (t *RenderUITool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "render_ui",
		Desc: "将业务数据渲染为前端可识别的 UI 组件代码块 (tgo-ui-widget)。调用前请确保已通过 get_ui_template 了解格式。",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"template_name": {
					Type:     schema.String,
					Desc:     "模板名称",
					Required: true,
				},
				"data": {
					Type:     schema.Object,
					Desc:     "要渲染的 JSON 数据对象，需符合模板定义的格式要求",
					Required: true,
				},
			},
		),
	}, nil
}

type renderUIInput struct {
	TemplateName string                 `json:"template_name"`
	Data         map[string]interface{} `json:"data"`
}

func (t *RenderUITool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input renderUIInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("parse arguments: %w", err)
	}

	result, err := RenderData(input.TemplateName, input.Data)
	if err != nil {
		return fmt.Sprintf("渲染错误: %s", err.Error()), nil
	}

	return result, nil
}

// ListUITemplatesTool returns the list_ui_templates tool
type ListUITemplatesTool struct{}

func (t *ListUITemplatesTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "list_ui_templates",
		Desc: "列出所有可用的 UI 模板及其简短描述，用于快速了解有哪些可用展示组件。",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{},
		),
	}, nil
}

func (t *ListUITemplatesTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	templates := GetAllTemplates()

	if len(templates) == 0 {
		return "暂无可用的 UI 模板", nil
	}

	var sb strings.Builder
	sb.WriteString("可用的 UI 模板:\n\n")

	for templateType, info := range templates {
		sb.WriteString(fmt.Sprintf("- **%s**: %s\n", templateType, info.Description))
	}

	return sb.String(), nil
}

// LoadUITemplateTools returns all UI template tools
func LoadUITemplateTools() []tool.BaseTool {
	return []tool.BaseTool{
		&GetUITemplateTool{},
		&RenderUITool{},
		&ListUITemplatesTool{},
	}
}
