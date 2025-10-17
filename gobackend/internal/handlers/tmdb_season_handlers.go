package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"GoComicMosaic-gobackend/gobackend/internal/models"
	"GoComicMosaic-gobackend/gobackend/internal/utils"
)

// GetTMDBSeasons 获取动漫季节信息
// @Summary 获取动漫的所有季信息
// @Description 根据TMDB ID获取动漫的所有季信息
// @Tags TMDB
// @Accept json
// @Produce json
// @Param series_id path int true "TMDB 系列ID"
// @Success 200 {object} utils.AnimeInfo
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/tmdb/seasons/{series_id} [get]
func GetTMDBSeasons(c *gin.Context) {
	// 获取路径参数
	seriesIDStr := c.Param("series_id")
	seriesID, err := strconv.Atoi(seriesIDStr)
	if err != nil {
		log.Printf("解析系列ID失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的系列ID"})
		return
	}

	// 获取季节信息
	animeInfo, err := utils.GetAnimeSeasons(seriesID)
	if err != nil {
		log.Printf("获取季节信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, animeInfo)
}

// GetTMDBEpisodes 获取季的所有集详情
// @Summary 获取某季的所有集详情
// @Description 根据TMDB ID和季号获取季的所有集详情
// @Tags TMDB
// @Accept json
// @Produce json
// @Param series_id path int true "TMDB 系列ID"
// @Param season_number path int true "季号"
// @Success 200 {object} utils.SeasonDetailsResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/tmdb/seasons/{series_id}/{season_number} [get]
func GetTMDBEpisodes(c *gin.Context) {
	// 获取路径参数
	seriesIDStr := c.Param("series_id")
	seriesID, err := strconv.Atoi(seriesIDStr)
	if err != nil {
		log.Printf("解析系列ID失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的系列ID"})
		return
	}

	seasonNumberStr := c.Param("season_number")
	seasonNumber, err := strconv.Atoi(seasonNumberStr)
	if err != nil {
		log.Printf("解析季号失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的季号"})
		return
	}

	// 获取集详情
	seasonDetails, err := utils.GetEpisodeDetails(seriesID, seasonNumber)
	if err != nil {
		log.Printf("获取集详情失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, seasonDetails)
}

// GetTMDBEpisodeImages 获取某集的剧照
// @Summary 获取某集的所有剧照
// @Description 根据TMDB ID、季号和集号获取剧集的剧照
// @Tags TMDB
// @Accept json
// @Produce json
// @Param series_id path int true "TMDB 系列ID"
// @Param season_number path int true "季号"
// @Param episode_number path int true "集号"
// @Success 200 {array} string
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/tmdb/seasons/{series_id}/{season_number}/{episode_number}/images [get]
func GetTMDBEpisodeImages(c *gin.Context) {
	// 获取路径参数
	seriesIDStr := c.Param("series_id")
	seriesID, err := strconv.Atoi(seriesIDStr)
	if err != nil {
		log.Printf("解析系列ID失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的系列ID"})
		return
	}

	seasonNumberStr := c.Param("season_number")
	seasonNumber, err := strconv.Atoi(seasonNumberStr)
	if err != nil {
		log.Printf("解析季号失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的季号"})
		return
	}

	episodeNumberStr := c.Param("episode_number")
	episodeNumber, err := strconv.Atoi(episodeNumberStr)
	if err != nil {
		log.Printf("解析集号失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的集号"})
		return
	}

	// 获取剧照
	imageURLs, err := utils.GetEpisodeImages(seriesID, seasonNumber, episodeNumber)
	if err != nil {
		log.Printf("获取剧照失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"images": imageURLs})
}

// GetTMDBEpisodeCredits 获取某集的演员信息
// @Summary 获取某集的演员信息
// @Description 根据TMDB ID、季号和集号获取剧集的演员信息
// @Tags TMDB
// @Accept json
// @Produce json
// @Param series_id path int true "TMDB 系列ID"
// @Param season_number path int true "季号"
// @Param episode_number path int true "集号"
// @Success 200 {object} utils.CreditsResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/tmdb/seasons/{series_id}/{season_number}/{episode_number}/credits [get]
func GetTMDBEpisodeCredits(c *gin.Context) {
	// 获取路径参数
	seriesIDStr := c.Param("series_id")
	seriesID, err := strconv.Atoi(seriesIDStr)
	if err != nil {
		log.Printf("解析系列ID失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的系列ID"})
		return
	}

	seasonNumberStr := c.Param("season_number")
	seasonNumber, err := strconv.Atoi(seasonNumberStr)
	if err != nil {
		log.Printf("解析季号失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的季号"})
		return
	}

	episodeNumberStr := c.Param("episode_number")
	episodeNumber, err := strconv.Atoi(episodeNumberStr)
	if err != nil {
		log.Printf("解析集号失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的集号"})
		return
	}

	// 获取演员信息
	credits, err := utils.GetEpisodeCredits(seriesID, seasonNumber, episodeNumber)
	if err != nil {
		log.Printf("获取演员信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, credits)
}

// GetEpisodeInfo 获取单集详细信息
// @Summary 获取单集的详细信息(包括剧照和演员)
// @Description 获取指定剧集的详细信息，包括基本信息、剧照和演员信息
// @Tags TMDB
// @Accept json
// @Produce json
// @Param series_id path int true "TMDB 系列ID"
// @Param season_number path int true "季号"
// @Param episode_number path int true "集号"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/tmdb/episode/{series_id}/{season_number}/{episode_number} [get]
func GetEpisodeInfo(c *gin.Context) {
	// 获取路径参数
	seriesIDStr := c.Param("series_id")
	seriesID, err := strconv.Atoi(seriesIDStr)
	if err != nil {
		log.Printf("解析系列ID失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的系列ID"})
		return
	}

	seasonNumberStr := c.Param("season_number")
	seasonNumber, err := strconv.Atoi(seasonNumberStr)
	if err != nil {
		log.Printf("解析季号失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的季号"})
		return
	}

	episodeNumberStr := c.Param("episode_number")
	episodeNumber, err := strconv.Atoi(episodeNumberStr)
	if err != nil {
		log.Printf("解析集号失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的集号"})
		return
	}

	// 使用 goroutine 并发获取数据
	type episodeResult struct {
		Episode    *utils.Episode
		SeasonName string
		SeasonID   int
		Err        error
	}

	type imagesResult struct {
		Images []string
		Err    error
	}

	type creditsResult struct {
		Credits *utils.CreditsResponse
		Err     error
	}

	// 创建通道接收结果
	episodeChan := make(chan episodeResult, 1)
	imagesChan := make(chan imagesResult, 1)
	creditsChan := make(chan creditsResult, 1)

	// 并发获取剧集详情
	go func() {
		seasonDetails, err := utils.GetEpisodeDetails(seriesID, seasonNumber)
		if err != nil {
			episodeChan <- episodeResult{Err: err}
			return
		}

		// 找到对应的集
		var targetEpisode *utils.Episode
		for i := range seasonDetails.Episodes {
			if seasonDetails.Episodes[i].EpisodeNumber == episodeNumber {
				targetEpisode = &seasonDetails.Episodes[i]
				break
			}
		}

		if targetEpisode == nil {
			episodeChan <- episodeResult{Err: fmt.Errorf("未找到对应的集: 第%d季第%d集", seasonNumber, episodeNumber)}
			return
		}

		episodeChan <- episodeResult{
			Episode:    targetEpisode,
			SeasonName: seasonDetails.Name,
			SeasonID:   seasonDetails.ID,
			Err:        nil,
		}
	}()

	// 并发获取剧照
	go func() {
		images, err := utils.GetEpisodeImages(seriesID, seasonNumber, episodeNumber)
		if err != nil {
			imagesChan <- imagesResult{Err: err}
			return
		}
		imagesChan <- imagesResult{Images: images, Err: nil}
	}()

	// 并发获取演员信息
	go func() {
		credits, err := utils.GetEpisodeCredits(seriesID, seasonNumber, episodeNumber)
		if err != nil {
			creditsChan <- creditsResult{Err: err}
			return
		}
		creditsChan <- creditsResult{Credits: credits, Err: nil}
	}()

	// 等待所有goroutine完成并获取结果
	episodeRes := <-episodeChan
	imagesRes := <-imagesChan
	creditsRes := <-creditsChan

	// 处理剧集详情错误
	if episodeRes.Err != nil {
		log.Printf("获取剧集详情失败: %v", episodeRes.Err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": episodeRes.Err.Error()})
		return
	}

	// 组装返回数据
	var images []string
	if imagesRes.Err != nil {
		log.Printf("获取剧照失败: %v", imagesRes.Err)
		images = []string{} // 如果获取图片失败，返回空数组
	} else {
		images = imagesRes.Images
	}

	var castResult []utils.Actor
	var guestStarsResult []utils.Actor

	if creditsRes.Err != nil {
		log.Printf("获取演员信息失败: %v", creditsRes.Err)
		castResult = []utils.Actor{}
		guestStarsResult = []utils.Actor{}
	} else if creditsRes.Credits != nil {
		castResult = creditsRes.Credits.Cast
		guestStarsResult = creditsRes.Credits.GuestStars
	}

	result := gin.H{
		"episode":     episodeRes.Episode,
		"images":      images,
		"cast":        castResult,
		"guest_stars": guestStarsResult,
		"season_name": episodeRes.SeasonName,
		"season_id":   episodeRes.SeasonID,
	}

	c.JSON(http.StatusOK, result)
}

