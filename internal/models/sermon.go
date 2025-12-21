// internal/models/sermon.go
package models

import "time"

// Sermon represents a church sermon (video message)
type Sermon struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Pastor      string    `gorm:"not null" json:"pastor"`
	Service     string    `gorm:"not null" json:"service"`
	Date        time.Time `gorm:"not null" json:"date"`
	YoutubeID   string    `gorm:"not null;unique" json:"youtubeId"`
	Duration    string    `json:"duration"`
	Description string    `gorm:"type:text" json:"description"`
	Published   bool      `gorm:"default:false" json:"published"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
