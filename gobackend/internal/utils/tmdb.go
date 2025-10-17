package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// TMDB API 配置
const (
	DEFAULT_TMDB_API_KEY = "" // 默认API密钥，仅作为回退使用
	BASE_URL     = "https://api.tmdb.org/3"
	POSTER_W     = "w500"
	BACKDROP_W   = "w1280"
	IMAGE_BASE_URL = "https://image.tmdb.org/t/p"
	TMDB_SETTINGS_KEY = "tmdb_config" // 管理后台配置键名
)

// 当前使用的TMDB API密钥
var currentTMDBAPIKey string

// 获取TMDB API密钥，按照优先级：
// 1. 使用已缓存的值（如果有）
// 2. 从环境变量获取
// 3. 使用默认值（如果前两者均未设置）
func GetTMDBAPIKey() string {
	// 如果已经有缓存的值，直接返回
	if currentTMDBAPIKey != "" {
		return currentTMDBAPIKey
	}
	
	// 从环境变量获取
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey != "" {
		currentTMDBAPIKey = apiKey
		return apiKey
	}
	
	// 使用默认值
	return DEFAULT_TMDB_API_KEY
}

// SetTMDBAPIKey 设置TMDB API密钥，可以由handlers包调用来更新密钥
func SetTMDBAPIKey(apiKey string) {
	currentTMDBAPIKey = apiKey
}

// TMDB中的类型ID映射到我们系统中的类型名称
var GENRES = map[int]string{
	// 原有映射
	16:    "幽默",
	35:    "讽刺",
	10759: "冒险",
	10765: "科幻",
	27:    "恐怖",
	80:    "犯罪",
	9648:  "悬疑",
	18:    "浪漫",
	
	// 添加电影常用分类
	12:    "冒险",  // Adventure
	28:    "动作",  // Action
	878:   "科幻",  // Science Fiction
	14:    "奇幻",  // Fantasy
	36:    "历史",  // History
	10751: "家庭",  // Family
	10749: "爱情",  // Romance
	53:    "惊悚",  // Thriller
	10752: "战争",  // War
	37:    "西部",  // Western
	99:    "纪录片", // Documentary
	10402: "音乐",  // Music
	10770: "电视电影", // TV Movie
	10762: "儿童",  // Kids
}

