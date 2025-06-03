package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sys/unix"
)

func main() {
	router := gin.Default()

	router.POST("/upload", handleFileUpload)
	router.GET("/files", handleListFiles)
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
	uploadPath := "/data/uploads"

	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	freeSpace, err := getFreeSpace(uploadPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Statfs failed: %v", err)})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	if uint64(file.Size) > freeSpace {
		c.JSON(http.StatusInsufficientStorage, gin.H{"error": "Not enough disk space to save file"})
		return
	}

	filename := filepath.Base(file.Filename)
	dst := filepath.Join(uploadPath, filename)

	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	escapedName := url.PathEscape(filename)

	c.JSON(http.StatusOK, gin.H{
		"message":      "File uploaded successfully",
		"filename":     filename,
		"path":         dst,
		"download_url": fmt.Sprintf("/download/%s", escapedName),
		"free_space":   fmt.Sprintf("%.2f MB", float64(freeSpace)/(1024*1024)),
		"file_size":    fmt.Sprintf("%.2f MB", float64(file.Size)/(1024*1024)),
	})
}

func handleListFiles(c *gin.Context) {
	uploadDir := "/data/uploads"

	files, err := os.ReadDir(uploadDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploads directory"})
		return
	}

	var totalUsed uint64
	var fileList []gin.H

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		size := uint64(info.Size())
		totalUsed += size

		fileList = append(fileList, gin.H{
			"name":         file.Name(),
			"size":         fmt.Sprintf("%.2f MB", float64(size)/(1024*1024)),
			"created":      info.ModTime().Format(time.RFC3339),
			"download_url": fmt.Sprintf("/download/%s", url.PathEscape(file.Name())),
		})
	}

	var stat unix.Statfs_t
	if err := unix.Statfs(uploadDir, &stat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read disk stats"})
		return
	}

	totalStorage := stat.Blocks * uint64(stat.Bsize)
	availableStorage := stat.Bavail * uint64(stat.Bsize)

	c.JSON(http.StatusOK, gin.H{
		"files":            fileList,
		"storage_used_mb":  fmt.Sprintf("%.2f", float64(totalUsed)/(1024*1024)),
		"storage_total_mb": fmt.Sprintf("%.2f", float64(totalStorage)/(1024*1024)),
		"storage_free_mb":  fmt.Sprintf("%.2f", float64(availableStorage)/(1024*1024)),
		"file_count":       len(fileList),
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
