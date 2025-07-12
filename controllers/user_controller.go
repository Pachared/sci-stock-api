package controllers

import (
	"io"
	"net/http"
	"strconv"
	"time"

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

func CreateUserRequestByAdmin(c *gin.Context) {
    if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่สามารถอ่านข้อมูลได้"})
        return
    }

    // ดึงสิทธิ์จาก JWT (middleware เก็บ userID ไว้)
    currentID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    var creator models.User
    if err := config.DB.First(&creator, currentID).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบผู้ใช้"})
        return
    }

    gmail := c.PostForm("gmail")
    password := c.PostForm("password")
    firstName := c.PostForm("first_name")
    lastName := c.PostForm("last_name")
    roleStr := c.PostForm("role_id")

    roleID, err := strconv.Atoi(roleStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "role_id ไม่ถูกต้อง"})
        return
    }

    if creator.RoleID == 2 && roleID <= 2 {
        c.JSON(http.StatusForbidden, gin.H{"error": "admin ไม่สามารถสร้าง admin หรือ superadmin ได้"})
        return
    }

    // ตรวจอีเมลซ้ำ
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

    file, _, _ := c.Request.FormFile("profile_image")
    var imageData []byte
    if file != nil {
        defer file.Close()
        imageData, _ = io.ReadAll(file)
    }

    hashedPass, _ := services.HashPassword(password)
    otp := services.GenerateOTP(6)
    now := time.Now().In(time.FixedZone("Asia/Bangkok", 7*3600))
    expire := now.Add(10 * time.Minute)

    // บันทึกชั่วคราว
    err = config.DB.Exec(`
        INSERT INTO user_verifications (gmail, password, first_name, last_name, role_id, profile_image, otp, otp_expires_at, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, gmail, hashedPass, firstName, lastName, roleID, imageData, otp, expire, now).Error
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกได้"})
        return
    }

    html, plain := services.GenerateEmailBodyForRegisterOTP(otp)
    if err := services.SendEmail(gmail, "ยืนยัน OTP สำหรับการสร้างบัญชี", html, plain); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "ส่งอีเมล OTP ไม่สำเร็จ"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "ส่ง OTP สำเร็จ กรุณายืนยันอีเมล"})
}

func VerifyAndActivateUser(c *gin.Context) {
    var input struct {
        Gmail string `json:"gmail" binding:"required,email"`
        OTP   string `json:"otp" binding:"required"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var verif struct {
        Gmail        string
        Password     string
        FirstName    string
        LastName     string
        RoleID       int
        Image        []byte `gorm:"column:profile_image"`
        OTP          string
        OtpExpiresAt time.Time
    }

    err := config.DB.Raw(`
        SELECT gmail, password, first_name, last_name, role_id, profile_image, otp, otp_expires_at
        FROM user_verifications
        WHERE gmail = ? ORDER BY id DESC LIMIT 1
    `, input.Gmail).Scan(&verif).Error

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบ OTP ได้"})
        return
    }

    if verif.OTP != input.OTP {
        c.JSON(http.StatusBadRequest, gin.H{"error": "รหัส OTP ไม่ถูกต้อง"})
        return
    }

    if time.Now().After(verif.OtpExpiresAt) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "รหัส OTP หมดอายุแล้ว"})
        return
    }

    // สร้างบัญชีจริง
    user := models.User{
        Gmail:        verif.Gmail,
        Password:     verif.Password,
        FirstName:    verif.FirstName,
        LastName:     verif.LastName,
        RoleID:       uint(verif.RoleID),
        ProfileImage: verif.Image,
    }

    if err := config.DB.Create(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "สร้างบัญชีไม่สำเร็จ"})
        return
    }

    config.DB.Exec("DELETE FROM user_verifications WHERE gmail = ?", input.Gmail)
    c.JSON(http.StatusOK, gin.H{"message": "ยืนยันสำเร็จ บัญชีถูกสร้างแล้ว"})
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
