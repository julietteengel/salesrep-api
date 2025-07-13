package conversation

import (
	"context"
	"gorm.io/gorm"
)

// IConversationRepository interface for conversation data access
type IConversationRepository interface {
	GetConversations(ctx context.Context, filters ConversationFilters) ([]Conversation, error)
	GetConversationByID(ctx context.Context, id uint) (*Conversation, error)
	CreateConversation(ctx context.Context, conversation *Conversation) (*Conversation, error)
	UpdateConversation(ctx context.Context, conversation *Conversation) (*Conversation, error)
	DeleteConversation(ctx context.Context, id uint) error
	GetConversationStats(ctx context.Context) (*ConversationStatsData, error)
	GetConversationDashboardStats(ctx context.Context) (*ConversationDashboardStatsData, error)
	GetRecentConversations(ctx context.Context, limit int) ([]Conversation, error)
	CountConversations(ctx context.Context, filters ConversationFilters) (int64, error)
}

// ConversationFilters for filtering conversations
type ConversationFilters struct {
	Search     string
	Status     string
	SalesStage string
	OwnerID    *uint
	Limit      int
	Offset     int
}

// ConversationStatsData for conversation statistics
type ConversationStatsData struct {
	TotalConversations int
	CompletedCount     int
	ScheduledCount     int
	CancelledCount     int
	AverageDealValue   float64
	TotalDealValue     float64
}

// ConversationDashboardStatsData for dashboard metrics
type ConversationDashboardStatsData struct {
	TotalRevenue           float64
	MonthlyRevenue         float64
	LastMonthRevenue       float64
	AverageDealValue       float64
	TotalConversations     int64
	CompletedConversations int64
}

// ConversationRepository implementation
type ConversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *gorm.DB) IConversationRepository {
	return &ConversationRepository{db: db}
}

