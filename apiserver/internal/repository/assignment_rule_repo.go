package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
)

type AssignmentRuleRepository struct {
	DB *gorm.DB
}

func NewAssignmentRuleRepository(db *gorm.DB) *AssignmentRuleRepository {
	return &AssignmentRuleRepository{DB: db}
}

func (r *AssignmentRuleRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID) (*model.AssignmentRule, error) {
	var rule model.AssignmentRule
	err := r.DB.WithContext(ctx).
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *AssignmentRuleRepository) Upsert(ctx context.Context, rule *model.AssignmentRule) error {
	return r.DB.WithContext(ctx).
		Where("project_id = ?", rule.ProjectID).
		Assign(rule).FirstOrCreate(rule).Error
}

func (r *AssignmentRuleRepository) Update(ctx context.Context, rule *model.AssignmentRule) error {
	return r.DB.WithContext(ctx).Save(rule).Error
}
