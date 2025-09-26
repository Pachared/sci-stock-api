package controllers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"sci-stock-api/config"
	"sci-stock-api/models"
)

var dashboardProductTables = []string{
	"dried_food",
	"fresh_food",
	"snack",
	"soft_drink",
	"stationery",
}

func GetTotalProducts(c *gin.Context) {
	result := make(map[string]int64)
	var grandTotal int64 = 0

	for _, table := range dashboardProductTables {
		var sum sql.NullInt64
		if err := config.DB.Table(table).Select("SUM(stock)").Scan(&sum).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"ข้อความผิดพลาด": "ไม่สามารถดึงข้อมูลได้จากตาราง " + table,
				"error":         err.Error(),
			})
			return
		}

		total := int64(0)
		if sum.Valid {
			total = sum.Int64
		}

		result[table] = total
		grandTotal += total
	}

	c.JSON(http.StatusOK, gin.H{
		"จำนวนสินค้าตามประเภท": result,
		"จำนวนสินค้าทั้งหมด":  grandTotal,
	})
}

func GetLowStockProducts(c *gin.Context) {
	lowStock := make(map[string][]models.ProductLow)
	totalLowCount := int64(0)

	for _, table := range dashboardProductTables {
		var products []models.ProductLow
		if err := config.DB.Table(table).
			Select("id, product_name, barcode, stock, reorder_level").
			Where("stock > 0 AND stock <= reorder_level").
			Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"ข้อความผิดพลาด": "ไม่สามารถดึงข้อมูลได้จากตาราง " + table,
				"error":         err.Error(),
			})
			return
		}

		if products == nil {
			products = []models.ProductLow{}
		}

		lowStock[table] = products

		var count int64
		if err := config.DB.Table(table).
			Where("stock > 0 AND stock <= reorder_level").
			Count(&count).Error; err == nil {
			totalLowCount += count
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"สินค้าใกล้หมดสต๊อก": lowStock,
		"จำนวนรวมทั้งหมด":   totalLowCount,
	})
}

func GetOutOfStockProducts(c *gin.Context) {
	outStock := make(map[string][]models.ProductLow)
	totalOutCount := int64(0)

	for _, table := range dashboardProductTables {
		var products []models.ProductLow
		if err := config.DB.Table(table).
			Select("id, product_name, barcode, stock").
			Where("stock = 0").
			Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"ข้อความผิดพลาด": "ไม่สามารถดึงข้อมูลได้จากตาราง " + table,
				"error":         err.Error(),
			})
			return
		}

		if products == nil {
			products = []models.ProductLow{}
		}

		outStock[table] = products

		var count int64
		if err := config.DB.Table(table).
			Where("stock = 0").
			Count(&count).Error; err == nil {
			totalOutCount += count
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"สินค้าหมดสต๊อก":     outStock,
		"จำนวนรวมทั้งหมด":   totalOutCount,
	})
}