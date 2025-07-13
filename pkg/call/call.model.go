package call

import (
	"database/sql/driver"
	"github.com/julietteengel/salesrep-api/pkg/common"
	"time"
)

// CallStatus - Call processing states
type CallStatus string

const (
	ProcessingCall CallStatus = "processing"
	CompletedCall  CallStatus = "completed"
	FailedCall     CallStatus = "failed"
)

func (cs *CallStatus) Scan(value interface{}) error {
	if str, ok := value.(string); ok {
		*cs = CallStatus(str)
	}
	return nil
}

func (cs CallStatus) Value() (driver.Value, error) {
	if string(cs) == "" {
		return ProcessingCall, nil
	}
	return string(cs), nil
}

type Call struct {
	common.BaseModelV2

	// Basic information
	Title    *string    `json:"title,omitempty"`
	Duration *int       `json:"duration,omitempty"` // in seconds
	Status   CallStatus `gorm:"type:varchar(50);default:'processing'" json:"status"`

	// Timing
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`

	// File storage
	AudioFileURL  *string `json:"audio_file_url,omitempty"`
	VideoFileURL  *string `json:"video_file_url,omitempty"`
	TranscriptURL *string `json:"transcript_url,omitempty"`

	// External references
	ExternalMeetingID *string `json:"external_meeting_id,omitempty"` // Zoom, Teams, etc.
	RecordingID       *string `json:"recording_id,omitempty"`

	// Relationships
	ConversationID uint `gorm:"not null" json:"conversation_id"`
	CreatedByID    uint `gorm:"not null" json:"created_by_id"`

	// Analysis data
	Metrics     []CallMetric     `gorm:"foreignKey:CallID" json:"metrics,omitempty"`
	Speakers    []CallSpeaker    `gorm:"foreignKey:CallID" json:"speakers,omitempty"`
	Transcripts []CallTranscript `gorm:"foreignKey:CallID" json:"transcripts,omitempty"`
	Analysis    *CallAnalysis    `gorm:"foreignKey:CallID" json:"analysis,omitempty"`
}

// ================================
// CALL ANALYSIS MODELS
// ================================

type CallMetric struct {
	common.BaseModelV2

	CallID uint `gorm:"not null;index" json:"call_id"`
	Call   Call `gorm:"foreignKey:CallID" json:"call"`

	MetricType string  `gorm:"not null;index" json:"metric_type"` // speaker_talk_time, interruptions, silence_percentage, etc.
	MetricName string  `gorm:"not null" json:"metric_name"`
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`                 // percentage, minutes, count, seconds
	SpeakerID  *string `json:"speaker_id,omitempty"` // For speaker-specific metrics

	Description *string  `json:"description,omitempty"`
	Threshold   *float64 `json:"threshold,omitempty"` // For flagging unusual values
}

type CallSpeaker struct {
	common.BaseModelV2

	CallID uint `gorm:"not null;index" json:"call_id"`
	Call   Call `gorm:"foreignKey:CallID" json:"call"`

	SpeakerID     string  `gorm:"not null" json:"speaker_id"` // AI-generated speaker identifier
	Name          string  `json:"name"`
	TalkTime      float64 `json:"talk_time"`    // in minutes
	TalkPercent   float64 `json:"talk_percent"` // percentage of total talk time
	WordCount     int     `json:"word_count"`
	Interruptions int     `json:"interruptions"`
	SpeakerType   string  `json:"speaker_type"` // sales_rep, client, other

	// Connection to system user if identified
	UserID *uint `json:"user_id,omitempty"`

	// Additional speaker insights
	SentimentScore   *float64 `json:"sentiment_score,omitempty"`  // -1 to 1
	EngagementLevel  *float64 `json:"engagement_level,omitempty"` // 0 to 1
	QuestionsAsked   *int     `json:"questions_asked,omitempty"`
	ObjectionsRaised *int     `json:"objections_raised,omitempty"`
}

type CallTranscript struct {
	common.BaseModelV2

	CallID uint `gorm:"not null;index" json:"call_id"`
	Call   Call `gorm:"foreignKey:CallID" json:"call"`

	SpeakerID   string  `gorm:"not null;index" json:"speaker_id"`
	SpeakerName string  `json:"speaker_name"`
	Text        string  `gorm:"type:TEXT;not null" json:"text"`
	StartTime   float64 `json:"start_time"` // seconds from call start
	EndTime     float64 `json:"end_time"`
	Confidence  float64 `json:"confidence"` // 0-1 transcription confidence

	// Additional context
	Sentiment   *string `json:"sentiment,omitempty"`                     // positive, negative, neutral
	KeywordTags *string `gorm:"type:JSON" json:"keyword_tags,omitempty"` // JSON array of relevant keywords
	IsQuestion  *bool   `json:"is_question,omitempty"`
	IsObjection *bool   `json:"is_objection,omitempty"`
}

type CallAnalysis struct {
	common.BaseModelV2

	CallID uint `gorm:"uniqueIndex;not null" json:"call_id"`
	Call   Call `gorm:"foreignKey:CallID" json:"call"`

	// General analysis
	Summary        *string  `gorm:"type:TEXT" json:"summary,omitempty"`
	KeyPoints      *string  `gorm:"type:JSON" json:"key_points,omitempty"`   // JSON array
	ActionItems    *string  `gorm:"type:JSON" json:"action_items,omitempty"` // JSON array
	Sentiment      *string  `json:"sentiment,omitempty"`                     // positive, negative, neutral
	SentimentScore *float64 `json:"sentiment_score,omitempty"`               // -1 to 1

	// Performance metrics
	OverallScore     *float64 `json:"overall_score,omitempty"`     // 0-10
	TalkTimeBalance  *float64 `json:"talk_time_balance,omitempty"` // How balanced was the conversation
	QuestioningScore *float64 `json:"questioning_score,omitempty"` // How well did rep ask questions
	ListeningScore   *float64 `json:"listening_score,omitempty"`   // How well did rep listen
	EnergyLevel      *float64 `json:"energy_level,omitempty"`      // Overall energy/enthusiasm

	// Sales-specific analysis
	SalesStage         *string  `json:"sales_stage,omitempty"`
	ObjectionsRaised   *string  `gorm:"type:JSON" json:"objections_raised,omitempty"`  // JSON array
	ObjectionsHandled  *string  `gorm:"type:JSON" json:"objections_handled,omitempty"` // JSON array
	NextSteps          *string  `gorm:"type:TEXT" json:"next_steps,omitempty"`
	SuccessProbability *float64 `json:"success_probability,omitempty"`             // 0-1
	BuyingSignals      *string  `gorm:"type:JSON" json:"buying_signals,omitempty"` // JSON array

	// Coaching insights
	Strengths        *string `gorm:"type:JSON" json:"strengths,omitempty"`         // JSON array
	AreasImprovement *string `gorm:"type:JSON" json:"areas_improvement,omitempty"` // JSON array
	CoachingTips     *string `gorm:"type:JSON" json:"coaching_tips,omitempty"`     // JSON array

	// Analysis metadata
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
	ProcessingTime *int       `json:"processing_time,omitempty"` // seconds
	AIModel        *string    `json:"ai_model,omitempty"`        // Which AI model was used
	Confidence     *float64   `json:"confidence,omitempty"`      // Overall analysis confidence
}
