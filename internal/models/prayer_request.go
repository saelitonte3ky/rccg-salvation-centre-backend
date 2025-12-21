// internal/models/prayer_request.go
package models

import "time"

type PrayerRequest struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Email       string    `gorm:"size:255;not null" json:"email"`
	Request     string    `gorm:"type:text;not null" json:"request"`
	Status      string    `gorm:"size:50;default:'pending'" json:"status"` // pending, prayed, archived
	SubmittedAt time.Time `gorm:"autoCreateTime" json:"submittedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
