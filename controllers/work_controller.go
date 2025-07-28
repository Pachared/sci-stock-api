package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"sci-stock-api/config"
	"sci-stock-api/models"
)

func GetWorkSchedules(c *gin.Context) {
	var schedules []models.WorkSchedule
	if err := config.DB.Find(&schedules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลได้"})
		return
	}
	c.JSON(http.StatusOK, schedules)
}

func CreateWorkSchedule(c *gin.Context) {
	var input struct {
		Title string `json:"title" binding:"required"`
		Date  string `json:"date" binding:"required"`
		Tag   string `json:"tag" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง: " + err.Error()})
		return
	}

	// แปลงวันที่ string เป็น time.Time
	dateParsed, err := time.Parse("2006-01-02", input.Date) // รูปแบบ YYYY-MM-DD
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รูปแบบวันที่ไม่ถูกต้อง"})
		return
	}

	schedule := models.WorkSchedule{
		Title: input.Title,
		Date:  dateParsed,
		Tag:   input.Tag,
	}

	if err := config.DB.Create(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกข้อมูลได้"})
		return
	}

	c.JSON(http.StatusCreated, schedule)
}

func UpdateWorkSchedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รหัสตารางงานไม่ถูกต้อง"})
		return
	}

	var input struct {
		Title string `json:"title" binding:"required"`
		Date  string `json:"date" binding:"required"`
		Tag   string `json:"tag" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง: " + err.Error()})
		return
	}

	dateParsed, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รูปแบบวันที่ไม่ถูกต้อง"})
		return
	}

	var schedule models.WorkSchedule
	if err := config.DB.First(&schedule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบตารางงานที่ระบุ"})
		return
	}

	schedule.Title = input.Title
	schedule.Date = dateParsed
	schedule.Tag = input.Tag

	if err := config.DB.Save(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอัปเดตข้อมูลได้"})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

func DeleteWorkSchedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รหัสตารางงานไม่ถูกต้อง"})
		return
	}

	if err := config.DB.Delete(&models.WorkSchedule{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถลบข้อมูลได้"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ลบข้อมูลเรียบร้อยแล้ว"})
}
