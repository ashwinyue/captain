package model

import (
	"time"

	"github.com/google/uuid"
)

// ProjectOnboardingProgress tracks onboarding progress for a project
type ProjectOnboardingProgress struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID      uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"project_id"`
	Step1Completed bool       `gorm:"default:false" json:"step_1_completed"`
	Step2Completed bool       `gorm:"default:false" json:"step_2_completed"`
	Step3Completed bool       `gorm:"default:false" json:"step_3_completed"`
	Step4Completed bool       `gorm:"default:false" json:"step_4_completed"`
	Step5Completed bool       `gorm:"default:false" json:"step_5_completed"`
	IsCompleted    bool       `gorm:"default:false" json:"is_completed"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

func (ProjectOnboardingProgress) TableName() string {
	return "api_project_onboarding_progress"
}

// OnboardingStep defines an onboarding step
type OnboardingStep struct {
	StepNumber    int    `json:"step_number"`
	StepName      string `json:"step_name"`
	Description   string `json:"description"`
	DescriptionZh string `json:"description_zh"`
	Route         string `json:"route"`
	StepType      string `json:"step_type"`
	Title         string `json:"title,omitempty"`
	TitleZh       string `json:"title_zh,omitempty"`
}

// OnboardingSteps defines the onboarding steps
var OnboardingSteps = []OnboardingStep{
	{
		StepNumber:    1,
		StepName:      "configure_ai_provider",
		Description:   "Configure your AI provider (e.g., OpenAI, Anthropic)",
		DescriptionZh: "配置 AI 服务提供商（如 OpenAI、Anthropic）",
		Route:         "/settings/providers",
		StepType:      "action",
		Title:         "Configure AI Provider",
		TitleZh:       "配置 AI 服务商",
	},
	{
		StepNumber:    2,
		StepName:      "set_default_models",
		Description:   "Set default models for chat and embedding",
		DescriptionZh: "设置默认的对话和嵌入模型",
		Route:         "/settings/providers",
		StepType:      "action",
		Title:         "Set Default Models",
		TitleZh:       "设置默认模型",
	},
	{
		StepNumber:    3,
		StepName:      "create_knowledge_base",
		Description:   "Create your first knowledge base collection",
		DescriptionZh: "创建您的第一个知识库",
		Route:         "/knowledge",
		StepType:      "action",
		Title:         "Create Knowledge Base",
		TitleZh:       "创建知识库",
	},
	{
		StepNumber:    4,
		StepName:      "create_agent",
		Description:   "Create an AI agent with knowledge base",
		DescriptionZh: "创建关联知识库的 AI 助手",
		Route:         "/ai/agents",
		StepType:      "action",
		Title:         "Create AI Agent",
		TitleZh:       "创建 AI 助手",
	},
	{
		StepNumber:    5,
		StepName:      "start_chat",
		Description:   "Start your first conversation",
		DescriptionZh: "开始您的第一次对话",
		Route:         "/chat",
		StepType:      "notify",
		Title:         "Start Chatting",
		TitleZh:       "开始对话",
	},
}
