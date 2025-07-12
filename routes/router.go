package routes

import (
	"sci-stock-api/controllers"
	"sci-stock-api/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.Use(middleware.CORSMiddleware())

	auth := r.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
		auth.POST("/enable-2fa", middleware.JWTAuthMiddleware(), controllers.EnableTwoFA)
		auth.POST("/confirm-2fa", middleware.JWTAuthMiddleware(), controllers.ConfirmEnableTwoFA)
		auth.GET("/profile", middleware.JWTAuthMiddleware(), controllers.Profile)
		auth.POST("/refresh", middleware.JWTAuthMiddleware(), controllers.RefreshToken)

		auth.POST("/forgot-password", controllers.ForgotPassword)
		auth.POST("/reset-password", controllers.ResetPassword)
		auth.POST("/verify-email", controllers.VerifyUser)

		auth.PUT("/change-password", middleware.JWTAuthMiddleware(), controllers.ChangeOwnPassword)
		auth.PUT("/users/:id/change-password", middleware.JWTAuthMiddleware(), controllers.AdminChangeUserPassword)
	}

	api := r.Group("/api")
	api.Use(middleware.JWTAuthMiddleware())
	{
		// Products CRUD
		api.GET("/products/:category", controllers.GetProductsByCategory)
		api.POST("/products/:category", controllers.CreateProductByCategory)

		// Orders
		api.POST("/orders", controllers.CreateOrder)
		api.GET("/orders", controllers.GetOrders)

		// Users (admin)
		api.GET("/users", controllers.GetUsers)
		api.PUT("/users/:id", controllers.UpdateUser)
		api.DELETE("/users/:id", controllers.DeleteUser)

		// สำหรับสร้าง user รอ OTP และยืนยัน OTP
		api.POST("/users/pending", controllers.CreateUserRequestByAdmin)
		api.POST("/users/approve", controllers.VerifyAndActivateUser)
	}
}
