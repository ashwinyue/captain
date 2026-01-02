package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// DateTimeTool provides current date/time information
type DateTimeTool struct {
	toolInfo *schema.ToolInfo
}

func NewDateTimeTool() *DateTimeTool {
	return &DateTimeTool{
		toolInfo: &schema.ToolInfo{
			Name: "get_current_time",
			Desc: "Get the current date and time. Optionally specify a timezone.",
			ParamsOneOf: schema.NewParamsOneOfByParams(
				map[string]*schema.ParameterInfo{
					"timezone": {
						Type: schema.String,
						Desc: "Timezone name (e.g., 'Asia/Shanghai', 'America/New_York'). Defaults to UTC.",
					},
				},
			),
		},
	}
}

func (t *DateTimeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return t.toolInfo, nil
}

type datetimeInput struct {
	Timezone string `json:"timezone"`
}

func (t *DateTimeTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input datetimeInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("parse arguments: %w", err)
	}

	loc := time.UTC
	if input.Timezone != "" {
		var err error
		loc, err = time.LoadLocation(input.Timezone)
		if err != nil {
			return "", fmt.Errorf("invalid timezone: %w", err)
		}
	}

	now := time.Now().In(loc)

	return fmt.Sprintf(
		"Current time: %s\nTimezone: %s\nUnix timestamp: %d\nDay of week: %s",
		now.Format("2006-01-02 15:04:05 MST"),
		loc.String(),
		now.Unix(),
		now.Weekday().String(),
	), nil
}
