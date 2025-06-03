package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"golang.org/x/sys/unix"
)

func main() {
	router := gin.Default()

	router.POST("/upload", handleFileUpload)
	router.GET("/download/:filename", handleFileDownload)
	router.DELETE("/delete/:filename", handleFileDelete)

	router.Run(":8080")
}

func getFreeSpace(path string) (uint64, error) {
	var stat unix.Statfs_t
	err := unix.Statfs(path, &stat)
	if err != nil {
		return 0, err
	}
	// Available blocks * size per block = available space in bytes
	return stat.Bavail * uint64(stat.Bsize), nil
}

func handleFileUpload(c *gin.Context) {
	// Optional: check available space first
	freeSpace, err := getFreeSpace("/data/uploads")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not determine available storage"})
		return
	}
	fmt.Printf("Free space: %.2f MB\n", float64(freeSpace)/(1024*1024))

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// Check file size vs available space
	if uint64(file.Size) > freeSpace {
		c.JSON(http.StatusInsufficientStorage, gin.H{"error": "Not enough disk space to save file"})
		return
	}

	filename := filepath.Base(file.Filename)
	dst := filepath.Join("/data/uploads", filename)

	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "File uploaded successfully",
		"filename":   filename,
		"path":       dst,
		"free_space": fmt.Sprintf("%.2f MB", float64(freeSpace)/(1024*1024)),
		"file_size":  fmt.Sprintf("%.2f MB", float64(file.Size)/(1024*1024)),
	})
}

func handleFileDownload(c *gin.Context) {
	filename := filepath.Base(c.Param("filename"))
	filePath := filepath.Join("/data/uploads", filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")

	c.File(filePath)
}

func handleFileDelete(c *gin.Context) {
	filename := filepath.Base(c.Param("filename"))
	filePath := filepath.Join("/data/uploads", filename)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully", "filename": filename})
}
