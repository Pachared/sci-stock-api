package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"sci-stock-api/models"
)

var DB *gorm.DB

func ConnectDatabase() {
	if err := godotenv.Load(); err != nil {
		log.Println("ไม่พบไฟล์ .env กำลังใช้ค่าจาก environment ของระบบแทน")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("MYSQLUSER"),
		os.Getenv("MYSQLPASSWORD"),
		os.Getenv("MYSQLHOST"),
		os.Getenv("MYSQLPORT"),
		os.Getenv("MYSQLDATABASE"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("เชื่อมต่อฐานข้อมูลไม่สำเร็จ:", err)
	}

	DB = db

	db.AutoMigrate(
		&models.Role{},
		&models.User{},
		&models.DriedFood{},
		&models.FreshFood{},
		&models.Snack{},
		&models.SoftDrink{},
		&models.Stationery{},
		&models.Checkin{},
		&models.StudentApplication{},
		&models.SaleToday{},
		&models.DailyPayment{},
		&models.WorkSchedule{},
		&models.RefreshToken{},
	)
}