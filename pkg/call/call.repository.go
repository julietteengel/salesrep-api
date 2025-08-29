package call

import (
	"context"
	"gorm.io/gorm"
)

type ICallRepository interface {
	Create(ctx context.Context, call *Call) error
	GetByID(ctx context.Context, id uint) (*Call, error)
	GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*Call, error)
	Update(ctx context.Context, call *Call) error
	CreateTranscript(ctx context.Context, transcript *CallTranscript) error
	GetTranscriptsByCallID(ctx context.Context, callID uint) ([]*CallTranscript, error)
}

type CallRepository struct {
	db *gorm.DB
}

func NewCallRepository(db *gorm.DB) ICallRepository {
	return &CallRepository{
		db: db,
	}
}

func (r *CallRepository) Create(ctx context.Context, call *Call) error {
	return r.db.WithContext(ctx).Create(call).Error
}

func (r *CallRepository) GetByID(ctx context.Context, id uint) (*Call, error) {
	var call Call
	err := r.db.WithContext(ctx).First(&call, id).Error
	if err != nil {
		return nil, err
	}
	return &call, nil
}

func (r *CallRepository) GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*Call, error) {
	var calls []*Call
	err := r.db.WithContext(ctx).
		Where("created_by_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&calls).Error
	return calls, err
}

func (r *CallRepository) Update(ctx context.Context, call *Call) error {
	return r.db.WithContext(ctx).Save(call).Error
}

func (r *CallRepository) CreateTranscript(ctx context.Context, transcript *CallTranscript) error {
	return r.db.WithContext(ctx).Create(transcript).Error
}

func (r *CallRepository) GetTranscriptsByCallID(ctx context.Context, callID uint) ([]*CallTranscript, error) {
	var transcripts []*CallTranscript
	err := r.db.WithContext(ctx).
		Where("call_id = ?", callID).
		Order("start_time ASC").
		Find(&transcripts).Error
	return transcripts, err
}