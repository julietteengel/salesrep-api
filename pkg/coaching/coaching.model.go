package coaching

import (
	"github.com/julietteengel/salesrep-api/pkg/common"
	"time"
)

type CoachingNote struct {
	common.BaseModelV2

	// Core relationships
	UserID    uint `gorm:"not null;index" json:"user_id"`    // The sales rep being coached
	ManagerID uint `gorm:"not null;index" json:"manager_id"` // The manager providing coaching

	// Content
	Title   string `gorm:"not null" json:"title"`
	Content string `gorm:"type:TEXT;not null" json:"content"`
	Type    string `gorm:"not null;index" json:"type"` // feedback, goal, improvement_area, strength, action_plan

	// Context (optional)
	CallID         *uint `json:"call_id,omitempty"`
	ConversationID *uint `json:"conversation_id,omitempty"`

	// Configuration
	IsPrivate   bool       `gorm:"default:false" json:"is_private"` // Only visible to manager
	DueDate     *time.Time `json:"due_date,omitempty"`              // For action items/goals
	CompletedAt *time.Time `json:"completed_at,omitempty"`          // When goal was achieved

	// Follow-up
	ParentNoteID *uint `json:"parent_note_id,omitempty"` // For follow-up notes
}
