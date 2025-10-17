package handlers

import (
	"fmt"

	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/aspnmy/GoComicMosaic-gobackend/gobackend/internal/auth"
	"github.com/aspnmy/GoComicMosaic-gobackend/gobackend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetUsers 获取所有用户 - 仅管理员可访问
func GetUsers(c *gin.Context) {
	var users []models.User
	err := models.DB.Select(&users, `SELECT id, username, is_admin, created_at FROM users`)
	if err != nil {
		log.Printf("获取用户列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetUserRoles 获取用户角色列表 - 仅管理员可访问
func GetUserRoles(c *gin.Context) {
	// 这里可以根据实际需求定义不同的角色
	roles := []string{"admin", "user"}
	c.JSON(http.StatusOK, roles)
}

// CreateUser 创建新用户 - 仅管理员可访问
func CreateUser(c *gin.Context) {
	var userCreate struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		IsAdmin  bool   `json:"is_admin"`
	}

	if err := c.ShouldBindJSON(&userCreate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 检查用户名是否已存在
	var count int
	err := models.DB.Get(&count, `SELECT COUNT(*) FROM users WHERE username = ?`, userCreate.Username)
	if err != nil {
		log.Printf("检查用户名是否存在失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		return
	}

	// 哈希密码
	hashedPassword, err := auth.GeneratePasswordHash(userCreate.Password)
	if err != nil {
		log.Printf("密码哈希失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	// 创建用户
	result, err := models.DB.Exec(
		`INSERT INTO users (username, hashed_password, is_admin, created_at) VALUES (?, ?, ?, ?)`,
		userCreate.Username, hashedPassword, userCreate.IsAdmin, time.Now(),
	)
	if err != nil {
		log.Printf("创建用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("创建用户失败: %v", err)})
		return
	}

	// 获取新用户ID
	userID, _ := result.LastInsertId()

	// 返回新创建的用户信息
	user := models.User{
		ID:        int(userID),
		Username:  userCreate.Username,
		IsAdmin:   userCreate.IsAdmin,
		CreatedAt: time.Now(),
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUser 更新用户信息 - 仅管理员可访问
func UpdateUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var userUpdate struct {
		Username string `json:"username"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"is_admin"`
	}

	if bindErr := c.ShouldBindJSON(&userUpdate); bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 检查用户是否存在
	var user models.User
	err = models.DB.Get(&user, `SELECT * FROM users WHERE id = ?`, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 如果更新用户名，检查新用户名是否已存在
	if userUpdate.Username != "" && userUpdate.Username != user.Username {
		var count int
		err = models.DB.Get(&count, `SELECT COUNT(*) FROM users WHERE username = ? AND id != ?`, userUpdate.Username, userID)
		if err != nil {
			log.Printf("检查用户名是否存在失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
			return
		}
	}

	// 准备更新语句
	updateFields := make([]string, 0)
	args := make([]interface{}, 0)

	// 更新用户名
	if userUpdate.Username != "" {
		updateFields = append(updateFields, "username = ?")
		args = append(args, userUpdate.Username)
	}

	// 更新密码
	if userUpdate.Password != "" {
		hashedPassword, hashErr := auth.GeneratePasswordHash(userUpdate.Password)
		if hashErr != nil {
			log.Printf("密码哈希失败: %v", hashErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败"})
			return
		}
		updateFields = append(updateFields, "hashed_password = ?")
		args = append(args, hashedPassword)
	}

	// 更新管理员状态
	updateFields = append(updateFields, "is_admin = ?")
	args = append(args, userUpdate.IsAdmin)

	// 如果没有需要更新的字段
	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有提供任何更新字段"})
		return
	}

	// 构建更新SQL
	query := "UPDATE users SET "
	for i, field := range updateFields {
		if i > 0 {
			query += ", "
		}
		query += field
	}
	query += " WHERE id = ?"
	args = append(args, userID)

	// 执行更新
	_, err = models.DB.Exec(query, args...)
	if err != nil {
		log.Printf("更新用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("更新用户失败: %v", err)})
		return
	}

	// 获取更新后的用户信息
	var updatedUser models.User
	err = models.DB.Get(&updatedUser, `SELECT id, username, is_admin, created_at FROM users WHERE id = ?`, userID)
	if err != nil {
		log.Printf("获取更新后的用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户成功，但获取更新后的信息失败"})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// DeleteUser 删除用户 - 仅管理员可访问
func DeleteUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 检查是否是当前登录用户
	currentUser, _ := GetCurrentUser(c)
	if currentUser != nil && currentUser.ID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除当前登录的用户"})
		return
	}

	// 检查用户是否存在
	var count int
	err = models.DB.Get(&count, `SELECT COUNT(*) FROM users WHERE id = ?`, userID)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 删除用户
	_, err = models.DB.Exec(`DELETE FROM users WHERE id = ?`, userID)
	if err != nil {
		log.Printf("删除用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("删除用户失败: %v", err)})
		return
	}

	c.Status(http.StatusNoContent)
}
