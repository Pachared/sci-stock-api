package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	"sci-stock-api/config"
	"sci-stock-api/models"
	"sci-stock-api/routes"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, reading environment variables from system")
	}

	config.Connect()
	db := config.DB

	if err := db.AutoMigrate(&models.Role{}, &models.User{}); err != nil {
		log.Fatal("Migration failed: ", err)
	}

	roles := []string{"superadmin", "admin", "employee", "user"}
	for _, r := range roles {
		db.FirstOrCreate(&models.Role{}, models.Role{Name: r})
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	routes.SetupRoutes(router)
	routes.BackupRoutes(router , db)
	router.Run(":8080")
}