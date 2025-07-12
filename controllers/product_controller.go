package controllers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"sci-stock-api/config"
	"sci-stock-api/models"
)

// map category => table model
var productTables = map[string]interface{}{
	"dried_food":  &[]models.DriedFood{},
	"fresh_food":  &[]models.FreshFood{},
	"snack":       &[]models.Snack{},
	"soft_drink":  &[]models.SoftDrink{},
	"stationery":  &[]models.Stationery{},
}

func GetProductsByCategory(c *gin.Context) {
	category := c.Param("category")
	products, exists := productTables[category]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
		return
	}

	if err := config.DB.Find(products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot fetch products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

func CreateProductByCategory(c *gin.Context) {
	category := c.Param("category")
	_, exists := productTables[category]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
		return
	}

	// Dynamic struct per table
	var input struct {
		ProductName   string  `form:"product_name" binding:"required"`
		Barcode       string  `form:"barcode" binding:"required"`
		Price         float64 `form:"price" binding:"required"`
		Cost          float64 `form:"cost" binding:"required"`
		Stock         int     `form:"stock"`
		ReorderLevel  int     `form:"reorder_level"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle image upload
	file, err := c.FormFile("image")
	var imageURL string
	if err == nil {
		path := "uploads/" + filepath.Base(file.Filename)
		if err := c.SaveUploadedFile(file, path); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
			return
		}
		imageURL = path
	}

	switch category {
	case "dried_food":
		config.DB.Create(&models.DriedFood{ProductName: input.ProductName, Barcode: input.Barcode, Price: input.Price, Cost: input.Cost, Stock: input.Stock, ReorderLevel: input.ReorderLevel, ImageURL: imageURL})
	case "fresh_food":
		config.DB.Create(&models.FreshFood{ProductName: input.ProductName, Barcode: input.Barcode, Price: input.Price, Cost: input.Cost, Stock: input.Stock, ReorderLevel: input.ReorderLevel, ImageURL: imageURL})
	case "snack":
		config.DB.Create(&models.Snack{ProductName: input.ProductName, Barcode: input.Barcode, Price: input.Price, Cost: input.Cost, Stock: input.Stock, ReorderLevel: input.ReorderLevel, ImageURL: imageURL})
	case "soft_drink":
		config.DB.Create(&models.SoftDrink{ProductName: input.ProductName, Barcode: input.Barcode, Price: input.Price, Cost: input.Cost, Stock: input.Stock, ReorderLevel: input.ReorderLevel, ImageURL: imageURL})
	case "stationery":
		config.DB.Create(&models.Stationery{ProductName: input.ProductName, Barcode: input.Barcode, Price: input.Price, Cost: input.Cost, Stock: input.Stock, ReorderLevel: input.ReorderLevel, ImageURL: imageURL})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "product created"})
}