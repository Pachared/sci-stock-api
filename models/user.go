package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Gmail        string    `gorm:"size:100;unique;not null" json:"gmail"`
	Password     string    `gorm:"size:255;not null" json:"-"`
	TwoFASecret  string    `gorm:"column:two_fa_secret"`
	TwoFAEnabled bool      `gorm:"column:two_fa_enabled"`
	FirstName    string    `gorm:"size:100" json:"first_name"`
	LastName     string    `gorm:"size:100" json:"last_name"`
	RoleID       uint      `json:"role_id"`
	ProfileImage []byte    `gorm:"type:longblob"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	Role         Role      `gorm:"foreignKey:RoleID" json:"role"`
}

type UserProfileResponse struct {
	Gmail        string `json:"gmail"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	RoleID       uint   `json:"role_id"`
	RoleName     string `json:"role_name"`
	ProfileImage string `json:"profile_image"`
}

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index;not null"`
	Token     string    `gorm:"size:255;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
}
