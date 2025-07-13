package conversation

import (
	"database/sql/driver"
	"github.com/julietteengel/salesrep-api/pkg/common"
	"time"
)

// ConversationStatus - Conversation states
type ConversationStatus string

const (
	ScheduledStatus  ConversationStatus = "scheduled"
	InProgressStatus ConversationStatus = "in_progress"
	CompletedStatus  ConversationStatus = "completed"
	CancelledStatus  ConversationStatus = "cancelled"
)

func (cs *ConversationStatus) Scan(value interface{}) error {
	if str, ok := value.(string); ok {
		*cs = ConversationStatus(str)
	}
	return nil
}

func (cs ConversationStatus) Value() (driver.Value, error) {
	if string(cs) == "" {
		return ScheduledStatus, nil
	}
	return string(cs), nil
}

type Conversation struct {
	common.BaseModelV2

	// Basic information
	Title       string             `gorm:"not null" json:"title"`
	Description *string            `json:"description,omitempty"`
	Status      ConversationStatus `gorm:"type:varchar(50);default:'scheduled'" json:"status"`

	// Timing
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
	Duration    *int       `json:"duration,omitempty"` // in seconds

	// Sales context
	ClientName    *string    `json:"client_name,omitempty"`
	ClientCompany *string    `json:"client_company,omitempty"`
	ClientEmail   *string    `json:"client_email,omitempty"`
	SalesStage    *string    `json:"sales_stage,omitempty"` // prospect, qualification, proposal, negotiation, closed
	DealValue     *float64   `json:"deal_value,omitempty"`
	ExpectedClose *time.Time `json:"expected_close,omitempty"`

	// Meeting context
	MeetingType      *string `json:"meeting_type,omitempty"` // discovery, demo, proposal, follow_up
	MeetingObjective *string `json:"meeting_objective,omitempty"`

	// Ownership
	OwnerID uint `gorm:"not null" json:"owner_id"` // The sales rep who owns this conversation
}
