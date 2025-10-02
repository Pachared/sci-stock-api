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
	RoleID       uint      `gorm:"not null"`
	ProfileImage []byte    `gorm:"type:longblob"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	Role         Role      `gorm:"foreignKey:RoleID"`
}

type UserProfileResponse struct {
	Gmail        string `json:"gmail"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	RoleID       uint32 `json:"roleId"`
	RoleName     string `json:"roleName"`
	ProfileImage string `json:"profileImage"`
}

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index;not null"`
	Token     string    `gorm:"size:255;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
}
