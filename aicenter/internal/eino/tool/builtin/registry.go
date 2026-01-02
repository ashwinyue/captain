package builtin

import (
	"github.com/cloudwego/eino/components/tool"
)

// ToolType represents the type of built-in tool
type ToolType string

const (
	ToolWebSearch  ToolType = "web_search"
	ToolCalculator ToolType = "calculator"
	ToolDateTime   ToolType = "datetime"
)

// GetBuiltinTools returns built-in tools by their types
func GetBuiltinTools(types ...ToolType) []tool.BaseTool {
	tools := make([]tool.BaseTool, 0, len(types))

	for _, t := range types {
		switch t {
		case ToolWebSearch:
			tools = append(tools, NewDuckDuckGoSearchTool())
		case ToolCalculator:
			tools = append(tools, NewCalculatorTool())
		case ToolDateTime:
			tools = append(tools, NewDateTimeTool())
		}
	}

	return tools
}

// GetAllBuiltinTools returns all available built-in tools
func GetAllBuiltinTools() []tool.BaseTool {
	return []tool.BaseTool{
		NewDuckDuckGoSearchTool(),
		NewCalculatorTool(),
		NewDateTimeTool(),
	}
}

// GetBuiltinToolNames returns all available built-in tool names
func GetBuiltinToolNames() []string {
	return []string{
		string(ToolWebSearch),
		string(ToolCalculator),
		string(ToolDateTime),
	}
}
