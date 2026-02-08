package models

import (
	"time"
)

type Checkin struct {
	ID             uint `gorm:"primaryKey"`
	UserID         uint `gorm:"index"`
	CheckinAt      time.Time
	CheckoutAt     *time.Time
	CheckinPhoto   []byte
	CheckoutPhoto  []byte
	IsAutoCheckin  bool    `gorm:"column:is_auto_checkin;default:false"`
	IsAutoCheckout bool    `gorm:"column:is_auto_checkout;default:false"`
	AutoReason     *string `gorm:"column:auto_reason"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
