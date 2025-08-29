package models

import (
	"github.com/julietteengel/salesrep-api/pkg/call"
	"github.com/julietteengel/salesrep-api/pkg/coaching"
	"github.com/julietteengel/salesrep-api/pkg/conversation"
	"github.com/julietteengel/salesrep-api/pkg/performance"
	"github.com/julietteengel/salesrep-api/pkg/user"
)

// UserWithRelations extends user.User with all relationships
type UserWithRelations struct {
	user.User
	Manager              *user.User                    `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	TeamMembers          []user.User                   `gorm:"foreignKey:ManagerID" json:"team_members,omitempty"`
	OwnedConversations   []conversation.Conversation   `gorm:"foreignKey:OwnerID" json:"owned_conversations,omitempty"`
	Conversations        []conversation.Conversation   `gorm:"many2many:conversation_participants" json:"conversations,omitempty"`
	CreatedCalls         []call.Call                   `gorm:"foreignKey:CreatedByID" json:"created_calls,omitempty"`
	CoachingNotes        []coaching.CoachingNote       `gorm:"foreignKey:UserID" json:"coaching_notes,omitempty"`
	CreatedCoachingNotes []coaching.CoachingNote       `gorm:"foreignKey:ManagerID" json:"created_coaching_notes,omitempty"`
	Performances         []performance.UserPerformance `gorm:"foreignKey:UserID" json:"performances,omitempty"`
}

// CallWithRelations extends call.Call with all relationships
type CallWithRelations struct {
	call.Call
	Conversation conversation.Conversation `gorm:"foreignKey:ConversationID" json:"conversation"`
	CreatedBy    user.User                 `gorm:"foreignKey:CreatedByID" json:"created_by"`
	Metrics      []call.CallMetric         `gorm:"foreignKey:CallID" json:"metrics,omitempty"`
	Speakers     []call.CallSpeaker        `gorm:"foreignKey:CallID" json:"speakers,omitempty"`
	Transcripts  []call.CallTranscript     `gorm:"foreignKey:CallID" json:"transcripts,omitempty"`
	Analysis     *call.CallAnalysis        `gorm:"foreignKey:CallID" json:"analysis,omitempty"`
}

// ConversationWithRelations extends conversation.Conversation with all relationships
type ConversationWithRelations struct {
	conversation.Conversation
	Owner        user.User   `gorm:"foreignKey:OwnerID" json:"owner"`
	Participants []user.User `gorm:"many2many:conversation_participants" json:"participants,omitempty"`
	Calls        []call.Call `gorm:"foreignKey:ConversationID" json:"calls,omitempty"`
}

// CoachingNoteWithRelations extends coaching.CoachingNote with all relationships
type CoachingNoteWithRelations struct {
	coaching.CoachingNote
	User         user.User                  `gorm:"foreignKey:UserID" json:"user"`
	Manager      user.User                  `gorm:"foreignKey:ManagerID" json:"manager"`
	Call         *call.Call                 `gorm:"foreignKey:CallID" json:"call,omitempty"`
	Conversation *conversation.Conversation `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`
	ParentNote   *coaching.CoachingNote     `gorm:"foreignKey:ParentNoteID" json:"parent_note,omitempty"`
	ChildNotes   []coaching.CoachingNote    `gorm:"foreignKey:ParentNoteID" json:"child_notes,omitempty"`
}

// CallSpeakerWithRelations extends call.CallSpeaker with user relationship
type CallSpeakerWithRelations struct {
	call.CallSpeaker
	User *user.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
