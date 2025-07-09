package main

import (
	"github.com/gin-gonic/gin"
	"sci-stock-api/config"
	"sci-stock-api/routes"
)

func main() {
	config.Connect()

	router := gin.Default()
	routes.RegisterRoutes(router)

	router.Run(":8080")
}