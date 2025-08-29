package conversation

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/pkg/errors"
)

// IConversationService interface for conversation business logic
type IConversationService interface {
	GetConversations(ctx context.Context, filters ConversationFilters) ([]Conversation, error)
	GetConversationByID(ctx context.Context, id uint) (*Conversation, error)
	CreateConversation(ctx context.Context, req CreateConversationRequest, userID uint) (*Conversation, error)
	UpdateConversation(ctx context.Context, id uint, req UpdateConversationRequest) (*Conversation, error)
	DeleteConversation(ctx context.Context, id uint) error
	GetConversationStats(ctx context.Context) (*ConversationStatsResponse, error)
	GetConversationDashboardStats(ctx context.Context) (*ConversationDashboardStatsResponse, error)
	GetRecentConversations(ctx context.Context, limit int) ([]Conversation, error)
	CountConversations(ctx context.Context, filters ConversationFilters) (int64, error)
}

// ConversationService implementation
type ConversationService struct {
	conversationRepository IConversationRepository
}

// NewConversationService creates a new conversation service
func NewConversationService(conversationRepository IConversationRepository) IConversationService {
	return &ConversationService{
		conversationRepository: conversationRepository,
	}
}

// GetConversations retrieves conversations with filters
func (s *ConversationService) GetConversations(ctx context.Context, filters ConversationFilters) ([]Conversation, error) {
	return s.conversationRepository.GetConversations(ctx, filters)
}

// CountConversations counts conversations with filters
func (s *ConversationService) CountConversations(ctx context.Context, filters ConversationFilters) (int64, error) {
	return s.conversationRepository.CountConversations(ctx, filters)
}

// GetConversationByID retrieves a conversation by ID
func (s *ConversationService) GetConversationByID(ctx context.Context, id uint) (*Conversation, error) {
	conversation, err := s.conversationRepository.GetConversationByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get conversation")
	}
	return conversation, nil
}

// CreateConversation creates a new conversation
func (s *ConversationService) CreateConversation(ctx context.Context, req CreateConversationRequest, userID uint) (*Conversation, error) {
	// Create conversation model
	conversation := &Conversation{
		Title:            req.Title,
		Description:      req.Description,
		Status:           ScheduledStatus,
		ClientName:       req.ClientName,
		ClientCompany:    req.ClientCompany,
		ClientEmail:      req.ClientEmail,
		SalesStage:       req.SalesStage,
		DealValue:        req.DealValue,
		MeetingType:      req.MeetingType,
		MeetingObjective: req.MeetingObjective,
		OwnerID:          userID,
	}

	// Create conversation
	createdConversation, err := s.conversationRepository.CreateConversation(ctx, conversation)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create conversation")
	}

	// Return with relations
	return s.conversationRepository.GetConversationByID(ctx, createdConversation.ID)
}

// UpdateConversation updates a conversation
func (s *ConversationService) UpdateConversation(ctx context.Context, id uint, req UpdateConversationRequest) (*Conversation, error) {
	// Get existing conversation
	existingConversation, err := s.conversationRepository.GetConversationByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get conversation for update")
	}

	// Update fields
	conversation := &Conversation{}
	conversation.ID = existingConversation.ID
	conversation.CreatedAt = existingConversation.CreatedAt
	conversation.UpdatedAt = time.Now()

	// Copy existing values
	conversation.Title = existingConversation.Title
	conversation.Description = existingConversation.Description
	conversation.Status = existingConversation.Status
	conversation.ClientName = existingConversation.ClientName
	conversation.ClientCompany = existingConversation.ClientCompany
	conversation.ClientEmail = existingConversation.ClientEmail
	conversation.SalesStage = existingConversation.SalesStage
	conversation.DealValue = existingConversation.DealValue
	conversation.MeetingType = existingConversation.MeetingType
	conversation.MeetingObjective = existingConversation.MeetingObjective
	conversation.OwnerID = existingConversation.OwnerID

	// Apply updates
	if req.Title != nil {
		conversation.Title = *req.Title
	}
	if req.Description != nil {
		conversation.Description = req.Description
	}
	if req.Status != nil {
		conversation.Status = ConversationStatus(*req.Status)
	}
	if req.ClientName != nil {
		conversation.ClientName = req.ClientName
	}
	if req.ClientCompany != nil {
		conversation.ClientCompany = req.ClientCompany
	}
	if req.ClientEmail != nil {
		conversation.ClientEmail = req.ClientEmail
	}
	if req.SalesStage != nil {
		conversation.SalesStage = req.SalesStage
	}
	if req.DealValue != nil {
		conversation.DealValue = req.DealValue
	}
	if req.MeetingType != nil {
		conversation.MeetingType = req.MeetingType
	}
	if req.MeetingObjective != nil {
		conversation.MeetingObjective = req.MeetingObjective
	}

	// Update conversation
	_, err = s.conversationRepository.UpdateConversation(ctx, conversation)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update conversation")
	}

	// Return updated conversation with relations
	return s.conversationRepository.GetConversationByID(ctx, id)
}

// DeleteConversation deletes a conversation
func (s *ConversationService) DeleteConversation(ctx context.Context, id uint) error {
	// Check if conversation exists
	_, err := s.conversationRepository.GetConversationByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "conversation not found")
	}

	// Delete conversation
	return s.conversationRepository.DeleteConversation(ctx, id)
}

// GetConversationStats retrieves conversation statistics
func (s *ConversationService) GetConversationStats(ctx context.Context) (*ConversationStatsResponse, error) {
	stats, err := s.conversationRepository.GetConversationStats(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get conversation stats")
	}

	return &ConversationStatsResponse{
		TotalConversations: stats.TotalConversations,
		CompletedCount:     stats.CompletedCount,
		ScheduledCount:     stats.ScheduledCount,
		CancelledCount:     stats.CancelledCount,
		AverageDealValue:   stats.AverageDealValue,
		TotalDealValue:     stats.TotalDealValue,
	}, nil
}

// GetConversationDashboardStats retrieves dashboard-specific statistics
func (s *ConversationService) GetConversationDashboardStats(ctx context.Context) (*ConversationDashboardStatsResponse, error) {
	stats, err := s.conversationRepository.GetConversationDashboardStats(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get conversation dashboard stats")
	}

	response := &ConversationDashboardStatsResponse{
		TotalRevenue:     stats.TotalRevenue,
		MonthlyRevenue:   stats.MonthlyRevenue,
		AverageDealValue: stats.AverageDealValue,
	}

	// Calculate trend
	if stats.LastMonthRevenue > 0 {
		trendPercent := ((stats.MonthlyRevenue - stats.LastMonthRevenue) / stats.LastMonthRevenue) * 100
		response.RevenueTrend = fmt.Sprintf("%.1f%%", math.Abs(trendPercent))
		if trendPercent >= 0 {
			response.TrendDirection = "up"
		} else {
			response.TrendDirection = "down"
		}
	} else {
		response.RevenueTrend = "0%"
		response.TrendDirection = "up"
	}

	// Calculate conversion rate
	if stats.TotalConversations > 0 {
		response.ConversionsRate = (float64(stats.CompletedConversations) / float64(stats.TotalConversations)) * 100
	}

	return response, nil
}

// GetRecentConversations retrieves recent conversations
func (s *ConversationService) GetRecentConversations(ctx context.Context, limit int) ([]Conversation, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	return s.conversationRepository.GetRecentConversations(ctx, limit)
}