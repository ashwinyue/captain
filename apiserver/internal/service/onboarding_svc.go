package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/pkg/aicenter"
	"github.com/tgo/captain/apiserver/internal/pkg/rag"
	"github.com/tgo/captain/apiserver/internal/repository"
)

type OnboardingService struct {
	repo      *repository.OnboardingRepository
	aiClient  *aicenter.Client
	ragClient *rag.Client
}

func NewOnboardingService(repo *repository.OnboardingRepository, aiClient *aicenter.Client, ragClient *rag.Client) *OnboardingService {
	return &OnboardingService{repo: repo, aiClient: aiClient, ragClient: ragClient}
}

// checkHasProvider checks if the project has at least one AI provider configured
func (s *OnboardingService) checkHasProvider(ctx context.Context, projectID uuid.UUID) bool {
	if s.aiClient == nil {
		return false
	}
	headers := map[string]string{"X-Project-ID": projectID.String()}
	data, status, err := s.aiClient.ListProviders(ctx, projectID.String(), headers)
	if err != nil || status >= 400 {
		return false
	}
	var resp struct {
		Data []interface{} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return false
	}
	return len(resp.Data) > 0
}

// checkHasDefaultModel checks if the project has default models configured
func (s *OnboardingService) checkHasDefaultModel(ctx context.Context, projectID uuid.UUID) bool {
	if s.aiClient == nil {
		return false
	}
	headers := map[string]string{"X-Project-ID": projectID.String()}
	data, status, err := s.aiClient.GetProjectAIConfig(ctx, projectID.String(), headers)
	if err != nil || status >= 400 {
		return false
	}
	var resp struct {
		DefaultChatProviderID      *string `json:"default_chat_provider_id"`
		DefaultChatModel           *string `json:"default_chat_model"`
		DefaultEmbeddingProviderID *string `json:"default_embedding_provider_id"`
		DefaultEmbeddingModel      *string `json:"default_embedding_model"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return false
	}
	// Consider step 2 done if at least chat model is configured
	return resp.DefaultChatProviderID != nil && resp.DefaultChatModel != nil &&
		*resp.DefaultChatProviderID != "" && *resp.DefaultChatModel != ""
}

// checkHasKnowledgeBase checks if the project has at least one knowledge base collection
func (s *OnboardingService) checkHasKnowledgeBase(ctx context.Context, projectID uuid.UUID) bool {
	if s.ragClient == nil {
		return false
	}
	data, status, err := s.ragClient.ListCollections(ctx, projectID.String(), nil)
	if err != nil || status >= 400 {
		return false
	}
	var resp struct {
		Data []interface{} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return false
	}
	return len(resp.Data) > 0
}

// checkHasAgent checks if the project has at least one AI agent configured
func (s *OnboardingService) checkHasAgent(ctx context.Context, projectID uuid.UUID) bool {
	if s.aiClient == nil {
		return false
	}
	headers := map[string]string{"X-Project-ID": projectID.String()}
	data, status, err := s.aiClient.ListAgents(ctx, projectID.String(), headers)
	if err != nil || status >= 400 {
		return false
	}
	var resp struct {
		Data []interface{} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return false
	}
	return len(resp.Data) > 0
}

// OnboardingStepStatus represents a step status
type OnboardingStepStatus struct {
	StepNumber    int    `json:"step_number"`
	StepName      string `json:"step_name"`
	IsCompleted   bool   `json:"is_completed"`
	Description   string `json:"description"`
	DescriptionZh string `json:"description_zh"`
	Route         string `json:"route"`
	StepType      string `json:"step_type"`
	Title         string `json:"title,omitempty"`
	TitleZh       string `json:"title_zh,omitempty"`
}

// OnboardingProgressResponse represents the progress response
type OnboardingProgressResponse struct {
	ID                 uuid.UUID              `json:"id"`
	ProjectID          uuid.UUID              `json:"project_id"`
	Steps              []OnboardingStepStatus `json:"steps"`
	CurrentStep        int                    `json:"current_step"`
	ProgressPercentage int                    `json:"progress_percentage"`
	IsCompleted        bool                   `json:"is_completed"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// GetProgress returns the onboarding progress for a project
func (s *OnboardingService) GetProgress(ctx context.Context, projectID uuid.UUID) (*OnboardingProgressResponse, error) {
	progress, err := s.repo.GetOrCreate(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Dynamically check actual configuration status
	step1Done := s.checkHasProvider(ctx, projectID)
	step2Done := s.checkHasDefaultModel(ctx, projectID)
	step3Done := s.checkHasKnowledgeBase(ctx, projectID)
	step4Done := s.checkHasAgent(ctx, projectID)

	// Update progress if changed
	if step1Done != progress.Step1Completed || step2Done != progress.Step2Completed ||
		step3Done != progress.Step3Completed || step4Done != progress.Step4Completed {
		progress.Step1Completed = step1Done
		progress.Step2Completed = step2Done
		progress.Step3Completed = step3Done
		progress.Step4Completed = step4Done
		// Check if all action steps are completed
		if progress.Step1Completed && progress.Step2Completed &&
			progress.Step3Completed && progress.Step4Completed {
			progress.IsCompleted = true
			now := time.Now()
			progress.CompletedAt = &now
		}
		_ = s.repo.Update(ctx, progress)
	}

	stepStatuses := []bool{
		progress.Step1Completed,
		progress.Step2Completed,
		progress.Step3Completed,
		progress.Step4Completed,
		progress.Step5Completed,
	}

	// Calculate current step (first incomplete action step)
	currentStep := 5
	for i := 0; i < 4; i++ {
		if !stepStatuses[i] {
			currentStep = i + 1
			break
		}
	}

	// Count completed action steps (1-4)
	completedCount := 0
	for i := 0; i < 4; i++ {
		if stepStatuses[i] {
			completedCount++
		}
	}

	// Build step status list
	steps := make([]OnboardingStepStatus, len(model.OnboardingSteps))
	for i, step := range model.OnboardingSteps {
		steps[i] = OnboardingStepStatus{
			StepNumber:    step.StepNumber,
			StepName:      step.StepName,
			IsCompleted:   stepStatuses[i],
			Description:   step.Description,
			DescriptionZh: step.DescriptionZh,
			Route:         step.Route,
			StepType:      step.StepType,
			Title:         step.Title,
			TitleZh:       step.TitleZh,
		}
	}

	progressPercentage := (completedCount * 100) / 4

	return &OnboardingProgressResponse{
		ID:                 progress.ID,
		ProjectID:          progress.ProjectID,
		Steps:              steps,
		CurrentStep:        currentStep,
		ProgressPercentage: progressPercentage,
		IsCompleted:        progress.IsCompleted,
		CompletedAt:        progress.CompletedAt,
		CreatedAt:          progress.CreatedAt,
		UpdatedAt:          progress.UpdatedAt,
	}, nil
}

// SkipStep skips a specific step or all steps
func (s *OnboardingService) SkipStep(ctx context.Context, projectID uuid.UUID, stepNumber *int) (*OnboardingProgressResponse, error) {
	progress, err := s.repo.GetOrCreate(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if stepNumber != nil {
		// Skip specific step
		switch *stepNumber {
		case 1:
			progress.Step1Completed = true
		case 2:
			progress.Step2Completed = true
		case 3:
			progress.Step3Completed = true
		case 4:
			progress.Step4Completed = true
		case 5:
			progress.Step5Completed = true
		}

		// Check if all action steps are completed
		if progress.Step1Completed && progress.Step2Completed &&
			progress.Step3Completed && progress.Step4Completed {
			progress.IsCompleted = true
			now := time.Now()
			progress.CompletedAt = &now
		}
	} else {
		// Skip all steps
		progress.Step1Completed = true
		progress.Step2Completed = true
		progress.Step3Completed = true
		progress.Step4Completed = true
		progress.Step5Completed = true
		progress.IsCompleted = true
		now := time.Now()
		progress.CompletedAt = &now
	}

	if err := s.repo.Update(ctx, progress); err != nil {
		return nil, err
	}

	return s.GetProgress(ctx, projectID)
}

// Reset resets the onboarding progress
func (s *OnboardingService) Reset(ctx context.Context, projectID uuid.UUID) (*OnboardingProgressResponse, error) {
	progress, err := s.repo.GetOrCreate(ctx, projectID)
	if err != nil {
		return nil, err
	}

	progress.Step1Completed = false
	progress.Step2Completed = false
	progress.Step3Completed = false
	progress.Step4Completed = false
	progress.Step5Completed = false
	progress.IsCompleted = false
	progress.CompletedAt = nil

	if err := s.repo.Update(ctx, progress); err != nil {
		return nil, err
	}

	return s.GetProgress(ctx, projectID)
}
