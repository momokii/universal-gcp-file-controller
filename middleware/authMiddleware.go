package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func CheckAuthUser(c *gin.Context) {

	tokenData := c.GetHeader("Authorization")
	if tokenData == "" {
		ErrorResponse(c, http.StatusBadRequest, "invalid input data request, please check your input. token must be filled")
		c.Abort()
		return
	}

	headerData := strings.Split(tokenData, " ")
	tokenData = headerData[1]
	if tokenData == "" {
		ErrorResponse(c, http.StatusBadRequest, "invalid input data request, please check your input. token must be filled with Bearer Token")
		return
	}

	// * decode token data
	claims := jwt.MapClaims{} // atau sesuaikan dengan tipe klaim yang Anda gunakan
	token, err := jwt.ParseWithClaims(tokenData, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_KEY")), nil
	})

	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if !token.Valid {
		ErrorResponse(c, http.StatusBadRequest, "invalid token")
		return
	}

	c.Set("userClaims", token)
	c.Next()
}
