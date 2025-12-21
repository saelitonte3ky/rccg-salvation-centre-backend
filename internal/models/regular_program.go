package models

import "time"

type RegularProgram struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:255;not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Day         string    `gorm:"size:50;not null" json:"day"`
	Frequency   string    `gorm:"size:100;not null" json:"frequency"`
	Time        string    `gorm:"size:50" json:"time"`
	Location    string    `gorm:"size:255" json:"location"`
	Type        string    `gorm:"size:100;not null" json:"type"`
	Active      bool      `gorm:"default:true" json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
