package routes

import (
	"sci-stock-api/controllers"
	"sci-stock-api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.Use(middleware.CORSMiddleware())

	auth := r.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}

	api := r.Group("/api")
	api.Use(middleware.JWTAuthMiddleware())
	{
		// Products CRUD
		api.GET("/products", controllers.GetProducts)
		api.POST("/products", controllers.CreateProduct)
		api.PUT("/products/:id", controllers.UpdateProduct)
		api.DELETE("/products/:id", controllers.DeleteProduct)

		// Orders
		api.POST("/orders", controllers.CreateOrder)
		api.GET("/orders", controllers.GetOrders)

		// Users (admin)
		api.GET("/users", controllers.GetUsers)
		api.PUT("/users/:id", controllers.UpdateUser)
		api.DELETE("/users/:id", controllers.DeleteUser)
	}
}

func RegisterProductRoutes(router *gin.Engine) {
	product := router.Group("/products")
	{
		product.GET("/", controllers.GetProducts)
		product.POST("/", controllers.CreateProduct)
	}
}