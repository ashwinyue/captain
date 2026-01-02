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

// VisitorInfoTool 访客信息工具
type VisitorInfoTool struct {
	client    *apiserver.Client
	projectID string
	visitorID string
}

// NewVisitorInfoTool 创建访客信息工具
func NewVisitorInfoTool(client *apiserver.Client, projectID, visitorID string) *VisitorInfoTool {
	return &VisitorInfoTool{
		client:    client,
		projectID: projectID,
		visitorID: visitorID,
	}
}

// Info 返回工具信息
func (t *VisitorInfoTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "update_visitor_info",
		Desc: "当访客提供联系方式或个人信息时，调用此工具以记录或更新访客资料。可收集的信息包括：邮箱、微信、电话、姓名、性别、公司、职位、地址、生日等；所有字段均为可选，支持部分更新。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"email":    {Type: "string", Desc: "邮箱（可选）"},
			"wechat":   {Type: "string", Desc: "微信号（可选）"},
			"phone":    {Type: "string", Desc: "手机号（可选）"},
			"name":     {Type: "string", Desc: "姓名（可选）"},
			"sex":      {Type: "string", Desc: "性别（可选）"},
			"age":      {Type: "string", Desc: "年龄（可选）"},
			"company":  {Type: "string", Desc: "公司（可选）"},
			"position": {Type: "string", Desc: "职位（可选）"},
			"address":  {Type: "string", Desc: "地址（可选）"},
			"birthday": {Type: "string", Desc: "生日（可选）"},
		}),
	}, nil
}

type visitorInfoInput struct {
	Email    string `json:"email,omitempty"`
	Wechat   string `json:"wechat,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Name     string `json:"name,omitempty"`
	Sex      string `json:"sex,omitempty"`
	Age      string `json:"age,omitempty"`
	Company  string `json:"company,omitempty"`
	Position string `json:"position,omitempty"`
	Address  string `json:"address,omitempty"`
	Birthday string `json:"birthday,omitempty"`
}

// InvokableRun 执行工具
func (t *VisitorInfoTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input visitorInfoInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("parse arguments: %w", err)
	}

	// 收集提供的字段
	provided := make(map[string]string)
	if input.Email != "" {
		provided["email"] = input.Email
	}
	if input.Wechat != "" {
		provided["wechat"] = input.Wechat
	}
	if input.Phone != "" {
		provided["phone"] = input.Phone
	}
	if input.Name != "" {
		provided["name"] = input.Name
	}
	if input.Sex != "" {
		provided["sex"] = input.Sex
	}
	if input.Age != "" {
		provided["age"] = input.Age
	}
	if input.Company != "" {
		provided["company"] = input.Company
	}
	if input.Position != "" {
		provided["position"] = input.Position
	}
	if input.Address != "" {
		provided["address"] = input.Address
	}
	if input.Birthday != "" {
		provided["birthday"] = input.Birthday
	}

	if len(provided) == 0 {
		return "请至少提供一个需要更新的访客信息字段，例如邮箱、电话、微信、姓名等。", nil
	}

	log.Printf("[VisitorInfo] Updating visitor %s with fields: %v", t.visitorID, provided)

	// 调用 apiserver 更新访客信息
	if t.client != nil && t.visitorID != "" {
		visitorUUID, err := uuid.Parse(t.visitorID)
		if err != nil {
			log.Printf("[VisitorInfo] Invalid visitor ID: %v", err)
			return "访客 ID 格式错误。", nil
		}

		// 转换为 interface{} map
		payload := make(map[string]interface{})
		for k, v := range provided {
			payload[k] = v
		}

		_, err = t.client.SendVisitorInfoUpdate(ctx, visitorUUID, payload)
		if err != nil {
			log.Printf("[VisitorInfo] Update failed: %v", err)
			return "抱歉，访客信息更新失败，请稍后重试。", nil
		}
	}

	keys := make([]string, 0, len(provided))
	for k := range provided {
		keys = append(keys, k)
	}
	return fmt.Sprintf("已提交访客信息更新：%s。感谢配合！", strings.Join(keys, ", ")), nil
}

// GetVisitorInfoTool 获取访客信息工具
type GetVisitorInfoTool struct {
	client    *apiserver.Client
	projectID string
	visitorID string
}

// NewGetVisitorInfoTool 创建获取访客信息工具
func NewGetVisitorInfoTool(client *apiserver.Client, projectID, visitorID string) *GetVisitorInfoTool {
	return &GetVisitorInfoTool{
		client:    client,
		projectID: projectID,
		visitorID: visitorID,
	}
}

// Info 返回工具信息
func (t *GetVisitorInfoTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "get_visitor_info",
		Desc:        "获取当前访客的详细背景资料，包括姓名、联系方式、公司职位、标签画像、来源渠道及最近活动记录。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

// InvokableRun 执行工具
func (t *GetVisitorInfoTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if t.visitorID == "" {
		return "无法确定当前访客 ID，请确保访客已初始化。", nil
	}

	log.Printf("[VisitorInfo] Getting visitor info for %s", t.visitorID)

	if t.client != nil {
		info, err := t.client.GetVisitorInfo(ctx, t.projectID, t.visitorID)
		if err != nil {
			log.Printf("[VisitorInfo] Get failed: %v", err)
			return "抱歉，获取访客信息失败，请稍后重试。", nil
		}

		data, _ := json.MarshalIndent(info, "", "  ")
		return string(data), nil
	}

	return "访客信息服务未配置。", nil
}
