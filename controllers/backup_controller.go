package controllers

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BackupController struct {
	DB *gorm.DB
}

func NewBackupController(db *gorm.DB) *BackupController {
	return &BackupController{DB: db}
}

var allowedTables = map[string]bool{
	"users":                true,
	"roles":                true,
	"student_applications": true,
	"checkins":             true,
	"dried_food":           true,
	"fresh_food":           true,
	"snack":                true,
	"soft_drink":           true,
	"stationery":           true,
	"sales_today":          true,
	"work_schedules":       true,
	"daily_payments":       true,
}

func normalizeValue(col string, v interface{}) interface{} {
	// ถ้าเป็น []byte
	b, ok := v.([]byte)
	if !ok {
		return v
	}

	// column ที่เป็น BLOB จริง ไม่ต้องแปลง
	if col == "profile_image" {
		return b
	}

	// ที่เหลือถือว่าเป็น string
	return string(b)
}

func (bc *BackupController) BackupTable(c *gin.Context) {
	table := c.Param("table")

	if !allowedTables[table] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ไม่อนุญาตให้ backup ตารางนี้",
		})
		return
	}

	rows, err := bc.DB.Raw("SELECT * FROM " + table).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	data := []map[string]interface{}{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		row := map[string]interface{}{}
		for i, col := range columns {
			row[col] = normalizeValue(col, values[i])
		}
		data = append(data, row)
	}

	c.Header("Content-Disposition", "attachment; filename="+table+"_backup.json")
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{
		"table": table,
		"data":  data,
	})
}

func (bc *BackupController) BackupAll(c *gin.Context) {
	backup := map[string]interface{}{}

	for table := range allowedTables {

		rows, err := bc.DB.Raw("SELECT * FROM " + table).Rows()
		if err != nil {
			continue
		}

		columns, _ := rows.Columns()
		data := []map[string]interface{}{}

		for rows.Next() {
			values := make([]interface{}, len(columns))
			ptrs := make([]interface{}, len(columns))
			for i := range values {
				ptrs[i] = &values[i]
			}

			if err := rows.Scan(ptrs...); err != nil {
				rows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			row := map[string]interface{}{}
			for i, col := range columns {
				row[col] = normalizeValue(col, values[i])
			}
			data = append(data, row)
		}

		rows.Close()
		backup[table] = data
	}

	filename := "backup_all.json"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/json")

	c.JSON(http.StatusOK, backup)
}

func (bc *BackupController) ImportData(c *gin.Context) {

	var payload struct {
		Table string                   `json:"table"`
		Data  []map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON ไม่ถูกต้อง"})
		return
	}

	if !allowedTables[payload.Table] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ไม่อนุญาตให้เพิ่มข้อมูลในตารางนี้",
		})
		return
	}

	tx := bc.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
		return
	}

	for _, row := range payload.Data {

		cols := ""
		vals := ""
		args := []interface{}{}
		i := 0

		for k, v := range row {

			if i > 0 {
				cols += ","
				vals += ","
			}

			cols += k
			vals += "?"

			if k == "profile_image" && v != nil {
				if s, ok := v.(string); ok && s != "" {
					decoded, err := base64.StdEncoding.DecodeString(s)
					if err != nil {
						tx.Rollback()
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "profile_image base64 ไม่ถูกต้อง",
						})
						return
					}
					args = append(args, decoded)
				} else {
					args = append(args, nil)
				}
			} else {
				args = append(args, v)
			}

			i++
		}

		query := "INSERT INTO " + payload.Table +
			" (" + cols + ") VALUES (" + vals + ")"

		if err := tx.Exec(query, args...).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "เพิ่มข้อมูลสำเร็จ",
		"rows":    len(payload.Data),
	})
}
