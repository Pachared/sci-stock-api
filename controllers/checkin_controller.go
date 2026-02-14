package controllers

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"sci-stock-api/models"
)

// func EmployeeCheckin(c *gin.Context) {

// 	dbAny, exists := c.Get("DB")
// 	if !exists {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่พบการเชื่อมต่อฐานข้อมูล"})
// 		return
// 	}
// 	db := dbAny.(*gorm.DB)

// 	userIDAny, exists := c.Get("userID")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบข้อมูลผู้ใช้"})
// 		return
// 	}

// 	userID, ok := userIDAny.(uint)
// 	if !ok {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "รูปแบบ userID ไม่ถูกต้อง"})
// 		return
// 	}

// 	roleAny, exists := c.Get("role")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบสิทธิ์ผู้ใช้"})
// 		return
// 	}

// 	role, ok := roleAny.(string)
// 	if !ok || role != "employee" {
// 		c.JSON(http.StatusForbidden, gin.H{
// 			"error": "เฉพาะพนักงานเท่านั้นที่สามารถเช็คอินได้",
// 		})
// 		return
// 	}

// 	if err := AutoCheckoutIfForgot(db, userID); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": "ไม่สามารถจัดการเช็คเอาท์อัตโนมัติได้",
// 		})
// 		return
// 	}

// 	now := time.Now()

// 	startOfDay := time.Date(
// 		now.Year(), now.Month(), now.Day(),
// 		0, 0, 0, 0,
// 		now.Location(),
// 	)
// 	endOfDay := startOfDay.Add(24 * time.Hour)

// 	var todayCheckin models.Checkin
// 	if err := db.Where(
// 		"user_id = ? AND checkin_at >= ? AND checkin_at < ?",
// 		userID, startOfDay, endOfDay,
// 	).First(&todayCheckin).Error; err == nil {

// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "วันนี้คุณได้เช็คอินไปแล้ว",
// 		})
// 		return
// 	}

// 	var existing models.Checkin
// 	if err := db.
// 		Where("user_id = ? AND checkout_at IS NULL", userID).
// 		First(&existing).Error; err == nil {

// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "คุณได้เช็คอินไปแล้ว และยังไม่ได้เช็คเอาท์",
// 		})
// 		return
// 	}

// 	file, err := c.FormFile("photo")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "กรุณาอัปโหลดรูปภาพ (field: photo)",
// 		})
// 		return
// 	}

// 	contentType := file.Header.Get("Content-Type")
// 	if !strings.HasPrefix(contentType, "image/") {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "ไฟล์ต้องเป็นรูปภาพเท่านั้น",
// 		})
// 		return
// 	}

// 	if file.Size > 5<<20 {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "ไฟล์รูปต้องมีขนาดไม่เกิน 5MB",
// 		})
// 		return
// 	}

// 	f, err := file.Open()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": "ไม่สามารถเปิดไฟล์รูปได้",
// 		})
// 		return
// 	}
// 	defer f.Close()

// 	photoBytes, err := io.ReadAll(f)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": "ไม่สามารถอ่านไฟล์รูปได้",
// 		})
// 		return
// 	}

// 	checkin := models.Checkin{
// 		UserID:       userID,
// 		CheckinAt:    now,
// 		CheckinPhoto: photoBytes,
// 	}

// 	if err := db.Create(&checkin).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": "ไม่สามารถบันทึกข้อมูลเช็คอินได้",
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message":    "เช็คอินสำเร็จ",
// 		"checkin_id": checkin.ID,
// 		"checkin_at": checkin.CheckinAt,
// 	})
// }

