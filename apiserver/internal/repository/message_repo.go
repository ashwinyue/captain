package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
)

type MessageRepository struct {
	DB *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{DB: db}
}

func (r *MessageRepository) Create(ctx context.Context, msg *model.Message) error {
	return r.DB.WithContext(ctx).Create(msg).Error
}

func (r *MessageRepository) FindByChannel(ctx context.Context, projectID uuid.UUID, channelID string, limit, offset int) ([]model.Message, int64, error) {
	var messages []model.Message
	var total int64

	query := r.DB.WithContext(ctx).Model(&model.Message{}).
		Where("project_id = ? AND channel_id = ? AND deleted_at IS NULL", projectID, channelID)

	query.Count(&total)
	err := query.Order("sent_at DESC").Limit(limit).Offset(offset).Find(&messages).Error
	return messages, total, err
}

func (r *MessageRepository) Revoke(ctx context.Context, projectID uuid.UUID, messageID string) error {
	return r.DB.WithContext(ctx).Model(&model.Message{}).
		Where("project_id = ? AND message_id = ?", projectID, messageID).
		Update("is_revoked", true).Error
}

type ConversationRepository struct {
	DB *gorm.DB
}

func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{DB: db}
}

func (r *ConversationRepository) Upsert(ctx context.Context, conv *model.Conversation) error {
	return r.DB.WithContext(ctx).
		Where("project_id = ? AND uid = ? AND channel_id = ?", conv.ProjectID, conv.UID, conv.ChannelID).
		Assign(conv).FirstOrCreate(conv).Error
}

func (r *ConversationRepository) FindByUID(ctx context.Context, projectID uuid.UUID, uid string, limit, offset int) ([]model.Conversation, int64, error) {
	var conversations []model.Conversation
	var total int64

	query := r.DB.WithContext(ctx).Model(&model.Conversation{}).
		Where("project_id = ? AND uid = ? AND deleted_at IS NULL", projectID, uid)

	query.Count(&total)
	err := query.Order("last_message_at DESC").Limit(limit).Offset(offset).Find(&conversations).Error
	return conversations, total, err
}

func (r *ConversationRepository) UpdateLastMessage(ctx context.Context, projectID uuid.UUID, uid, channelID, messageID, content string) error {
	now := time.Now()
	return r.DB.WithContext(ctx).Model(&model.Conversation{}).
		Where("project_id = ? AND uid = ? AND channel_id = ?", projectID, uid, channelID).
		Updates(map[string]interface{}{
			"last_message_id": messageID,
			"last_message":    content,
			"last_message_at": now,
		}).Error
}

func (r *ConversationRepository) FindAll(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.Conversation, int64, error) {
	var conversations []model.Conversation
	var total int64

	query := r.DB.WithContext(ctx).Model(&model.Conversation{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID)

	query.Count(&total)
	err := query.Order("last_message_at DESC").Limit(limit).Offset(offset).Find(&conversations).Error
	return conversations, total, err
}

func (r *ConversationRepository) FindRecent(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.Conversation, int64, error) {
	var conversations []model.Conversation
	var total int64

	query := r.DB.WithContext(ctx).Model(&model.Conversation{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID)

	query.Count(&total)
	err := query.Order("last_message_at DESC").Limit(limit).Offset(offset).Find(&conversations).Error
	return conversations, total, err
}
