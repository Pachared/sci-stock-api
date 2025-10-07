package controllers

import (
	"net/http"
	"strconv"
	"encoding/base64"

	"github.com/gin-gonic/gin"
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
		"message":    "สมัครนักศึกษาเป็นพนักงานสำเร็จ",
		"firstName":  firstName,
		"lastName":   lastName,
		"gmail":      gmail,
		"studentId":  studentId,
		"contact":    contact,
		"fileSize":   len(fileBytes),
	})
}

// ดึงข้อมูลใบสมัครทั้งหมด
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
            ID:        app.ID,
            FirstName: app.FirstName,
            LastName:  app.LastName,
            Gmail:     app.Gmail,
            StudentID: app.StudentID,
            Schedule:  scheduleBase64,
            Contact:   app.Contact,
            Status:    app.Status,
            CreatedAt: app.CreatedAt,
            UpdatedAt: app.UpdatedAt,
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
		if err := config.DB.Exec(`
			UPDATE student_applications SET status = 'อนุมัติ' WHERE id = ?
		`, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอัปเดตข้อมูลได้"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "อนุมัติเรียบร้อยแล้ว"})
		return
	}

	if action == "reject" {
		if err := config.DB.Exec(`
			DELETE FROM student_applications WHERE id = ?
		`, id).Error; err != nil {
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

    // ลบเฉพาะรายการที่ status = 'อนุมัติ'
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