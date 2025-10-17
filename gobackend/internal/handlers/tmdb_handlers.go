package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aspnmy/GoComicMosaic-gobackend/gobackend/internal/models"
	"github.com/aspnmy/GoComicMosaic-gobackend/gobackend/internal/utils"
)

// TMDBSearchRequest TMDB搜索请求
type TMDBSearchRequest struct {
	Query string `json:"query" binding:"required"`
	// TMDB ID
	ID int `json:"id"`
	// 添加自定义字段，支持自定义资源创建
	Title        string                         `json:"title"`
	TitleEn      string                         `json:"title_en"`
	Description  string                         `json:"description"`
	ResourceType string                         `json:"resource_type"`
	PosterImage  string                         `json:"poster_image"`
	Images       []string                       `json:"images"`
	Links        map[string][]map[string]string `json:"links"`
	MediaType    string                         `json:"media_type"` // 媒体类型：movie, tv
	IsCustom     bool                           `json:"is_custom"`  // 标识是否为自定义资源
}

// SearchTMDB 搜索TMDB API
// @Summary 搜索TMDB API
// @Description 根据查询字符串搜索TMDB API获取动画信息
// @Tags TMDB
// @Accept json
// @Produce json
// @Param query query string true "搜索关键词"
// @Success 200 {object} utils.TMDBResource
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/tmdb/search [get]
func SearchTMDB(c *gin.Context) {
	// 从URL查询参数获取查询字符串
	query := c.Query("query")

	// 检查查询字符串是否为空
	if strings.TrimSpace(query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "查询字符串不能为空"})
		return
	}

	// 使用TMDB工具搜索
	resource, err := utils.SearchTMDB(query)
	if err != nil {
		log.Printf("TMDB搜索失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "TMDB搜索失败"})
		return
	}

	c.JSON(http.StatusOK, resource)
}

// CreateResourceFromTMDB 从TMDB搜索结果创建资源
// @Summary 从TMDB创建资源
// @Description 根据TMDB搜索结果或用户自定义内容创建新的资源
// @Tags TMDB
// @Accept json
// @Produce json
// @Param request body TMDBSearchRequest true "搜索请求"
// @Success 200 {object} models.Resource
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/tmdb/create [post]
func CreateResourceFromTMDB(c *gin.Context) {
	var req TMDBSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("解析请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 检查查询字符串是否为空
	if strings.TrimSpace(req.Query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "查询字符串不能为空"})
		return
	}

	// 转换为需要插入数据库的资源
	now := time.Now()
	defaultStatus := models.ResourceStatusPending
	var resource *models.Resource

	// 根据是否为自定义资源处理不同的逻辑
	if req.IsCustom {
		// 自定义资源处理逻辑
		if strings.TrimSpace(req.Title) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "自定义资源的标题不能为空"})
			return
		}

		// 确保PosterImage不为空
		var posterImage *string
		if req.PosterImage != "" {
			posterImage = &req.PosterImage
		} else if len(req.Images) > 0 {
			posterImage = &req.Images[0]
		}

		// 将 req.Links 转换为 models.JsonMap 类型
		linksMap := make(models.JsonMap)
		for key, value := range req.Links {
			linksMap[key] = value
		}

		// 处理TMDB ID
		var tmdbID *int
		if req.ID > 0 {
			tmdbID = &req.ID
		}

		// 处理媒体类型
		var mediaType *string
		if req.MediaType != "" {
			mediaType = &req.MediaType
		}

		// 创建自定义资源对象
		resource = &models.Resource{
			Title:        req.Title,
			TitleEn:      req.TitleEn,
			Description:  req.Description,
			ResourceType: req.ResourceType,
			PosterImage:  posterImage,
			Images:       req.Images,
			Links:        linksMap,
			Status:       defaultStatus,
			TmdbID:       tmdbID,
			MediaType:    mediaType,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	} else {
		// 标准TMDB资源处理逻辑 - 使用前端传递的数据而不是重新搜索
		log.Printf("处理TMDB资源导入请求: ID=%d, 标题=%s, 类型=%s",
			req.ID, req.Title, req.MediaType)

		// 验证必要字段
		if req.ID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "TMDB ID不能为空"})
			return
		}

		if req.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "资源标题不能为空"})
			return
		}

		// 确保PosterImage不为空
		var posterImage *string
		if req.PosterImage != "" {
			posterImage = &req.PosterImage
		} else if len(req.Images) > 0 {
			posterImage = &req.Images[0]
		}

		// 将 req.Links 转换为 models.JsonMap 类型
		linksMap := make(models.JsonMap)
		for key, value := range req.Links {
			linksMap[key] = value
		}

		// 处理TMDB ID
		tmdbID := req.ID

		// 处理媒体类型
		var mediaType *string
		if req.MediaType != "" {
			mediaType = &req.MediaType
		}

		// 创建资源对象 - 直接使用前端提供的数据
		resource = &models.Resource{
			Title:        req.Title,
			TitleEn:      req.TitleEn,
			Description:  req.Description,
			ResourceType: req.ResourceType,
			PosterImage:  posterImage,
			Images:       req.Images,
			Links:        linksMap,
			Status:       defaultStatus,
			TmdbID:       &tmdbID,
			MediaType:    mediaType,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	}

	// 执行SQL插入
	result, err := models.DB.NamedExec(`
		INSERT INTO resources (
			title, title_en, description, resource_type, poster_image, 
			images, links, status, tmdb_id, media_type, created_at, updated_at
		) VALUES (
			:title, :title_en, :description, :resource_type, :poster_image, 
			:images, :links, :status, :tmdb_id, :media_type, :created_at, :updated_at
		)
	`, resource)

	if err != nil {
		log.Printf("插入资源失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建资源失败"})
		return
	}

	// 获取新插入记录的ID
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("获取插入ID失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取新资源ID失败"})
		return
	}

	// 设置返回资源的ID
	resource.ID = int(id)

	c.JSON(http.StatusOK, resource)
}

