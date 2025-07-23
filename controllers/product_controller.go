package controllers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"sci-stock-api/config"
	"sci-stock-api/models"
)

var productTables = map[string]interface{}{
	"dried_food":  &[]models.DriedFood{},
	"fresh_food":  &[]models.FreshFood{},
	"snack":       &[]models.Snack{},
	"soft_drink":  &[]models.SoftDrink{},
	"stationery":  &[]models.Stationery{},
}

func GetProductsByCategory(c *gin.Context) {
	category := c.Param("category")

	switch category {
	case "dried_food":
		var products []models.DriedFood
		if err := config.DB.Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot fetch products"})
			return
		}
		c.JSON(http.StatusOK, products)

	case "fresh_food":
		var products []models.FreshFood
		if err := config.DB.Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot fetch products"})
			return
		}
		c.JSON(http.StatusOK, products)

	case "snack":
		var products []models.Snack
		if err := config.DB.Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot fetch products"})
			return
		}
		c.JSON(http.StatusOK, products)

	case "soft_drink":
		var products []models.SoftDrink
		if err := config.DB.Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot fetch products"})
			return
		}
		c.JSON(http.StatusOK, products)

	case "stationery":
		var products []models.Stationery
		if err := config.DB.Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot fetch products"})
			return
		}
		c.JSON(http.StatusOK, products)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
	}
}

func CreateProductByCategory(c *gin.Context) {
	category := c.Param("category")
	_, exists := productTables[category]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
		return
	}

	var input models.ProductInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var imageURL string
	file, err := c.FormFile("image")
	if err == nil {
		filename := filepath.Base(file.Filename)
		path := "uploads/" + filename
		if err := c.SaveUploadedFile(file, path); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
			return
		}
		imageURL = path
	}

	if err := createProduct(category, input, imageURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "product created"})
}

func createProduct(category string, input models.ProductInput, imageURL string) error {
	switch category {
	case "dried_food":
		return config.DB.Create(&models.DriedFood{
			ProductName:  input.ProductName,
			Barcode:      input.Barcode,
			Price:        input.Price,
			Cost:         input.Cost,
			Stock:        input.Stock,
			ReorderLevel: input.ReorderLevel,
			ImageURL:     imageURL,
		}).Error
	case "fresh_food":
		return config.DB.Create(&models.FreshFood{
			ProductName:  input.ProductName,
			Barcode:      input.Barcode,
			Price:        input.Price,
			Cost:         input.Cost,
			Stock:        input.Stock,
			ReorderLevel: input.ReorderLevel,
			ImageURL:     imageURL,
		}).Error
	case "snack":
		return config.DB.Create(&models.Snack{
			ProductName:  input.ProductName,
			Barcode:      input.Barcode,
			Price:        input.Price,
			Cost:         input.Cost,
			Stock:        input.Stock,
			ReorderLevel: input.ReorderLevel,
			ImageURL:     imageURL,
		}).Error
	case "soft_drink":
		return config.DB.Create(&models.SoftDrink{
			ProductName:  input.ProductName,
			Barcode:      input.Barcode,
			Price:        input.Price,
			Cost:         input.Cost,
			Stock:        input.Stock,
			ReorderLevel: input.ReorderLevel,
			ImageURL:     imageURL,
		}).Error
	case "stationery":
		return config.DB.Create(&models.Stationery{
			ProductName:  input.ProductName,
			Barcode:      input.Barcode,
			Price:        input.Price,
			Cost:         input.Cost,
			Stock:        input.Stock,
			ReorderLevel: input.ReorderLevel,
			ImageURL:     imageURL,
		}).Error
	}
	return nil
}
