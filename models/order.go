package models

import "time"

type Order struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	OrderDate time.Time   `gorm:"autoCreateTime" json:"order_date"`
	Items     []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

type OrderItem struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	OrderID     uint    `json:"order_id"`
	ProductName string  `gorm:"size:255" json:"product_name"`
	Barcode     string  `gorm:"size:50" json:"barcode"`
	Price       float64 `gorm:"type:decimal(10,2)" json:"price"`
	Quantity    int     `json:"quantity"`
	ImageURL    string  `gorm:"type:longtext" json:"image_url"`
}
