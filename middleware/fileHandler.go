package middleware

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ProcessFileMiddleware(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Error when getting file, please check your file again")
		return
	}

	// * check file must be exist
	files := form.File["file"]
	if len(files) == 0 {
		ErrorResponse(c, http.StatusBadRequest, "Please Insert File")
		return
	}

	// * check file max size if used this controller config
	file := files[0]
	// maxSize := "2" // ex using 2MB
	maxSizeFile, _ := strconv.ParseInt("2", 10, 64)
	if file.Size > maxSizeFile*1024*1024 {
		ErrorResponse(c, http.StatusBadRequest, "File too large, please insert file with size less than 2MB")
		return
	}

	// * check file extension if using this config
	ext := filepath.Ext(file.Filename)
	if IsInvalidExt(ext) {
		ErrorResponse(c, http.StatusBadRequest, "File extension not allowed, please insert file with extension .jpg, .jpeg, .png, .gif")
		return
	}

	fileData, err := file.Open()
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Error when opening file, please check your file again")
		return
	}
	defer fileData.Close()

	// * read and buffered the file data
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, fileData); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Error when buffered file, please check your file again")
		return
	}

	c.Set("fileBuffer", buffer.Bytes())
	c.Set("FileExt", ext)
	c.Next()
}
