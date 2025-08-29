package transcription

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/julietteengel/salesrep-api/pkg/common"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// CallStatus - Call processing states (duplicated to avoid circular import)
type CallStatus string

const (
	ProcessingCall CallStatus = "processing"
	CompletedCall  CallStatus = "completed"
	FailedCall     CallStatus = "failed"
)

// CallTranscript represents a transcript segment (duplicated to avoid circular import)
type CallTranscript struct {
	common.BaseModelV2

	CallID uint `gorm:"not null;index" json:"call_id"`

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

// Call represents the minimal call structure needed for transcription (duplicated to avoid circular import)
type Call struct {
	ID     uint       `json:"id"`
	Status CallStatus `json:"status"`
}

// ICallRepository interface duplicated to avoid circular import
type ICallRepository interface {
	GetByID(ctx context.Context, id uint) (*Call, error)
	Update(ctx context.Context, call *Call) error
	CreateTranscript(ctx context.Context, transcript *CallTranscript) error
	GetTranscriptsByCallID(ctx context.Context, callID uint) ([]*CallTranscript, error)
}

// CallRepository adapter to work with the actual call repository
type CallRepositoryAdapter struct {
	db *gorm.DB
}

func NewCallRepositoryAdapter(db *gorm.DB) ICallRepository {
	return &CallRepositoryAdapter{db: db}
}

func (r *CallRepositoryAdapter) GetByID(ctx context.Context, id uint) (*Call, error) {
	var call Call
	if err := r.db.WithContext(ctx).Table("calls").First(&call, id).Error; err != nil {
		return nil, err
	}
	return &call, nil
}

func (r *CallRepositoryAdapter) Update(ctx context.Context, call *Call) error {
	return r.db.WithContext(ctx).Table("calls").Where("id = ?", call.ID).Update("status", call.Status).Error
}

func (r *CallRepositoryAdapter) CreateTranscript(ctx context.Context, transcript *CallTranscript) error {
	return r.db.WithContext(ctx).Table("call_transcripts").Create(transcript).Error
}

func (r *CallRepositoryAdapter) GetTranscriptsByCallID(ctx context.Context, callID uint) ([]*CallTranscript, error) {
	var transcripts []*CallTranscript
	if err := r.db.WithContext(ctx).Table("call_transcripts").Where("call_id = ?", callID).Order("start_time").Find(&transcripts).Error; err != nil {
		return nil, err
	}
	return transcripts, nil
}

type ITranscriptionService interface {
	TranscribeCall(ctx context.Context, callID uint, audioFileURL string) error
	GetTranscriptionByCallID(ctx context.Context, callID uint) ([]*CallTranscript, error)
}

type TranscriptionService struct {
	callRepo   ICallRepository
	openAIKey  string
	httpClient *http.Client
}

type OpenAITranscriptionRequest struct {
	File           io.Reader `json:"file"`
	Model          string    `json:"model"`
	Language       string    `json:"language,omitempty"`
	Prompt         string    `json:"prompt,omitempty"`
	ResponseFormat string    `json:"response_format,omitempty"`
	Temperature    float64   `json:"temperature,omitempty"`
}

type OpenAITranscriptionResponse struct {
	Text     string                    `json:"text"`
	Language string                    `json:"language,omitempty"`
	Duration float64                   `json:"duration,omitempty"`
	Segments []OpenAITranscriptSegment `json:"segments,omitempty"`
}

type OpenAITranscriptSegment struct {
	ID               int     `json:"id"`
	Seek             int     `json:"seek"`
	Start            float64 `json:"start"`
	End              float64 `json:"end"`
	Text             string  `json:"text"`
	Tokens           []int   `json:"tokens"`
	Temperature      float64 `json:"temperature"`
	AvgLogprob       float64 `json:"avg_logprob"`
	CompressionRatio float64 `json:"compression_ratio"`
	NoSpeechProb     float64 `json:"no_speech_prob"`
}

func NewTranscriptionService(db *gorm.DB) ITranscriptionService {
	return &TranscriptionService{
		callRepo:  NewCallRepositoryAdapter(db),
		openAIKey: viper.GetString("OPENAI_API_KEY"),
		httpClient: &http.Client{
			Timeout: 10 * time.Minute, // Whisper can take time for long files
		},
	}
}

func (s *TranscriptionService) TranscribeCall(ctx context.Context, callID uint, audioFileURL string) error {
	// Get the call record
	existingCall, err := s.callRepo.GetByID(ctx, callID)
	if err != nil {
		return fmt.Errorf("failed to get call: %w", err)
	}

	// Update call status to transcribing
	existingCall.Status = ProcessingCall
	if err := s.callRepo.Update(ctx, existingCall); err != nil {
		return fmt.Errorf("failed to update call status: %w", err)
	}

	// Download audio file from S3/storage
	audioContent, err := s.downloadAudioFile(ctx, audioFileURL)
	if err != nil {
		return fmt.Errorf("failed to download audio file: %w", err)
	}

	// Transcribe with OpenAI Whisper
	transcriptionResp, err := s.callOpenAIWhisper(ctx, audioContent)
	if err != nil {
		// Update call status to failed
		existingCall.Status = FailedCall
		s.callRepo.Update(ctx, existingCall)
		return fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Store transcription segments in database
	err = s.storeTranscription(ctx, callID, transcriptionResp)
	if err != nil {
		return fmt.Errorf("failed to store transcription: %w", err)
	}

	// Update call status to completed
	existingCall.Status = CompletedCall
	if err := s.callRepo.Update(ctx, existingCall); err != nil {
		return fmt.Errorf("failed to update call status to completed: %w", err)
	}

	return nil
}

func (s *TranscriptionService) callOpenAIWhisper(ctx context.Context, audioContent []byte) (*OpenAITranscriptionResponse, error) {
	url := "https://api.openai.com/v1/audio/transcriptions"

	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file
	fileWriter, err := writer.CreateFormFile("file", "audio.mp3")
	if err != nil {
		return nil, err
	}
	_, err = fileWriter.Write(audioContent)
	if err != nil {
		return nil, err
	}

	// Add other form fields
	writer.WriteField("model", "whisper-1")
	writer.WriteField("response_format", "verbose_json") // Get segments with timestamps
	writer.WriteField("language", "en")                  // Can be auto-detected by omitting this

	writer.Close()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.openAIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Make request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var transcriptionResp OpenAITranscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&transcriptionResp); err != nil {
		return nil, err
	}

	return &transcriptionResp, nil
}

func (s *TranscriptionService) downloadAudioFile(ctx context.Context, audioFileURL string) ([]byte, error) {
	// Make HTTP request to download the file from S3 URL
	req, err := http.NewRequestWithContext(ctx, "GET", audioFileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: status %d", resp.StatusCode)
	}

	// Read the entire file content
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return content, nil
}

func (s *TranscriptionService) storeTranscription(ctx context.Context, callID uint, resp *OpenAITranscriptionResponse) error {
	// Store each segment as a separate transcript record
	for _, segment := range resp.Segments {
		transcript := &CallTranscript{
			CallID:      callID,
			SpeakerID:   fmt.Sprintf("speaker_%d", segment.ID%2), // Simple alternating speaker assignment for now
			SpeakerName: fmt.Sprintf("Speaker %d", segment.ID%2+1), // Speaker 1 or Speaker 2
			Text:        segment.Text,
			StartTime:   segment.Start,
			EndTime:     segment.End,
			Confidence:  1.0 - segment.NoSpeechProb, // Approximate confidence from no_speech_prob
		}

		err := s.callRepo.CreateTranscript(ctx, transcript)
		if err != nil {
			return fmt.Errorf("failed to store transcript segment: %w", err)
		}
	}

	return nil
}

func (s *TranscriptionService) GetTranscriptionByCallID(ctx context.Context, callID uint) ([]*CallTranscript, error) {
	return s.callRepo.GetTranscriptsByCallID(ctx, callID)
}