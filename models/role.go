package models

type Role struct {
	ID   uint32 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"size:50;not null" json:"name"`
}