package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/aicenter/internal/model"
)

type AgentRepository struct {
	db *gorm.DB
}

func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) List(ctx context.Context, projectID uuid.UUID, opts ...ListOption) ([]model.Agent, int64, error) {
	var agents []model.Agent
	var total int64

	query := r.db.WithContext(ctx).Where("project_id = ?", projectID)

	// Apply options
	o := &listOptions{}
	for _, opt := range opts {
		opt(o)
	}

	if o.teamID != nil {
		query = query.Where("team_id = ?", *o.teamID)
	}

	// Count total
	if err := query.Model(&model.Agent{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if o.limit > 0 {
		query = query.Limit(o.limit)
	}
	if o.offset > 0 {
		query = query.Offset(o.offset)
	}

	if err := query.Find(&agents).Error; err != nil {
		return nil, 0, err
	}

	// Load relations separately for each agent
	for i := range agents {
		var tools []model.AgentTool
		r.db.WithContext(ctx).Where("agent_id = ?", agents[i].ID).Find(&tools)
		agents[i].Tools = tools

		var collections []model.AgentCollection
		r.db.WithContext(ctx).Where("agent_id = ?", agents[i].ID).Find(&collections)
		agents[i].Collections = collections

		if agents[i].LLMProviderID != nil {
			var llm model.LLMProvider
			if err := r.db.WithContext(ctx).First(&llm, agents[i].LLMProviderID).Error; err == nil {
				agents[i].LLMProvider = &llm
			}
		}
	}

	return agents, total, nil
}

func (r *AgentRepository) GetByID(ctx context.Context, projectID, agentID uuid.UUID) (*model.Agent, error) {
	var agent model.Agent
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND id = ?", projectID, agentID).
		First(&agent).Error
	if err != nil {
		return nil, err
	}

	// Load relations separately
	var tools []model.AgentTool
	r.db.WithContext(ctx).Where("agent_id = ?", agent.ID).Find(&tools)
	agent.Tools = tools

	var collections []model.AgentCollection
	r.db.WithContext(ctx).Where("agent_id = ?", agent.ID).Find(&collections)
	agent.Collections = collections

	if agent.LLMProviderID != nil {
		var llm model.LLMProvider
		if err := r.db.WithContext(ctx).First(&llm, agent.LLMProviderID).Error; err == nil {
			agent.LLMProvider = &llm
		}
	}

	if agent.TeamID != nil {
		var team model.Team
		if err := r.db.WithContext(ctx).First(&team, agent.TeamID).Error; err == nil {
			agent.Team = &team
		}
	}

	return &agent, nil
}

func (r *AgentRepository) Create(ctx context.Context, agent *model.Agent) error {
	return r.db.WithContext(ctx).Create(agent).Error
}

func (r *AgentRepository) Update(ctx context.Context, agent *model.Agent) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Save main agent fields
		if err := tx.Save(agent).Error; err != nil {
			return err
		}

		// Handle Collections association
		if agent.Collections != nil {
			// Delete existing collections
			if err := tx.Where("agent_id = ?", agent.ID).Delete(&model.AgentCollection{}).Error; err != nil {
				return err
			}
			// Create new collections with fresh IDs
			for _, coll := range agent.Collections {
				newColl := model.AgentCollection{
					AgentID:      agent.ID,
					CollectionID: coll.CollectionID,
					IsEnabled:    coll.IsEnabled,
				}
				if err := tx.Create(&newColl).Error; err != nil {
					return err
				}
			}
		}

		// Handle Tools association
		if agent.Tools != nil {
			// Delete existing tools
			if err := tx.Where("agent_id = ?", agent.ID).Delete(&model.AgentTool{}).Error; err != nil {
				return err
			}
			// Create new tools with fresh IDs
			for _, tool := range agent.Tools {
				newTool := model.AgentTool{
					AgentID:      agent.ID,
					ToolProvider: tool.ToolProvider,
					ToolName:     tool.ToolName,
					IsEnabled:    tool.IsEnabled,
					Config:       tool.Config,
				}
				if err := tx.Create(&newTool).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *AgentRepository) Delete(ctx context.Context, projectID, agentID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND id = ?", projectID, agentID).
		Delete(&model.Agent{}).Error
}

func (r *AgentRepository) Exists(ctx context.Context, projectID uuid.UUID) (bool, int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Agent{}).
		Where("project_id = ?", projectID).
		Count(&count).Error
	return count > 0, count, err
}

func (r *AgentRepository) SetToolEnabled(ctx context.Context, projectID, agentID, toolID uuid.UUID, enabled bool) error {
	return r.db.WithContext(ctx).
		Model(&model.AgentTool{}).
		Where("agent_id = ? AND id = ?", agentID, toolID).
		Update("is_enabled", enabled).Error
}

func (r *AgentRepository) SetCollectionEnabled(ctx context.Context, projectID, agentID uuid.UUID, collectionID string, enabled bool) error {
	return r.db.WithContext(ctx).
		Model(&model.AgentCollection{}).
		Where("agent_id = ? AND collection_id = ?", agentID, collectionID).
		Update("is_enabled", enabled).Error
}

// List options
type listOptions struct {
	teamID *uuid.UUID
	limit  int
	offset int
}

type ListOption func(*listOptions)

func WithTeamID(teamID uuid.UUID) ListOption {
	return func(o *listOptions) {
		o.teamID = &teamID
	}
}

func WithPagination(limit, offset int) ListOption {
	return func(o *listOptions) {
		o.limit = limit
		o.offset = offset
	}
}