// GetResourceByTMDBID 通过TMDB ID查找本地资源
// @Summary 通过TMDB ID查找本地资源
// @Description 根据TMDB ID查找本地已导入的资源
// @Tags TMDB
// @Accept json
// @Produce json
// @Param tmdb_id path int true "TMDB ID"
// @Success 200 {object} models.Resource
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/tmdb/resource/{tmdb_id} [get]
func GetResourceByTMDBID(c *gin.Context) {
	// 获取路径参数
	tmdbIDStr := c.Param("tmdb_id")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		log.Printf("解析TMDB ID失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的TMDB ID"})
		return
	}

	// 查询数据库
	var resource models.Resource
	err = models.DB.Get(&resource, `
		SELECT * FROM resources WHERE tmdb_id = ? LIMIT 1
	`, tmdbID)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到对应的资源"})
			return
		}
		log.Printf("查询资源失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询资源失败"})
		return
	}

	c.JSON(http.StatusOK, resource)
}

// GetBatchEpisodeInfo 批量获取多个剧集详细信息
// @Summary 批量获取多个剧集的详细信息
// @Description 批量获取多个剧集的详细信息，包括基本信息、剧照和演员信息
// @Tags TMDB
// @Accept json
// @Produce json
// @Param request body BatchEpisodeRequest true "批量请求详情"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/tmdb/episodes/batch [post]
func GetBatchEpisodeInfo(c *gin.Context) {
	// 批量请求结构
	type EpisodeRequest struct {
		SeriesID      int `json:"series_id"`
		SeasonNumber  int `json:"season_number"`
		EpisodeNumber int `json:"episode_number"`
	}

	type BatchEpisodeRequest struct {
		Episodes []EpisodeRequest `json:"episodes" binding:"required"`
	}

	// 解析请求体
	var req BatchEpisodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("解析批量请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 检查请求参数
	if len(req.Episodes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供剧集请求"})
		return
	}

	// 限制批量请求数量，避免过大的请求
	if len(req.Episodes) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "批量请求剧集数量不能超过10个"})
		return
	}

	// 批量结果结构
	type EpisodeData struct {
		Episode    *utils.Episode `json:"episode"`
		Images     []string       `json:"images"`
		Cast       []utils.Actor  `json:"cast"`
		GuestStars []utils.Actor  `json:"guest_stars"`
		SeasonName string         `json:"season_name"`
		SeasonID   int            `json:"season_id"`
	}

	// 创建结果映射
	results := make(map[string]EpisodeData)

	// 使用goroutine并发处理每个请求
	type resultWithKey struct {
		Key   string
		Data  EpisodeData
		Error error
	}

	// 创建通道接收结果，缓冲大小为请求数量
	resultChan := make(chan resultWithKey, len(req.Episodes))

	// 处理每个剧集请求
	for _, episodeReq := range req.Episodes {
		// 为每个请求创建一个goroutine
		go func(er EpisodeRequest) {
			// 生成结果键
			resultKey := fmt.Sprintf("%d_%d_%d", er.SeriesID, er.SeasonNumber, er.EpisodeNumber)

			// 获取季节详情
			seasonDetails, err := utils.GetEpisodeDetails(er.SeriesID, er.SeasonNumber)
			if err != nil {
				resultChan <- resultWithKey{Key: resultKey, Error: fmt.Errorf("获取剧集详情失败: %w", err)}
				return
			}

			// 找到对应的集
			var targetEpisode *utils.Episode
			for i := range seasonDetails.Episodes {
				if seasonDetails.Episodes[i].EpisodeNumber == er.EpisodeNumber {
					targetEpisode = &seasonDetails.Episodes[i]
					break
				}
			}

			if targetEpisode == nil {
				resultChan <- resultWithKey{Key: resultKey, Error: fmt.Errorf("未找到对应的集: 第%d季第%d集", er.SeasonNumber, er.EpisodeNumber)}
				return
			}

			// 并发获取图片和演员信息
			type imagesResult struct {
				Images []string
				Err    error
			}

			type creditsResult struct {
				Credits *utils.CreditsResponse
				Err     error
			}

			// 创建通道
			imagesChan := make(chan imagesResult, 1)
			creditsChan := make(chan creditsResult, 1)

			// 获取图片
			go func() {
				images, err := utils.GetEpisodeImages(er.SeriesID, er.SeasonNumber, er.EpisodeNumber)
				if err != nil {
					imagesChan <- imagesResult{Images: []string{}, Err: err}
					return
				}
				imagesChan <- imagesResult{Images: images, Err: nil}
			}()

			// 获取演员信息
			go func() {
				credits, err := utils.GetEpisodeCredits(er.SeriesID, er.SeasonNumber, er.EpisodeNumber)
				if err != nil {
					creditsChan <- creditsResult{Credits: nil, Err: err}
					return
				}
				creditsChan <- creditsResult{Credits: credits, Err: nil}
			}()

			// 等待所有结果
			imagesRes := <-imagesChan
			creditsRes := <-creditsChan

			// 处理结果
			var images []string
			if imagesRes.Err != nil {
				log.Printf("获取剧照失败: %v", imagesRes.Err)
				images = []string{} // 如果获取图片失败，返回空数组
			} else {
				images = imagesRes.Images
			}

			var castResult []utils.Actor
			var guestStarsResult []utils.Actor

			if creditsRes.Err != nil {
				log.Printf("获取演员信息失败: %v", creditsRes.Err)
				castResult = []utils.Actor{}
				guestStarsResult = []utils.Actor{}
			} else if creditsRes.Credits != nil {
				castResult = creditsRes.Credits.Cast
				guestStarsResult = creditsRes.Credits.GuestStars
			}

			// 发送结果
			resultChan <- resultWithKey{
				Key: resultKey,
				Data: EpisodeData{
					Episode:    targetEpisode,
					Images:     images,
					Cast:       castResult,
					GuestStars: guestStarsResult,
					SeasonName: seasonDetails.Name,
					SeasonID:   seasonDetails.ID,
				},
				Error: nil,
			}
		}(episodeReq)
	}

	// 收集所有结果
	errors := make([]string, 0)
	for i := 0; i < len(req.Episodes); i++ {
		result := <-resultChan
		if result.Error != nil {
			errors = append(errors, result.Error.Error())
		} else {
			results[result.Key] = result.Data
		}
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"errors":  errors,
	})
}
