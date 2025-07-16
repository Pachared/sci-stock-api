package main

import (
	"github.com/gin-gonic/gin"
	"sci-stock-api/config"
	"sci-stock-api/routes"
)

func main() {
	config.Connect() // เซ็ต config.DB

	router := gin.Default()
	routes.SetupRoutes(router)  // ส่งแค่ router

	router.Run(":8080")
}
