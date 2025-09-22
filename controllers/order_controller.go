package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
	"sci-stock-api/models"
	"regexp"
	"log"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"gorm.io/gorm"
)

const (
	spreadsheetId = "1_PkqV97P468hg0PODas0VT7e-Bvyl7u6dDfrVPFHHWU"
	sheetName     = "Sheet1"
)

var (
	cachedBarcodes     []string
	lastCacheUpdatedAt time.Time
)

const cacheDuration = 0

func getSheetsService() (*sheets.Service, error) {
	ctx := context.Background()
	b, err := os.ReadFile("api-sci-next-e83942db1165.json")
	if err != nil {
		return nil, err
	}
	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, err
	}
	client := config.Client(ctx)
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return srv, nil
}

func updateCachedBarcodes() error {
	srv, err := getSheetsService()
	if err != nil {
		return err
	}

	readRange := sheetName + "!B2:B"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		return err
	}

	var barcodes []string
	for _, row := range resp.Values {
		if len(row) > 0 {
			barcodes = append(barcodes, fmt.Sprintf("%v", row[0]))
		}
	}

	cachedBarcodes = barcodes
	lastCacheUpdatedAt = time.Now()
	return nil
}

func FindProductByBarcode(db *gorm.DB, barcode string) (*models.Product, string, error) {
	tables := []string{"dried_food", "fresh_food", "snack", "soft_drink", "stationery"}
	for _, table := range tables {
		var p models.Product
		err := db.Table(table).Where("barcode = ?", barcode).First(&p).Error
		if err == nil {
			fmt.Printf("DEBUG: Table: %s, ProductName: %s, Barcode: %s\n", table, p.ProductName, p.Barcode)
			p.Category = table
			return &p, table, nil
		}
	}
	return nil, "", fmt.Errorf("ไม่พบสินค้า")
}

func GetProductByBarcode(c *gin.Context) {
	db := c.MustGet("DB").(*gorm.DB)
	barcode := c.Param("barcode")
	
	if barcode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Barcode ไม่ถูกต้อง"})
		return
	}

	product, table, err := FindProductByBarcode(db, barcode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบสินค้า"})
		return
	}

	resp := models.ProductResponse{
		ID:          product.ID,
		ProductName: product.ProductName,
		Barcode:     product.Barcode,
		Price:       product.Price,
		ImageURL:    product.ImageURL,
		Category:    table,
	}

	c.JSON(http.StatusOK, resp)
}

func GetProductsFromSheet(c *gin.Context) {
	db := c.MustGet("DB").(*gorm.DB)

	if time.Since(lastCacheUpdatedAt) > cacheDuration {
		if err := updateCachedBarcodes(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "โหลดข้อมูลจาก Google Sheets ไม่ได้"})
			return
		}
	}

	var results []models.Product
	for _, barcode := range cachedBarcodes {
		product, _, err := FindProductByBarcode(db, barcode)
		if err == nil {
			results = append(results, *product)
		}
	}

	c.JSON(http.StatusOK, gin.H{"products": results})
}

func isValidBarcode(barcode string) bool {
	re := regexp.MustCompile(`^\d{8,13}$`)
	return re.MatchString(barcode)
}

func isBarcodeInSheet(barcode string) bool {
	if time.Since(lastCacheUpdatedAt) > cacheDuration {
		_ = updateCachedBarcodes()
	}

	for _, b := range cachedBarcodes {
		if b == barcode {
			return true
		}
	}
	return false
}

