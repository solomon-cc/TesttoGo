package controller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"testogo/internal/model/entity"
	"testogo/internal/model/request"
	"testogo/pkg/database"

	"github.com/gin-gonic/gin"
)

const (
	ImportMaxFileSize = 50 << 20 // 50MB
	ImportDir         = "./uploads/imports"
)

var importAllowedExtensions = map[string]bool{
	".pdf":  true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".xlsx": true,
	".xls":  true,
	".docx": true,
	".doc":  true,
}

// ImportQuestionResponse 导入题目响应
type ImportQuestionResponse struct {
	ID            uint     `json:"id"`
	Title         string   `json:"title"`
	Type          string   `json:"type"`
	DetectedType  string   `json:"detected_type"`  // 检测到的题目类型
	Confidence    float64  `json:"confidence"`     // 识别置信度
	Options       []string `json:"options"`
	Answer        string   `json:"answer"`
	MediaURLs     []string `json:"media_urls"`
	Status        string   `json:"status"`    // pending, approved, rejected
	ErrorMessage  string   `json:"error_message,omitempty"`
	OriginalIndex int      `json:"original_index"` // 在原文件中的位置
}

// @Summary 上传文件进行题目导入
// @Description 支持PDF、图片、Excel等格式的题目批量导入
// @Tags 题目导入
// @Accept multipart/form-data
// @Produce json
// @Security BasicAuth
// @Param file formData file true "要导入的文件"
// @Param auto_detect formData boolean false "是否自动检测题型"
// @Param default_grade formData string false "默认年级"
// @Param default_subject formData string false "默认科目"
// @Success 200 {object} map[string]interface{} "返回解析的题目列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/import/questions [post]
func ImportQuestions(c *gin.Context) {
	// 确保导入目录存在
	if err := os.MkdirAll(ImportDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建导入目录失败"})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败"})
		return
	}

	// 验证文件大小
	if file.Size > ImportMaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件超过最大大小限制(50MB)"})
		return
	}

	// 验证文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !importAllowedExtensions[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件格式"})
		return
	}

	// 保存上传的文件
	timestamp := time.Now().Unix()
	savedFileName := fmt.Sprintf("%d_%s", timestamp, file.Filename)
	savedFilePath := filepath.Join(ImportDir, savedFileName)

	if err := c.SaveUploadedFile(file, savedFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	// 获取可选参数
	autoDetect := c.DefaultPostForm("auto_detect", "true") == "true"
	defaultGrade := c.DefaultPostForm("default_grade", "grade1")
	defaultSubject := c.DefaultPostForm("default_subject", "math")

	// 解析文件内容
	questions, err := parseImportFile(savedFilePath, ext, autoDetect, defaultGrade, defaultSubject)
	if err != nil {
		// 清理上传的文件
		os.Remove(savedFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("解析文件失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "文件解析成功",
		"file_id":        savedFileName,
		"questions":      questions,
		"total_count":    len(questions),
		"detected_count": countDetectedQuestions(questions),
	})
}

// @Summary 确认导入题目
// @Description 确认并保存解析的题目到数据库
// @Tags 题目导入
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param request body request.ConfirmImportRequest true "确认导入请求参数"
// @Success 200 {object} map[string]interface{} "返回导入结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/import/questions/confirm [post]
func ConfirmImportQuestions(c *gin.Context) {
	var req request.ConfirmImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")
	var createdQuestions []entity.Question
	var errors []string

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, questionData := range req.Questions {
		if questionData.Status != "approved" {
			continue // 跳过未批准的题目
		}

		// 创建题目实体
		question := entity.Question{
			Title:       questionData.Title,
			Type:        entity.QuestionType(questionData.Type),
			Difficulty:  questionData.Difficulty,
			Grade:       questionData.Grade,
			Subject:     questionData.Subject,
			Topic:       questionData.Topic,
			Options:     strings.Join(questionData.Options, ","), // 简单处理，实际应该用JSON
			Answer:      questionData.Answer,
			Explanation: questionData.Explanation,
			CreatorID:   userID,
			MediaURLs:   strings.Join(questionData.MediaURLs, ","), // 简单处理，实际应该用JSON
			LayoutType:  questionData.LayoutType,
			ElementData: questionData.ElementData,
			Tags:        questionData.Tags,
		}

		if err := tx.Create(&question).Error; err != nil {
			errors = append(errors, fmt.Sprintf("创建题目失败: %s - %v", questionData.Title, err))
			continue
		}

		createdQuestions = append(createdQuestions, question)
	}

	if len(errors) > 0 {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "导入过程中出现错误",
			"details": errors,
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存题目失败"})
		return
	}

	// 清理临时文件
	if req.FileID != "" {
		filePath := filepath.Join(ImportDir, req.FileID)
		os.Remove(filePath)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "题目导入成功",
		"imported_count": len(createdQuestions),
		"total_requested": len(req.Questions),
		"questions":     createdQuestions,
	})
}

// 解析导入文件
func parseImportFile(filePath, ext string, autoDetect bool, defaultGrade, defaultSubject string) ([]ImportQuestionResponse, error) {
	switch ext {
	case ".pdf":
		return parsePDFFile(filePath, autoDetect, defaultGrade, defaultSubject)
	case ".jpg", ".jpeg", ".png":
		return parseImageFile(filePath, autoDetect, defaultGrade, defaultSubject)
	case ".xlsx", ".xls":
		return parseExcelFile(filePath, autoDetect, defaultGrade, defaultSubject)
	case ".docx", ".doc":
		return parseWordFile(filePath, autoDetect, defaultGrade, defaultSubject)
	default:
		return nil, fmt.Errorf("不支持的文件格式: %s", ext)
	}
}

// 解析PDF文件 (简化版本，实际需要PDF解析库)
func parsePDFFile(filePath string, autoDetect bool, defaultGrade, defaultSubject string) ([]ImportQuestionResponse, error) {
	// TODO: 实际实现需要使用PDF解析库，如 github.com/ledongthuc/pdf
	// 这里提供一个模拟的实现
	questions := []ImportQuestionResponse{
		{
			ID:            0,
			Title:         "示例题目：蝴蝶比花朵（）",
			Type:          "comparison",
			DetectedType:  "comparison",
			Confidence:    0.85,
			Options:       []string{"多", "少", "一样"},
			Answer:        "多",
			MediaURLs:     []string{},
			Status:        "pending",
			OriginalIndex: 1,
		},
	}

	return questions, nil
}

// 解析图片文件
func parseImageFile(filePath string, autoDetect bool, defaultGrade, defaultSubject string) ([]ImportQuestionResponse, error) {
	// TODO: 实际实现需要OCR和图像识别
	return []ImportQuestionResponse{}, nil
}

// 解析Excel文件
func parseExcelFile(filePath string, autoDetect bool, defaultGrade, defaultSubject string) ([]ImportQuestionResponse, error) {
	// TODO: 实际实现需要Excel解析库
	return []ImportQuestionResponse{}, nil
}

// 解析Word文件
func parseWordFile(filePath string, autoDetect bool, defaultGrade, defaultSubject string) ([]ImportQuestionResponse, error) {
	// TODO: 实际实现需要Word解析库
	return []ImportQuestionResponse{}, nil
}

// 统计检测到的题目数量
func countDetectedQuestions(questions []ImportQuestionResponse) int {
	count := 0
	for _, q := range questions {
		if q.Confidence > 0.5 { // 置信度阈值
			count++
		}
	}
	return count
}