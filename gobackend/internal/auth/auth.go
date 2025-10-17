package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"GoComicMosaic-gobackend/gobackend/internal/models"
)

// 配置常量
const (
	// 密钥，生产环境应该使用更安全的方式存储
	SecretKey           = "your-secret-key-replace-in-production"
	AccessTokenDuration = time.Hour * 24 * 7 // 一周
)

// Claims 定义JWT的声明
type Claims struct {
	Username string `json:"sub"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// GeneratePasswordHash 生成密码哈希
func GeneratePasswordHash(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword 验证密码
func VerifyPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// GenerateToken 生成JWT令牌
func GenerateToken(username string, isAdmin bool) (string, error) {
	// 创建令牌声明
	claims := &Claims{
		Username: username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名令牌
	return token.SignedString([]byte(SecretKey))
}

// VerifyToken 验证JWT令牌
func VerifyToken(tokenString string) (*Claims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名方法: %v", token.Header["alg"])
		}
		return []byte(SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证令牌声明
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的令牌")
}

// AuthenticateUser 认证用户
func AuthenticateUser(username, password string) (*models.User, error) {
	// 查询用户
	var user models.User
	err := models.DB.Get(&user, "SELECT * FROM users WHERE username = ?", username)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 验证密码
	if !VerifyPassword(user.HashedPassword, password) {
		return nil, errors.New("密码错误")
	}

	return &user, nil
}

// IsAdmin 检查当前用户是否为管理员
func IsAdmin(c *gin.Context) bool {
	// 从请求头中获取令牌
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false
	}

	// 提取令牌
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		// 没有找到Bearer前缀
		return false
	}

	// 验证令牌
	claims, err := VerifyToken(tokenString)
	if err != nil {
		return false
	}

	// 检查是否为管理员
	return claims.IsAdmin
}