func EmployeeCheckin(c *gin.Context) {

	dbAny, exists := c.Get("DB")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่พบการเชื่อมต่อฐานข้อมูล"})
		return
	}
	db := dbAny.(*gorm.DB)

	userIDAny, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบข้อมูลผู้ใช้"})
		return
	}

	userID, ok := userIDAny.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "รูปแบบ userID ไม่ถูกต้อง"})
		return
	}

	// ❌ ปิด check role ชั่วคราว
	/*
		roleAny, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบสิทธิ์ผู้ใช้"})
			return
		}

		role, ok := roleAny.(string)
		if !ok || role != "employee" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "เฉพาะพนักงานเท่านั้นที่สามารถเช็คอินได้",
			})
			return
		}
	*/

	// ❌ ปิด Auto checkout ชั่วคราว
	/*
		if err := AutoCheckoutIfForgot(db, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "ไม่สามารถจัดการเช็คเอาท์อัตโนมัติได้",
			})
			return
		}
	*/

	now := time.Now()

	// ❌ ปิด check เช็คอินซ้ำในวันเดียว
	/*
		startOfDay := time.Date(
			now.Year(), now.Month(), now.Day(),
			0, 0, 0, 0,
			now.Location(),
		)
		endOfDay := startOfDay.Add(24 * time.Hour)

		var todayCheckin models.Checkin
		if err := db.Where(
			"user_id = ? AND checkin_at >= ? AND checkin_at < ?",
			userID, startOfDay, endOfDay,
		).First(&todayCheckin).Error; err == nil {

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "วันนี้คุณได้เช็คอินไปแล้ว",
			})
			return
		}
	*/

	// ❌ ปิด check ยังไม่ checkout
	/*
		var existing models.Checkin
		if err := db.
			Where("user_id = ? AND checkout_at IS NULL", userID).
			First(&existing).Error; err == nil {

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "คุณได้เช็คอินไปแล้ว และยังไม่ได้เช็คเอาท์",
			})
			return
		}
	*/

	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "กรุณาอัปโหลดรูปภาพ (field: photo)",
		})
		return
	}

	contentType := file.Header.Get("Content-Type")
	filename := strings.ToLower(file.Filename)
	ext := strings.ToLower(filepath.Ext(filename))

	allowedExt := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}

	isImageByCT := strings.HasPrefix(strings.ToLower(contentType), "image/")
	isImageByExt := allowedExt[ext]

	if !isImageByCT && !isImageByExt {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ไฟล์ต้องเป็นรูปภาพเท่านั้น",
			"debug": gin.H{
				"content_type": contentType,
				"filename":     file.Filename,
				"ext":          ext,
			},
		})
		return
	}

	if file.Size > 5<<20 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ไฟล์รูปต้องมีขนาดไม่เกิน 5MB",
		})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ไม่สามารถเปิดไฟล์รูปได้",
		})
		return
	}
	defer f.Close()

	photoBytes, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ไม่สามารถอ่านไฟล์รูปได้",
		})
		return
	}

	checkin := models.Checkin{
		UserID:       userID,
		CheckinAt:    now,
		CheckinPhoto: photoBytes,
	}

	if err := db.Create(&checkin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "ไม่สามารถบันทึกข้อมูลเช็คอินได้",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "เช็คอินสำเร็จ (TEST MODE)",
		"checkin_id": checkin.ID,
		"checkin_at": checkin.CheckinAt,
	})
}

// func EmployeeCheckout(c *gin.Context) {

// 	dbAny, exists := c.Get("DB")
// 	if !exists {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่พบการเชื่อมต่อฐานข้อมูล"})
// 		return
// 	}
// 	db := dbAny.(*gorm.DB)

// 	userIDAny, exists := c.Get("userID")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบข้อมูลผู้ใช้"})
// 		return
// 	}

// 	userID, ok := userIDAny.(uint)
// 	if !ok {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "รูปแบบ userID ไม่ถูกต้อง"})
// 		return
// 	}

// 	roleAny, exists := c.Get("role")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบสิทธิ์ผู้ใช้"})
// 		return
// 	}

// 	role, ok := roleAny.(string)
// 	if !ok || role != "employee" {
// 		c.JSON(http.StatusForbidden, gin.H{"error": "เฉพาะพนักงานเท่านั้น"})
// 		return
// 	}

// 	var checkin models.Checkin
// 	if err := db.
// 		Where("user_id = ? AND checkout_at IS NULL", userID).
// 		First(&checkin).Error; err != nil {

// 		if err == gorm.ErrRecordNotFound {
// 			c.JSON(http.StatusBadRequest, gin.H{
// 				"error": "ยังไม่มีการเช็คอิน หรือเช็คเอาท์ไปแล้ว",
// 			})
// 			return
// 		}

// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": "เกิดข้อผิดพลาดในการค้นหาข้อมูล",
// 		})
// 		return
// 	}

// 	file, err := c.FormFile("photo")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "กรุณาอัปโหลดรูปภาพ (field: photo)",
// 		})
// 		return
// 	}

