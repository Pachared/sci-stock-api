package models

import "time"

type SaleToday struct {
	ID          int       `json:"id" gorm:"column:id;primaryKey"`
	ProductName string    `json:"product_name" gorm:"column:product_name"`
	Barcode     string    `json:"barcode" gorm:"column:barcode"`
	Cost        float64   `json:"cost" gorm:"column:cost"`
	Price       float64   `json:"price" gorm:"column:price"`
	Quantity    int       `json:"quantity" gorm:"column:quantity"`
	SoldAt      time.Time `json:"sold_at" gorm:"column:sold_at"`
	ImageURL    *string   `json:"image_url,omitempty" gorm:"column:image_url"`
}

type DailyPayment struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ItemName    string    `json:"item_name" gorm:"column:item_name;not null"`
	Amount      float64   `json:"amount" gorm:"not null"`
	PaymentDate string    `json:"payment_date" gorm:"not null;type:date"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type DailyExpense struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    ItemName    string    `json:"item_name"`
    Amount      float64   `json:"amount"`
    PaymentDate time.Time `json:"payment_date"`
    CreatedAt   time.Time `json:"created_at"`
}