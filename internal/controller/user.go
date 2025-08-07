package controller

import (
	"net/http"
	"strconv"

	"testogo/internal/model/entity"
	"testogo/pkg/database"

	"github.com/gin-gonic/gin"
)

// @Summary 获取用户列表
// @Description 获取用户列表，支持分页
// @Tags 用户
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param page query integer false "页码，默认1"
// @Param page_size query integer false "每页数量，默认10"
// @Success 200 {object} map[string]interface{} "用户列表和总数"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/users [get]
func ListUsers(c *gin.Context) {
	var users []entity.User
	query := database.DB.Order("id desc")

	// 分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	var total int64
	database.DB.Model(&entity.User{}).Count(&total)

	err := query.Select("id, username, role, created_at, updated_at").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": users,
	})
}

// @Summary 更新用户角色
// @Description 更新指定用户的角色
// @Tags 用户
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param id path integer true "用户ID"
// @Param request body map[string]string true "角色信息 {\"role\": \"user|teacher|admin\"}"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/users/{id}/role [put]
func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Role string `json:"role" binding:"required,oneof=user teacher admin"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户是否存在
	var user entity.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 更新用户角色
	if err := database.DB.Model(&user).Update("role", input.Role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户角色失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	// 检查用户是否存在
	var user entity.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 不允许删除管理员账户
	if user.Role == entity.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能删除管理员账户"})
		return
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
