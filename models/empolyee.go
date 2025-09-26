package models

import "time"

type StudentApplicationResponse struct {
	ID         uint      `json:"id"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	StudentID  string    `json:"studentId"`
	Schedule   string    `json:"schedule"` // Base64 string
	Contact    string    `json:"contactInfo"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type StudentApplication struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    FirstName string    `json:"first_name"`
    LastName  string    `json:"last_name"`
    StudentID string    `json:"student_id"`
    Schedule  []byte    `json:"schedule"`
    Contact   string    `gorm:"column:contact_info" json:"contact_info"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}