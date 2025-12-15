package routes

import (
	"sci-stock-api/config"
	"sci-stock-api/controllers"
	"sci-stock-api/middleware"

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
		auth.POST("/register", controllers.Register)                                                                // POST /auth/register // ลงทะเบียนผู้ใช้ใหม่ (ยังไม่สร้างบัญชีจริง รอ OTP)
		auth.POST("/login", controllers.Login)                                                                      // POST /auth/login // เข้าสู่ระบบ รับ token
		auth.GET("/profile", middleware.JWTAuthMiddleware(), controllers.Profile)                                   // GET /auth/profile // ดูข้อมูลโปรไฟล์ผู้ใช้ (ต้อง login)
		auth.PUT("/profile", middleware.JWTAuthMiddleware(), controllers.UpdateOwnProfile)                          // PUT /auth/profile // อัปเดตข้อมูลโปรไฟล์ผู้ใช้ (ต้อง login)
		auth.POST("/refresh", middleware.JWTAuthMiddleware(), controllers.RefreshToken)                             // POST /auth/refresh // รีเฟรช access token (ต้อง login)
		auth.POST("/forgot-password", controllers.ForgotPassword)                                                   // POST /auth/forgot-password // ขอรีเซ็ตรหัสผ่าน
		auth.POST("/reset-password", controllers.ResetPassword)                                                     // POST /auth/reset-password // รีเซ็ตรหัสผ่าน
		auth.POST("/verify-email", controllers.VerifyUser)                                                          // POST /auth/verify-email // ยืนยันอีเมลผู้ใช้
		auth.PUT("/change-password", middleware.JWTAuthMiddleware(), controllers.ChangeOwnPassword)                 // PUT /auth/change-password // เปลี่ยนรหัสผ่านตัวเอง (ต้อง login)
		auth.PUT("/users/:id/change-password", middleware.JWTAuthMiddleware(), controllers.AdminChangeUserPassword) // PUT /auth/users/:id/change-password // admin เปลี่ยนรหัสผ่านผู้ใช้อื่น (ต้อง login)
	}

	// สมัครพนักงานใหม่ (ไม่ต้อง login)
	r.POST("/api/employees/register", controllers.HandleEmployeeRegister) // POST /api/employees/register // สมัครพนักงานใหม่ (อัปโหลดไฟล์แนบได้)

	// กลุ่ม route สำหรับ API ที่ต้อง login ทุกครั้ง
	api := r.Group("/api")
	api.Use(middleware.JWTAuthMiddleware()) // ใช้ JWT ตรวจสอบ token ทุก route ในกลุ่มนี้
	{
		// จัดการสินค้าตามหมวดหมู่
		api.GET("/products/:category", controllers.GetProductsByCategory)               // GET /api/products/:category // ดึงสินค้าในหมวดหมู่
		api.POST("/products/:category", controllers.CreateProductByCategory)            // POST /api/products/:category // เพิ่มสินค้าในหมวดหมู่
		api.POST("/products/:category/bulk", controllers.CreateProductsBulkByCategory)  // POST /api/products/:category/bulk // เพิ่มสินค้าจำนวนมากในหมวดหมู่
		api.PUT("/products/:category/:barcode", controllers.UpdateProductByCategory)    // PUT /api/products/:category/:barcode // อัปเดตสินค้าตามหมวดหมู่และบาร์โค้ด
		api.DELETE("/products/:category/:barcode", controllers.DeleteProductByCategory) // DELETE /api/products/:category/:barcode // ลบสินค้าตามหมวดหมู่และบาร์โค้ด

		// จัดการแดชบอร์ด
		api.GET("/dashboard/total", controllers.GetTotalProducts)             // GET /api/dashboard/total // ดึงข้อมูลสรุปจำนวนสินค้าทั้งหมด
		api.GET("/dashboard/low-stock", controllers.GetLowStockProducts)      // GET /api/dashboard/low-stock // ดึงข้อมูลสินค้าที่ใกล้หมดสต๊อก
		api.GET("/dashboard/out-of-stock", controllers.GetOutOfStockProducts) // GET /api/dashboard/out-of-stock // ดึงข้อมูลสินค้าที่หมดสต๊อก
		api.GET("/dashboard/sales-summary", controllers.GetMonthlySalesSummary) // GET /api/dashboard/sales-summary // ดึงข้อมูลสรุปยอดขายรายเดือน
		api.GET("/dashboard/sales-weekly", controllers.GetWeeklySalesCurrentMonth) // GET /api/dashboard/sales-weekly // ดึงข้อมูลสรุปยอดขายรายสัปดาห์
		api.GET("/dashboard/top-selling-products", controllers.GetTopSellingProductsCurrentMonth,) // GET /api/dashboard/top-selling-products // ดึงข้อมูลสินค้าขายดี 3 อันดับแรกของเดือน

		// จัดการคำสั่งซื้อ
		api.POST("/orders", controllers.SellProduct)                  // POST /api/orders // สร้างคำสั่งซื้อ (ขายสินค้า)
		api.GET("/fromsheet", controllers.GetProductsFromSheet)       // GET /api/fromsheet // ดึงข้อมูลสินค้าจาก Google Sheets
		api.GET("/sales_today", controllers.GetSalesToday)            // GET /api/sales_today // ดึงยอดขายวันนี้
		api.GET("/expenses/today", controllers.GetDailyExpenses)      // GET /api/expenses/today // ดึงข้อมูลรายจ่ายวันนี้
		api.POST("/sell-local", controllers.SellProductLocal)         // POST /api/sell-local // ขายสินค้าแบบออฟไลน์ (บันทึกคำสั่งซื้อจากเครื่องลูกข่าย)
		api.GET("/product/:barcode", controllers.GetProductByBarcode) // GET /api/product/:barcode // ดึงข้อมูลสินค้าตามบาร์โค้ด
		api.POST("/daily-payments", controllers.CreateDailyPayment)   // POST /api/daily-payments // บันทึกการชำระเงินรายวัน

		// จัดการแคชสินค้า
		api.GET("/refresh-cache", middleware.AdminOrSuperAdmin(), controllers.RefreshCache) // GET /api/refresh-cache // รีเฟรชแคชสินค้า (จำกัดสิทธิ์เฉพาะ admin/superadmin)
		api.POST("/auth/refresh", controllers.RefreshAccessToken)                           // กลุ่มจัดการผู้ใช้ (จำกัดสิทธิ์ admin หรือ superadmin เท่านั้น)

		// จัดการใบสมัครพนักงาน
		api.GET("/employees/applications", controllers.GetStudentApplications)                           // GET /api/employees/applications // ดึงข้อมูลใบสมัครพนักงานทั้งหมด
		api.PUT("/employees/applications/:id", controllers.ApproveStudentApplication)                    // PUT /api/employees/applications/:id // อนุมัติใบสมัครพนักงานตาม id
		api.DELETE("/employees/applications/approved/:id", controllers.DeleteApprovedStudentApplication) // DELETE /api/employees/applications/approved/:id // ลบใบสมัครพนักงานที่อนุมัติแล้วตาม id
		api.POST("/employees/check-or-add", controllers.CheckOrAddEmployee) 						  // POST /api/employees/check-or-add // ตรวจสอบหรือเพิ่มพนักงานใหม่จากข้อมูลใบสมัคร
		api.DELETE("employees/:gmail", controllers.DeleteEmployeeByGmail) 					// DELETE /api/employees/:gmail // ลบพนักงานตาม gmail (จำกัดสิทธิ์เฉพาะ admin/superadmin)

		// จัดการตารางการทำงาน
		api.GET("/work-schedules", controllers.GetWorkSchedules)          // GET /api/work-schedules // ดึงตารางการทำงานทั้งหมด
		api.POST("/work-schedules", controllers.CreateWorkSchedule)       // POST /api/work-schedules // สร้างตารางการทำงานใหม่
		api.PUT("/work-schedules/:id", controllers.UpdateWorkSchedule)    // PUT /api/work-schedules/:id // อัปเดตตารางการทำงานตาม id
		api.DELETE("/work-schedules/:id", controllers.DeleteWorkSchedule) // DELETE /api/work-schedules/:id // ลบตารางการทำงานตาม id

		// จัดการผู้ใช้
		usersGroup := api.Group("/users")              // กลุ่มจัดการผู้ใช้ (จำกัดสิทธิ์ admin หรือ superadmin เท่านั้น)
		usersGroup.Use(middleware.AdminOrSuperAdmin()) // middleware ตรวจสอบ role

		{
			// CRUD ผู้ใช้
			usersGroup.GET("", controllers.GetUsers)          // GET /api/users // ดึงรายชื่อผู้ใช้ทั้งหมด
			usersGroup.PUT("/:gmail", controllers.UpdateUser)    // PUT /api/users/:gmail // อัปเดตข้อมูลผู้ใช้ตาม gmail
			usersGroup.DELETE("/gmail/:gmail", controllers.DeleteUser) // DELETE /api/users/gmail/:gmail // ลบผู้ใช้ตาม gmail

			// จัดการผู้ใช้ที่ยังรออนุมัติ (สร้างคำขอและยืนยัน OTP)
			usersGroup.POST("/requests", controllers.CreateUserRequestByAdmin)     // POST /api/users/requests // สร้างคำขอสร้างผู้ใช้ใหม่ (รอ OTP)
			usersGroup.POST("/requests/verify", controllers.VerifyAndActivateUser) // POST /api/users/requests/verify // ยืนยัน OTP เพื่อสร้างผู้ใช้จริง
		}
	}
}