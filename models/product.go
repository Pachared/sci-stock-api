package models

type DriedFood struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	ProductName  string  `gorm:"column:product_name" json:"product_name"`
	Barcode      string  `json:"barcode"`
	Price        float64 `json:"price"`
	Cost         float64 `json:"cost"`
	Stock        int     `json:"stock"`
	ReorderLevel int     `json:"reorder_level"`
	ImageURL     string  `json:"image_url"`
}
func (DriedFood) TableName() string {
	return "dried_food"
}

type FreshFood struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	ProductName  string  `gorm:"column:product_name" json:"product_name"`
	Barcode      string  `json:"barcode"`
	Price        float64 `json:"price"`
	Cost         float64 `json:"cost"`
	Stock        int     `json:"stock"`
	ReorderLevel int     `json:"reorder_level"`
	ImageURL     string  `json:"image_url"`
}
func (FreshFood) TableName() string {
	return "fresh_food"
}

type Snack struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	ProductName  string  `gorm:"column:product_name" json:"product_name"`
	Barcode      string  `json:"barcode"`
	Price        float64 `json:"price"`
	Cost         float64 `json:"cost"`
	Stock        int     `json:"stock"`
	ReorderLevel int     `json:"reorder_level"`
	ImageURL     string  `json:"image_url"`
}
func (Snack) TableName() string {
	return "snack"
}

type SoftDrink struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	ProductName  string  `gorm:"column:product_name" json:"product_name"`
	Barcode      string  `json:"barcode"`
	Price        float64 `json:"price"`
	Cost         float64 `json:"cost"`
	Stock        int     `json:"stock"`
	ReorderLevel int     `json:"reorder_level"`
	ImageURL     string  `json:"image_url"`
}
func (SoftDrink) TableName() string {
	return "soft_drink"
}

type Stationery struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	ProductName  string  `gorm:"column:product_name" json:"product_name"`
	Barcode      string  `json:"barcode"`
	Price        float64 `json:"price"`
	Cost         float64 `json:"cost"`
	Stock        int     `json:"stock"`
	ReorderLevel int     `json:"reorder_level"`
	ImageURL     string  `json:"image_url"`
}
func (Stationery) TableName() string {
	return "stationery"
}