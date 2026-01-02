package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/pkg/wukongim"
	"github.com/tgo/captain/apiserver/internal/repository"
)

// TransferResult represents the result of a transfer operation
type TransferResult struct {
	Success         bool                       `json:"success"`
	AssignedStaffID *uuid.UUID                 `json:"assigned_staff_id,omitempty"`
	QueuePosition   *int                       `json:"queue_position,omitempty"`
	WaitingQueue    *model.VisitorWaitingQueue `json:"waiting_queue,omitempty"`
	Message         string                     `json:"message"`
}

// AssignmentSource represents the source of assignment
type AssignmentSource string

const (
	AssignmentSourceManual   AssignmentSource = "manual"
	AssignmentSourceRule     AssignmentSource = "rule"
	AssignmentSourceLLM      AssignmentSource = "llm"
	AssignmentSourceTransfer AssignmentSource = "transfer"
)

// TransferService handles visitor transfer to staff
type TransferService struct {
	staffRepo   *repository.StaffRepository
	visitorRepo *repository.VisitorRepository
	queueRepo   *repository.QueueRepository
	wkClient    *wukongim.Client
}

// NewTransferService creates a new transfer service
func NewTransferService(
	staffRepo *repository.StaffRepository,
	visitorRepo *repository.VisitorRepository,
	queueRepo *repository.QueueRepository,
	wkClient *wukongim.Client,
) *TransferService {
	return &TransferService{
		staffRepo:   staffRepo,
		visitorRepo: visitorRepo,
		queueRepo:   queueRepo,
		wkClient:    wkClient,
	}
}

// TransferRequest represents a transfer request
type TransferRequest struct {
	VisitorID           uuid.UUID
	ProjectID           uuid.UUID
	Source              AssignmentSource
	TargetStaffID       *uuid.UUID
	AddToQueueIfNoStaff bool
	AIDisabled          *bool
}

// TransferToStaff transfers a visitor to staff service
func (s *TransferService) TransferToStaff(ctx context.Context, req *TransferRequest) (*TransferResult, error) {
	// 1. Validate visitor exists
	visitor, err := s.visitorRepo.FindByIDAndProject(ctx, req.ProjectID, req.VisitorID)
	if err != nil {
		return &TransferResult{
			Success: false,
			Message: "Visitor not found",
		}, nil
	}

	// 2. Get available staff
	var assignedStaffID *uuid.UUID

	if req.TargetStaffID != nil {
		// Direct assignment to specific staff
		staff, err := s.staffRepo.FindByID(ctx, *req.TargetStaffID)
		if err == nil && staff.IsActive && !staff.ServicePaused {
			assignedStaffID = &staff.ID
		}
	} else {
		// Find available staff using load balancing
		availableStaff, err := s.staffRepo.FindAvailableStaff(ctx, req.ProjectID)
		if err == nil && len(availableStaff) > 0 {
			// Simple load balancing: pick first available staff
			// TODO: Add more sophisticated load balancing based on current chat count
			assignedStaffID = &availableStaff[0].ID
		}
	}

	// 3. Handle case when no staff is assigned
	if assignedStaffID == nil && req.AddToQueueIfNoStaff {
		// Add to waiting queue
		queueEntry, position, err := s.addToWaitingQueue(ctx, req.ProjectID, req.VisitorID, visitor)
		if err != nil {
			return &TransferResult{
				Success: false,
				Message: "Failed to add to waiting queue: " + err.Error(),
			}, nil
		}

		return &TransferResult{
			Success:       true,
			QueuePosition: &position,
			WaitingQueue:  queueEntry,
			Message:       "Added to waiting queue",
		}, nil
	}

	// 4. No staff and not adding to queue
	if assignedStaffID == nil {
		return &TransferResult{
			Success: false,
			Message: "No available staff",
		}, nil
	}

	// 5. Staff assigned - update visitor and add staff to channel
	channelID := BuildVisitorChannelID(visitor.ID)

	// Add staff as subscriber to the visitor's channel
	if s.wkClient != nil {
		staffUID := assignedStaffID.String()
		s.wkClient.AddSubscribers(ctx, &wukongim.SubscribersRequest{
			ChannelID:   channelID,
			ChannelType: 2, // Customer service channel
			Subscribers: []string{staffUID},
		})
	}

	// 6. Update AI disabled status if provided
	if req.AIDisabled != nil {
		visitor.AIEnabled = !*req.AIDisabled
		s.visitorRepo.Update(ctx, visitor)
	}

	// 7. Remove from waiting queue if exists
	if queueEntry, _ := s.queueRepo.FindByVisitorID(ctx, req.ProjectID, req.VisitorID); queueEntry != nil {
		s.queueRepo.Assign(ctx, req.ProjectID, queueEntry.ID, *assignedStaffID)
	}

	return &TransferResult{
		Success:         true,
		AssignedStaffID: assignedStaffID,
		Message:         "Transfer successful",
	}, nil
}

// addToWaitingQueue adds a visitor to the waiting queue
func (s *TransferService) addToWaitingQueue(ctx context.Context, projectID, visitorID uuid.UUID, visitor *model.Visitor) (*model.VisitorWaitingQueue, int, error) {
	// Check if already in queue
	existing, _ := s.queueRepo.FindByVisitorID(ctx, projectID, visitorID)
	if existing != nil {
		position, _ := s.queueRepo.GetPosition(ctx, projectID, existing.ID)
		return existing, int(position), nil
	}

	// Create new queue entry
	channelID := BuildVisitorChannelID(visitorID)
	now := time.Now()
	entry := &model.VisitorWaitingQueue{
		ProjectID:     projectID,
		VisitorID:     visitorID,
		ChannelID:     channelID,
		Status:        model.QueueStatusWaiting,
		Priority:      0,
		Source:        "transfer",
		WaitStartedAt: now,
	}

	if err := s.queueRepo.Create(ctx, entry); err != nil {
		return nil, 0, err
	}

	position, _ := s.queueRepo.GetPosition(ctx, projectID, entry.ID)
	return entry, int(position), nil
}

// AcceptVisitorFromQueue accepts a visitor from the queue
func (s *TransferService) AcceptVisitorFromQueue(ctx context.Context, projectID, queueEntryID, staffID uuid.UUID) (*TransferResult, error) {
	// Find queue entry
	entry, err := s.queueRepo.FindByID(ctx, projectID, queueEntryID)
	if err != nil {
		return &TransferResult{
			Success: false,
			Message: "Queue entry not found",
		}, nil
	}

	// Transfer the visitor
	return s.TransferToStaff(ctx, &TransferRequest{
		VisitorID:           entry.VisitorID,
		ProjectID:           projectID,
		Source:              AssignmentSourceManual,
		TargetStaffID:       &staffID,
		AddToQueueIfNoStaff: false,
	})
}

// GetProjectIDByAPIKey gets project ID by platform API key
func (s *TransferService) GetProjectIDByAPIKey(ctx context.Context, apiKey string) (uuid.UUID, error) {
	platform, err := s.visitorRepo.GetPlatformByAPIKey(ctx, apiKey)
	if err != nil {
		return uuid.Nil, err
	}
	return platform.ProjectID, nil
}
