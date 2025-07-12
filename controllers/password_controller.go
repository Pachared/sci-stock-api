package controllers

import (
	"net/http"
	"sci-stock-api/config"
	"sci-stock-api/models"
	"sci-stock-api/services"
	"time"

	"github.com/gin-gonic/gin"
)

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

	hashed, err := services.HashPassword(input.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เข้ารหัสรหัสผ่านไม่สำเร็จ"})
		return
	}

	if err := config.DB.Model(&models.User{}).Where("gmail = ?", input.Gmail).Update("password", hashed).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเปลี่ยนรหัสผ่านได้"})
		return
	}

	config.DB.Exec("DELETE FROM password_resets WHERE gmail = ?", input.Gmail)

	c.JSON(http.StatusOK, gin.H{"message": "เปลี่ยนรหัสผ่านเรียบร้อยแล้ว"})
}

func ChangeOwnPassword(c *gin.Context) {
	var input struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(uint)
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบผู้ใช้"})
		return
	}

	if !services.CheckPasswordHash(input.OldPassword, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รหัสผ่านเดิมไม่ถูกต้อง"})
		return
	}

	hashed, err := services.HashPassword(input.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เข้ารหัสรหัสผ่านไม่สำเร็จ"})
		return
	}

	config.DB.Model(&user).Update("password", hashed)
	c.JSON(http.StatusOK, gin.H{"message": "เปลี่ยนรหัสผ่านเรียบร้อยแล้ว"})
}

func AdminChangeUserPassword(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var currentUser models.User
	if err := config.DB.First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if currentUser.RoleID > 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "คุณไม่มีสิทธิ์เปลี่ยนรหัสผ่านให้ผู้อื่น"})
		return
	}

	targetID := c.Param("id")

	var input struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var targetUser models.User
	if err := config.DB.First(&targetUser, targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบผู้ใช้เป้าหมาย"})
		return
	}

	if currentUser.RoleID == 2 && targetUser.RoleID <= 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin ไม่สามารถเปลี่ยนรหัสผ่านของ Admin หรือ Superadmin ได้"})
		return
	}

	hashed, err := services.HashPassword(input.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เข้ารหัสรหัสผ่านไม่สำเร็จ"})
		return
	}

	if err := config.DB.Model(&targetUser).Update("password", hashed).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเปลี่ยนรหัสผ่านได้"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "เปลี่ยนรหัสผ่านสำเร็จ"})
}
