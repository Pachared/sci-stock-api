package controllers

import (
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"sci-stock-api/config"
	"sci-stock-api/models"
)

func HandleEmployeeRegister(c *gin.Context) {
	firstName := c.PostForm("firstName")
	lastName := c.PostForm("lastName")
	gmail := c.PostForm("gmail")
	studentId := c.PostForm("studentId")
	contact := c.PostForm("contact")

	var fileBytes []byte
	file, err := c.FormFile("resumeFile")
	if file != nil && err == nil {
		f, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านไฟล์ได้"})
			return
		}
		defer f.Close()

		buf := make([]byte, file.Size)
		_, err = f.Read(buf)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านไฟล์ได้"})
			return
		}
		fileBytes = buf
	}

	result := config.DB.Exec(`
		INSERT INTO student_applications
			(first_name, last_name, gmail, student_id, schedule, contact_info, status)
		VALUES (?, ?, ?, ?, ?, ?, 'รออนุมัติ')
	`, firstName, lastName, gmail, studentId, fileBytes, contact)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกลงฐานข้อมูลได้"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "สมัครนักศึกษาเป็นพนักงานสำเร็จ",
		"firstName": firstName,
		"lastName":  lastName,
		"gmail":     gmail,
		"studentId": studentId,
		"contact":   contact,
		"fileSize":  len(fileBytes),
	})
}

func GetStudentApplications(c *gin.Context) {
	var applications []models.StudentApplication
	if err := config.DB.Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลได้"})
		return
	}

	result := make([]models.StudentApplicationResponse, 0)

	for _, app := range applications {
		var scheduleBase64 string
		if len(app.Schedule) > 0 {
			scheduleBase64 = "data:image/png;base64," + base64.StdEncoding.EncodeToString(app.Schedule)
		}
		result = append(result, models.StudentApplicationResponse{
			ID:         app.ID,
			FirstName:  app.FirstName,
			LastName:   app.LastName,
			Gmail:      app.Gmail,
			StudentID:  app.StudentID,
			Schedule:   scheduleBase64,
			Contact:    app.Contact,
			Status:     app.Status,
			IsEmployee: app.IsEmployee,
			CreatedAt:  app.CreatedAt,
			UpdatedAt:  app.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

func ApproveStudentApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID ไม่ถูกต้อง"})
		return
	}

	action := c.Query("action")
	if action == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ต้องระบุ action=approve หรือ action=reject"})
		return
	}

	if action == "approve" {
		if err := config.DB.Exec(`UPDATE student_applications SET status = 'อนุมัติ' WHERE id = ?`, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอัปเดตข้อมูลได้"})
			return
		}

		var student models.ApprovedStudent
		if err := config.DB.Where("id = ?", id).First(&student).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลล่าสุดได้"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "อนุมัติเรียบร้อยแล้ว",
			"student": student,
		})
		return
	}

	if action == "reject" {
		if err := config.DB.Exec(`DELETE FROM student_applications WHERE id = ?`, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถลบข้อมูลได้"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "ลบข้อมูลเรียบร้อยแล้ว"})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "action ไม่ถูกต้อง ต้องเป็น approve หรือ reject"})
}

func DeleteApprovedStudentApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID ไม่ถูกต้อง"})
		return
	}

	result := config.DB.Exec(`
        DELETE FROM student_applications 
        WHERE id = ? AND status = 'อนุมัติ'
    `, id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถลบข้อมูลได้"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่พบข้อมูลอนุมัติสำหรับ ID นี้"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ลบข้อมูลอนุมัติเรียบร้อยแล้ว"})
}

func hashPassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return password
	}
	return string(hashed)
}

func CheckOrAddEmployee(c *gin.Context) {
	db := config.DB

	var req struct {
		FirstName    string `json:"firstName"`
		LastName     string `json:"lastName"`
		Gmail        string `json:"gmail"`
		Password     string `json:"password"`
		RoleID       uint   `json:"roleId"`
		ProfileImage string `json:"profileimage"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var profileBytes []byte
	if req.ProfileImage != "" {
		decoded, err := base64.StdEncoding.DecodeString(req.ProfileImage)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "รูปโปรไฟล์ไม่ถูกต้อง"})
			return
		}
		profileBytes = decoded
	}

	var existingUser models.User
	if err := db.Where("gmail = ?", req.Gmail).First(&existingUser).Error; err == nil {
		existingUser.FirstName = req.FirstName
		existingUser.LastName = req.LastName
		if req.Password != "" {
			existingUser.Password = hashPassword(req.Password)
		}
		if len(profileBytes) > 0 {
			existingUser.ProfileImage = profileBytes
		}
		db.Save(&existingUser)

		db.Model(&models.ApprovedStudent{}).
			Where("gmail = ?", req.Gmail).
			Update("is_employee", true)

		c.JSON(http.StatusOK, gin.H{
			"message":     "อัปเดตข้อมูลพนักงานเรียบร้อย",
			"isEmployee":  true,
			"employee_id": existingUser.ID,
		})
		return
	}

	user := models.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Gmail:        req.Gmail,
		Password:     hashPassword(req.Password),
		RoleID:       uint64(req.RoleID),
		ProfileImage: profileBytes,
	}
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	db.Model(&models.ApprovedStudent{}).
		Where("gmail = ?", req.Gmail).
		Update("is_employee", true)

	c.JSON(http.StatusOK, gin.H{
		"message":     "เพิ่มพนักงานเรียบร้อย",
		"isEmployee":  true,
		"employee_id": user.ID,
	})
}

func DeleteEmployeeByGmail(c *gin.Context) {
    db := config.DB
    gmail := c.Param("gmail")
    if gmail == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "ต้องระบุ Gmail"})
        return
    }

    err := db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Exec(`DELETE FROM student_applications WHERE gmail = ?`, gmail).Error; err != nil {
            return err
        }
        if err := tx.Exec(`DELETE FROM users WHERE gmail = ?`, gmail).Error; err != nil {
            return err
        }
        return nil
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถลบข้อมูลได้: " + err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "ลบข้อมูลนักศึกษา/พนักงานเรียบร้อยแล้ว",
        "gmail":   gmail,
    })
}
