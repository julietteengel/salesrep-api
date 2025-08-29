package call

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/julietteengel/salesrep-api/pkg/storage"
)

type ICallService interface {
	UploadCall(ctx context.Context, userID uint, file multipart.File, header *multipart.FileHeader, metadata *CallUploadMetadata) (*Call, error)
	GetCallByID(ctx context.Context, callID, userID uint) (*Call, error)
	GetUserCalls(ctx context.Context, userID uint, page, pageSize int) ([]*Call, error)
}

type CallService struct {
	repo    ICallRepository
	storage storage.StorageService
}

type CallWithTranscript struct {
	*Call
	Transcripts []*CallTranscript `json:"transcripts,omitempty"`
	Insights    interface{}       `json:"insights,omitempty"`
}

type CallUploadMetadata struct {
	Title          string `json:"title"`
	ClientName     string `json:"client_name"`
	ClientCompany  string `json:"client_company"`
	ConversationID uint   `json:"conversation_id"`
}

func NewCallService(repo ICallRepository, storage storage.StorageService) ICallService {
	return &CallService{
		repo:    repo,
		storage: storage,
	}
}

func (s *CallService) UploadCall(ctx context.Context, userID uint, file multipart.File, header *multipart.FileHeader, metadata *CallUploadMetadata) (*Call, error) {
	// Validate file type
	if !isValidAudioVideo(header.Filename) {
		return nil, errors.New("invalid file type. Only audio/video files are allowed")
	}

	// Validate file size (max 500MB)
	if header.Size > 500*1024*1024 {
		return nil, errors.New("file size exceeds 500MB limit")
	}

	// Upload file to storage service
	fileKey, err := s.storage.UploadFile(ctx, file, header, "calls", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create call record in database
	call := &Call{
		Title:          &metadata.Title,
		Status:         ProcessingCall,
		AudioFileURL:   &fileKey,
		ConversationID: metadata.ConversationID,
		CreatedByID:    userID,
	}

	if err := s.repo.Create(ctx, call); err != nil {
		// Clean up file if database insert fails
		s.storage.DeleteFile(ctx, fileKey)
		return nil, fmt.Errorf("failed to create call record: %w", err)
	}

	return call, nil
}

func (s *CallService) GetCallByID(ctx context.Context, callID, userID uint) (*Call, error) {
	call, err := s.repo.GetByID(ctx, callID)
	if err != nil {
		return nil, err
	}

	// Verify the call belongs to the user
	if call.CreatedByID != userID {
		return nil, errors.New("unauthorized access to call")
	}

	// Convert S3 keys to presigned URLs
	if call.AudioFileURL != nil && *call.AudioFileURL != "" {
		audioURL, err := s.storage.GetFileURL(ctx, *call.AudioFileURL)
		if err == nil {
			call.AudioFileURL = &audioURL
		}
	}

	if call.VideoFileURL != nil && *call.VideoFileURL != "" {
		videoURL, err := s.storage.GetFileURL(ctx, *call.VideoFileURL)
		if err == nil {
			call.VideoFileURL = &videoURL
		}
	}

	return call, nil
}

func (s *CallService) GetUserCalls(ctx context.Context, userID uint, page, pageSize int) ([]*Call, error) {
	offset := (page - 1) * pageSize
	calls, err := s.repo.GetByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	// Convert S3 keys to presigned URLs for all calls
	for _, call := range calls {
		if call.AudioFileURL != nil && *call.AudioFileURL != "" {
			audioURL, err := s.storage.GetFileURL(ctx, *call.AudioFileURL)
			if err == nil {
				call.AudioFileURL = &audioURL
			}
		}

		if call.VideoFileURL != nil && *call.VideoFileURL != "" {
			videoURL, err := s.storage.GetFileURL(ctx, *call.VideoFileURL)
			if err == nil {
				call.VideoFileURL = &videoURL
			}
		}
	}

	return calls, nil
}

// Helper function to validate file extensions
func isValidAudioVideo(filename string) bool {
	validExtensions := map[string]bool{
		".mp3":  true,
		".wav":  true,
		".m4a":  true,
		".aac":  true,
		".ogg":  true,
		".mp4":  true,
		".avi":  true,
		".mov":  true,
		".webm": true,
		".mkv":  true,
	}

	ext := filepath.Ext(filename)
	return validExtensions[ext]
}

