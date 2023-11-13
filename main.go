package main

import (
	"fmt"
	"net/http"
	"os"
	"universal-gcp-file-aapi-go/controllers"
	"universal-gcp-file-aapi-go/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/unrolled/secure"
)

func main() {

	secureMiddleware := secure.New(secure.Options{
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		FrameDeny:             true,
		ContentSecurityPolicy: "default-src 'self'",
	})

	// * using env on dev only
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	// * gin setting
	// gin.SetMode(gin.ReleaseMode) // * use on production
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Use(func(c *gin.Context) {
		secureMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
		c.Next()
	})

	// * routing
	r.POST("/files", controllers.GetListFile)
	r.POST("/files/upload", middleware.ProcessFileMiddleware, controllers.UploadFile)
	r.POST("/files/delete", controllers.DeleteFileController)
	r.POST("/auth/check", controllers.CheckAccConnection)
	r.POST("/auth", controllers.CreateAuthToken)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	err = r.Run(":" + port)
	if err != nil {
		fmt.Println("Error when trying start server")
	}

}