// 	contentType := strings.ToLower(file.Header.Get("Content-Type"))
// 	ext := strings.ToLower(filepath.Ext(file.Filename))

// 	allowedExt := map[string]bool{
// 		".jpg":  true,
// 		".jpeg": true,
// 		".png":  true,
// 		".webp": true,
// 	}

// 	isImageByCT := strings.HasPrefix(contentType, "image/")
// 	isImageByExt := allowedExt[ext]

// 	if !isImageByCT && !isImageByExt {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "ไฟล์ต้องเป็นรูปภาพเท่านั้น",
// 			"debug": gin.H{
// 				"content_type": contentType,
// 				"filename":     file.Filename,
// 				"ext":          ext,
// 			},
// 		})
// 		return
// 	}

// 	if file.Size > 5<<20 {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "ไฟล์รูปต้องมีขนาดไม่เกิน 5MB",
// 		})
// 		return
// 	}

// 	f, err := file.Open()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": "ไม่สามารถเปิดไฟล์รูปได้",
// 		})
// 		return
// 	}
// 	defer f.Close()

// 	photoBytes, err := io.ReadAll(f)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": "ไม่สามารถอ่านไฟล์รูปได้",
// 		})
// 		return
// 	}

// 	now := time.Now()

// 	checkinDate := checkin.CheckinAt.Format("2006-01-02")
// 	today := now.Format("2006-01-02")
// 	if checkinDate != today {
// 		c.JSON(http.StatusForbidden, gin.H{
// 			"error": "ไม่สามารถเช็คเอาท์ข้ามวันได้",
// 		})
// 		return
// 	}

// 	worked := now.Sub(checkin.CheckinAt)
// 	if worked < 2*time.Hour {
// 		c.JSON(http.StatusForbidden, gin.H{
// 			"error": "ต้องทำงานอย่างน้อย 2 ชั่วโมง",
// 		})
// 		return
// 	}

// 	if now.Hour() < 13 {
// 		c.JSON(http.StatusForbidden, gin.H{
// 			"error": "สามารถเช็คเอาท์ได้หลัง 13:00 เท่านั้น",
// 		})
// 		return
// 	}

// 	checkin.CheckoutAt = &now
// 	checkin.CheckoutPhoto = photoBytes

// 	if err := db.Save(&checkin).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": "ไม่สามารถบันทึกข้อมูลเช็คเอาท์ได้",
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message":     "เช็คเอาท์สำเร็จ",
// 		"checkin_id":  checkin.ID,
// 		"checkout_at": now,
// 	})
// }

