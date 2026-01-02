package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/pkg/wukongim"
	"github.com/tgo/captain/apiserver/internal/repository"
)

type ChannelService struct {
	channelRepo *repository.ChannelRepository
	memberRepo  *repository.ChannelMemberRepository
	wkClient    *wukongim.Client
}

func NewChannelService(channelRepo *repository.ChannelRepository, memberRepo *repository.ChannelMemberRepository, wkClient *wukongim.Client) *ChannelService {
	return &ChannelService{channelRepo: channelRepo, memberRepo: memberRepo, wkClient: wkClient}
}

func (s *ChannelService) List(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.Channel, int64, error) {
	return s.channelRepo.FindByProjectID(ctx, projectID, limit, offset)
}

func (s *ChannelService) GetByID(ctx context.Context, projectID uuid.UUID, channelID string) (*model.Channel, error) {
	return s.channelRepo.FindByChannelID(ctx, projectID, channelID)
}

func (s *ChannelService) Create(ctx context.Context, channel *model.Channel, subscribers []string) error {
	if err := s.wkClient.CreateChannel(ctx, &wukongim.CreateChannelRequest{
		ChannelID:   channel.ChannelID,
		ChannelType: int(channel.ChannelType),
		Subscribers: subscribers,
	}); err != nil {
		return err
	}

	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return err
	}

	return s.memberRepo.AddMembers(ctx, channel.ProjectID, channel.ChannelID, subscribers)
}

func (s *ChannelService) Delete(ctx context.Context, projectID uuid.UUID, channelID string) error {
	return s.channelRepo.Delete(ctx, projectID, channelID)
}

func (s *ChannelService) AddMembers(ctx context.Context, projectID uuid.UUID, channelID string, uids []string) error {
	channel, err := s.GetByID(ctx, projectID, channelID)
	if err != nil {
		return err
	}

	if err := s.wkClient.AddSubscribers(ctx, &wukongim.SubscribersRequest{
		ChannelID:   channelID,
		ChannelType: int(channel.ChannelType),
		Subscribers: uids,
	}); err != nil {
		return err
	}

	return s.memberRepo.AddMembers(ctx, projectID, channelID, uids)
}

func (s *ChannelService) RemoveMembers(ctx context.Context, projectID uuid.UUID, channelID string, uids []string) error {
	channel, err := s.GetByID(ctx, projectID, channelID)
	if err != nil {
		return err
	}

	if err := s.wkClient.RemoveSubscribers(ctx, &wukongim.SubscribersRequest{
		ChannelID:   channelID,
		ChannelType: int(channel.ChannelType),
		Subscribers: uids,
	}); err != nil {
		return err
	}

	return s.memberRepo.RemoveMembers(ctx, projectID, channelID, uids)
}
