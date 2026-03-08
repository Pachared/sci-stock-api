package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"sci-stock-api/config"
	"sci-stock-api/models"
)

const productsCacheTTL = 10 * time.Minute

func productsCacheKey(category string) string {
	return fmt.Sprintf("products:category:%s", category)
}

func invalidateAllProductsCache() {
	if config.RDB == nil {
		return
	}

	keys, err := config.RDB.Keys(config.Ctx, "products:category:*").Result()
	if err != nil || len(keys) == 0 {
		return
	}

	_ = config.RDB.Del(config.Ctx, keys...).Err()
}

var productTables = map[string]interface{}{
	"dried_food": &[]models.DriedFood{},
	"fresh_food": &[]models.FreshFood{},
	"snack":      &[]models.Snack{},
	"soft_drink": &[]models.SoftDrink{},
	"stationery": &[]models.Stationery{},
}

func getProductSliceByCategory(category string) (interface{}, error) {
	switch category {
	case "dried_food":
		return &[]models.DriedFood{}, nil
	case "fresh_food":
		return &[]models.FreshFood{}, nil
	case "snack":
		return &[]models.Snack{}, nil
	case "soft_drink":
		return &[]models.SoftDrink{}, nil
	case "stationery":
		return &[]models.Stationery{}, nil
	default:
		return nil, fmt.Errorf("invalid category")
	}
}

func getProductModelByCategory(category string) (interface{}, error) {
	switch category {
	case "dried_food":
		return &models.DriedFood{}, nil
	case "fresh_food":
		return &models.FreshFood{}, nil
	case "snack":
		return &models.Snack{}, nil
	case "soft_drink":
		return &models.SoftDrink{}, nil
	case "stationery":
		return &models.Stationery{}, nil
	default:
		return nil, fmt.Errorf("invalid category")
	}
}

func applyProductInput(product interface{}, input models.ProductInput) error {
	switch p := product.(type) {
	case *models.DriedFood:
		p.ProductName = input.ProductName
		p.Price = input.Price
		p.Cost = input.Cost
		p.Stock = input.Stock
		p.ReorderLevel = input.ReorderLevel
		p.ImageURL = input.ImageURL
	case *models.FreshFood:
		p.ProductName = input.ProductName
		p.Price = input.Price
		p.Cost = input.Cost
		p.Stock = input.Stock
		p.ReorderLevel = input.ReorderLevel
		p.ImageURL = input.ImageURL
	case *models.Snack:
		p.ProductName = input.ProductName
		p.Price = input.Price
		p.Cost = input.Cost
		p.Stock = input.Stock
		p.ReorderLevel = input.ReorderLevel
		p.ImageURL = input.ImageURL
	case *models.SoftDrink:
		p.ProductName = input.ProductName
		p.Price = input.Price
		p.Cost = input.Cost
		p.Stock = input.Stock
		p.ReorderLevel = input.ReorderLevel
		p.ImageURL = input.ImageURL
	case *models.Stationery:
		p.ProductName = input.ProductName
		p.Price = input.Price
		p.Cost = input.Cost
		p.Stock = input.Stock
		p.ReorderLevel = input.ReorderLevel
		p.ImageURL = input.ImageURL
	default:
		return fmt.Errorf("unsupported product model")
	}

	return nil
}

func createProduct(category string, input models.ProductInput) error {
	model, err := getProductModelByCategory(category)
	if err != nil {
		return err
	}

	if err := applyProductInput(model, input); err != nil {
		return err
	}

	switch p := model.(type) {
	case *models.DriedFood:
		p.Barcode = input.Barcode
	case *models.FreshFood:
		p.Barcode = input.Barcode
	case *models.Snack:
		p.Barcode = input.Barcode
	case *models.SoftDrink:
		p.Barcode = input.Barcode
	case *models.Stationery:
		p.Barcode = input.Barcode
	default:
		return fmt.Errorf("unsupported product model")
	}

	return config.DB.Create(model).Error
}

func GetProductsByCategory(c *gin.Context) {
	category := c.Param("category")

	if config.RDB != nil {
		if cached, err := config.RDB.Get(config.Ctx, productsCacheKey(category)).Bytes(); err == nil && len(cached) > 0 {
			c.Data(http.StatusOK, "application/json; charset=utf-8", cached)
			return
		}
	}

	result, err := getProductSliceByCategory(category)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
		return
	}

	if err := config.DB.Find(result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot fetch products"})
		return
	}

	if config.RDB != nil {
		if data, err := json.Marshal(result); err == nil {
			_ = config.RDB.Set(config.Ctx, productsCacheKey(category), data, productsCacheTTL).Err()
		}
	}

	c.JSON(http.StatusOK, result)
}

func CreateProductByCategory(c *gin.Context) {
	category := c.Param("category")
	_, exists := productTables[category]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read request body"})
		return
	}

	var inputs []models.ProductInput
	if err := json.Unmarshal(bodyBytes, &inputs); err != nil {
		var singleInput models.ProductInput
		if err2 := json.Unmarshal(bodyBytes, &singleInput); err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input format"})
			return
		}
		inputs = []models.ProductInput{singleInput}
	}

	for _, input := range inputs {
		if err := createProduct(category, input); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	invalidateAllProductsCache()
	c.JSON(http.StatusCreated, gin.H{"message": "product(s) created"})
}

func CreateProductsBulkByCategory(c *gin.Context) {
	category := c.Param("category")

	_, exists := productTables[category]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
		return
	}

	var inputs []models.ProductInput
	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, input := range inputs {
		if err := createProduct(category, input); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	invalidateAllProductsCache()
	c.JSON(http.StatusCreated, gin.H{"message": "bulk products created"})
}

func UpdateProductByCategory(c *gin.Context) {
	category := c.Param("category")
	barcode := c.Param("barcode")

	var input models.ProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	model, err := getProductModelByCategory(category)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
		return
	}

	err = config.DB.Where("barcode = ?", barcode).First(model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot find product"})
		return
	}

	if err := applyProductInput(model, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Save(model).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update product"})
		return
	}

	invalidateAllProductsCache()
	c.JSON(http.StatusOK, gin.H{"message": "product updated"})
}

func DeleteProductByCategory(c *gin.Context) {
	category := c.Param("category")
	barcode := c.Param("barcode")

	model, err := getProductModelByCategory(category)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
		return
	}

	tx := config.DB.Where("barcode = ?", barcode).Delete(model)
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot delete product"})
		return
	}

	if tx.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	invalidateAllProductsCache()
	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}
