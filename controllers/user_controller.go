package controllers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"sci-stock-api/config"
	"sci-stock-api/models"
	"sci-stock-api/services"
)

func GetUsers(c *gin.Context) {
	var users []models.User
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func CreateUser(c *gin.Context) {
	// รับค่าแบบ multipart
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot parse form"})
		return
	}

	// ดึงค่าจาก JWT middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var currentUser models.User
	if err := config.DB.First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// รับค่าจาก form
	gmail := c.PostForm("gmail")
	password := c.PostForm("password")
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	roleIDStr := c.PostForm("role_id")

	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	// ตรวจสิทธิ์
	if currentUser.RoleID == 2 && roleID <= 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin cannot create admin or superadmin"})
		return
	}

	// อ่านรูป
	file, _, err := c.Request.FormFile("profile_image")
	var imageData []byte
	if err == nil {
		defer file.Close()
		imageData, _ = io.ReadAll(file)
	}

	// Hash password
	hashedPass, err := services.HashPassword(password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot hash password"})
		return
	}

	newUser := models.User{
		Gmail:        gmail,
		Password:     hashedPass,
		FirstName:    firstName,
		LastName:     lastName,
		RoleID:       uint(roleID),
		ProfileImage: imageData,
	}

	if err := config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot parse form"})
		return
	}

	user.FirstName = c.PostForm("first_name")
	user.LastName = c.PostForm("last_name")

	file, _, err := c.Request.FormFile("profile_image")
	if err == nil {
		defer file.Close()
		imageData, _ := io.ReadAll(file)
		user.ProfileImage = imageData
	}

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user updated"})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}
