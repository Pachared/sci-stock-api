package models

import "time"

type WorkSchedule struct {
	ID    uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title string    `gorm:"type:varchar(255);not null" json:"title"`
	Date  time.Time `gorm:"type:date;not null" json:"date"`
	Tag   string    `gorm:"type:varchar(50);not null" json:"tag"`
}