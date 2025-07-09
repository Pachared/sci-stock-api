package models

type Product struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	Category     string  `gorm:"size:50;not null" json:"category"`
	ProductName  string  `gorm:"size:255;not null" json:"product_name"`
	Barcode      string  `gorm:"size:100;not null" json:"barcode"`
	Price        float64 `gorm:"type:decimal(10,2);not null" json:"price"`
	Cost         float64 `gorm:"type:decimal(10,2);not null" json:"cost"`
	Stock        int     `gorm:"not null;default:0" json:"stock"`
	ReorderLevel int     `gorm:"not null;default:0" json:"reorder_level"`
	ImageURL     string  `gorm:"type:longtext" json:"image_url"`
}