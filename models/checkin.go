package models

import (
	"time"
)

type Checkin struct {
	ID             uint       `gorm:"primaryKey"`
	UserID         uint       `gorm:"index;column:user_id"`
	CheckinAt      time.Time  `gorm:"column:checkin_at"`
	CheckoutAt     *time.Time `gorm:"column:checkout_at"`
	CheckinPhoto   []byte     `gorm:"column:checkin_photo;type:longblob"`
	CheckoutPhoto  []byte     `gorm:"column:checkout_photo;type:longblob"`
	IsAutoCheckout bool       `gorm:"column:is_auto_checkout;default:false"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

type CheckinWithUser struct {
	UserID        uint       `json:"user_id"`
	FirstName     string     `json:"first_name"`
	CheckinAt     time.Time  `json:"checkin_at"`
	CheckoutAt    *time.Time `json:"checkout_at"`
	CheckinPhoto  []byte     `json:"checkin_photo"`
	CheckoutPhoto []byte     `json:"checkout_photo"`
}
