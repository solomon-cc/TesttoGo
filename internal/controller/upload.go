package controller

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	MaxFileSize = 10 << 20 // 10MB
	UploadDir   = "./uploads/images"
)

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

// @Summary 上传图片
// @Description 上传题目相关的图片文件，支持多文件上传
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce json
// @Security BasicAuth
// @Param images formData file true "图片文件，支持多个"
// @Success 200 {object} map[string]interface{} "返回上传的文件URL列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/upload/images [post]
func UploadImages(c *gin.Context) {
	// 确保上传目录存在
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败"})
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的图片文件"})
		return
	}

	var uploadedURLs []string

	for _, file := range files {
		// 验证文件大小
		if file.Size > MaxFileSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("文件 %s 超过最大大小限制(10MB)", file.Filename)})
			return
		}

		// 验证文件扩展名
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowedExtensions[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("不支持的文件格式: %s", ext)})
			return
		}

		// 生成唯一文件名
		timestamp := time.Now().Unix()
		filename := fmt.Sprintf("%d_%s", timestamp, file.Filename)
		filePath := filepath.Join(UploadDir, filename)

		// 保存文件
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("保存文件 %s 失败", file.Filename)})
			return
		}

		// 生成访问URL
		fileURL := fmt.Sprintf("/api/v1/media/%s", filename)
		uploadedURLs = append(uploadedURLs, fileURL)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "上传成功",
		"urls":    uploadedURLs,
		"count":   len(uploadedURLs),
	})
}

// @Summary 获取媒体文件
// @Description 获取上传的图片文件
// @Tags 文件上传
// @Param filename path string true "文件名"
// @Success 200 "返回图片文件"
// @Failure 404 {object} map[string]interface{} "文件不存在"
// @Router /api/v1/media/{filename} [get]
func ServeMedia(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件名不能为空"})
		return
	}

	// 安全检查：防止路径遍历攻击
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "非法文件名"})
		return
	}

	filePath := filepath.Join(UploadDir, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 设置适当的Content-Type
	ext := strings.ToLower(filepath.Ext(filename))
	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=3600") // 1小时缓存

	// 打开并传输文件
	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}
	defer file.Close()

	// 传输文件内容
	if _, err := io.Copy(c.Writer, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "传输文件失败"})
		return
	}
}