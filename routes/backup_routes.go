package routes

import (
	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
	"sci-stock-api/controllers"
	"sci-stock-api/middleware"
)

func BackupRoutes(r *gin.Engine, db *gorm.DB) {

	superadmin := r.Group("/api/superadmin")
	superadmin.Use(
		middleware.JWTAuthMiddleware(),
		middleware.RoleAuthorization("superadmin"),
	)
	{
		ctrl := controllers.NewBackupController(db)
		superadmin.GET("/backup/:table", ctrl.BackupTable)
		superadmin.GET("/backup-all", ctrl.BackupAll)
		superadmin.POST("/import-data", ctrl.ImportData)
	}
}
