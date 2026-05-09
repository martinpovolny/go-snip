package models

import (
	"time"

	"gorm.io/gorm"
)

type Person struct {
	gorm.Model  `json:"-"` // hides id, created_at, updated_at, deleted_at
	ExternalID  string     `json:"external_id"   gorm:"uniqueIndex; default:null; not null"`
	Name        string     `json:"name"          gorm:"default:null; not null"`
	Email       string     `json:"email"         gorm:"default:null; not null"`
	DateOfBirth time.Time  `json:"date_of_birth" gorm:"default:null; not null"`
}
