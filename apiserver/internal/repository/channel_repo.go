package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
)

type ChannelRepository struct {
	DB *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{DB: db}
}

func (r *ChannelRepository) Create(ctx context.Context, channel *model.Channel) error {
	return r.DB.WithContext(ctx).Create(channel).Error
}

func (r *ChannelRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.Channel, int64, error) {
	var channels []model.Channel
	var total int64

	query := r.DB.WithContext(ctx).Model(&model.Channel{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID)

	query.Count(&total)
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&channels).Error
	return channels, total, err
}

func (r *ChannelRepository) FindByChannelID(ctx context.Context, projectID uuid.UUID, channelID string) (*model.Channel, error) {
	var channel model.Channel
	err := r.DB.WithContext(ctx).
		Where("channel_id = ? AND project_id = ? AND deleted_at IS NULL", channelID, projectID).
		First(&channel).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (r *ChannelRepository) Delete(ctx context.Context, projectID uuid.UUID, channelID string) error {
	return r.DB.WithContext(ctx).
		Where("channel_id = ? AND project_id = ?", channelID, projectID).
		Delete(&model.Channel{}).Error
}

type ChannelMemberRepository struct {
	DB *gorm.DB
}

func NewChannelMemberRepository(db *gorm.DB) *ChannelMemberRepository {
	return &ChannelMemberRepository{DB: db}
}

func (r *ChannelMemberRepository) AddMembers(ctx context.Context, projectID uuid.UUID, channelID string, uids []string) error {
	members := make([]model.ChannelMember, len(uids))
	for i, uid := range uids {
		members[i] = model.ChannelMember{
			ProjectID: projectID,
			ChannelID: channelID,
			UID:       uid,
		}
	}
	return r.DB.WithContext(ctx).Create(&members).Error
}

func (r *ChannelMemberRepository) RemoveMembers(ctx context.Context, projectID uuid.UUID, channelID string, uids []string) error {
	return r.DB.WithContext(ctx).
		Where("project_id = ? AND channel_id = ? AND uid IN ?", projectID, channelID, uids).
		Delete(&model.ChannelMember{}).Error
}

func (r *ChannelMemberRepository) GetMembers(ctx context.Context, projectID uuid.UUID, channelID string) ([]model.ChannelMember, error) {
	var members []model.ChannelMember
	err := r.DB.WithContext(ctx).
		Where("project_id = ? AND channel_id = ? AND deleted_at IS NULL", projectID, channelID).
		Find(&members).Error
	return members, err
}
