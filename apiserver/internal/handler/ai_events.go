package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/pkg/wukongim"
)

// Event type constants
const (
	ManualServiceEvent    = "manual_service.request"
	VisitorInfoEvent      = "visitor_info.update"
	VisitorSentimentEvent = "visitor_sentiment.update"
	VisitorTagEvent       = "visitor_tag.add"
)

// Manual service tag constants
const (
	ManualServiceTagID     = "00000000-0000-0000-0000-000000000001"
	ManualServiceTagName   = "Manual Service"
	ManualServiceTagNameZh = "转人工"
)

// AIServiceEvent represents an event from AI service
type AIServiceEvent struct {
	EventType string                 `json:"event_type" binding:"required"`
	VisitorID *uuid.UUID             `json:"visitor_id"`
	Payload   map[string]interface{} `json:"payload"`
}

// AIEventsHandler handles internal AI events
type AIEventsHandler struct {
	db       *gorm.DB
	wkClient *wukongim.Client
}

// NewAIEventsHandler creates a new AI events handler
func NewAIEventsHandler(db *gorm.DB, wkClient *wukongim.Client) *AIEventsHandler {
	return &AIEventsHandler{db: db, wkClient: wkClient}
}

// IngestEvent handles AI service events (internal endpoint, no auth required)
func (h *AIEventsHandler) IngestEvent(c *gin.Context) {
	var event AIServiceEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate visitor_id is provided
	if event.VisitorID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "visitor_id is required in event payload"})
		return
	}

	// Query visitor to get project
	var visitor model.Visitor
	if err := h.db.WithContext(c.Request.Context()).
		Where("id = ? AND deleted_at IS NULL", event.VisitorID).
		First(&visitor).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Visitor not found: " + event.VisitorID.String()})
		return
	}

	// Get project from visitor
	var project model.Project
	if err := h.db.WithContext(c.Request.Context()).
		Where("id = ? AND deleted_at IS NULL", visitor.ProjectID).
		First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Project not found for visitor"})
		return
	}

	eventType := strings.TrimSpace(strings.ToLower(event.EventType))

	// Handle different event types
	switch eventType {
	case ManualServiceEvent:
		result, err := h.handleManualServiceRequest(c, &event, &project, &visitor)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"event_type": eventType, "result": result})

	case VisitorInfoEvent:
		result, err := h.handleVisitorInfoUpdate(c, &event, &project, &visitor)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"event_type": eventType, "result": result})

	case VisitorSentimentEvent:
		result, err := h.handleVisitorSentimentUpdate(c, &event, &project, &visitor)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"event_type": eventType, "result": result})

	case VisitorTagEvent:
		result, err := h.handleVisitorTag(c, &event, &project, &visitor)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"event_type": eventType, "result": result})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Unsupported AI event_type: " + event.EventType})
	}
}

// handleManualServiceRequest handles manual service request events
func (h *AIEventsHandler) handleManualServiceRequest(c *gin.Context, event *AIServiceEvent, project *model.Project, visitor *model.Visitor) (map[string]interface{}, error) {
	ctx := c.Request.Context()

	// Ensure manual service tag exists and is applied to visitor
	h.ensureManualServiceTag(ctx, project.ID, visitor)

	// Get reason from payload
	reason := ""
	if event.Payload != nil {
		if r, ok := event.Payload["reason"].(string); ok {
			reason = strings.TrimSpace(r)
		}
	}

	// Find available staff
	var staffs []struct {
		ID uuid.UUID
	}
	if err := h.db.WithContext(ctx).Table("staff").
		Where("project_id = ? AND is_active = true AND service_paused = false AND deleted_at IS NULL", project.ID).
		Find(&staffs).Error; err != nil || len(staffs) == 0 {
		// No available staff, add to queue
		channelID := "cs_" + visitor.ID.String()
		queueEntry := &model.VisitorWaitingQueue{
			ProjectID:     project.ID,
			VisitorID:     visitor.ID,
			ChannelID:     channelID,
			Status:        model.QueueStatusWaiting,
			Source:        reason,
			WaitStartedAt: time.Now(),
		}
		if err := h.db.WithContext(ctx).Create(queueEntry).Error; err != nil {
			return nil, err
		}

		// Get queue position
		var position int64
		h.db.WithContext(ctx).Model(&model.VisitorWaitingQueue{}).
			Where("project_id = ? AND status = ? AND wait_started_at < ? AND deleted_at IS NULL",
				project.ID, model.QueueStatusWaiting, queueEntry.WaitStartedAt).
			Count(&position)

		return map[string]interface{}{
			"entry_id":     queueEntry.ID.String(),
			"status":       string(queueEntry.Status),
			"position":     position + 1,
			"channel_id":   channelID,
			"channel_type": 2,
			"message":      "Added to waiting queue",
		}, nil
	}

	// Assign first available staff
	assignedStaffID := staffs[0].ID
	channelID := "cs_" + visitor.ID.String()

	// Add staff as subscriber to the channel
	if h.wkClient != nil {
		staffUID := assignedStaffID.String()
		h.wkClient.AddSubscribers(ctx, &wukongim.SubscribersRequest{
			ChannelID:   channelID,
			ChannelType: 2,
			Subscribers: []string{staffUID},
		})
	}

	// Update visitor AI status (disable AI after manual service request)
	h.db.WithContext(ctx).Model(&model.Visitor{}).
		Where("id = ?", visitor.ID).
		Update("ai_enabled", false)

	return map[string]interface{}{
		"assigned_staff_id": assignedStaffID.String(),
		"channel_id":        channelID,
		"channel_type":      2,
		"message":           "Transfer successful",
	}, nil
}

