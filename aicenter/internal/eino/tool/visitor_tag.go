package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"github.com/tgo/captain/aicenter/internal/pkg/apiserver"
)

// VisitorTagTool 访客标签工具
type VisitorTagTool struct {
	client    *apiserver.Client
	projectID string
	visitorID string
}

// NewVisitorTagTool 创建访客标签工具
func NewVisitorTagTool(client *apiserver.Client, projectID, visitorID string) *VisitorTagTool {
	return &VisitorTagTool{
		client:    client,
		projectID: projectID,
		visitorID: visitorID,
	}
}

// Info 返回工具信息
func (t *VisitorTagTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "add_visitor_tag",
		Desc: "当你从对话中识别出访客特征或需要对访客进行分类时，调用此工具为访客添加标签。标签可用于后续的访客分析和个性化服务。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"tags": {
				Type: "array",
				Desc: "要添加的标签列表，每个标签包含 name（标签名称）和可选的 value（标签值）",
			},
		}),
	}, nil
}

type visitorTagInput struct {
	Tags []struct {
		Name  string `json:"name"`
		Value string `json:"value,omitempty"`
	} `json:"tags"`
}

// InvokableRun 执行工具
func (t *VisitorTagTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input visitorTagInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("parse arguments: %w", err)
	}

	if len(input.Tags) == 0 {
		return "请提供至少一个标签。", nil
	}

	// 验证标签
	var tagNames []string
	tags := make([]map[string]string, 0, len(input.Tags))
	for _, tag := range input.Tags {
		if tag.Name == "" {
			continue
		}
		tagMap := map[string]string{"name": tag.Name}
		if tag.Value != "" {
			tagMap["value"] = tag.Value
		}
		tags = append(tags, tagMap)
		tagNames = append(tagNames, tag.Name)
	}

	if len(tags) == 0 {
		return "标签名称不能为空。", nil
	}

	log.Printf("[VisitorTag] Adding tags to visitor %s: %v", t.visitorID, tagNames)

	// 调用 apiserver 添加标签
	if t.client != nil && t.visitorID != "" {
		visitorUUID, err := uuid.Parse(t.visitorID)
		if err != nil {
			log.Printf("[VisitorTag] Invalid visitor ID: %v", err)
			return "访客 ID 格式错误。", nil
		}

		_, err = t.client.SendVisitorTagAdd(ctx, visitorUUID, tags)
		if err != nil {
			log.Printf("[VisitorTag] Add failed: %v", err)
			return "抱歉，标签添加未能成功提交。请稍后重试。", nil
		}
	}

	return fmt.Sprintf("已为访客添加标签：%s。", strings.Join(tagNames, ", ")), nil
}
