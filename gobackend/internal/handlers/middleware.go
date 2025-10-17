package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"GoComicMosaic-gobackend/gobackend/internal/auth"
	"GoComicMosaic-gobackend/gobackend/internal/models"
)

// JWTAuthMiddleware 验证JWT令牌的中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			log.Printf("缺少Authorization头")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "需要认证"})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			log.Printf("无效的认证格式: %s", authHeader)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证格式"})
			c.Abort()
			return
		}

		// 验证令牌
		tokenString := parts[1]
		log.Printf("令牌: %s", tokenString)

		claims, err := auth.VerifyToken(tokenString)
		if err != nil {
			log.Printf("令牌验证失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌"})
			c.Abort()
			return
		}

		// log.Printf("令牌验证成功，用户: %s, 是否管理员: %v", claims.Username, claims.IsAdmin)

		// 存储用户信息到上下文
		c.Set("username", claims.Username)
		c.Set("is_admin", claims.IsAdmin)
		c.Next()
	}
}

// AdminAuthMiddleware 检查用户是否为管理员的中间件
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先确保用户已认证
		isAdmin, exists := c.Get("is_admin")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "需要认证"})
			c.Abort()
			return
		}

		// 检查是否为管理员
		if isAdminBool, ok := isAdmin.(bool); !ok || !isAdminBool {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUser 从数据库获取当前用户信息
func GetCurrentUser(c *gin.Context) (*models.User, error) {
	username, exists := c.Get("username")
	if !exists {
		return nil, nil
	}

	usernameStr, ok := username.(string)
	if !ok {
		return nil, nil
	}

	var user models.User
	err := models.DB.Get(&user, "SELECT * FROM users WHERE username = ?", usernameStr)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
