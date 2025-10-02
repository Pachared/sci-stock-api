package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"sci-stock-api/config"
	"sci-stock-api/models"
	"sci-stock-api/routes"
)

func main() {
	config.Connect()
	db := config.DB

	err := db.AutoMigrate(&models.Role{}, &models.User{})
	if err != nil {
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
	router.Run(":8080")
}