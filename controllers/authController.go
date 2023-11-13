package controllers

import (
	"net/http"
	"os"
	"time"
	"universal-gcp-file-aapi-go/middleware"
	"universal-gcp-file-aapi-go/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func CreateAuthToken(c *gin.Context) {
	var userSAData models.InputServiceAccount
	typeSA := "service_account"

	if err := c.BindJSON(&userSAData); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "Error when binding json data, please check your input data again and make sure all data is filled with string format")
		return
	}

	// * create token for user with sa data account
	sign := jwt.New(jwt.GetSigningMethod("HS256"))
	claim := sign.Claims.(jwt.MapClaims)
	claim["type"] = typeSA
	claim["project_id"] = userSAData.ProjectID
	claim["private_key"] = userSAData.PrivateKey
	claim["client_email"] = userSAData.ClientEmail
	claim["exp"] = time.Now().Add(time.Hour * 168).Unix() // 1 week exp token time

	token, err := sign.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errors":  false,
		"message": "success auth data",
		"data": gin.H{
			"token":      token,
			"token_type": "JWT",
		},
	})
}

func CheckAccConnection(c *gin.Context) {
	var inputUser models.InputUserAll
	if err := c.ShouldBindJSON(&inputUser); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	client, err, projectId := middleware.InitCS_GCP(c, &inputUser)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	defer client.Close()

	c.JSON(http.StatusOK, gin.H{
		"errors":  false,
		"message": "success check connection to gcp",
		"data": gin.H{
			"project_id": projectId,
		},
	})
}
