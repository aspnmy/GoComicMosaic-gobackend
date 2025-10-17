package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"

	"GoComicMosaic-gobackend/gobackend/internal/models"
	"GoComicMosaic-gobackend/gobackend/internal/utils"
)

// UploadImage 处理单个图片上传
func UploadImage(c *gin.Context) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件上传"})
		return
	}
	defer file.Close()

	// 检查文件类型
	ext := filepath.Ext(header.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件类型，仅支持jpg、jpeg、png、gif和webp"})
		return
	}

	// 为了计算哈希值，我们需要完整地读取文件内容
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("读取文件内容失败: %v", err)})
		return
	}

	// 创建一个字节读取器，这样可以多次读取文件内容
	fileReader := bytes.NewReader(fileBytes)

	// 计算文件哈希值
	fileHash, err := utils.CalculateFileHash(fileReader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("计算文件哈希失败: %v", err)})
		return
	}

	// 重新创建一个新的字节读取器用于保存文件
	fileReader.Seek(0, io.SeekStart) // 确保重置到开头

	// 保存文件
	savedPath, err := utils.SaveUploadedFile(fileReader, header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("保存文件失败: %v", err)})
		return
	}

	// 返回文件信息
	c.JSON(http.StatusOK, gin.H{
		"filename": savedPath, // 与Python版本保持一致，使用filename字段
		"url":      savedPath, // 保留url字段以兼容可能的前端代码
		"hash":     fileHash,
		"name":     header.Filename,
		"size":     header.Size,
		"type":     filepath.Ext(header.Filename),
	})
}

// UploadMultipleImages 处理多个图片上传（批量上传）
func UploadMultipleImages(c *gin.Context) {
	// 获取上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件上传"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有文件上传"})
		return
	}

	results := make([]gin.H, 0, len(files))

	for _, fileHeader := range files {
		// 打开文件
		file, err := fileHeader.Open()
		if err != nil {
			results = append(results, gin.H{
				"name":  fileHeader.Filename,
				"error": "打开文件失败",
			})
			continue
		}
		defer file.Close()

		// 检查文件类型
		ext := filepath.Ext(fileHeader.Filename)
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
			results = append(results, gin.H{
				"name":  fileHeader.Filename,
				"error": "不支持的文件类型，仅支持jpg、jpeg、png、gif和webp",
			})
			continue
		}

		// 为了计算哈希值，我们需要完整地读取文件内容
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			results = append(results, gin.H{
				"name":  fileHeader.Filename,
				"error": fmt.Sprintf("读取文件内容失败: %v", err),
			})
			continue
		}

		// 创建一个字节读取器，这样可以多次读取文件内容
		fileReader := bytes.NewReader(fileBytes)

		// 计算文件哈希值
		fileHash, err := utils.CalculateFileHash(fileReader)
		if err != nil {
			results = append(results, gin.H{
				"name":  fileHeader.Filename,
				"error": fmt.Sprintf("计算文件哈希失败: %v", err),
			})
			continue
		}

		// 重置文件指针
		fileReader.Seek(0, io.SeekStart)

		// 保存文件
		savedPath, err := utils.SaveUploadedFile(fileReader, fileHeader.Filename)
		if err != nil {
			results = append(results, gin.H{
				"name":  fileHeader.Filename,
				"error": fmt.Sprintf("保存文件失败: %v", err),
			})
			continue
		}

		// 添加结果
		results = append(results, gin.H{
			"filename": savedPath, // 与Python版本保持一致
			"url":      savedPath, // 保留url字段以兼容可能的前端代码
			"hash":     fileHash,
			"name":     fileHeader.Filename,
			"size":     fileHeader.Size,
			"type":     filepath.Ext(fileHeader.Filename),
		})
	}

	c.JSON(http.StatusOK, results)
}

// LikeResource 增加资源喜欢计数
func LikeResource(c *gin.Context) {
	// 获取路径参数
	resourceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的资源ID"})
		return
	}

	// 增加喜欢计数
	_, err = models.DB.Exec(
		`UPDATE resources SET likes_count = likes_count + 1 WHERE id = ?`,
		resourceID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "增加喜欢计数失败"})
		return
	}

	c.Status(http.StatusOK)
}

// UnlikeResource 减少资源喜欢计数
func UnlikeResource(c *gin.Context) {
	// 获取路径参数
	resourceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的资源ID"})
		return
	}

	// 减少喜欢计数，但不小于0
	_, err = models.DB.Exec(
		`UPDATE resources SET likes_count = CASE WHEN likes_count > 0 THEN likes_count - 1 ELSE 0 END WHERE id = ?`,
		resourceID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "减少喜欢计数失败"})
		return
	}

	c.Status(http.StatusOK)
}
