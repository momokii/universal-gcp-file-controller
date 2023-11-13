package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"universal-gcp-file-aapi-go/middleware"
	"universal-gcp-file-aapi-go/models"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

func GetListFile(c *gin.Context) {

	var gcpFileList []models.GetFile
	var InputUser models.InputUserAll

	err := c.ShouldBindJSON(&InputUser)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "invalid input data request")
		return
	}

	client, err, _ := middleware.InitCS_GCP(c, &InputUser)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer client.Close()

	// * get folderpath from query dat
	if InputUser.BucketName == "" {
		middleware.ErrorResponse(c, http.StatusBadRequest, "invalid input data request, please check your input. bucketname must be filled")
		return
	}

	// * get list file from gcp
	bucket := client.Bucket(InputUser.BucketName)
	if bucket == nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "bucketname not found, check your input data")
		return
	}
	// * check bucket if exists or not
	_, err = bucket.Attrs(context.Background())
	if err != nil {
		fmt.Println("masik sini")
		if err == storage.ErrBucketNotExist {
			middleware.ErrorResponse(c, http.StatusBadRequest, "bucketname not found, check your input data")
			return
		} else {
			middleware.ErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	allFile := bucket.Objects(
		context.Background(),
		&storage.Query{
			Prefix: InputUser.Folderpath,
		})

	for {
		data, err := allFile.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			middleware.ErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		publicUrl := "https://storage.googleapis.com/" + InputUser.BucketName + "/" + data.Name

		fileData := models.GetFile{
			Name:        data.Name,
			PublicURL:   publicUrl,
			DownloadURL: data.MediaLink,
		}

		gcpFileList = append(gcpFileList, fileData)
	}

	if gcpFileList == nil {
		gcpFileList = []models.GetFile{}
	}

	if InputUser.WithInfo {
		c.JSON(http.StatusOK, gin.H{
			"errors":  false,
			"message": "success get list file",
			"data": gin.H{
				"bucketname":        InputUser.BucketName,
				"prefix_folderpath": InputUser.Folderpath,
				"data":              gcpFileList,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errors":  false,
		"message": "success get list file",
		"data":    gcpFileList,
	})
}

func UploadFile(c *gin.Context) {
	var inputUser models.InputUserAll
	// * get user input for data auth
	inputUser.Token = c.Request.FormValue("token")
	if inputUser.Token == "" {
		middleware.ErrorResponse(c, http.StatusBadRequest, "invalid input data request, please check your input. token must be filled")
		return
	}
	// inputUser.ProjectID = c.Request.FormValue("project_id")
	// inputUser.PrivateKey = c.Request.FormValue("private_key")
	// inputUser.ClientEmail = c.Request.FormValue("client_email")
	inputUser.UsingToken = true //c.Request.FormValue("using_token") == "true"

	// * form input
	bucketName := c.Request.FormValue("bucketName")
	foldername := c.Request.FormValue("folderName")
	if foldername != "" {
		foldername = foldername + "/"
	}
	filename := c.Request.FormValue("filename")
	fileExt := c.GetString("fileExt")
	if (bucketName == "") || (filename == "") {
		middleware.ErrorResponse(c, http.StatusBadRequest, "invalid input data request, please check your input. foldername, filename, bucketname must be filled")
		return
	}
	objectName := foldername + filename + fileExt

	// * delete filepath if exist (delete old file) and replace with new file
	deleteFilePath := c.Request.FormValue("deleteFilePath")
	// ? filepath is optional, if filepath not exist, then deleteFilePath is empty string
	// ? filepath format is foldername/filename.fileExt

	client, err, _ := middleware.InitCS_GCP(c, &inputUser)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer client.Close()

	if deleteFilePath != "" {
		// * check if file need delete is exist
		err = middleware.IsObjectExist(c, deleteFilePath, bucketName, client)
		if err != nil {
			middleware.ErrorResponse(c, http.StatusBadRequest, "File need to delete is not exist, please check your input filepath or the bucketname")
			return
		}
		// * if exists, delete file
		err = client.Bucket(bucketName).Object(deleteFilePath).Delete(context.Background())
		if err != nil {
			middleware.ErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// * get buffered file
	fileBuffer, ok := c.MustGet("fileBuffer").([]byte)
	if !ok {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "fileBuffer not exist")
		return
	}
	reader := bytes.NewReader(fileBuffer) // * convert fileBuffer to reader

	// * upload file to gcp
	objectWriter := client.Bucket(bucketName).Object(objectName).NewWriter(context.Background())
	if _, err := io.Copy(objectWriter, reader); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	if err := objectWriter.Close(); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)

	c.JSON(http.StatusOK, gin.H{
		"errors":  false,
		"message": "success upload file",
		"data": gin.H{
			"url": url,
		},
	})
}

func DeleteFileController(c *gin.Context) {
	var inputUser models.InputUserAll

	err := c.ShouldBindJSON(&inputUser)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "invalid input data request")
		return
	}

	// * check client connection
	client, err, _ := middleware.InitCS_GCP(c, &inputUser)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	client.Close()

	if inputUser.BucketName == "" || inputUser.Filepath == "" {
		middleware.ErrorResponse(c, http.StatusBadRequest, "invalid input data request, please check your input. bucketname and filepath must be filled")
		return
	}

	// * check if file need delete is exist
	err = middleware.IsObjectExist(c, inputUser.Filepath, inputUser.BucketName, client)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "File need to delete is not exist, please check your input filepath or the bucketname")
		return
	}

	// * if exists, delete file
	err = client.Bucket(inputUser.BucketName).Object(inputUser.Filepath).Delete(context.Background())
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errors":  false,
		"message": "success delete file",
	})
}
