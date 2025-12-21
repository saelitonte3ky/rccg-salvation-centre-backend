// internal/models/attendance.go
package models

import "time"

type Attendance struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Date        time.Time `gorm:"not null" json:"date"`
	ServiceType string    `gorm:"size:100;not null" json:"serviceType"`
	Adults      int       `json:"adults"`
	Children    int       `json:"children"`
	Total       int       `json:"total"`
	FirstTimers int       `json:"firstTimers"`
	Visitors    int       `json:"visitors"`
	Members     int       `json:"members"`
	Notes       string    `gorm:"type:text" json:"notes"`
	RecordedBy  string    `gorm:"size:100" json:"recordedBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
