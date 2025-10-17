package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// 缓存结构
type cacheItem struct {
	data      interface{}
	timestamp time.Time
}

// 内存缓存
var (
	// 季节缓存
	seasonDetailsCache sync.Map
	// 剧照缓存
	episodeImagesCache sync.Map
	// 演员信息缓存
	episodeCreditsCache sync.Map
	// 缓存过期时间
	cacheTTL = 1 * time.Hour
)

// Season 季节信息结构体
type Season struct {
	ID           int    `json:"id"`
	SeasonNumber int    `json:"season_number"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	AirDate      string `json:"air_date"`
	EpisodeCount int    `json:"episode_count"`
	PosterPath   string `json:"poster_path"`
}

// Episode 剧集信息结构体
type Episode struct {
	ID            int     `json:"id"`
	EpisodeNumber int     `json:"episode_number"`
	Name          string  `json:"name"`
	Overview      string  `json:"overview"`
	StillPath     string  `json:"still_path"`
	AirDate       string  `json:"air_date"`
	SeasonNumber  int     `json:"season_number"`
	Runtime       int     `json:"runtime"`
	ShowID        int     `json:"show_id"`
	VoteAverage   float64 `json:"vote_average"`
}

// EpisodeImage 剧集图片信息
type EpisodeImage struct {
	AspectRatio float64 `json:"aspect_ratio"`
	Height      int     `json:"height"`
	Width       int     `json:"width"`
	FilePath    string  `json:"file_path"`
	VoteAverage float64 `json:"vote_average"`
	VoteCount   int     `json:"vote_count"`
}

// EpisodeImagesResponse 剧集图片响应
type EpisodeImagesResponse struct {
	ID     int            `json:"id"`
	Stills []EpisodeImage `json:"stills"`
}

// Actor 演员信息结构体
type Actor struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Character    string `json:"character"`
	ProfilePath  string `json:"profile_path,omitempty"`
	Gender       int    `json:"gender,omitempty"`
	Popularity   float64 `json:"popularity,omitempty"`
}

// CreditsResponse 演员信息响应
type CreditsResponse struct {
	ID         int            `json:"id"`
	Cast       []Actor        `json:"cast"`
	Crew       []json.RawMessage `json:"crew"`
	GuestStars []Actor        `json:"guest_stars,omitempty"`
}

// AnimeInfo 动漫详情和季节信息
type AnimeInfo struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Seasons []Season `json:"seasons"`
}

// SeasonDetailsResponse 季详情响应
type SeasonDetailsResponse struct {
	ID       int       `json:"id"`
	AirDate  string    `json:"air_date"`
	Name     string    `json:"name"`
	Overview string    `json:"overview"`
	Episodes []Episode `json:"episodes"`
}

// GetAnimeSeasons 获取动漫的所有季信息
func GetAnimeSeasons(seriesID int) (*AnimeInfo, error) {
	// 构建URL，使用GetTMDBAPIKey()获取API密钥
	requestURL := fmt.Sprintf("%s/tv/%d?api_key=%s&language=%s", BASE_URL, seriesID, GetTMDBAPIKey(), "zh-CN")
	
	// 发送请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("获取动漫季节信息失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应
	var animeInfo AnimeInfo
	err = json.NewDecoder(resp.Body).Decode(&animeInfo)
	if err != nil {
		return nil, fmt.Errorf("解析动漫季节信息失败: %w", err)
	}
	
	return &animeInfo, nil
}

// GetEpisodeDetails 获取某季的所有集详情
func GetEpisodeDetails(seriesID int, seasonNumber int) (*SeasonDetailsResponse, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("season_%d_%d", seriesID, seasonNumber)
	
	// 从缓存中获取
	if cachedItem, found := seasonDetailsCache.Load(cacheKey); found {
		if item, ok := cachedItem.(cacheItem); ok {
			// 检查缓存是否过期
			if time.Since(item.timestamp) < cacheTTL {
				return item.data.(*SeasonDetailsResponse), nil
			}
		}
	}
	
	// 构建URL，使用GetTMDBAPIKey()获取API密钥
	requestURL := fmt.Sprintf("%s/tv/%d/season/%d?api_key=%s&language=%s", BASE_URL, seriesID, seasonNumber, GetTMDBAPIKey(), "zh-CN")
	
	// 发送请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("获取季详情失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应
	var seasonDetails SeasonDetailsResponse
	err = json.NewDecoder(resp.Body).Decode(&seasonDetails)
	if err != nil {
		return nil, fmt.Errorf("解析季详情失败: %w", err)
	}
	
	// 如果季节概要为空，尝试获取英文季节概要
	needEnglishData := seasonDetails.Overview == ""
	
	// 检查是否有剧集缺少概要
	for _, episode := range seasonDetails.Episodes {
		if episode.Overview == "" {
			needEnglishData = true
			break
		}
	}
	
	// 如果需要英文数据，获取英文版本
	if needEnglishData {
		englishDetails, err := getEnglishEpisodeDetails(seriesID, seasonNumber)
		if err == nil {
			// 如果季节概要为空，使用英文版
			if seasonDetails.Overview == "" {
				seasonDetails.Overview = englishDetails.Overview
			}
			
			// 为每个缺少概要的剧集填充英文概要
			for i := range seasonDetails.Episodes {
				if seasonDetails.Episodes[i].Overview == "" {
					// 在英文版本中查找对应的集
					for _, engEpisode := range englishDetails.Episodes {
						if engEpisode.EpisodeNumber == seasonDetails.Episodes[i].EpisodeNumber {
							seasonDetails.Episodes[i].Overview = engEpisode.Overview
							break
						}
					}
				}
			}
		}
	}
	
	// 保存到缓存
	seasonDetailsCache.Store(cacheKey, cacheItem{
		data:      &seasonDetails,
		timestamp: time.Now(),
	})
	
	return &seasonDetails, nil
}

// getEnglishEpisodeDetails 获取英文版剧集详情，用于中文版没有概要时
func getEnglishEpisodeDetails(seriesID int, seasonNumber int) (SeasonDetailsResponse, error) {
	// 构建URL，使用GetTMDBAPIKey()获取API密钥
	requestURL := fmt.Sprintf("%s/tv/%d/season/%d?api_key=%s&language=%s", BASE_URL, seriesID, seasonNumber, GetTMDBAPIKey(), "en-US")
	
	// 发送请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return SeasonDetailsResponse{}, fmt.Errorf("获取英文季详情失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return SeasonDetailsResponse{}, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应
	var seasonDetails SeasonDetailsResponse
	err = json.NewDecoder(resp.Body).Decode(&seasonDetails)
	if err != nil {
		return SeasonDetailsResponse{}, fmt.Errorf("解析英文季详情失败: %w", err)
	}
	
	return seasonDetails, nil
}

// GetEpisodeImages 获取某集的剧照列表
func GetEpisodeImages(seriesID int, seasonNumber int, episodeNumber int) ([]string, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("images_%d_%d_%d", seriesID, seasonNumber, episodeNumber)
	
	// 从缓存中获取
	if cachedItem, found := episodeImagesCache.Load(cacheKey); found {
		if item, ok := cachedItem.(cacheItem); ok {
			// 检查缓存是否过期
			if time.Since(item.timestamp) < cacheTTL {
				return item.data.([]string), nil
			}
		}
	}
	
	// 构建URL，使用GetTMDBAPIKey()获取API密钥
	requestURL := fmt.Sprintf("%s/tv/%d/season/%d/episode/%d/images?api_key=%s", BASE_URL, seriesID, seasonNumber, episodeNumber, GetTMDBAPIKey())
	
	// 发送请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("获取集剧照失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应
	var imagesResp EpisodeImagesResponse
	err = json.NewDecoder(resp.Body).Decode(&imagesResp)
	if err != nil {
		return nil, fmt.Errorf("解析剧照数据失败: %w", err)
	}
	
	// 处理图片URL
	imageURLs := make([]string, 0, len(imagesResp.Stills))
	for _, still := range imagesResp.Stills {
		imageURL := fmt.Sprintf("%s/%s%s", IMAGE_BASE_URL, "w780", still.FilePath)
		imageURLs = append(imageURLs, imageURL)
	}
	
	// 保存到缓存
	episodeImagesCache.Store(cacheKey, cacheItem{
		data:      imageURLs,
		timestamp: time.Now(),
	})
	
	return imageURLs, nil
}

// GetEpisodeCredits 获取演员信息
func GetEpisodeCredits(seriesID int, seasonNumber int, episodeNumber int) (*CreditsResponse, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("credits_%d_%d_%d", seriesID, seasonNumber, episodeNumber)
	
	// 从缓存中获取
	if cachedItem, found := episodeCreditsCache.Load(cacheKey); found {
		if item, ok := cachedItem.(cacheItem); ok {
			// 检查缓存是否过期
			if time.Since(item.timestamp) < cacheTTL {
				return item.data.(*CreditsResponse), nil
			}
		}
	}
	
	// 构建URL，使用GetTMDBAPIKey()获取API密钥
	requestURL := fmt.Sprintf("%s/tv/%d/season/%d/episode/%d/credits?api_key=%s&language=%s", BASE_URL, seriesID, seasonNumber, episodeNumber, GetTMDBAPIKey(), "zh-CN")
	
	// 发送请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("获取演员信息失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应
	var creditsResp CreditsResponse
	err = json.NewDecoder(resp.Body).Decode(&creditsResp)
	if err != nil {
		return nil, fmt.Errorf("解析演员信息失败: %w", err)
	}
	
	// 保存到缓存
	episodeCreditsCache.Store(cacheKey, cacheItem{
		data:      &creditsResp,
		timestamp: time.Now(),
	})
	
	return &creditsResp, nil
} 