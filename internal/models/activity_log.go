// internal/models/activity_log.go
package models

import "time"

type ActivityLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	AdminID    uint      `json:"adminId"`
	AdminEmail string    `json:"adminEmail"`
	Action     string    `gorm:"not null" json:"action"` // e.g., "Created sermon", "Approved testimony"
	Details    string    `gorm:"type:text" json:"details,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}
