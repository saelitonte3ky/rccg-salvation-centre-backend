// internal/models/first_timer.go
package models

import "time"

type FirstTimer struct {
	ID                     uint      `gorm:"primaryKey" json:"id"`
	FirstName              string    `gorm:"size:100;not null" json:"firstName"`
	LastName               string    `gorm:"size:100;not null" json:"lastName"`
	Email                  string    `gorm:"size:100" json:"email"`
	Phone                  string    `gorm:"size:50" json:"phone"`
	Address                string    `gorm:"size:255" json:"address"`
	City                   string    `gorm:"size:100" json:"city"`
	State                  string    `gorm:"size:100" json:"state"`
	DateOfBirth            string    `gorm:"size:10" json:"dateOfBirth"` // YYYY-MM-DD
	Gender                 string    `gorm:"size:20" json:"gender"`
	MaritalStatus          string    `gorm:"size:50" json:"maritalStatus"`
	Occupation             string    `gorm:"size:100" json:"occupation"`
	VisitDate              time.Time `gorm:"not null" json:"visitDate"`
	HowDidYouHear          string    `gorm:"size:255" json:"howDidYouHear"`
	PrayerRequest          string    `gorm:"type:text" json:"prayerRequest"`
	InterestedInMembership bool      `json:"interestedInMembership"`
	FollowUpStatus         string    `gorm:"default:'pending'" json:"followUpStatus"` // pending, contacted, joined, etc.
	Status                 string    `gorm:"default:'new'" json:"status"`             // new, followed up, member
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}
