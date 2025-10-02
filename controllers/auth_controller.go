package controllers

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"sci-stock-api/config"
	"sci-stock-api/models"
	"sci-stock-api/services"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	var input struct {
		Gmail    string `json:"gmail" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลที่ส่งมาไม่ถูกต้อง: " + err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Preload("Role").Where("gmail = ?", input.Gmail).First(&user).Error; err != nil {
		log.Printf("Failed login attempt for gmail: %s from IP: %s", input.Gmail, c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "อีเมลหรือรหัสผ่านไม่ถูกต้อง"})
		return
	}

	if !services.CheckPasswordHash(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "อีเมลหรือรหัสผ่านไม่ถูกต้อง"})
		return
	}

	token, err := services.GenerateJWT(user.ID, user.Role.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถสร้าง access token ได้"})
		return
	}

	refreshToken, err := services.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถสร้าง refresh token ได้"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         token,
		"refresh_token": refreshToken,
		"role":          user.Role.Name,
	})
}

func Profile(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var user models.User
	if err := config.DB.Preload("Role").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่พบข้อมูลผู้ใช้"})
		return
	}

	var profileImage string
	if len(user.ProfileImage) > 0 {
		profileImage = base64.StdEncoding.EncodeToString(user.ProfileImage)
	} else {
		profileImage = ""
	}

	var roleID uint32 = uint32(user.RoleID)

	resp := models.UserProfileResponse{
		Gmail:        user.Gmail,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		RoleID:       roleID,
		RoleName:     user.Role.Name,
		ProfileImage: profileImage,
	}

	c.JSON(http.StatusOK, resp)
}

func UpdateOwnProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่สามารถอ่านข้อมูลได้"})
		return
	}

	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")

	file, _, err := c.Request.FormFile("profile_image")
	var imageData []byte
	if err == nil {
		defer file.Close()
		imageData, err = io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่สามารถอ่านรูปภาพได้"})
			return
		}
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบผู้ใช้"})
		return
	}

	user.FirstName = firstName
	user.LastName = lastName
	if len(imageData) > 0 {
		user.ProfileImage = imageData
	}

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกข้อมูลได้"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "อัปเดตข้อมูลสำเร็จ"})
}

func RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่พบข้อมูล Authorization header"})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รูปแบบ Authorization header ไม่ถูกต้อง"})
		return
	}
	refreshToken := parts[1]

	claims, err := services.ParseRefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token ไม่ถูกต้องหรือหมดอายุ"})
		return
	}

	userID := claims.UserID

	var tokenRecord models.RefreshToken
	err = config.DB.Where("user_id = ? AND token = ?", userID, refreshToken).First(&tokenRecord).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token นี้ไม่ถูกจดจำในระบบ"})
		return
	}

	if tokenRecord.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token หมดอายุ"})
		return
	}

	var user models.User
	if err := config.DB.Preload("Role").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่พบข้อมูลผู้ใช้"})
		return
	}

	newAccessToken, err := services.GenerateJWT(user.ID, user.Role.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถสร้าง Access token ใหม่ได้"})
		return
	}

	newRefreshToken, err := services.GenerateRefreshToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถสร้าง Refresh token ใหม่ได้"})
		return
	}

	if err := config.DB.Model(&tokenRecord).Update("token", newRefreshToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอัปเดต Refresh token ใหม่ได้"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "อีเมลนี้กำลังรอยืนยัน OTP อยู่"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดในการเข้ารหัสรหัสผ่าน"})
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

func VerifyUser(c *gin.Context) {
	var input struct {
		Gmail string `json:"gmail" binding:"required,email"`
		OTP   string `json:"otp" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง: " + err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "รหัส OTP ไม่ถูกต้อง"})
		return
	}

	if time.Now().In(loc).After(userVerif.OtpExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รหัส OTP หมดอายุแล้ว"})
		return
	}

	roleID := uint64(4)
	user := models.User{
		Gmail:        userVerif.Gmail,
		Password:     userVerif.Password,
		FirstName:    userVerif.FirstName,
		LastName:     userVerif.LastName,
		RoleID:       roleID,
		ProfileImage: userVerif.Image,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถสร้างบัญชีผู้ใช้ได้"})
		return
	}

	config.DB.Exec("DELETE FROM user_verifications WHERE gmail = ?", input.Gmail)

	c.JSON(http.StatusOK, gin.H{"message": "สมัครสมาชิกและยืนยันสำเร็จ"})
}

func RefreshAccessToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "คำขอไม่ถูกต้อง"})
		return
	}

	claims, err := services.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token ไม่ถูกต้องหรือหมดอายุ"})
		return
	}

	var rt models.RefreshToken
	if err := config.DB.Where("user_id = ? AND token = ?", claims.UserID, req.RefreshToken).First(&rt).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบ Refresh token ในระบบ"})
		return
	}

	user := models.User{}
	config.DB.First(&user, claims.UserID)

	newToken, err := services.GenerateJWT(user.ID, user.Role.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถสร้าง Access token ได้"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": newToken,
	})
}