// GetConversations retrieves conversations with filters
func (r *ConversationRepository) GetConversations(ctx context.Context, filters ConversationFilters) ([]Conversation, error) {
	var conversations []Conversation

	query := r.db.WithContext(ctx).
		Preload("Owner").
		Preload("Calls")

	// Apply filters
	if filters.Search != "" {
		query = query.Where(
			"title ILIKE ? OR client_name ILIKE ? OR client_company ILIKE ?",
			"%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%",
		)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.SalesStage != "" {
		query = query.Where("sales_stage = ?", filters.SalesStage)
	}

	if filters.OwnerID != nil {
		query = query.Where("owner_id = ?", *filters.OwnerID)
	}

	if err := query.Limit(filters.Limit).Offset(filters.Offset).Find(&conversations).Error; err != nil {
		return nil, err
	}

	return conversations, nil
}

// CountConversations counts conversations with filters
func (r *ConversationRepository) CountConversations(ctx context.Context, filters ConversationFilters) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&Conversation{})

	// Apply same filters as GetConversations
	if filters.Search != "" {
		query = query.Where(
			"title ILIKE ? OR client_name ILIKE ? OR client_company ILIKE ?",
			"%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%",
		)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.SalesStage != "" {
		query = query.Where("sales_stage = ?", filters.SalesStage)
	}

	if filters.OwnerID != nil {
		query = query.Where("owner_id = ?", *filters.OwnerID)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// GetConversationByID retrieves a conversation by ID
func (r *ConversationRepository) GetConversationByID(ctx context.Context, id uint) (*Conversation, error) {
	var conversation Conversation

	if err := r.db.WithContext(ctx).
		Preload("Owner").
		Preload("Participants").
		Preload("Calls").
		First(&conversation, id).Error; err != nil {
		return nil, err
	}

	return &conversation, nil
}

// CreateConversation creates a new conversation
func (r *ConversationRepository) CreateConversation(ctx context.Context, conversation *Conversation) (*Conversation, error) {
	if err := r.db.WithContext(ctx).Create(conversation).Error; err != nil {
		return nil, err
	}
	return conversation, nil
}

// UpdateConversation updates a conversation
func (r *ConversationRepository) UpdateConversation(ctx context.Context, conversation *Conversation) (*Conversation, error) {
	if err := r.db.WithContext(ctx).Save(conversation).Error; err != nil {
		return nil, err
	}
	return conversation, nil
}

// DeleteConversation soft deletes a conversation
func (r *ConversationRepository) DeleteConversation(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Conversation{}, id).Error
}

// GetConversationStats retrieves conversation statistics
func (r *ConversationRepository) GetConversationStats(ctx context.Context) (*ConversationStatsData, error) {
	var stats ConversationStatsData

	// Get total count
	var totalCount int64
	if err := r.db.WithContext(ctx).Model(&Conversation{}).Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats.TotalConversations = int(totalCount)

	// Get counts by status
	var statusCounts []struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}

	if err := r.db.WithContext(ctx).Model(&Conversation{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return nil, err
	}

	// Parse status counts
	for _, sc := range statusCounts {
		switch sc.Status {
		case "completed":
			stats.CompletedCount = sc.Count
		case "scheduled":
			stats.ScheduledCount = sc.Count
		case "cancelled":
			stats.CancelledCount = sc.Count
		}
	}

	// Get deal value statistics
	type DealStats struct {
		Total   *float64 `json:"total"`
		Average *float64 `json:"average"`
	}

	var dealStats DealStats
	if err := r.db.WithContext(ctx).Model(&Conversation{}).
		Select("SUM(deal_value) as total, AVG(deal_value) as average").
		Where("deal_value IS NOT NULL").
		Scan(&dealStats).Error; err != nil {
		return nil, err
	}

	if dealStats.Total != nil {
		stats.TotalDealValue = *dealStats.Total
	}
	if dealStats.Average != nil {
		stats.AverageDealValue = *dealStats.Average
	}

	return &stats, nil
}

// GetConversationDashboardStats retrieves dashboard-specific statistics
func (r *ConversationRepository) GetConversationDashboardStats(ctx context.Context) (*ConversationDashboardStatsData, error) {
	var stats ConversationDashboardStatsData

	// Get total revenue
	if err := r.db.WithContext(ctx).Model(&Conversation{}).
		Select("SUM(deal_value)").
		Where("deal_value IS NOT NULL AND status = ?", "completed").
		Scan(&stats.TotalRevenue).Error; err != nil {
		return nil, err
	}

	// Get current month revenue
	if err := r.db.WithContext(ctx).Model(&Conversation{}).
		Select("SUM(deal_value)").
		Where("deal_value IS NOT NULL AND status = ? AND updated_at >= DATE_TRUNC('month', CURRENT_DATE)", "completed").
		Scan(&stats.MonthlyRevenue).Error; err != nil {
		return nil, err
	}

	// Get last month revenue
	if err := r.db.WithContext(ctx).Model(&Conversation{}).
		Select("SUM(deal_value)").
		Where("deal_value IS NOT NULL AND status = ? AND updated_at >= DATE_TRUNC('month', CURRENT_DATE) - INTERVAL '1 month' AND updated_at < DATE_TRUNC('month', CURRENT_DATE)", "completed").
		Scan(&stats.LastMonthRevenue).Error; err != nil {
		return nil, err
	}

	// Get average deal value
	if err := r.db.WithContext(ctx).Model(&Conversation{}).
		Select("AVG(deal_value)").
		Where("deal_value IS NOT NULL").
		Scan(&stats.AverageDealValue).Error; err != nil {
		return nil, err
	}

	// Get conversion rate data
	if err := r.db.WithContext(ctx).Model(&Conversation{}).Count(&stats.TotalConversations).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Model(&Conversation{}).Where("status = ?", "completed").Count(&stats.CompletedConversations).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetRecentConversations retrieves recent conversations
func (r *ConversationRepository) GetRecentConversations(ctx context.Context, limit int) ([]Conversation, error) {
	var conversations []Conversation

	if err := r.db.WithContext(ctx).
		Preload("Owner").
		Preload("Calls").
		Order("updated_at DESC").
		Limit(limit).
		Find(&conversations).Error; err != nil {
		return nil, err
	}

	return conversations, nil
}
