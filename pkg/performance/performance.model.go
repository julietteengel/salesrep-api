package performance

import (
	"github.com/julietteengel/salesrep-api/pkg/common"
	"github.com/julietteengel/salesrep-api/pkg/user"
	"time"
)

type UserPerformance struct {
	common.BaseModelV2

	// Time period
	PeriodStart time.Time `gorm:"not null;index" json:"period_start"`
	PeriodEnd   time.Time `gorm:"not null;index" json:"period_end"`
	PeriodType  string    `gorm:"not null;index" json:"period_type"` // daily, weekly, monthly, quarterly

	// User reference
	UserID uint      `gorm:"not null;index" json:"user_id"`
	User   user.User `gorm:"foreignKey:UserID" json:"user"`

	// Call metrics
	TotalCalls         int     `gorm:"default:0" json:"total_calls"`
	TotalConversations int     `gorm:"default:0" json:"total_conversations"`
	AverageScore       float64 `gorm:"default:0" json:"average_score"`
	TotalTalkTime      float64 `gorm:"default:0" json:"total_talk_time"` // in minutes
	AverageTalkPercent float64 `gorm:"default:0" json:"average_talk_percent"`

	// Sales metrics
	DealsWon         int     `gorm:"default:0" json:"deals_won"`
	DealsLost        int     `gorm:"default:0" json:"deals_lost"`
	ConversionRate   float64 `gorm:"default:0" json:"conversion_rate"`
	TotalRevenue     float64 `gorm:"default:0" json:"total_revenue"`
	AverageDealValue float64 `gorm:"default:0" json:"average_deal_value"`

	// Performance trends
	ScoreTrend    *string `json:"score_trend,omitempty"`    // improving, declining, stable
	ActivityTrend *string `json:"activity_trend,omitempty"` // increasing, decreasing, stable
	RevenueTrend  *string `json:"revenue_trend,omitempty"`  // up, down, flat

	// Coaching insights
	AreasImprovement *string `gorm:"type:JSON" json:"areas_improvement,omitempty"` // JSON array
	StrengthAreas    *string `gorm:"type:JSON" json:"strength_areas,omitempty"`    // JSON array
	CoachingGoals    *string `gorm:"type:JSON" json:"coaching_goals,omitempty"`    // JSON array

	// Targets vs actual
	CallsTarget    *int     `json:"calls_target,omitempty"`
	RevenueTarget  *float64 `json:"revenue_target,omitempty"`
	CallsActual    int      `gorm:"default:0" json:"calls_actual"`
	RevenueActual  float64  `gorm:"default:0" json:"revenue_actual"`
	TargetProgress float64  `gorm:"default:0" json:"target_progress"` // percentage
}

type TeamPerformance struct {
	common.BaseModelV2

	// Time period
	PeriodStart time.Time `gorm:"not null;index" json:"period_start"`
	PeriodEnd   time.Time `gorm:"not null;index" json:"period_end"`
	PeriodType  string    `gorm:"not null;index" json:"period_type"`

	// Manager reference
	ManagerID uint      `gorm:"not null;index" json:"manager_id"`
	Manager   user.User `gorm:"foreignKey:ManagerID" json:"manager"`

	// Aggregated team metrics
	TotalCalls         int     `gorm:"default:0" json:"total_calls"`
	TotalConversations int     `gorm:"default:0" json:"total_conversations"`
	AverageScore       float64 `gorm:"default:0" json:"average_score"`
	TotalTalkTime      float64 `gorm:"default:0" json:"total_talk_time"`
	TeamSize           int     `gorm:"default:0" json:"team_size"`

	// Team performance
	ConversionRate   float64 `gorm:"default:0" json:"conversion_rate"`
	AverageDealValue float64 `gorm:"default:0" json:"average_deal_value"`
	TotalRevenue     float64 `gorm:"default:0" json:"total_revenue"`

	// Team health metrics
	TopPerformerID    *uint      `json:"top_performer_id,omitempty"`
	TopPerformer      *user.User `gorm:"foreignKey:TopPerformerID" json:"top_performer,omitempty"`
	TeamMorale        *float64   `json:"team_morale,omitempty"`        // 0-10 based on various factors
	CoachingFrequency *float64   `json:"coaching_frequency,omitempty"` // notes per member per period

	// Detailed breakdown (JSON)
	MemberPerformances *string `gorm:"type:JSON" json:"member_performances,omitempty"` // JSON object with member details
	TopChallenges      *string `gorm:"type:JSON" json:"top_challenges,omitempty"`      // JSON array
	TopWins            *string `gorm:"type:JSON" json:"top_wins,omitempty"`            // JSON array
}
