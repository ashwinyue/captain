package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/repository"
)

type TagService struct {
	tagRepo        *repository.TagRepository
	visitorTagRepo *repository.VisitorTagRepository
}

func NewTagService(tagRepo *repository.TagRepository, visitorTagRepo *repository.VisitorTagRepository) *TagService {
	return &TagService{tagRepo: tagRepo, visitorTagRepo: visitorTagRepo}
}

func (s *TagService) List(ctx context.Context, projectID uuid.UUID, category string, limit, offset int) ([]model.Tag, int64, error) {
	return s.tagRepo.FindByProjectID(ctx, projectID, category, limit, offset)
}

func (s *TagService) GetByID(ctx context.Context, projectID, tagID uuid.UUID) (*model.Tag, error) {
	return s.tagRepo.FindByIDAndProject(ctx, projectID, tagID)
}

func (s *TagService) Create(ctx context.Context, tag *model.Tag) error {
	return s.tagRepo.Create(ctx, tag)
}

func (s *TagService) Update(ctx context.Context, tag *model.Tag) error {
	return s.tagRepo.Update(ctx, tag)
}

func (s *TagService) Delete(ctx context.Context, projectID, tagID uuid.UUID) error {
	// Remove all visitor-tag associations first
	s.visitorTagRepo.RemoveTagFromVisitor(ctx, projectID, uuid.Nil, tagID)
	return s.tagRepo.Delete(ctx, tagID)
}

func (s *TagService) AddToVisitor(ctx context.Context, projectID, visitorID, tagID uuid.UUID) error {
	return s.visitorTagRepo.AddTagToVisitor(ctx, projectID, visitorID, tagID)
}

func (s *TagService) RemoveFromVisitor(ctx context.Context, projectID, visitorID, tagID uuid.UUID) error {
	return s.visitorTagRepo.RemoveTagFromVisitor(ctx, projectID, visitorID, tagID)
}

func (s *TagService) GetVisitorTags(ctx context.Context, projectID, visitorID uuid.UUID) ([]model.Tag, error) {
	return s.visitorTagRepo.GetVisitorTags(ctx, projectID, visitorID)
}
