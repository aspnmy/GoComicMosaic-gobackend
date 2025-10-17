package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TMDBMultiSearchResult 表示Multi Search API的单个结果项
type TMDBMultiSearchResult struct {
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	ID               int     `json:"id"`
	Title            string  `json:"title,omitempty"`           // 电影标题
	OriginalTitle    string  `json:"original_title,omitempty"`  // 电影原始标题
	Name             string  `json:"name,omitempty"`            // 电视剧标题
	OriginalName     string  `json:"original_name,omitempty"`   // 电视剧原始标题
	Overview         string  `json:"overview"`
	PosterPath       string  `json:"poster_path"`
	MediaType        string  `json:"media_type"`                // movie 或 tv
	OriginalLanguage string  `json:"original_language"`
	GenreIDs         []int   `json:"genre_ids"`
	Popularity       float64 `json:"popularity"`
	ReleaseDate      string  `json:"release_date,omitempty"`    // 电影发行日期
	FirstAirDate     string  `json:"first_air_date,omitempty"`  // 电视剧首播日期
	Video            bool    `json:"video,omitempty"`           // 仅电影有此字段
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	OriginCountry    []string `json:"origin_country,omitempty"` // 仅电视剧有此字段
}

// TMDBMultiSearchResponse 表示Multi Search API的响应
type TMDBMultiSearchResponse struct {
	Page         int                    `json:"page"`
	Results      []TMDBMultiSearchResult `json:"results"`
	TotalPages   int                    `json:"total_pages"`
	TotalResults int                    `json:"total_results"`
}

// MultiSearch 执行多类型搜索，调用TMDB API获取结果
// 参数:
// - query: 搜索关键词
// - page: 页码（默认为1）
// 返回:
// - *TMDBMultiSearchResponse: 搜索结果
// - error: 错误信息
func MultiSearch(query string, page ...int) (*TMDBMultiSearchResponse, error) {
	// URL编码查询参数
	encodedQuery := url.QueryEscape(query)
	
	// 处理页码参数
	pageNum := 1
	if len(page) > 0 && page[0] > 0 {
		pageNum = page[0]
	}
	
	// 构建URL，使用GetTMDBAPIKey()获取API密钥
	requestURL := fmt.Sprintf("%s/search/multi?api_key=%s&query=%s&language=zh-CN&page=%d", 
		BASE_URL, GetTMDBAPIKey(), encodedQuery, pageNum)
	
	// 创建HTTP客户端并设置超时
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	// 发送请求
	log.Printf("发送TMDB Multi Search请求: %s", requestURL)
	resp, err := client.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("TMDB API请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应
	var searchResp TMDBMultiSearchResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResp)
	if err != nil {
		return nil, fmt.Errorf("解析TMDB API响应失败: %w", err)
	}
	
	// 过滤结果，只保留电影和电视剧
	var filteredResults []TMDBMultiSearchResult
	for _, result := range searchResp.Results {
		if result.MediaType == "movie" || result.MediaType == "tv" {
			filteredResults = append(filteredResults, result)
		}
	}
	searchResp.Results = filteredResults
	
	log.Printf("TMDB Multi Search成功，找到 %d 个结果，当前页 %d，总页数 %d", 
		len(searchResp.Results), searchResp.Page, searchResp.TotalPages)
	return &searchResp, nil
}

// TMDBMediaDetails 表示媒体详情响应
type TMDBMediaDetails struct {
	ID            int         `json:"id"`
	Images        TMDBMediaImages  `json:"images"`
	Genres        []TMDBGenre `json:"genres"`
	// 其他字段根据需要添加
}

// TMDBMediaImages 表示媒体图片集合
type TMDBMediaImages struct {
	Backdrops []TMDBImage `json:"backdrops"`
	Posters   []TMDBImage `json:"posters"`
}