// UpdateResourceTmdbID 更新资源的TMDB ID
// @Summary 更新资源的TMDB ID
// @Description 根据资源ID更新对应资源的TMDB ID
// @Tags TMDB
// @Accept json
// @Produce json
// @Param id path int true "资源ID"
// @Param tmdb_id path int true "TMDB ID"
// @Success 200 {object} models.Resource
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/tmdb/update-resource-id/{id}/{tmdb_id} [put]
func UpdateResourceTmdbID(c *gin.Context) {
	// 获取路径参数
	resourceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的资源ID"})
		return
	}

	tmdbID, err := strconv.Atoi(c.Param("tmdb_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的TMDB ID"})
		return
	}

	log.Printf("开始更新资源ID: %d 的TMDB ID为: %d", resourceID, tmdbID)

	// 检查资源是否存在
	var resource models.Resource
	err = models.DB.Get(&resource, `SELECT * FROM resources WHERE id = ?`, resourceID)
	if err != nil {
		log.Printf("无法找到资源: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "资源未找到"})
		return
	}

	log.Printf("已找到资源: %+v", resource)

	// 更新TMDB ID
	resource.TmdbID = &tmdbID
	resource.UpdatedAt = time.Now()

	// 更新记录
	_, err = models.DB.Exec(
		`UPDATE resources SET tmdb_id = ?, updated_at = ? WHERE id = ?`,
		resource.TmdbID, resource.UpdatedAt, resourceID,
	)

	if err != nil {
		log.Printf("更新资源失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新资源失败"})
		return
	}

	log.Printf("成功更新资源ID: %d 的TMDB ID为: %d", resourceID, tmdbID)

	// 返回更新后的资源
	c.JSON(http.StatusOK, resource)
}

// SearchTmdbId 仅搜索TMDB ID
// @Summary 仅搜索TMDB ID
// @Description 根据查询字符串搜索TMDB API获取ID，适用于剧集探索等场景
// @Tags TMDB
// @Accept json
// @Produce json
// @Param query query string true "搜索关键词"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/tmdb/search_id [get]
func SearchTmdbId(c *gin.Context) {
	// 从URL查询参数获取查询字符串
	query := c.Query("query")

	// 检查查询字符串是否为空
	if strings.TrimSpace(query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "查询字符串不能为空"})
		return
	}

	// 使用TMDB工具仅搜索ID
	id, err := utils.GetTmdbIdByQuery(query)
	if err != nil {
		log.Printf("TMDB ID搜索失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// CheckResourceExists 检查资源是否已存在
// @Summary 检查资源是否已存在
// @Description 根据TMDB ID或标题检查资源是否已存在
// @Tags TMDB
// @Accept json
// @Produce json
// @Param tmdb_id query int false "TMDB ID"
// @Param title query string false "资源标题"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/tmdb/check-exists [get]
func CheckResourceExists(c *gin.Context) {
	tmdbIDStr := c.Query("tmdb_id")
	title := c.Query("title")

	// 至少需要提供一个参数
	if tmdbIDStr == "" && title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "必须提供TMDB ID或标题"})
		return
	}

	var exists bool
	var existingResource models.Resource

	// 首先按TMDB ID查询
	if tmdbIDStr != "" {
		tmdbID, err := strconv.Atoi(tmdbIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的TMDB ID"})
			return
		}

		// 查询数据库
		err = models.DB.Get(&existingResource, `SELECT * FROM resources WHERE tmdb_id = ?`, tmdbID)
		if err == nil {
			// 找到资源
			exists = true
		}
	}

	// 如果按TMDB ID没找到，再按标题查询
	if !exists && title != "" {
		// 查询数据库
		err := models.DB.Get(&existingResource, `SELECT * FROM resources WHERE title = ? OR title_en = ?`, title, title)
		if err == nil {
			// 找到资源
			exists = true
		}
	}

	if exists {
		c.JSON(http.StatusOK, gin.H{
			"exists":   true,
			"resource": existingResource,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"exists": false,
		})
	}
}
