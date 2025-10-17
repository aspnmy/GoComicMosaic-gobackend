package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"GoComicMosaic-gobackend/gobackend/internal/config"
	"GoComicMosaic-gobackend/gobackend/internal/models"
)

// GetAllPosts 获取所有文章列表
func GetAllPosts(c *gin.Context) {
	posts, err := models.GetAllPosts(config.AssetPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, posts)
}

// GetPostByID 根据ID获取文章
func GetPostByID(c *gin.Context) {
	id := c.Param("id")
	post, err := models.GetPostByID(id, config.AssetPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// GetPostBySlug 根据Slug获取文章
func GetPostBySlug(c *gin.Context) {
	slug := c.Param("slug")
	post, err := models.GetPostBySlug(slug, config.AssetPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// CreatePost 创建新文章
func CreatePost(c *gin.Context) {

	var post models.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证必填字段
	if post.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "标题不能为空"})
		return
	}

	// 保存文章
	if err := models.SavePost(&post, config.AssetPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
}

// UpdatePost 更新文章
func UpdatePost(c *gin.Context) {
	// 获取ID参数
	id := c.Param("id")

	// 检查文章是否存在
	_, err := models.GetPostByID(id, config.AssetPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	var post models.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 确保ID一致
	post.ID = id

	// 保存文章
	if err := models.SavePost(&post, config.AssetPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

// DeletePost 删除文章
func DeletePost(c *gin.Context) {
	id := c.Param("id")

	// 删除文章
	if err := models.DeletePost(id, config.AssetPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文章已删除"})
}

// SearchPosts 搜索文章
func SearchPosts(c *gin.Context) {
	query := c.Query("q")

	posts, err := models.SearchPosts(query, config.AssetPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, posts)
}

// UploadPostImage 上传文章图片
func UploadPostImage(c *gin.Context) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法获取上传的文件"})
		return
	}
	defer file.Close()

	// 检查文件类型
	if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只允许上传图片文件"})
		return
	}

	// 生成唯一文件名
	filename := uuid.New().String() + filepath.Ext(header.Filename)

	// 确保目录存在
	imgsDir := filepath.Join(config.AssetPath, "posts", "imgs")
	if err := os.MkdirAll(imgsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建目录"})
		return
	}

	// 创建目标文件
	dst, err := os.Create(filepath.Join(imgsDir, filename))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建文件"})
		return
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法保存文件"})
		return
	}

	// 返回图片URL路径
	imageURL := "/api/assets/posts/imgs/" + filename
	c.JSON(http.StatusOK, gin.H{"url": imageURL})
}

// UploadPostFile 上传文章附件
func UploadPostFile(c *gin.Context) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法获取上传的文件"})
		return
	}
	defer file.Close()

	// 生成唯一文件名
	filename := uuid.New().String() + filepath.Ext(header.Filename)

	// 确保目录存在
	filesDir := filepath.Join(config.AssetPath, "posts", "files")
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建目录"})
		return
	}

	// 创建目标文件
	dst, err := os.Create(filepath.Join(filesDir, filename))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建文件"})
		return
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法保存文件"})
		return
	}

	// 返回文件URL路径
	fileURL := "/api/assets/posts/files/" + filename
	originalName := header.Filename
	c.JSON(http.StatusOK, gin.H{
		"url":  fileURL,
		"name": originalName,
	})
}