// GetMediaDetails 获取媒体详情
// 参数:
// - mediaType: 媒体类型 (movie 或 tv)
// - mediaID: 媒体ID
// 返回:
// - map[string]interface{}: 详情数据
// - error: 错误信息
func GetMediaDetails(mediaType string, mediaID int) (map[string]interface{}, error) {
	// 构建URL，使用正确的API接口路径
	requestURL := fmt.Sprintf("%s/%s/%d?api_key=%s&append_to_response=images", 
		BASE_URL, mediaType, mediaID, GetTMDBAPIKey())
	
	// 创建HTTP客户端并设置超时
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// 发送请求
	log.Printf("发送TMDB %s详情请求: %s", mediaType, requestURL)
	resp, err := client.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("TMDB API请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应
	var details TMDBMediaDetails
	err = json.NewDecoder(resp.Body).Decode(&details)
	if err != nil {
		return nil, fmt.Errorf("解析TMDB API响应失败: %w", err)
	}

	// 处理图片数据
	var imageURLs []string
	if len(details.Images.Backdrops) > 0 {
		// 按照投票数排序
		for i := 0; i < len(details.Images.Backdrops); i++ {
			for j := i + 1; j < len(details.Images.Backdrops); j++ {
				if details.Images.Backdrops[i].VoteCount < details.Images.Backdrops[j].VoteCount {
					details.Images.Backdrops[i], details.Images.Backdrops[j] = details.Images.Backdrops[j], details.Images.Backdrops[i]
				}
			}
		}
		
		// 最多取10张图片
		count := min(10, len(details.Images.Backdrops))
		for i := 0; i < count; i++ {
			imageURL := fmt.Sprintf("https://image.tmdb.org/t/p/w1280%s", details.Images.Backdrops[i].FilePath)
			imageURLs = append(imageURLs, imageURL)
		}
	}
	
	// 获取海报URL
	posterURL := ""
	if len(details.Images.Posters) > 0 {
		posterURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", details.Images.Posters[0].FilePath)
	}
	
	// 处理类型，转换为中文分类
	var genres []string
	// 创建中文化的genres数组，替换原有的英文genres
	var chineseGenres []map[string]interface{}
	
	for _, genre := range details.Genres {
		// 对于resource_type字段使用中文名称拼接
		if genreName, ok := GENRES[genre.ID]; ok {
			genres = append(genres, genreName)
			
			// 创建一个新的genre对象，但使用中文名称
			chineseGenre := map[string]interface{}{
				"id": genre.ID,
				"name": genreName,
			}
			chineseGenres = append(chineseGenres, chineseGenre)
		} else {
			// 如果在映射表中找不到对应的中文名称，保留原始英文名称
			chineseGenres = append(chineseGenres, map[string]interface{}{
				"id": genre.ID,
				"name": genre.Name,
			})
		}
	}
	
	// 构建返回结果
	result := map[string]interface{}{
		"id":       details.ID,
		"genres":   chineseGenres, // 使用中文化的genres数组
		"images":   imageURLs,
		"poster_path": posterURL,
		"media_type": mediaType,
		"resource_type": strings.Join(genres, ","), // 添加中文分类字段
	}
	
	// 根据媒体类型添加不同的字段
	if mediaType == "movie" {
		// 再次发起一个multi_search请求，获取中文标题和简介
		searchResp, err := MultiSearch(fmt.Sprintf("id:%d", mediaID), 1)
		if err == nil && len(searchResp.Results) > 0 {
			for _, item := range searchResp.Results {
				if item.MediaType == "movie" && item.ID == mediaID {
					result["title"] = item.Title
					result["original_title"] = item.OriginalTitle
					result["overview"] = item.Overview
					break
				}
			}
		} else {
			// 如果无法通过multi_search获取，直接从详情API解析
			var movieDetails struct {
				Title        string `json:"title"`
				OriginalTitle string `json:"original_title"`
				Overview     string `json:"overview"`
			}
			resp, err := client.Get(requestURL)
			if err == nil {
				defer resp.Body.Close()
				json.NewDecoder(resp.Body).Decode(&movieDetails)
				result["title"] = movieDetails.Title
				result["original_title"] = movieDetails.OriginalTitle
				result["overview"] = movieDetails.Overview
			}
		}
	} else if mediaType == "tv" {
		// 再次发起一个multi_search请求，获取中文标题和简介
		searchResp, err := MultiSearch(fmt.Sprintf("id:%d", mediaID), 1)
		if err == nil && len(searchResp.Results) > 0 {
			for _, item := range searchResp.Results {
				if item.MediaType == "tv" && item.ID == mediaID {
					result["name"] = item.Name
					result["original_name"] = item.OriginalName
					result["overview"] = item.Overview
					break
				}
			}
		} else {
			// 如果无法通过multi_search获取，直接从详情API解析
			var tvDetails struct {
				Name        string `json:"name"`
				OriginalName string `json:"original_name"`
				Overview     string `json:"overview"`
				Seasons      []interface{} `json:"seasons"`
			}
			resp, err := client.Get(requestURL)
			if err == nil {
				defer resp.Body.Close()
				json.NewDecoder(resp.Body).Decode(&tvDetails)
				result["name"] = tvDetails.Name
				result["original_name"] = tvDetails.OriginalName
				result["overview"] = tvDetails.Overview
				result["seasons"] = tvDetails.Seasons
			}
		}
	}

	return result, nil
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
} 