// ensureManualServiceTag ensures the visitor has the manual service tag
func (h *AIEventsHandler) ensureManualServiceTag(ctx interface{}, projectID uuid.UUID, visitor *model.Visitor) {
	tagID, _ := uuid.Parse(ManualServiceTagID)

	// Check if tag exists
	var tag model.Tag
	err := h.db.Where("id = ? AND project_id = ?", tagID, projectID).First(&tag).Error
	if err != nil {
		// Create manual service tag
		tag = model.Tag{
			ProjectID:   projectID,
			Name:        ManualServiceTagName,
			Color:       "#3B82F6",
			Description: "Flag visitors who requested human assistance",
			Category:    "visitor",
		}
		tag.ID = tagID
		h.db.Create(&tag)
	}

	// Check if visitor already has this tag
	var visitorTag model.VisitorTag
	err = h.db.Where("visitor_id = ? AND tag_id = ?", visitor.ID, tagID).First(&visitorTag).Error
	if err != nil {
		// Create visitor-tag association
		visitorTag = model.VisitorTag{
			ProjectID: projectID,
			VisitorID: visitor.ID,
			TagID:     tagID,
		}
		h.db.Create(&visitorTag)
	}
}

// handleVisitorInfoUpdate handles visitor info update events
func (h *AIEventsHandler) handleVisitorInfoUpdate(c *gin.Context, event *AIServiceEvent, project *model.Project, visitor *model.Visitor) (map[string]interface{}, error) {
	ctx := c.Request.Context()

	if event.Payload == nil {
		return map[string]interface{}{"updated_fields": []string{}}, nil
	}

	// Field mapping
	fieldMap := map[string]string{
		"name":         "name",
		"nickname":     "nickname",
		"email":        "email",
		"phone":        "phone_number",
		"phone_number": "phone_number",
		"company":      "company",
		"job_title":    "job_title",
		"avatar":       "avatar_url",
		"avatar_url":   "avatar_url",
		"note":         "note",
		"source":       "source",
	}

	updates := make(map[string]interface{})
	updatedFields := []string{}

	for payloadKey, dbField := range fieldMap {
		if value, ok := event.Payload[payloadKey]; ok && value != nil {
			if strValue, ok := value.(string); ok && strValue != "" {
				updates[dbField] = strValue
				updatedFields = append(updatedFields, dbField)
			}
		}
	}

	if len(updates) > 0 {
		if err := h.db.WithContext(ctx).Model(&model.Visitor{}).
			Where("id = ?", visitor.ID).
			Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return map[string]interface{}{
		"visitor_id":     visitor.ID.String(),
		"updated_fields": updatedFields,
	}, nil
}

// handleVisitorSentimentUpdate handles visitor sentiment update events
func (h *AIEventsHandler) handleVisitorSentimentUpdate(c *gin.Context, event *AIServiceEvent, project *model.Project, visitor *model.Visitor) (map[string]interface{}, error) {
	// For now, just acknowledge the event
	// Full implementation would update sentiment metrics
	return map[string]interface{}{
		"visitor_id": visitor.ID.String(),
		"message":    "Sentiment update acknowledged",
	}, nil
}

// handleVisitorTag handles visitor tag add events
func (h *AIEventsHandler) handleVisitorTag(c *gin.Context, event *AIServiceEvent, project *model.Project, visitor *model.Visitor) (map[string]interface{}, error) {
	ctx := c.Request.Context()

	if event.Payload == nil {
		return map[string]interface{}{"added_tags": []interface{}{}, "skipped_tags": []interface{}{}}, nil
	}

	var tagItems []map[string]interface{}
	if tags, ok := event.Payload["tags"].([]interface{}); ok {
		for _, t := range tags {
			if tagMap, ok := t.(map[string]interface{}); ok {
				tagItems = append(tagItems, tagMap)
			}
		}
	}

	addedTags := []map[string]string{}
	skippedTags := []map[string]string{}

	for _, tagItem := range tagItems {
		tagName, _ := tagItem["name"].(string)
		tagNameZh, _ := tagItem["name_zh"].(string)
		if tagName == "" {
			continue
		}

		// Find or create tag
		var tag model.Tag
		err := h.db.WithContext(ctx).
			Where("project_id = ? AND name = ? AND deleted_at IS NULL", project.ID, tagName).
			First(&tag).Error
		if err != nil {
			// Create new tag
			tag = model.Tag{
				ProjectID: project.ID,
				Name:      tagName,
				Color:     "#3B82F6",
				Category:  "visitor",
			}
			h.db.WithContext(ctx).Create(&tag)
		}

		// Check if visitor already has this tag
		var visitorTag model.VisitorTag
		err = h.db.WithContext(ctx).
			Where("visitor_id = ? AND tag_id = ? AND deleted_at IS NULL", visitor.ID, tag.ID).
			First(&visitorTag).Error

		tagInfo := map[string]string{"name": tagName}
		if tagNameZh != "" {
			tagInfo["name_zh"] = tagNameZh
		}

		if err != nil {
			// Create visitor-tag association
			visitorTag = model.VisitorTag{
				ProjectID: project.ID,
				VisitorID: visitor.ID,
				TagID:     tag.ID,
			}
			h.db.WithContext(ctx).Create(&visitorTag)
			addedTags = append(addedTags, tagInfo)
		} else {
			skippedTags = append(skippedTags, tagInfo)
		}
	}

	return map[string]interface{}{
		"visitor_id":      visitor.ID.String(),
		"added_tags":      addedTags,
		"skipped_tags":    skippedTags,
		"total_requested": len(tagItems),
	}, nil
}
