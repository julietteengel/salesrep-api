package common

import (
	"time"

	"gorm.io/gorm"
)

type BaseModelV2 struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedBy *uint
	UpdatedBy *uint
	DeletedBy *uint
}

type BaseTranslationMandatory struct {
	Fr string `json:"fr,omitempty"`
	En string `json:"en,omitempty"`
}
