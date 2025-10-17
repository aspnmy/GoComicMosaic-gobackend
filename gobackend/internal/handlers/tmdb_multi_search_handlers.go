package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"GoComicMosaic-gobackend/gobackend/internal/utils"
)

// MultiSearchTMDB 多类型搜索TMDB API
// @Summary 多类型搜索TMDB API
// @Description 根据查询字符串搜索TMDB API获取电影和电视剧信息
// @Tags TMDB
// @Accept json
// @Produce json
// @Param query query string true "搜索关键词"
// @Param page query int false "页码，默认为1"
// @Success 200 {object} utils.TMDBMultiSearchResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/tmdb/multi_search [get]
func MultiSearchTMDB(c *gin.Context) {
	// 从URL查询参数获取查询字符串
	query := c.Query("query")

	// 检查查询字符串是否为空
	if strings.TrimSpace(query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "查询字符串不能为空"})
		return
	}

	// 获取页码参数，默认为1
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// 使用TMDB工具搜索
	response, err := utils.MultiSearch(query, page)
	if err != nil {
		log.Printf("TMDB多类型搜索失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "TMDB多类型搜索失败"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetMediaDetails 获取媒体详情
// @Summary 获取媒体详情
// @Description 根据媒体类型和ID获取详细信息，包括海报和剧照
// @Tags TMDB
// @Accept json
// @Produce json
// @Param media_type path string true "媒体类型(movie或tv)"
// @Param media_id path int true "媒体ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/tmdb/details/{media_type}/{media_id} [get]
func GetMediaDetails(c *gin.Context) {
	// 获取路径参数
	mediaType := c.Param("media_type")
	mediaIDStr := c.Param("media_id")

	// 验证媒体类型
	if mediaType != "movie" && mediaType != "tv" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的媒体类型，必须是 movie 或 tv"})
		return
	}

	// 解析媒体ID
	mediaID, err := strconv.Atoi(mediaIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的媒体ID"})
		return
	}

	// 获取媒体详情
	details, err := utils.GetMediaDetails(mediaType, mediaID)
	if err != nil {
		log.Printf("获取媒体详情失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取媒体详情失败"})
		return
	}

	c.JSON(http.StatusOK, details)
}
