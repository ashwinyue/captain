package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/repository"
)

type AssignmentRuleService struct {
	repo *repository.AssignmentRuleRepository
}

func NewAssignmentRuleService(repo *repository.AssignmentRuleRepository) *AssignmentRuleService {
	return &AssignmentRuleService{repo: repo}
}

const DefaultAssignmentPrompt = `You are a customer service routing assistant. Based on the visitor's message and available staff members, determine the best staff member to handle this conversation.

Consider:
1. Staff member expertise and specialization
2. Current workload (prefer staff with fewer active chats)
3. Language compatibility
4. Previous interaction history

Respond with the staff member ID that should handle this conversation.`

func (s *AssignmentRuleService) Get(ctx context.Context, projectID uuid.UUID) (*model.AssignmentRule, error) {
	rule, err := s.repo.FindByProjectID(ctx, projectID)
	if err == gorm.ErrRecordNotFound {
		// Return default rule
		return &model.AssignmentRule{
			ProjectID:            projectID,
			Name:                 "Default Rule",
			IsEnabled:            false,
			LLMAssignmentEnabled: false,
			Timezone:             "UTC",
			ServiceWeekdays:      []int{1, 2, 3, 4, 5},
			ServiceStartTime:     "09:00",
			ServiceEndTime:       "18:00",
			MaxConcurrentChats:   5,
			AutoCloseHours:       24,
		}, nil
	}
	return rule, err
}

func (s *AssignmentRuleService) Upsert(ctx context.Context, projectID uuid.UUID, updates map[string]interface{}) (*model.AssignmentRule, error) {
	rule, err := s.repo.FindByProjectID(ctx, projectID)

	if err == gorm.ErrRecordNotFound {
		// Create new rule
		rule = &model.AssignmentRule{
			ProjectID:            projectID,
			Name:                 "Default Rule",
			IsEnabled:            true,
			LLMAssignmentEnabled: false,
			Timezone:             "UTC",
			ServiceWeekdays:      []int{1, 2, 3, 4, 5},
			ServiceStartTime:     "09:00",
			ServiceEndTime:       "18:00",
			MaxConcurrentChats:   5,
			AutoCloseHours:       24,
		}
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "ai_provider_id":
			if v, ok := value.(*uuid.UUID); ok {
				rule.AIProviderID = v
			}
		case "model":
			if v, ok := value.(string); ok {
				rule.Model = v
			}
		case "prompt":
			if v, ok := value.(string); ok {
				rule.Prompt = v
			}
		case "llm_assignment_enabled":
			if v, ok := value.(bool); ok {
				rule.LLMAssignmentEnabled = v
			}
		case "timezone":
			if v, ok := value.(string); ok {
				rule.Timezone = v
			}
		case "service_weekdays":
			if v, ok := value.([]int); ok {
				rule.ServiceWeekdays = v
			}
		case "service_start_time":
			if v, ok := value.(string); ok {
				rule.ServiceStartTime = v
			}
		case "service_end_time":
			if v, ok := value.(string); ok {
				rule.ServiceEndTime = v
			}
		case "max_concurrent_chats":
			if v, ok := value.(int); ok {
				rule.MaxConcurrentChats = v
			}
		case "auto_close_hours":
			if v, ok := value.(int); ok {
				rule.AutoCloseHours = v
			}
		}
	}

	rule.UpdatedAt = time.Now()

	err = s.repo.Upsert(ctx, rule)
	return rule, err
}

func (s *AssignmentRuleService) GetDefaultPrompt() string {
	return DefaultAssignmentPrompt
}