// TMDBSearchResult TMDB搜索结果
type TMDBSearchResult struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`           // 电视剧名称
	Title         string `json:"title"`          // 电影标题
	OriginalName  string `json:"original_name"`  // 电视剧原名
	OriginalTitle string `json:"original_title"` // 电影原标题
	Overview      string `json:"overview"`
}

// TMDBSearchResponse TMDB搜索响应
type TMDBSearchResponse struct {
	Page         int               `json:"page"`
	Results      []TMDBSearchResult `json:"results"`
	TotalPages   int               `json:"total_pages"`
	TotalResults int               `json:"total_results"`
}

// TMDBGenre TMDB类型
type TMDBGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TMDBDetails TMDB详情
type TMDBDetails struct {
	ID           int         `json:"id"`
	Name         string      `json:"name"`           // 电视剧名称
	Title        string      `json:"title"`          // 电影标题
	OriginalName string      `json:"original_name"`  // 电视剧原名
	OriginalTitle string     `json:"original_title"` // 电影原标题
	Overview     string      `json:"overview"`
	Genres       []TMDBGenre `json:"genres"`
}

// TMDBImage 图片信息
type TMDBImage struct {
	FilePath  string  `json:"file_path"`
	VoteCount int     `json:"vote_count"`
}

// TMDBImageResponse 图片响应
type TMDBImageResponse struct {
	ID        int         `json:"id"`
	Backdrops []TMDBImage `json:"backdrops"`
	Posters   []TMDBImage `json:"posters"`
}

// TMDBResource 最终整合的TMDB结果，用于替代models.ResourceCreate
type TMDBResource struct {
	ID          int       `json:"id"`            // TMDB ID
	Title       string    `json:"title"`         // 标题（电影Title或电视剧Name）
	TitleEn     string    `json:"title_en"`      // 英文标题（电影OriginalTitle或电视剧OriginalName）
	Description string    `json:"description"`   // 描述
	ResourceType string   `json:"resource_type"` // 资源类型，逗号分隔的类型名称
	PosterImage string    `json:"poster_image"`  // 海报图片URL
	Images      []string  `json:"images"`        // 所有图片URL列表
	Links       map[string]interface{} `json:"links"` // 链接信息
	MediaType   string    `json:"media_type"`    // 媒体类型：movie, tv
}

// SearchAnime 搜索动画，返回动画ID
func SearchAnime(query string, language string) (int, error) {
	if language == "" {
		language = "zh-CN"
	}
	
	// URL编码查询参数
	encodedQuery := url.QueryEscape(query)
	
	// 构建URL，使用GetTMDBAPIKey()获取API密钥
	requestURL := fmt.Sprintf("%s/search/tv?api_key=%s&query=%s&language=%s", BASE_URL, GetTMDBAPIKey(), encodedQuery, language)
	
	// 创建HTTP客户端并设置超时
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	
	// 发送请求
	resp, err := client.Get(requestURL)
	if err != nil {
		return 0, fmt.Errorf("搜索失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应
	var searchResp TMDBSearchResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResp)
	if err != nil {
		return 0, fmt.Errorf("解析搜索结果失败: %w", err)
	}
	
	// 检查是否有结果
	if len(searchResp.Results) == 0 {
		return 0, errors.New("未找到匹配的动画")
	}
	
	// 返回第一个结果的ID
	return searchResp.Results[0].ID, nil
}

// GetAnimeDetails 获取动画详情
func GetAnimeDetails(animeID int) (TMDBDetails, error) {
	// 构建URL，使用GetTMDBAPIKey()获取API密钥
	requestURL := fmt.Sprintf("%s/tv/%d?api_key=%s&language=zh-CN", BASE_URL, animeID, GetTMDBAPIKey())
	
	// 发送请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return TMDBDetails{}, fmt.Errorf("获取详情失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return TMDBDetails{}, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应
	var details TMDBDetails
	err = json.NewDecoder(resp.Body).Decode(&details)
	if err != nil {
		return TMDBDetails{}, fmt.Errorf("解析详情失败: %w", err)
	}
	
	return details, nil
}

// GetImages 获取海报和背景图片
func GetImages(animeID int) (string, []string, error) {
	// 构建URL，使用GetTMDBAPIKey()获取API密钥
	requestURL := fmt.Sprintf("%s/tv/%d/images?api_key=%s", BASE_URL, animeID, GetTMDBAPIKey())
	
	// 发送请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return "", nil, fmt.Errorf("获取图片失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应
	var imageResp TMDBImageResponse
	err = json.NewDecoder(resp.Body).Decode(&imageResp)
	if err != nil {
		return "", nil, fmt.Errorf("解析图片数据失败: %w", err)
	}
	
	// 处理海报图片
	var posterURL string
	if len(imageResp.Posters) > 0 {
		// 按投票数排序
		sortedPosters := imageResp.Posters
		// 简单的冒泡排序
		for i := 0; i < len(sortedPosters)-1; i++ {
			for j := 0; j < len(sortedPosters)-i-1; j++ {
				if sortedPosters[j].VoteCount < sortedPosters[j+1].VoteCount {
					sortedPosters[j], sortedPosters[j+1] = sortedPosters[j+1], sortedPosters[j]
				}
			}
		}
		posterURL = fmt.Sprintf("%s/%s%s", IMAGE_BASE_URL, POSTER_W, sortedPosters[0].FilePath)
	}
	
	// 处理背景图片
	var backdropURLs []string
	if len(imageResp.Backdrops) > 0 {
		// 按投票数排序
		sortedBackdrops := imageResp.Backdrops
		// 简单的冒泡排序
		for i := 0; i < len(sortedBackdrops)-1; i++ {
			for j := 0; j < len(sortedBackdrops)-i-1; j++ {
				if sortedBackdrops[j].VoteCount < sortedBackdrops[j+1].VoteCount {
					sortedBackdrops[j], sortedBackdrops[j+1] = sortedBackdrops[j+1], sortedBackdrops[j]
				}
			}
		}
		
		// 取前10张背景图片
		count := len(sortedBackdrops)
		if count > 10 {
			count = 10
		}
		
		backdropURLs = make([]string, count+1) // +1 是为了加入海报图片
		backdropURLs[0] = posterURL // 将海报图片加入到第一位
		
		for i := 0; i < count; i++ {
			backdropURLs[i+1] = fmt.Sprintf("%s/%s%s", IMAGE_BASE_URL, BACKDROP_W, sortedBackdrops[i].FilePath)
		}
	} else {
		// 如果没有背景图片，至少返回海报图片
		backdropURLs = []string{posterURL}
	}
	
	return posterURL, backdropURLs, nil
}

// SearchMovie 搜索电影，返回电影ID
func SearchMovie(query string, language string) (int, error) {
	if language == "" {
		language = "zh-CN"
	}
	
	// URL编码查询参数
	encodedQuery := url.QueryEscape(query)
	
	// 构建URL，使用GetTMDBAPIKey()获取API密钥
	requestURL := fmt.Sprintf("%s/search/movie?api_key=%s&query=%s&language=%s", BASE_URL, GetTMDBAPIKey(), encodedQuery, language)
	
	// 创建HTTP客户端并设置超时
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	
	// 发送请求
	resp, err := client.Get(requestURL)
	if err != nil {
		return 0, fmt.Errorf("搜索失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应
	var searchResp TMDBSearchResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResp)
	if err != nil {
		return 0, fmt.Errorf("解析搜索结果失败: %w", err)
	}
	
	// 检查是否有结果
	if len(searchResp.Results) == 0 {
		return 0, errors.New("未找到匹配的电影")
	}
	
	// 返回第一个结果的ID
	return searchResp.Results[0].ID, nil
}

// 获取电影详情
func GetMovieDetails(movieID int) (TMDBDetails, error) {
	// 构建URL，使用GetTMDBAPIKey()获取API密钥
	requestURL := fmt.Sprintf("%s/movie/%d?api_key=%s&language=zh-CN", BASE_URL, movieID, GetTMDBAPIKey())
	
	// 发送请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return TMDBDetails{}, fmt.Errorf("获取详情失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return TMDBDetails{}, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应
	var details TMDBDetails
	err = json.NewDecoder(resp.Body).Decode(&details)
	if err != nil {
		return TMDBDetails{}, fmt.Errorf("解析详情失败: %w", err)
	}
	
	return details, nil
}

// 获取电影图片
func GetMovieImages(movieID int) (string, []string, error) {
	// 构建URL，使用GetTMDBAPIKey()获取API密钥
	requestURL := fmt.Sprintf("%s/movie/%d/images?api_key=%s", BASE_URL, movieID, GetTMDBAPIKey())
	
	// 发送请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return "", nil, fmt.Errorf("获取图片失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应
	var imageResp TMDBImageResponse
	err = json.NewDecoder(resp.Body).Decode(&imageResp)
	if err != nil {
		return "", nil, fmt.Errorf("解析图片数据失败: %w", err)
	}
	
	// 处理海报图片
	var posterURL string
	if len(imageResp.Posters) > 0 {
		// 按投票数排序
		sortedPosters := imageResp.Posters
		// 简单的冒泡排序
		for i := 0; i < len(sortedPosters)-1; i++ {
			for j := 0; j < len(sortedPosters)-i-1; j++ {
				if sortedPosters[j].VoteCount < sortedPosters[j+1].VoteCount {
					sortedPosters[j], sortedPosters[j+1] = sortedPosters[j+1], sortedPosters[j]
				}
			}
		}
		posterURL = fmt.Sprintf("%s/%s%s", IMAGE_BASE_URL, POSTER_W, sortedPosters[0].FilePath)
	}
	
	// 处理背景图片
	var backdropURLs []string
	if len(imageResp.Backdrops) > 0 {
		// 按投票数排序
		sortedBackdrops := imageResp.Backdrops
		// 简单的冒泡排序
		for i := 0; i < len(sortedBackdrops)-1; i++ {
			for j := 0; j < len(sortedBackdrops)-i-1; j++ {
				if sortedBackdrops[j].VoteCount < sortedBackdrops[j+1].VoteCount {
					sortedBackdrops[j], sortedBackdrops[j+1] = sortedBackdrops[j+1], sortedBackdrops[j]
				}
			}
		}
		
		// 取前10张背景图片
		count := len(sortedBackdrops)
		if count > 10 {
			count = 10
		}
		
		backdropURLs = make([]string, count+1) // +1 是为了加入海报图片
		backdropURLs[0] = posterURL // 将海报图片加入到第一位
		
		for i := 0; i < count; i++ {
			backdropURLs[i+1] = fmt.Sprintf("%s/%s%s", IMAGE_BASE_URL, BACKDROP_W, sortedBackdrops[i].FilePath)
		}
	} else {
		// 如果没有背景图片，至少返回海报图片
		backdropURLs = []string{posterURL}
	}
	
	return posterURL, backdropURLs, nil
}

// SearchTMDB 搜索TMDB并返回适合资源表结构的结果
func SearchTMDB(query string) (*TMDBResource, error) {
	// 尝试作为电影搜索
	movieID, movieErr := SearchMovie(query, "zh-CN")
	
	// 如果电影搜索失败，尝试作为电视剧搜索
	if movieErr != nil {
		animeID, err := SearchAnime(query, "zh-CN")
		if err != nil {
			return nil, fmt.Errorf("TMDB搜索失败: %w", err)
		}
		
		// 获取电视剧详情
		details, err := GetAnimeDetails(animeID)
		if err != nil {
			return nil, fmt.Errorf("获取TMDB详情失败: %w", err)
		}
		
		// 获取海报和背景图片
		posterURL, imageURLs, err := GetImages(animeID)
		if err != nil {
			return nil, fmt.Errorf("获取TMDB图片失败: %w", err)
		}
		
		// 处理类型
		var genres []string
		for _, genre := range details.Genres {
			if genreName, ok := GENRES[genre.ID]; ok {
				genres = append(genres, genreName)
			}
		}
		
		// 构建适合资源表的结构
		resource := &TMDBResource{
			ID:          animeID,
			Title:       details.Name,
			TitleEn:     details.OriginalName,
			Description: details.Overview,
			ResourceType: strings.Join(genres, ","),
			PosterImage: posterURL,
			Images:      imageURLs,
			Links:       map[string]interface{}{},
			MediaType:   "tv", // 电视剧类型
		}
		
		return resource, nil
	}
	
	// 电影搜索成功，获取电影详情
	details, err := GetMovieDetails(movieID)
	if err != nil {
		return nil, fmt.Errorf("获取电影详情失败: %w", err)
	}
	
	// 获取海报和背景图片
	posterURL, imageURLs, err := GetMovieImages(movieID)
	if err != nil {
		return nil, fmt.Errorf("获取电影图片失败: %w", err)
	}
	
	// 处理类型
	var genres []string
	for _, genre := range details.Genres {
		if genreName, ok := GENRES[genre.ID]; ok {
			genres = append(genres, genreName)
		}
	}
	
	// 构建适合资源表的结构
	resource := &TMDBResource{
		ID:          movieID,
		Title:       details.Title,               // 使用电影标题
		TitleEn:     details.OriginalTitle,       // 使用电影原标题
		Description: details.Overview,
		ResourceType: strings.Join(genres, ","),
		PosterImage: posterURL,
		Images:      imageURLs,
		Links:       map[string]interface{}{},
		MediaType:   "movie", // 电影类型
	}
	
	return resource, nil
}

// GetTMDBResource 直接获取TMDB资源，适用于调试
func GetTMDBResource(query string) (*TMDBResource, error) {
	// 使用上面的函数获取资源
	resource, err := SearchTMDB(query)
	if err != nil {
		return nil, err
	}
	
	return resource, nil
}

// GetTmdbIdByQuery 简单快速地获取TMDB ID（仅用于剧集探索）
func GetTmdbIdByQuery(query string) (int, error) {
	// 直接调用SearchAnime函数，仅获取ID
	animeID, err := SearchAnime(query, "zh-CN")
	if err != nil {
		return 0, fmt.Errorf("TMDB搜索ID失败: %w", err)
	}
	
	return animeID, nil
} 