func SellProduct(c *gin.Context) {
	db := c.MustGet("DB").(*gorm.DB)

	var req struct {
		Barcode  string `json:"barcode"`
		Quantity int    `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "ข้อมูลไม่ถูกต้อง"})
		return
	}

	if req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "จำนวนสินค้า (Quantity) ต้องเป็นจำนวนบวก"})
		return
	}

	if !isValidBarcode(req.Barcode) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Barcode ไม่ถูกต้อง"})
		return
	}

	if !isBarcodeInSheet(req.Barcode) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "ไม่สามารถขายสินค้าได้ เนื่องจากไม่อยู่ในรายการ Google Sheets"})
		return
	}

	product, table, err := FindProductByBarcode(db, req.Barcode)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "ไม่พบสินค้า"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Table(table).
			Where("id = ? AND stock >= ?", product.ID, req.Quantity).
			UpdateColumn("stock", gorm.Expr("stock - ?", req.Quantity))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("สต๊อกไม่เพียงพอ")
		}

		if err := tx.Exec(`
			INSERT INTO sales_today (product_name, barcode, price, quantity, sold_at, image_url)
			VALUES (?, ?, ?, ?, ?, ?)`,
			product.ProductName, product.Barcode, product.Price, req.Quantity, time.Now(), product.ImageURL).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	if err := deleteBarcodeFromSheet(product.Barcode); err != nil {
		log.Printf("ลบ barcode %s จาก Google Sheet ไม่สำเร็จ: %v\n", product.Barcode, err)
	} else {
		clearCache()
	}

	resp := models.ProductResponse{
		ID:          product.ID,
		ProductName: product.ProductName,
		Barcode:     product.Barcode,
		Price:       product.Price,
		ImageURL:    product.ImageURL,
		Category:    product.Category,
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "ตัดสต๊อกสำเร็จ",
		"product":          resp,
		"stock_remaining": product.Stock - req.Quantity,
	})
}

func SellProductLocal(c *gin.Context) {
    db := c.MustGet("DB").(*gorm.DB)

    var req struct {
        Barcode  string `json:"barcode"`
        Quantity int    `json:"quantity"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "ข้อมูลไม่ถูกต้อง"})
        return
    }

    if req.Quantity <= 0 {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "จำนวนสินค้า (Quantity) ต้องเป็นจำนวนบวก"})
        return
    }

    if !isValidBarcode(req.Barcode) {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Barcode ไม่ถูกต้อง"})
        return
    }

    // ดึงสินค้าจากฐานข้อมูลเหมือนเดิม
    product, table, err := FindProductByBarcode(db, req.Barcode)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "ไม่พบสินค้า"})
        return
    }

    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        res := tx.Table(table).
            Where("id = ? AND stock >= ?", product.ID, req.Quantity).
            UpdateColumn("stock", gorm.Expr("stock - ?", req.Quantity))
        if res.Error != nil {
            return res.Error
        }
        if res.RowsAffected == 0 {
            return fmt.Errorf("สต๊อกไม่เพียงพอ")
        }

        if err := tx.Exec(`
            INSERT INTO sales_today (product_name, barcode, price, quantity, sold_at, image_url)
            VALUES (?, ?, ?, ?, ?, ?)`,
            product.ProductName, product.Barcode, product.Price, req.Quantity, time.Now(), product.ImageURL).Error; err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
        return
    }

    resp := models.ProductResponse{
        ID:          product.ID,
        ProductName: product.ProductName,
        Barcode:     product.Barcode,
        Price:       product.Price,
        ImageURL:    product.ImageURL,
        Category:    product.Category,
    }

    c.JSON(http.StatusOK, gin.H{
        "message":          "ตัดสต๊อกสำเร็จ (Local)",
        "product":          resp,
        "stock_remaining": product.Stock - req.Quantity,
    })
}

func deleteBarcodeFromSheet(barcode string) error {
	srv, err := getSheetsService()
	if err != nil {
		return err
	}

	readRange := sheetName + "!B2:B"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		return err
	}

	var rowIndex int64 = -1
	for i, row := range resp.Values {
		if len(row) > 0 && fmt.Sprintf("%v", row[0]) == barcode {
			rowIndex = int64(i) + 2
			break
		}
	}

	if rowIndex == -1 {
		return fmt.Errorf("ไม่พบ barcode ใน Google Sheet")
	}

	requests := []*sheets.Request{
		{
			DeleteDimension: &sheets.DeleteDimensionRequest{
				Range: &sheets.DimensionRange{
					SheetId:    0,
					Dimension:  "ROWS",
					StartIndex: rowIndex - 1,
					EndIndex:   rowIndex,
				},
			},
		},
	}
	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}).Do()

	return err
}

func clearCache() {
	cachedBarcodes = nil
	lastCacheUpdatedAt = time.Time{}
}

func RefreshCache(c *gin.Context) {
	if err := updateCachedBarcodes(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "โหลดข้อมูลใหม่จาก Google Sheets ไม่สำเร็จ",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "โหลดข้อมูลใหม่จาก Google Sheets สำเร็จ",
		"count":   len(cachedBarcodes),
	})
}

func GetSalesToday(c *gin.Context) {
    db := c.MustGet("DB").(*gorm.DB)

    var sales []models.SaleToday

    fmt.Println("DB instance:", db)

    if err := db.Table("sales_today").Find(&sales).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "ไม่สามารถดึงข้อมูล sales_today ได้: " + err.Error(),
        })
        return
    }

    fmt.Printf("Found %d sales\n", len(sales))

    c.JSON(http.StatusOK, gin.H{
        "sales_today": sales,
    })
}

func CreateDailyPayment(c *gin.Context) {
	db := c.MustGet("DB").(*gorm.DB)

	var payment models.DailyPayment

	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง: " + err.Error()})
		return
	}

	// ถ้าไม่มี PaymentDate ให้ใส่วันที่วันนี้
	if payment.PaymentDate == "" {
		payment.PaymentDate = time.Now().Format("2006-01-02")
	}

	fmt.Printf("Received payment: %+v\n", payment) // log ตรวจสอบ

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("daily_payments").Create(&payment).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Saved payment: %+v\n", payment) // log ตรวจสอบ

	c.JSON(http.StatusOK, gin.H{
		"message": "บันทึกรายการสำเร็จ",
		"data":    payment,
	})
}