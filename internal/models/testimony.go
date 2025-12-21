// internal/models/testimony.go
package models

import "time"

type TestimonyStatus string

const (
	Pending  TestimonyStatus = "pending"
	Approved TestimonyStatus = "approved"
	Rejected TestimonyStatus = "rejected"
)

type Testimony struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	Name        string          `gorm:"not null" json:"name"`
	Title       string          `gorm:"not null" json:"title"`
	Message     string          `gorm:"type:text;not null" json:"message"`
	Email       string          `json:"email,omitempty"`
	Phone       string          `json:"phone,omitempty"`
	Status      TestimonyStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	ApprovedAt  *time.Time      `json:"approvedAt,omitempty"`
	RejectedAt  *time.Time      `json:"rejectedAt,omitempty"`
	SubmittedAt time.Time       `gorm:"autoCreateTime" json:"submittedAt"`
}
