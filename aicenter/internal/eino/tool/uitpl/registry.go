package uitpl

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TemplateInfo contains metadata about a template
type TemplateInfo struct {
	Type        TemplateType
	Description string
	Example     map[string]interface{}
}

// Registry manages UI template registration and lookup
type Registry struct {
	templates map[TemplateType]TemplateInfo
}

// defaultRegistry is the global template registry
var defaultRegistry *Registry

func init() {
	defaultRegistry = NewRegistry()
	defaultRegistry.registerBuiltinTemplates()
}

// NewRegistry creates a new template registry
func NewRegistry() *Registry {
	return &Registry{
		templates: make(map[TemplateType]TemplateInfo),
	}
}

// registerBuiltinTemplates registers all built-in templates
func (r *Registry) registerBuiltinTemplates() {
	templates := []Template{
		&OrderTemplate{},
		&ProductTemplate{},
		&ProductListTemplate{},
		&LogisticsTemplate{},
		&PriceComparisonTemplate{},
	}

	for _, t := range templates {
		r.Register(t)
	}
}

// Register adds a template to the registry
func (r *Registry) Register(t Template) {
	r.templates[t.GetType()] = TemplateInfo{
		Type:        t.GetType(),
		Description: t.GetDescription(),
		Example:     t.GetExample(),
	}
}

// Get returns template info by type
func (r *Registry) Get(templateType TemplateType) (TemplateInfo, bool) {
	info, ok := r.templates[templateType]
	return info, ok
}

// List returns all registered template types
func (r *Registry) List() []TemplateType {
	types := make([]TemplateType, 0, len(r.templates))
	for t := range r.templates {
		types = append(types, t)
	}
	return types
}

// GetAll returns all registered templates
func (r *Registry) GetAll() map[TemplateType]TemplateInfo {
	return r.templates
}

// GetTemplate returns template info from default registry
func GetTemplate(templateType string) (TemplateInfo, bool) {
	return defaultRegistry.Get(TemplateType(templateType))
}

// ListTemplates returns all template types from default registry
func ListTemplates() []TemplateType {
	return defaultRegistry.List()
}

// GetAllTemplates returns all templates from default registry
func GetAllTemplates() map[TemplateType]TemplateInfo {
	return defaultRegistry.GetAll()
}

// ValidateData validates data against a template schema
func ValidateData(templateType string, data map[string]interface{}) error {
	info, ok := GetTemplate(templateType)
	if !ok {
		return fmt.Errorf("unknown template: %s", templateType)
	}

	// Basic validation - check required fields based on template type
	switch info.Type {
	case TemplateOrder:
		if _, ok := data["order_id"]; !ok {
			return fmt.Errorf("missing required field: order_id")
		}
		if _, ok := data["status"]; !ok {
			return fmt.Errorf("missing required field: status")
		}
	case TemplateProduct:
		if _, ok := data["product_id"]; !ok {
			return fmt.Errorf("missing required field: product_id")
		}
		if _, ok := data["name"]; !ok {
			return fmt.Errorf("missing required field: name")
		}
	case TemplateProductList:
		if _, ok := data["products"]; !ok {
			return fmt.Errorf("missing required field: products")
		}
	case TemplateLogistics:
		if _, ok := data["tracking_no"]; !ok {
			return fmt.Errorf("missing required field: tracking_no")
		}
	case TemplatePriceComparison:
		if _, ok := data["product_name"]; !ok {
			return fmt.Errorf("missing required field: product_name")
		}
		if _, ok := data["prices"]; !ok {
			return fmt.Errorf("missing required field: prices")
		}
	}

	return nil
}

// RenderData renders data as a tgo-ui-widget markdown block
func RenderData(templateType string, data map[string]interface{}) (string, error) {
	if err := ValidateData(templateType, data); err != nil {
		return "", err
	}

	// Ensure type field is set
	data["type"] = templateType

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	return fmt.Sprintf("```tgo-ui-widget\n%s\n```", string(jsonBytes)), nil
}

// GenerateTemplateDetail generates detailed documentation for a template
func GenerateTemplateDetail(templateType string) string {
	info, ok := GetTemplate(templateType)
	if !ok {
		available := make([]string, 0)
		for _, t := range ListTemplates() {
			available = append(available, string(t))
		}
		return fmt.Sprintf("未知模板 '%s'。可用模板: %s", templateType, strings.Join(available, ", "))
	}

	exampleJSON, _ := json.MarshalIndent(info.Example, "", "  ")

	return fmt.Sprintf(`## %s 模板

**描述**: %s

### 示例数据:
%sjson
%s
%s

### 使用说明:
1. 准备符合上述格式的数据
2. 调用 render_ui 工具渲染为 UI 组件
3. 将返回的 tgo-ui-widget 代码块包含在回复中`,
		templateType,
		info.Description,
		"```",
		string(exampleJSON),
		"```",
	)
}
