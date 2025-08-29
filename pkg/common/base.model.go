package common

import (
	"time"

	"gorm.io/gorm"
)

type BaseModelV2 struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *uint          `json:"created_by,omitempty"`
	UpdatedBy *uint          `json:"updated_by,omitempty"`
	DeletedBy *uint          `json:"deleted_by,omitempty"`
}

type BaseTranslationMandatory struct {
	Fr string `json:"fr,omitempty"`
	En string `json:"en,omitempty"`
}
