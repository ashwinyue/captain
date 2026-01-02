package service

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/tgo/captain/rag/internal/config"
	"github.com/tgo/captain/rag/internal/model"
	"github.com/tgo/captain/rag/internal/repository"
)

type FileService struct {
	fileRepo     *repository.FileRepository
	documentRepo *repository.DocumentRepository
	cfg          *config.Config
}

func NewFileService(fileRepo *repository.FileRepository, documentRepo *repository.DocumentRepository, cfg *config.Config) *FileService {
	return &FileService{fileRepo: fileRepo, documentRepo: documentRepo, cfg: cfg}
}

func (s *FileService) List(ctx context.Context, projectID uuid.UUID, collectionID *uuid.UUID, status string, limit, offset int) ([]model.File, int64, error) {
	return s.fileRepo.FindByProjectID(ctx, projectID, collectionID, status, limit, offset)
}

func (s *FileService) GetByID(ctx context.Context, id uuid.UUID) (*model.File, error) {
	return s.fileRepo.FindByID(ctx, id)
}

func (s *FileService) Upload(ctx context.Context, projectID uuid.UUID, collectionID *uuid.UUID, filename string, contentType string, size int64, reader io.Reader) (*model.File, error) {
	// Generate unique storage path
	fileID := uuid.New()
	storagePath := filepath.Join(s.cfg.StoragePath, projectID.String(), fileID.String(), filename)

	// Create directory
	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		return nil, err
	}

	// Save file
	dst, err := os.Create(storagePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		return nil, err
	}

	// Create file record
	file := &model.File{
		ProjectID:    projectID,
		CollectionID: collectionID,
		FileName:     filename,
		OriginalName: filename,
		ContentType:  contentType,
		Size:         size,
		StoragePath:  storagePath,
		Status:       model.FileStatusPending,
	}
	file.ID = fileID

	if err := s.fileRepo.Create(ctx, file); err != nil {
		os.Remove(storagePath)
		return nil, err
	}

	// TODO: Trigger async processing

	return file, nil
}

func (s *FileService) Delete(ctx context.Context, id uuid.UUID) error {
	file, err := s.fileRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete associated documents
	s.documentRepo.DeleteByFileID(ctx, id)

	// Delete physical file
	if file.StoragePath != "" {
		os.Remove(file.StoragePath)
	}

	return s.fileRepo.Delete(ctx, id)
}

func (s *FileService) GetFilePath(ctx context.Context, id uuid.UUID) (string, error) {
	file, err := s.fileRepo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	return file.StoragePath, nil
}

func (s *FileService) UpdateStatus(ctx context.Context, id uuid.UUID, status model.FileStatus, errorMsg string) error {
	file, err := s.fileRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	file.Status = status
	if errorMsg != "" {
		file.ErrorMessage = errorMsg
	}
	if status == model.FileStatusCompleted {
		now := time.Now()
		file.ProcessedAt = &now
	}

	return s.fileRepo.Update(ctx, file)
}

// ListDocuments returns documents for a file
func (s *FileService) ListDocuments(ctx context.Context, fileID uuid.UUID, limit, offset int) ([]model.Document, int64, error) {
	return s.documentRepo.FindByFileID(ctx, fileID, limit, offset)
}
