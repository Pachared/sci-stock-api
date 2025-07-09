package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Gmail        string    `gorm:"size:100;unique;not null" json:"gmail"`
	Password     string    `gorm:"size:255;not null" json:"-"`
	FirstName    string    `gorm:"size:100" json:"first_name"`
	LastName     string    `gorm:"size:100" json:"last_name"`
	RoleID       uint      `json:"role_id"`
	ProfileImage []byte    `json:"-"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	Role         Role      `gorm:"foreignKey:RoleID" json:"role"`
}