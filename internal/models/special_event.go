// internal/models/special_event.go
package models

import "time"

type SpecialEvent struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:255;not null" json:"title"`
	Type        string    `gorm:"size:100;not null" json:"type"`
	Description string    `gorm:"type:text" json:"description"`
	Date        time.Time `gorm:"not null" json:"date"`
	StartTime   string    `gorm:"size:20" json:"startTime"`
	EndTime     string    `gorm:"size:20" json:"endTime"`
	Location    string    `gorm:"size:255" json:"location"`
	Published   bool      `gorm:"default:false" json:"published"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
