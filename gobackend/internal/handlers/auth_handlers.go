package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"GoComicMosaic-gobackend/gobackend/internal/auth"
	"GoComicMosaic-gobackend/gobackend/internal/models"
)

// LoginRequest 登录请求结构，遵循OAuth2标准
type LoginRequest struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
	// 下面是OAuth2可选字段，但我们不使用
	GrantType    string `form:"grant_type" json:"grant_type"`
	Scope        string `form:"scope" json:"scope"`
	ClientID     string `form:"client_id" json:"client_id"`
	ClientSecret string `form:"client_secret" json:"client_secret"`
}

// Login 处理用户登录请求
func Login(c *gin.Context) {
	var loginReq LoginRequest

	// 检测Content-Type
	contentType := c.GetHeader("Content-Type")
	log.Printf("登录请求Content-Type: %s", contentType)

	// 根据Content-Type选择不同的绑定方式
	var err error
	if c.ContentType() == "application/json" {
		// JSON绑定
		err = c.ShouldBindJSON(&loginReq)
	} else {
		// 表单绑定(支持普通表单和multipart/form-data)
		err = c.ShouldBind(&loginReq)
	}

	if err != nil {
		log.Printf("绑定请求参数失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	log.Printf("尝试登录用户: %s, Content-Type: %s", loginReq.Username, contentType)

	// 认证用户
	user, err := auth.AuthenticateUser(loginReq.Username, loginReq.Password)
	if err != nil {
		log.Printf("认证失败: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码不正确"})
		return
	}

	// 生成令牌
	token, err := auth.GenerateToken(user.Username, user.IsAdmin)
	if err != nil {
		log.Printf("生成令牌失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	log.Printf("用户 %s 登录成功, token长度: %d", user.Username, len(token))

	// 返回令牌 - 注意要按照OAuth2标准格式返回
	c.JSON(http.StatusOK, models.TokenResponse{
		AccessToken: token,
		TokenType:   "bearer",
	})
}

// GetCurrentUserInfo 获取当前用户信息
func GetCurrentUserInfo(c *gin.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	if user == nil {
		log.Printf("未找到用户信息")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	log.Printf("成功获取用户信息: %s", user.Username)
	c.JSON(http.StatusOK, user)
}

// UpdatePassword 更新用户密码
func UpdatePassword(c *gin.Context) {
	var passwordReq models.PasswordUpdate

	// 根据Content-Type选择不同的绑定方式
	var err error
	if c.ContentType() == "application/json" {
		// JSON绑定
		err = c.ShouldBindJSON(&passwordReq)
	} else {
		// 表单绑定
		err = c.ShouldBind(&passwordReq)
	}

	if err != nil {
		log.Printf("绑定密码更新请求参数失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 获取当前用户
	user, err := GetCurrentUser(c)
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 验证当前密码
	if !auth.VerifyPassword(user.HashedPassword, passwordReq.CurrentPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前密码不正确"})
		return
	}

	// 生成新密码哈希
	hashedPassword, err := auth.GeneratePasswordHash(passwordReq.NewPassword)
	if err != nil {
		log.Printf("生成密码哈希失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成密码哈希失败"})
		return
	}

	// 更新密码
	_, err = models.DB.Exec(
		"UPDATE users SET hashed_password = ? WHERE id = ?",
		hashedPassword, user.ID,
	)
	if err != nil {
		log.Printf("更新密码失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
		return
	}

	log.Printf("用户 %s 密码已更新", user.Username)

	// 更新内存中的用户信息
	user.HashedPassword = hashedPassword

	c.JSON(http.StatusOK, user)
}
