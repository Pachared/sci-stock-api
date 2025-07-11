package controllers

import "github.com/gin-gonic/gin"

func CreateOrder(c *gin.Context) {
    c.JSON(200, gin.H{"message": "order created"})
}

func GetOrders(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Orders list",
	})
}
