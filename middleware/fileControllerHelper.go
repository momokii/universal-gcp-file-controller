package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"universal-gcp-file-aapi-go/models"

	"cloud.google.com/go/storage"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"google.golang.org/api/option"
)

func IsInvalidExt(extInput string) bool {
	validExt := []string{".jpg", ".jpeg", ".png", ".gif"}
	for _, ext := range validExt {
		if strings.ToLower(extInput) == "."+ext {
			return true
		}
	}
	return false
}

func InitCS_GCP(c *gin.Context, inputUser *models.InputUserAll) (*storage.Client, error, string) {
	typeSA := "service_account"
	var SAData models.InputServiceAccount // * service account data

	// * check user input json if not using binding JSON from another controller, try to bind json
	if inputUser == nil {
		// * if not using binding json, try to bind json (* if using check connection so will bind from here)
		var inputUserData models.InputUserData // * input user data
		if err := c.ShouldBindJSON(&inputUserData); err != nil {
			return nil, err, ""
		}

		// * copy binding user data to inputUser Data
		inputUser = &models.InputUserAll{}
		copier.Copy(&inputUser, &inputUserData)
	}

	if inputUser.UsingToken {
		// * if using token check token
		if inputUser.Token == "" {
			return nil, fmt.Errorf("token is empty, please insert your token"), ""
		}

		decode_token, err := jwt.Parse(inputUser.Token, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_KEY")), nil
		})
		if err != nil {
			return nil, fmt.Errorf("invalid token, please check your token again"), ""
		}

		token, ok := decode_token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("token is invalid/expired please check again your token or create new token"), ""
		}

		SAData.ProjectID = token["project_id"].(string)
		SAData.PrivateKey = token["private_key"].(string)
		SAData.ClientEmail = token["client_email"].(string)

	} else {
		// * if not using token so bind user input to all field below
		SAData.ProjectID = inputUser.ProjectID
		SAData.PrivateKey = inputUser.PrivateKey
		SAData.ClientEmail = inputUser.ClientEmail
	}

	jsonKeyData := map[string]interface{}{
		"type":         typeSA,
		"project_id":   SAData.ProjectID,
		"private_key":  SAData.PrivateKey,
		"client_email": SAData.ClientEmail,
	}

	jsonKey, err := json.Marshal(jsonKeyData)
	if err != nil {
		return nil, err, ""
	}

	// * Load the service account JSON key file
	opt := option.WithCredentialsJSON(jsonKey)

	// * Create a Google Cloud Storage client
	ctx := context.Background()
	client, err := storage.NewClient(ctx, opt)
	if err != nil {
		return nil, fmt.Errorf("error when trying connect with your input data, please check again your input/token data"), ""
	}

	// * check connection to gcp with login client
	_, err = client.ServiceAccount(ctx, SAData.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("error to get data & connect to your gcp account, please check your input data again"), ""
	}

	return client, nil, SAData.ProjectID
}

// * check if object exist
func IsObjectExist(c *gin.Context, filepath string, buckeName string, client *storage.Client) error {
	_, err := client.Bucket(buckeName).Object(filepath).Attrs(context.Background())
	if err == storage.ErrObjectNotExist {
		return fmt.Errorf("object not exist")
	}

	return nil
}
