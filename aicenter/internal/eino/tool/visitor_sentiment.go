package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"github.com/tgo/captain/aicenter/internal/pkg/apiserver"
)

// VisitorSentimentTool 访客情绪/状态工具
type VisitorSentimentTool struct {
	client    *apiserver.Client
	projectID string
	visitorID string
}

// NewVisitorSentimentTool 创建访客情绪工具
func NewVisitorSentimentTool(client *apiserver.Client, projectID, visitorID string) *VisitorSentimentTool {
	return &VisitorSentimentTool{
		client:    client,
		projectID: projectID,
		visitorID: visitorID,
	}
}

// Info 返回工具信息
func (t *VisitorSentimentTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "update_visitor_sentiment",
		Desc: "当你在对话中识别到访客满意度、情绪或意图发生变化时，调用此工具以记录/更新访客状态。可跟踪的信息包括：满意度（0-5数值）、情绪（0-5数值）、意图（如 purchase/inquiry/complaint/support）；字段均为可选，支持部分更新。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"satisfaction": {Type: "integer", Desc: "满意度，0-5数值（0=未知，1=非常不满意，5=非常满意）"},
			"emotion":      {Type: "integer", Desc: "情绪，0-5数值（0=未知，1=非常消极，5=非常积极）"},
			"intent":       {Type: "string", Desc: "意图（如 purchase/inquiry/complaint/support）"},
		}),
	}, nil
}

type visitorSentimentInput struct {
	Satisfaction interface{} `json:"satisfaction,omitempty"`
	Emotion      interface{} `json:"emotion,omitempty"`
	Intent       string      `json:"intent,omitempty"`
}

// parseScale 解析并验证 0-5 范围的数值
func parseScale(val interface{}) (int, error) {
	if val == nil {
		return -1, fmt.Errorf("nil value")
	}

	switch v := val.(type) {
	case float64:
		iv := int(v)
		if iv < 0 || iv > 5 {
			return -1, fmt.Errorf("value out of range 0-5")
		}
		return iv, nil
	case int:
		if v < 0 || v > 5 {
			return -1, fmt.Errorf("value out of range 0-5")
		}
		return v, nil
	case string:
		iv, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return -1, err
		}
		if iv < 0 || iv > 5 {
			return -1, fmt.Errorf("value out of range 0-5")
		}
		return iv, nil
	default:
		return -1, fmt.Errorf("unsupported type")
	}
}

// InvokableRun 执行工具
func (t *VisitorSentimentTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input visitorSentimentInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("parse arguments: %w", err)
	}

	provided := make(map[string]interface{})

	// 解析满意度
	if input.Satisfaction != nil {
		satisfaction, err := parseScale(input.Satisfaction)
		if err != nil {
			return "满意度和情绪的数值必须在0-5之间，0表示未知。", nil
		}
		provided["satisfaction"] = satisfaction
	}

	// 解析情绪
	if input.Emotion != nil {
		emotion, err := parseScale(input.Emotion)
		if err != nil {
			return "满意度和情绪的数值必须在0-5之间，0表示未知。", nil
		}
		provided["emotion"] = emotion
	}

	// 解析意图
	if input.Intent != "" {
		provided["intent"] = input.Intent
	}

	if len(provided) == 0 {
		return "请至少提供一个需要更新的访客状态字段，例如满意度、情绪或意图。", nil
	}

	log.Printf("[VisitorSentiment] Updating visitor %s with: %v", t.visitorID, provided)

	// 调用 apiserver 更新访客情绪
	if t.client != nil && t.visitorID != "" {
		visitorUUID, err := uuid.Parse(t.visitorID)
		if err != nil {
			log.Printf("[VisitorSentiment] Invalid visitor ID: %v", err)
			return "访客 ID 格式错误。", nil
		}

		_, err = t.client.SendVisitorSentimentUpdate(ctx, visitorUUID, provided)
		if err != nil {
			log.Printf("[VisitorSentiment] Update failed: %v", err)
			return "抱歉，访客状态更新未能成功提交。请稍后重试。", nil
		}
	}

	keys := make([]string, 0, len(provided))
	for k := range provided {
		keys = append(keys, k)
	}
	return fmt.Sprintf("已记录访客状态更新：%s。", strings.Join(keys, ", ")), nil
}
