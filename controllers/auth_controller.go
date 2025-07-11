package controllers

import (
	"io"
	"net/http"
	"sci-stock-api/config"
	"sci-stock-api/models"
	"sci-stock-api/services"
	"time"

	"github.com/gin-gonic/gin"
)

// Login รับ email กับ password แล้วตรวจสอบ จากนั้นส่ง JWT token กลับ
func Login(c *gin.Context) {
	var input struct {
		Gmail    string `json:"gmail" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("gmail = ?", input.Gmail).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !services.CheckPasswordHash(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := services.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Profile ดึงข้อมูล user ตาม userID ที่อยู่ใน JWT
func Profile(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// RefreshToken สร้าง JWT token ใหม่
func RefreshToken(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	newToken, err := services.GenerateJWT(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": newToken})
}

// Register รับสมัครและส่ง OTP ยืนยันอีเมล
func Register(c *gin.Context) {

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่สามารถอ่านข้อมูลที่ส่งมาได้"})
		return
	}

	gmail := c.PostForm("gmail")
	password := c.PostForm("password")
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")

	if gmail == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณากรอกอีเมลและรหัสผ่าน"})
		return
	}

	var count int64
	config.DB.Model(&models.User{}).Where("gmail = ?", gmail).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "อีเมลนี้ถูกใช้งานแล้ว"})
		return
	}
	config.DB.Table("user_verifications").Where("gmail = ?", gmail).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "อีเมลนี้กำลังรอยืนยัน OTP"})
		return
	}

	file, _, err := c.Request.FormFile("profile_image")
	var imageData []byte
	if err == nil {
		defer file.Close()
		imageData, _ = io.ReadAll(file)
	}

	hashedPass, err := services.HashPassword(password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เข้ารหัสรหัสผ่านไม่สำเร็จ"})
		return
	}

	otp := services.GenerateOTP(6)
	loc, _ := time.LoadLocation("Asia/Bangkok")
	expire := time.Now().In(loc).Add(10 * time.Minute)
	now := time.Now().In(loc)

	err = config.DB.Exec(`
		INSERT INTO user_verifications (gmail, password, first_name, last_name, profile_image, otp, otp_expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, gmail, hashedPass, firstName, lastName, imageData, otp, expire, now).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกข้อมูลได้"})
		return
	}

	html, plain := services.GenerateEmailBodyForRegisterOTP(otp)
	if err := services.SendEmail(gmail, "ยืนยันอีเมลสำหรับสมัครสมาชิก", html, plain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ส่งอีเมล OTP ไม่สำเร็จ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ส่ง OTP ไปยังอีเมลแล้ว กรุณายืนยันเพื่อสมัครสมาชิก"})
}

// VerifyUser ตรวจสอบ OTP และยืนยันการสมัคร
func VerifyUser(c *gin.Context) {
	var input struct {
		Gmail string `json:"gmail" binding:"required,email"`
		OTP   string `json:"otp" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loc, _ := time.LoadLocation("Asia/Bangkok")

	var userVerif struct {
		Gmail        string
		Password     string
		FirstName    string
		LastName     string
		Image        []byte `gorm:"column:profile_image"`
		OTP          string
		OtpExpiresAt time.Time
	}

	err := config.DB.Raw(`
		SELECT gmail, password, first_name, last_name, profile_image, otp, otp_expires_at
		FROM user_verifications
		WHERE gmail = ? ORDER BY id DESC LIMIT 1
	`, input.Gmail).Scan(&userVerif).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดในการตรวจสอบ OTP"})
		return
	}

	if userVerif.OTP != input.OTP {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP ไม่ถูกต้อง"})
		return
	}

	if time.Now().In(loc).After(userVerif.OtpExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP หมดอายุแล้ว"})
		return
	}

	user := models.User{
		Gmail:        userVerif.Gmail,
		Password:     userVerif.Password,
		FirstName:    userVerif.FirstName,
		LastName:     userVerif.LastName,
		RoleID:       4,
		ProfileImage: userVerif.Image,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถสร้างบัญชีผู้ใช้ได้"})
		return
	}

	config.DB.Exec("DELETE FROM user_verifications WHERE gmail = ?", input.Gmail)

	c.JSON(http.StatusOK, gin.H{"message": "สมัครสมาชิกและยืนยันสำเร็จ"})
}

func ForgotPassword(c *gin.Context) {
	var input struct {
		Gmail string `json:"gmail" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	otp := services.GenerateOTP(6)
	loc, _ := time.LoadLocation("Asia/Bangkok")
	expire := time.Now().In(loc).Add(10 * time.Minute)

	err := config.DB.Exec(`
		INSERT INTO password_resets (gmail, otp, expires_at, created_at)
		VALUES (?, ?, ?, ?)
	`, input.Gmail, otp, expire, time.Now().In(loc)).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึก OTP ได้"})
		return
	}

	html, plain := services.GenerateEmailBodyForOTP(otp)
	if err := services.SendEmail(input.Gmail, "Reset Password OTP", html, plain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ส่งอีเมลไม่สำเร็จ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ส่งรหัส OTP ไปยังอีเมลแล้ว"})
}

func ResetPassword(c *gin.Context) {
	var input struct {
		Gmail       string `json:"gmail" binding:"required,email"`
		OTP         string `json:"otp" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loc, _ := time.LoadLocation("Asia/Bangkok")
	var otpEntry struct {
		OTP       string
		ExpiresAt time.Time
	}

	err := config.DB.Raw(`
		SELECT otp, expires_at
		FROM password_resets
		WHERE gmail = ? ORDER BY created_at DESC LIMIT 1
	`, input.Gmail).Scan(&otpEntry).Error

	if err != nil || otpEntry.OTP != input.OTP {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP ไม่ถูกต้อง"})
		return
	}

	if time.Now().In(loc).After(otpEntry.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP หมดอายุแล้ว"})
		return
	}

	hashed, _ := services.HashPassword(input.NewPassword)
	config.DB.Model(&models.User{}).Where("gmail = ?", input.Gmail).Update("password", hashed)

	c.JSON(http.StatusOK, gin.H{"message": "เปลี่ยนรหัสผ่านเรียบร้อยแล้ว"})
}