func EmployeeCheckout(c *gin.Context) {

	dbAny, exists := c.Get("DB")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่พบการเชื่อมต่อฐานข้อมูล"})
		return
	}
	db := dbAny.(*gorm.DB)

	userIDAny, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบข้อมูลผู้ใช้"})
		return
	}

	userID, ok := userIDAny.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "รูปแบบ userID ไม่ถูกต้อง"})
		return
	}

	// ❌ ปิด check role ชั่วคราว (ถ้าต้องการ)
	/*
		roleAny, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบสิทธิ์ผู้ใช้"})
			return
		}

		role, ok := roleAny.(string)
		if !ok || role != "employee" {
			c.JSON(http.StatusForbidden, gin.H{"error": "เฉพาะพนักงานเท่านั้น"})
			return
		}
	*/

	var checkin models.Checkin
	if err := db.
		Where("user_id = ? AND checkout_at IS NULL", userID).
		First(&checkin).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "ยังไม่มีการเช็คอิน หรือเช็คเอาท์ไปแล้ว",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "เกิดข้อผิดพลาดในการค้นหาข้อมูล",
		})
		return
	}

	// ✅ TEST MODE: อนุญาตไม่ส่งรูปก็ได้
	var photoBytes []byte

	file, err := c.FormFile("photo")
	if err == nil {
		// ถ้าส่งรูปมา ค่อยตรวจ/อ่าน
		contentType := strings.ToLower(file.Header.Get("Content-Type"))
		ext := strings.ToLower(filepath.Ext(file.Filename))

		allowedExt := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
			".webp": true,
		}

		isImageByCT := strings.HasPrefix(contentType, "image/")
		isImageByExt := allowedExt[ext]

		if !isImageByCT && !isImageByExt {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "ไฟล์ต้องเป็นรูปภาพเท่านั้น",
				"debug": gin.H{
					"content_type": contentType,
					"filename":     file.Filename,
					"ext":          ext,
				},
			})
			return
		}

		if file.Size > 5<<20 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "ไฟล์รูปต้องมีขนาดไม่เกิน 5MB",
			})
			return
		}

		f, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "ไม่สามารถเปิดไฟล์รูปได้",
			})
			return
		}
		defer f.Close()

		photoBytes, err = io.ReadAll(f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "ไม่สามารถอ่านไฟล์รูปได้",
			})
			return
		}
	}

	now := time.Now()

	// ❌ ปิดเงื่อนไข “ห้ามข้ามวัน”
	/*
		checkinDate := checkin.CheckinAt.Format("2006-01-02")
		today := now.Format("2006-01-02")
		if checkinDate != today {
			c.JSON(http.StatusForbidden, gin.H{"error": "ไม่สามารถเช็คเอาท์ข้ามวันได้"})
			return
		}
	*/

	// ❌ ปิดเงื่อนไข “ต้องทำงานอย่างน้อย 2 ชั่วโมง”
	/*
		worked := now.Sub(checkin.CheckinAt)
		if worked < 2*time.Hour {
			c.JSON(http.StatusForbidden, gin.H{"error": "ต้องทำงานอย่างน้อย 2 ชั่วโมง"})
			return
		}
	*/

	// ❌ ปิดเงื่อนไข “เช็คเอาท์ได้หลัง 13:00”
	/*
		if now.Hour() < 13 {
			c.JSON(http.StatusForbidden, gin.H{"error": "สามารถเช็คเอาท์ได้หลัง 13:00 เท่านั้น"})
			return
		}
	*/

	checkin.CheckoutAt = &now

	// ถ้าไม่ส่งรูปมา photoBytes จะเป็น nil -> จะไม่แก้ค่ารูปเดิม (หรือจะบันทึกเป็นว่างก็ได้)
	if len(photoBytes) > 0 {
		checkin.CheckoutPhoto = photoBytes
	}

	if err := db.Save(&checkin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "ไม่สามารถบันทึกข้อมูลเช็คเอาท์ได้",
			"details": err.Error(), // debug ชั่วคราว
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "เช็คเอาท์สำเร็จ (TEST MODE)",
		"checkin_id":  checkin.ID,
		"checkout_at": now,
	})
}

func GetCheckinsWithFirstName(c *gin.Context) {
	db := c.MustGet("DB").(*gorm.DB)

	rows := make([]models.CheckinWithUser, 0) // ✅ สำคัญ

	err := db.Table("checkins").
		Select(`
      checkins.user_id,
      users.first_name,
      checkins.checkin_at,
      checkins.checkout_at,
      checkins.checkin_photo,
      checkins.checkout_photo
    `).
		Joins("JOIN users ON users.id = checkins.user_id").
		Order("checkins.checkin_at DESC").
		Scan(&rows).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูล checkins ได้", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func AutoCheckoutIfForgot(db *gorm.DB, userID uint) error {
	now := time.Now()

	var checkin models.Checkin
	err := db.
		Where("user_id = ? AND checkout_at IS NULL", userID).
		First(&checkin).Error

	if err != nil {
		return nil
	}

	if checkin.CheckinAt.Format("2006-01-02") == now.Format("2006-01-02") {
		return nil
	}

	endOfDay := time.Date(
		checkin.CheckinAt.Year(),
		checkin.CheckinAt.Month(),
		checkin.CheckinAt.Day(),
		23, 59, 59, 0,
		checkin.CheckinAt.Location(),
	)

	checkin.CheckoutAt = &endOfDay
	checkin.IsAutoCheckout = true

	return db.Save(&checkin).Error
}

func CheckinStatus(c *gin.Context) {
	db := c.MustGet("DB").(*gorm.DB)
	userID := c.MustGet("userID").(uint)

	var checkin models.Checkin
	err := db.
		Where("user_id = ? AND checkout_at IS NULL", userID).
		First(&checkin).Error

	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusOK, gin.H{
			"checked_in": false,
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ไม่สามารถตรวจสอบสถานะได้",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"checked_in": true,
		"checkin_at": checkin.CheckinAt,
	})
}
