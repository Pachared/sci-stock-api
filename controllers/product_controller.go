package controllers

import (
	"fmt"
	"net/http"
	"time"

	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"sci-stock-api/config"
	"sci-stock-api/models"
)

const productsCacheTTL = 10 * time.Minute

func productsCacheKey(category string) string {
	return fmt.Sprintf("products:category:%s", category)
}

func invalidateProductsCache(category string) {
	if config.RDB == nil {
		return
	}

	_ = config.RDB.Del(config.Ctx, productsCacheKey(category)).Err()
}

var productTables = map[string]interface{}{
	"dried_food": &[]models.DriedFood{},
	"fresh_food": &[]models.FreshFood{},
	"snack":      &[]models.Snack{},
	"soft_drink": &[]models.SoftDrink{},
	"stationery": &[]models.Stationery{},
}

func createProduct(category string, input models.ProductInput) error {
	switch category {
	case "dried_food":
		return config.DB.Create(&models.DriedFood{
			ProductName:  input.ProductName,
			Barcode:      input.Barcode,
			Price:        input.Price,
			Cost:         input.Cost,
			Stock:        input.Stock,
			ReorderLevel: input.ReorderLevel,
			ImageURL:     input.ImageURL,
		}).Error
	case "fresh_food":
		return config.DB.Create(&models.FreshFood{
			ProductName:  input.ProductName,
			Barcode:      input.Barcode,
			Price:        input.Price,
			Cost:         input.Cost,
			Stock:        input.Stock,
			ReorderLevel: input.ReorderLevel,
			ImageURL:     input.ImageURL,
		}).Error
	case "snack":
		return config.DB.Create(&models.Snack{
			ProductName:  input.ProductName,
			Barcode:      input.Barcode,
			Price:        input.Price,
			Cost:         input.Cost,
			Stock:        input.Stock,
			ReorderLevel: input.ReorderLevel,
			ImageURL:     input.ImageURL,
		}).Error
	case "soft_drink":
		return config.DB.Create(&models.SoftDrink{
			ProductName:  input.ProductName,
			Barcode:      input.Barcode,
			Price:        input.Price,
			Cost:         input.Cost,
			Stock:        input.Stock,
			ReorderLevel: input.ReorderLevel,
			ImageURL:     input.ImageURL,
		}).Error
	case "stationery":
		return config.DB.Create(&models.Stationery{
			ProductName:  input.ProductName,
			Barcode:      input.Barcode,
			Price:        input.Price,
			Cost:         input.Cost,
			Stock:        input.Stock,
			ReorderLevel: input.ReorderLevel,
			ImageURL:     input.ImageURL,
		}).Error
	}
	return nil
}

func GetProductsByCategory(c *gin.Context) {
	category := c.Param("category")

	if config.RDB != nil {
		if cached, err := config.RDB.Get(config.Ctx, productsCacheKey(category)).Bytes(); err == nil && len(cached) > 0 {
			c.Data(http.StatusOK, "application/json; charset=utf-8", cached)
			return
		}
	}

	var result interface{}

	switch category {
	case "dried_food":
		result = &[]models.DriedFood{}
	case "fresh_food":
		result = &[]models.FreshFood{}
	case "snack":
		result = &[]models.Snack{}
	case "soft_drink":
		result = &[]models.SoftDrink{}
	case "stationery":
		result = &[]models.Stationery{}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
		return
	}

	if err := config.DB.Find(result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot fetch products"})
		return
	}

	if  config.RDB != nil {
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

	invalidateProductsCache(category)
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

	invalidateProductsCache(category)
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

	var err error
	switch category {
	case "dried_food":
		var product models.DriedFood
		err = config.DB.Where("barcode = ?", barcode).First(&product).Error
		if err == nil {
			product.ProductName = input.ProductName
			product.Price = input.Price
			product.Cost = input.Cost
			product.Stock = input.Stock
			product.ReorderLevel = input.ReorderLevel
			product.ImageURL = input.ImageURL
			err = config.DB.Save(&product).Error
		}
	case "fresh_food":
		var product models.FreshFood
		err = config.DB.Where("barcode = ?", barcode).First(&product).Error
		if err == nil {
			product.ProductName = input.ProductName
			product.Price = input.Price
			product.Cost = input.Cost
			product.Stock = input.Stock
			product.ReorderLevel = input.ReorderLevel
			product.ImageURL = input.ImageURL
			err = config.DB.Save(&product).Error
		}
	case "snack":
		var product models.Snack
		err = config.DB.Where("barcode = ?", barcode).First(&product).Error
		if err == nil {
			product.ProductName = input.ProductName
			product.Price = input.Price
			product.Cost = input.Cost
			product.Stock = input.Stock
			product.ReorderLevel = input.ReorderLevel
			product.ImageURL = input.ImageURL
			err = config.DB.Save(&product).Error
		}
	case "soft_drink":
		var product models.SoftDrink
		err = config.DB.Where("barcode = ?", barcode).First(&product).Error
		if err == nil {
			product.ProductName = input.ProductName
			product.Price = input.Price
			product.Cost = input.Cost
			product.Stock = input.Stock
			product.ReorderLevel = input.ReorderLevel
			product.ImageURL = input.ImageURL
			err = config.DB.Save(&product).Error
		}
	case "stationery":
		var product models.Stationery
		err = config.DB.Where("barcode = ?", barcode).First(&product).Error
		if err == nil {
			product.ProductName = input.ProductName
			product.Price = input.Price
			product.Cost = input.Cost
			product.Stock = input.Stock
			product.ReorderLevel = input.ReorderLevel
			product.ImageURL = input.ImageURL
			err = config.DB.Save(&product).Error
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update product"})
		return
	}

	invalidateProductsCache(category)
	c.JSON(http.StatusOK, gin.H{"message": "product updated"})
}

func DeleteProductByCategory(c *gin.Context) {
	category := c.Param("category")
	barcode := c.Param("barcode")

	var err error

	switch category {
	case "dried_food":
		err = config.DB.Where("barcode = ?", barcode).Delete(&models.DriedFood{}).Error
	case "fresh_food":
		err = config.DB.Where("barcode = ?", barcode).Delete(&models.FreshFood{}).Error
	case "snack":
		err = config.DB.Where("barcode = ?", barcode).Delete(&models.Snack{}).Error
	case "soft_drink":
		err = config.DB.Where("barcode = ?", barcode).Delete(&models.SoftDrink{}).Error
	case "stationery":
		err = config.DB.Where("barcode = ?", barcode).Delete(&models.Stationery{}).Error
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot delete product"})
		return
	}

	invalidateProductsCache(category)
	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}
