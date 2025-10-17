package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aspnmy/GoComicMosaic-gobackend/gobackend/internal/config"
	"github.com/aspnmy/GoComicMosaic-gobackend/gobackend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ImageEnhanceRequest 图像增强请求结构
type ImageEnhanceRequest struct {
	SaveResult bool `json:"save_result" form:"save_result"`
}

// ImageEnhanceResponse 图像增强响应结构
type ImageEnhanceResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message,omitempty"`
	OriginalURL    string `json:"original_url,omitempty"`
	EnhancedURL    string `json:"enhanced_url,omitempty"`
	OriginalPath   string `json:"original_path,omitempty"`
	EnhancedPath   string `json:"enhanced_path,omitempty"`
	ProcessingTime string `json:"processing_time,omitempty"`
}

// EnhanceImageHandler 处理图像增强请求
func EnhanceImageHandler(c *gin.Context) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请提供图像文件",
		})
		return
	}
	defer file.Close()

	// 解析请求参数
	var req ImageEnhanceRequest
	if bindErr := c.ShouldBind(&req); bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求参数",
		})
		return
	}

	// 确保 handles 目录存在
	handlesDir := filepath.Join(config.AssetPath, "handles")
	if mkdirErr := os.MkdirAll(handlesDir, 0755); mkdirErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("创建目录失败: %v", mkdirErr),
		})
		return
	}

	// 生成唯一文件名
	fileID := uuid.New().String()
	filename := filepath.Base(header.Filename)
	ext := filepath.Ext(filename)
	originalFilename := fmt.Sprintf("original_%s%s", fileID, ext)
	originalPath := filepath.Join(handlesDir, originalFilename)

	// 保存上传的原始图片
	originalFile, err := os.Create(originalPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("创建文件失败: %v", err),
		})
		return
	}
	defer originalFile.Close()

	// 将上传的文件内容写入磁盘
	_, err = io.Copy(originalFile, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("保存上传文件失败: %v", err),
		})
		return
	}
	// 确保文件内容已写入
	originalFile.Close()

	// 开始计时
	startTime := time.Now()

	// 调用图像增强功能
	result, err := utils.EnhanceImage(originalPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("图像处理失败: %v", err),
		})
		return
	}

	// 结束计时
	duration := time.Since(startTime)

	// 如果处理失败
	if !result.Success {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": result.ErrorMessage,
		})
		return
	}

	// 构造处理后的图片保存路径
	enhancedFilename := fmt.Sprintf("enhanced_%s%s", fileID, ext)
	enhancedPath := filepath.Join(handlesDir, enhancedFilename)

	// 下载并保存增强后的图片
	err = downloadEnhancedImage(result.EnhancedImage, enhancedPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("保存增强图片失败: %v", err),
		})
		return
	}

	// 构造URL和相对路径（用于前端显示）
	baseURL := c.GetHeader("Origin")

	originalRelPath := filepath.Join("assets", "handles", originalFilename)
	enhancedRelPath := filepath.Join("assets", "handles", enhancedFilename)

	originalURL := fmt.Sprintf("%s/%s", baseURL, originalRelPath)
	enhancedURL := fmt.Sprintf("%s/%s", baseURL, enhancedRelPath)

	// 构造响应
	response := ImageEnhanceResponse{
		Success:        true,
		OriginalURL:    originalURL,
		EnhancedURL:    enhancedURL,
		OriginalPath:   originalRelPath,
		EnhancedPath:   enhancedRelPath,
		ProcessingTime: duration.String(),
	}

	c.JSON(http.StatusOK, response)
}

// 下载增强后的图片并保存到本地
func downloadEnhancedImage(url, savePath string) error {
	// 创建HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("请求图片失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求图片返回非200状态码: %d", resp.StatusCode)
	}

	// 创建输出文件
	out, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer out.Close()

	// 将响应内容写入文件
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("写入图片数据失败: %v", err)
	}

	return nil
}
