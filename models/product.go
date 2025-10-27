package models

type DriedFood struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	ProductName  string  `gorm:"column:product_name" json:"product_name"`
	Barcode      string  `json:"barcode"`
	Price        float64 `json:"price"`
	Cost         float64 `json:"cost"`
	Stock        int     `json:"stock"`
	ReorderLevel int     `json:"reorder_level"`
	ImageURL     string  `gorm:"column:image_url" json:"image_url"`
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
	ImageURL     string  `gorm:"column:image_url" json:"image_url"`
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
	ImageURL     string  `gorm:"column:image_url" json:"image_url"`
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
	ImageURL     string  `gorm:"column:image_url" json:"image_url"`
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
	ImageURL     string  `gorm:"column:image_url" json:"image_url"`
}

func (Stationery) TableName() string {
	return "stationery"
}

type Product struct {
	ID          uint    `gorm:"column:id" json:"id"`
	ProductName string  `gorm:"column:product_name" json:"product_name"`
	Barcode     string  `gorm:"column:barcode" json:"barcode"`
	Price       float64 `gorm:"column:price" json:"price"`
	Cost        float64 `gorm:"column:cost" json:"cost"`
	Stock       int     `gorm:"column:stock" json:"stock"`
	ImageURL    string  `gorm:"column:image_url" json:"image_url"`
	Category    string  `gorm:"-"`
}

type ProductResponse struct {
	ID          uint    `json:"id"`
	ProductName string  `json:"product_name"`
	Barcode     string  `json:"barcode"`
	Price       float64 `json:"price"`
	Cost        float64 `json:"cost"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ProductInput struct {
	ProductName  string  `json:"product_name" binding:"required"`
	Barcode      string  `json:"barcode" binding:"required"`
	Price        float64 `json:"price" binding:"required"`
	Cost         float64 `json:"cost" binding:"required"`
	Stock        int     `json:"stock"`
	ReorderLevel int     `json:"reorder_level"`
	ImageURL     string  `json:"image_url"`
}

type ProductDashboard struct {
	ID           uint  `json:"id"`
	ProductName  string `json:"product_name"`
	Barcode      string `json:"barcode"`
	Stock        int  `json:"stock"`
	ReorderLevel int  `json:"reorder_level"`
}