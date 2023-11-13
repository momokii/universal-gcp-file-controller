package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func ErrorResponse(c *gin.Context, statusCode int, message string) {
	fmt.Println(message)
	c.JSON(statusCode, gin.H{
		"errors":  true,
		"message": message,
	})
	c.Abort()

}
