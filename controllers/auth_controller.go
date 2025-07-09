package controllers

import (
	"net/http"
	"sci-stock-api/config"
	"sci-stock-api/models"
	"sci-stock-api/services"

	"github.com/gin-gonic/gin"
	"io"
)

func Register(c *gin.Context) {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่สามารถอ่านข้อมูลที่ส่งมาได้"})
		return
	}

	gmail := c.PostForm("gmail")
	password := c.PostForm("password")
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")

	// ตรวจสอบข้อมูล
	if gmail == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณากรอกอีเมลและรหัสผ่าน"})
		return
	}

	// รับไฟล์ภาพ
	file, _, err := c.Request.FormFile("profile_image")
	var imageData []byte
	if err == nil {
		defer file.Close()
		imageData, _ = io.ReadAll(file)
	}

	// เข้ารหัสรหัสผ่าน
	hashedPass, err := services.HashPassword(password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเข้ารหัสรหัสผ่านได้"})
		return
	}

	// สร้าง User
	user := models.User{
		Gmail:        gmail,
		Password:     hashedPass,
		FirstName:    firstName,
		LastName:     lastName,
		RoleID:       3, // user role
		ProfileImage: imageData, // เก็บรูป
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "อีเมลนี้ถูกใช้งานแล้ว"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "สมัครสมาชิกสำเร็จ"})
}

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