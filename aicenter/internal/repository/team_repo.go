package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/aicenter/internal/model"
)

type TeamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) List(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.Team, int64, error) {
	var teams []model.Team
	var total int64

	query := r.db.WithContext(ctx).Where("project_id = ?", projectID)

	if err := query.Model(&model.Team{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&teams).Error; err != nil {
		return nil, 0, err
	}

	return teams, total, nil
}

func (r *TeamRepository) GetByID(ctx context.Context, projectID, teamID uuid.UUID) (*model.Team, error) {
	var team model.Team
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND id = ?", projectID, teamID).
		First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepository) GetWithAgents(ctx context.Context, projectID, teamID uuid.UUID) (*model.Team, error) {
	var team model.Team
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND id = ?", projectID, teamID).
		First(&team).Error
	if err != nil {
		return nil, err
	}

	// Load agents separately
	var agents []model.Agent
	r.db.WithContext(ctx).
		Where("team_id = ?", teamID).
		Preload("Tools").
		Preload("LLMProvider").
		Find(&agents)
	team.Agents = agents

	// Load supervisor LLM if set
	if team.SupervisorLLMID != nil {
		var llm model.LLMProvider
		if err := r.db.WithContext(ctx).First(&llm, team.SupervisorLLMID).Error; err == nil {
			team.SupervisorLLM = &llm
		}
	}

	return &team, nil
}

func (r *TeamRepository) GetDefault(ctx context.Context, projectID uuid.UUID) (*model.Team, error) {
	var team model.Team
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND is_default = ?", projectID, true).
		First(&team).Error
	if err != nil {
		return nil, err
	}

	// Load agents separately (Collections/Tools/LLMProvider have gorm:"-", load manually)
	var agents []model.Agent
	r.db.WithContext(ctx).
		Where("team_id = ?", team.ID).
		Find(&agents)

	// Load related data for each agent
	for i := range agents {
		// Load Tools
		var tools []model.AgentTool
		r.db.WithContext(ctx).Where("agent_id = ?", agents[i].ID).Find(&tools)
		agents[i].Tools = tools

		// Load Collections
		var collections []model.AgentCollection
		r.db.WithContext(ctx).Where("agent_id = ?", agents[i].ID).Find(&collections)
		agents[i].Collections = collections

		// Load LLMProvider
		if agents[i].LLMProviderID != nil {
			var provider model.LLMProvider
			if err := r.db.WithContext(ctx).First(&provider, agents[i].LLMProviderID).Error; err == nil {
				agents[i].LLMProvider = &provider
			}
		}
	}
	team.Agents = agents

	// Load supervisor LLM if set
	if team.SupervisorLLMID != nil {
		var llm model.LLMProvider
		if err := r.db.WithContext(ctx).First(&llm, team.SupervisorLLMID).Error; err == nil {
			team.SupervisorLLM = &llm
		}
	}

	return &team, nil
}

func (r *TeamRepository) Create(ctx context.Context, team *model.Team) error {
	return r.db.WithContext(ctx).Create(team).Error
}

func (r *TeamRepository) Update(ctx context.Context, team *model.Team) error {
	return r.db.WithContext(ctx).Save(team).Error
}

func (r *TeamRepository) Delete(ctx context.Context, projectID, teamID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND id = ?", projectID, teamID).
		Delete(&model.Team{}).Error
}
