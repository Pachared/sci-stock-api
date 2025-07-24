package routes

import (
	"sci-stock-api/controllers"
	"sci-stock-api/middleware"
	"sci-stock-api/config"  // import config เพื่อใช้ config.DB

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// เปิดใช้งาน CORS middleware
	r.Use(middleware.CORSMiddleware())
	// ใส่ database instance ลง context ให้ทุก request ใช้งานได้
	r.Use(middleware.DBMiddleware(config.DB))

	// กลุ่ม route สำหรับระบบ Authentication เช่น ลงทะเบียน, เข้าสู่ระบบ
	auth := r.Group("/auth")
	{
		auth.POST("/register", controllers.Register)   // POST /auth/register // ลงทะเบียนผู้ใช้ใหม่ (ยังไม่สร้างบัญชีจริง รอ OTP)
		auth.POST("/login", controllers.Login)   // POST /auth/login // เข้าสู่ระบบ รับ token
		auth.POST("/enable-2fa", middleware.JWTAuthMiddleware(), controllers.EnableTwoFA)  // POST /auth/enable-2fa // เปิดใช้งาน 2FA (ต้อง login)
		auth.POST("/confirm-2fa", middleware.JWTAuthMiddleware(), controllers.ConfirmEnableTwoFA) // POST /auth/confirm-2fa // ยืนยันการเปิดใช้งาน 2FA (ต้อง login)
		auth.GET("/profile", middleware.JWTAuthMiddleware(), controllers.Profile)   // GET /auth/profile // ดูข้อมูลโปรไฟล์ผู้ใช้ (ต้อง login)
		auth.PUT("/profile", middleware.JWTAuthMiddleware(), controllers.UpdateOwnProfile) // PUT /auth/profile // อัปเดตข้อมูลโปรไฟล์ผู้ใช้ (ต้อง login)
		auth.POST("/refresh", middleware.JWTAuthMiddleware(), controllers.RefreshToken) // POST /auth/refresh // รีเฟรช access token (ต้อง login)
		auth.POST("/forgot-password", controllers.ForgotPassword)   // POST /auth/forgot-password // ขอรีเซ็ตรหัสผ่าน
		auth.POST("/reset-password", controllers.ResetPassword)     // POST /auth/reset-password // รีเซ็ตรหัสผ่าน
		auth.POST("/verify-email", controllers.VerifyUser)  // POST /auth/verify-email // ยืนยันอีเมลผู้ใช้
		auth.PUT("/change-password", middleware.JWTAuthMiddleware(), controllers.ChangeOwnPassword) // PUT /auth/change-password // เปลี่ยนรหัสผ่านตัวเอง (ต้อง login)
		auth.PUT("/users/:id/change-password", middleware.JWTAuthMiddleware(), controllers.AdminChangeUserPassword) // PUT /auth/users/:id/change-password // admin เปลี่ยนรหัสผ่านผู้ใช้อื่น (ต้อง login)
	}

	// กลุ่ม route สำหรับ API ที่ต้อง login ทุกครั้ง
	api := r.Group("/api")
	api.Use(middleware.JWTAuthMiddleware()) // ใช้ JWT ตรวจสอบ token ทุก route ในกลุ่มนี้
	{
		// จัดการสินค้าตามหมวดหมู่
		api.GET("/products/:category", controllers.GetProductsByCategory)    // GET /api/products/:category // ดึงสินค้าในหมวดหมู่
		api.POST("/products/:category", controllers.CreateProductByCategory) // POST /api/products/:category // เพิ่มสินค้าในหมวดหมู่
		api.POST("/products/:category/bulk", controllers.CreateProductsBulkByCategory)
	
		// จัดการคำสั่งซื้อ
		api.POST("/orders", controllers.SellProduct)                         // POST /api/orders // สร้างคำสั่งซื้อ (ขายสินค้า)
		api.GET("/fromsheet", controllers.GetProductsFromSheet)              // GET /api/fromsheet // ดึงข้อมูลสินค้าจาก Google Sheets
		api.GET("/sales_today", controllers.GetSalesToday)

		api.GET("/refresh-cache", middleware.AdminOrSuperAdmin(), controllers.RefreshCache) // GET /api/refresh-cache // รีเฟรชแคชสินค้า (จำกัดสิทธิ์เฉพาะ admin/superadmin)
		api.POST("/auth/refresh", controllers.RefreshAccessToken)
		// กลุ่มจัดการผู้ใช้ (จำกัดสิทธิ์ admin หรือ superadmin เท่านั้น)
		usersGroup := api.Group("/users")
		usersGroup.Use(middleware.AdminOrSuperAdmin()) // middleware ตรวจสอบ role

		{
			usersGroup.GET("", controllers.GetUsers)                  // GET /api/users // ดึงรายชื่อผู้ใช้ทั้งหมด
			usersGroup.PUT("/:id", controllers.UpdateUser)            // PUT /api/users/:id // อัปเดตข้อมูลผู้ใช้ตาม id
			usersGroup.DELETE("/:id", controllers.DeleteUser)         // DELETE /api/users/:id // ลบผู้ใช้ตาม id

			// จัดการผู้ใช้ที่ยังรออนุมัติ (สร้างคำขอและยืนยัน OTP)
			usersGroup.POST("/requests", controllers.CreateUserRequestByAdmin)     // POST /api/users/requests // สร้างคำขอสร้างผู้ใช้ใหม่ (รอ OTP)
			usersGroup.POST("/requests/verify", controllers.VerifyAndActivateUser) // POST /api/users/requests/verify // ยืนยัน OTP เพื่อสร้างผู้ใช้จริง
		}
	}
}


