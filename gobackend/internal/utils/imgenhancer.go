package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// 基础请求头
var baseHeaders = map[string]string{
	"X-AppID":        "808645",
	"X-Platform":     "h5",
	"X-Version":      "8.9.7",
	"X-SessionToken": "",
	"X-UniqueID":     "",
	"X-DeviceID":     "db16163954eac9c5ef7e28f8934db9f4",
	"X-MCC":          "zh-TW",
	"X-GhostID":      "9558ca811dcc2cfce851d57a0f092a3a",
}

// 常量
const (
	sigPrefix = "XX"
	sigSecret = "Pg@photo_photogrid#20250225"
)

// GenerateHashedString 计算请求签名
func GenerateHashedString(headers map[string]string) string {
	// 按字母顺序排序 headers 的键
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接键值对
	var sb strings.Builder
	for _, key := range keys {
		sb.WriteString(key)
		sb.WriteString(headers[key])
	}

	// 拼接密钥
	sb.WriteString(sigSecret)
	concatenated := sb.String()

	// 尝试 SHA-256 哈希，如果失败则使用 MD5
	sha256Hash := sha256.Sum256([]byte(concatenated))
	return sigPrefix + hex.EncodeToString(sha256Hash[:])
}

// UploadURLResponse 获取上传URL的响应结构
type UploadURLResponse struct {
	Code    int    `json:"code"`
	Errmsg  string `json:"errmsg"`
	Data    *UploadURLData `json:"data,omitempty"`
}

// UploadURLData 获取上传URL的数据结构
type UploadURLData struct {
	ImgURL    string `json:"img_url"`
	UploadURL string `json:"upload_url"`
}

// GetUploadURL 获取上传图片的URL
func GetUploadURL() (*UploadURLResponse, error) {
	// 构造请求体
	body := map[string]string{
		"ext":    "png",
		"method": "wn_superresolution",
	}

	// 合并 headers 和 body 用于签名
	signHeaders := make(map[string]string)
	for k, v := range baseHeaders {
		signHeaders[k] = v
	}
	signHeaders["ext"] = body["ext"]
	signHeaders["method"] = body["method"]

	// 生成签名
	signature := GenerateHashedString(signHeaders)

	// 构造请求 headers，添加签名
	requestHeaders := make(map[string]string)
	for k, v := range baseHeaders {
		requestHeaders[k] = v
	}
	requestHeaders["sig"] = signature
	requestHeaders["Content-Type"] = "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW"

	// 构造 multipart/form-data 请求体
	var requestBody bytes.Buffer
	boundary := "----WebKitFormBoundary7MA4YWxkTrZu0gW"
	
	// 添加 ext 字段
	requestBody.WriteString("--" + boundary + "\r\n")
	requestBody.WriteString("Content-Disposition: form-data; name=\"ext\"\r\n\r\n")
	requestBody.WriteString(body["ext"] + "\r\n")
	
	// 添加 method 字段
	requestBody.WriteString("--" + boundary + "\r\n")
	requestBody.WriteString("Content-Disposition: form-data; name=\"method\"\r\n\r\n")
	requestBody.WriteString(body["method"] + "\r\n")
	
	// 结束边界
	requestBody.WriteString("--" + boundary + "--\r\n")

	// 创建请求
	url := "https://api.grid.plus/v1/ai/web/nologin/getuploadurl"
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	for k, v := range requestHeaders {
		req.Header.Set(k, v)
	}

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var response UploadURLResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &response, nil
}

// UploadImageResponse 上传图片的响应结构
type UploadImageResponse struct {
	Code       int    `json:"code"`
	Errmsg     string `json:"errmsg"`
	StatusCode int    `json:"status_code,omitempty"`
}

// UploadImage 上传图片到指定URL
func UploadImage(imagePath string, uploadURL string) (*UploadImageResponse, error) {
	// 读取图片文件
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return &UploadImageResponse{
			Code:   -1,
			Errmsg: fmt.Sprintf("读取图片文件失败: %v", err),
		}, nil
	}

	// 构造请求头
	headers := map[string]string{
		"Referer":        "https://www.photogrid.app/",
		"X-AppID":        "808645",
		"X-Platform":     "h5",
		"X-Version":      "8.9.7",
		"X-SessionToken": "",
		"X-UniqueID":     "",
		"X-DeviceID":     "db16163954eac9c5ef7e28f8934db9f4",
		"X-MCC":          "zh-TW",
		"X-GhostID":      "9558ca811dcc2cfce851d57a0f092a3a",
		"Content-Type":   "image/png",
	}

	// 创建请求
	req, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(imageData))
	if err != nil {
		return &UploadImageResponse{
			Code:   -1,
			Errmsg: fmt.Sprintf("创建请求失败: %v", err),
		}, nil
	}

	// 设置请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &UploadImageResponse{
			Code:   -1,
			Errmsg: fmt.Sprintf("上传失败: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	return &UploadImageResponse{
		Code:       0,
		Errmsg:     "",
		StatusCode: resp.StatusCode,
	}, nil
}

// TaskIDResponse 获取任务ID的响应结构
type TaskIDResponse struct {
	Code    int    `json:"code"`
	Errmsg  string `json:"errmsg"`
	TaskID  string `json:"task_id,omitempty"`
}

// GetTaskID 获取处理任务的ID
func GetTaskID(imageURL string) (*TaskIDResponse, error) {
	// 构造请求体
	body := map[string]string{
		"url":    imageURL,
		"method": "wn_superresolution",
	}

	// 合并 headers 和 body 用于签名
	signHeaders := make(map[string]string)
	for k, v := range baseHeaders {
		signHeaders[k] = v
	}
	signHeaders["url"] = body["url"]
	signHeaders["method"] = body["method"]

	// 生成签名
	signature := GenerateHashedString(signHeaders)

	// 构造请求 headers，添加签名
	requestHeaders := make(map[string]string)
	for k, v := range baseHeaders {
		requestHeaders[k] = v
	}
	requestHeaders["sig"] = signature
	requestHeaders["Referer"] = "https://www.photogrid.app/"
	requestHeaders["Content-Type"] = "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW"

	// 构造 multipart/form-data 请求体
	var requestBody bytes.Buffer
	boundary := "----WebKitFormBoundary7MA4YWxkTrZu0gW"
	
	// 添加 url 字段
	requestBody.WriteString("--" + boundary + "\r\n")
	requestBody.WriteString("Content-Disposition: form-data; name=\"url\"\r\n\r\n")
	requestBody.WriteString(body["url"] + "\r\n")
	
	// 添加 method 字段
	requestBody.WriteString("--" + boundary + "\r\n")
	requestBody.WriteString("Content-Disposition: form-data; name=\"method\"\r\n\r\n")
	requestBody.WriteString(body["method"] + "\r\n")
	
	// 结束边界
	requestBody.WriteString("--" + boundary + "--\r\n")

	// 创建请求
	url := "https://api.grid.plus/v1/ai/web/super_resolution/nologinupload"
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	for k, v := range requestHeaders {
		req.Header.Set(k, v)
	}

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var response TaskIDResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &response, nil
}

// TaskData 任务数据结构
type TaskData struct {
	Errmsg        string   `json:"errmsg"`
	ImageURL      []string `json:"image_url"`
	ShowWatermark bool     `json:"show_watermark"`
	SrcURL        []string `json:"src_url"`
	Status        int      `json:"status"`
	TaskID        string   `json:"task_id"`
	TimeEnd       int64    `json:"time_end"`
	TimeStart     int64    `json:"time_start"`
	Version       int      `json:"version"`
}

// TaskResultResponse 任务结果响应结构
type TaskResultResponse struct {
	Code    int        `json:"code"`
	Data    []TaskData `json:"data,omitempty"`
	Errmsg  string     `json:"errmsg"`
	ImageURLs []string `json:"image_urls,omitempty"`
}

// PollTaskResult 轮询处理结果
func PollTaskResult(taskIDs []string, pollInterval int, maxAttempts int) (*TaskResultResponse, error) {
	if pollInterval <= 0 {
		pollInterval = 2
	}
	if maxAttempts <= 0 {
		maxAttempts = 30
	}

	// 构造请求体
	bodyMap := map[string][]string{
		"task_ids": taskIDs,
	}
	
	// 将请求体序列化为JSON
	bodyJSON, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %v", err)
	}

	// 合并 headers 和 body 用于签名
	signHeaders := make(map[string]string)
	for k, v := range baseHeaders {
		signHeaders[k] = v
	}
	taskIDsJSON, _ := json.Marshal(taskIDs)
	signHeaders["task_ids"] = string(taskIDsJSON)

	// 生成签名
	signature := GenerateHashedString(signHeaders)

	// 构造请求 headers，添加签名
	requestHeaders := make(map[string]string)
	for k, v := range baseHeaders {
		requestHeaders[k] = v
	}
	requestHeaders["sig"] = signature
	requestHeaders["Content-Type"] = "application/json"

	// 轮询逻辑
	url := "https://api.grid.plus/v1/ai/web/super_resolution/nologinbatchresult"
	attempts := 0
	client := &http.Client{Timeout: 10 * time.Second}

	for attempts < maxAttempts {
		// 创建请求
		req, err := http.NewRequest("POST", url, bytes.NewReader(bodyJSON))
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %v", err)
		}

		// 设置请求头
		for k, v := range requestHeaders {
			req.Header.Set(k, v)
		}

		// 发送请求
		resp, err := client.Do(req)
		if err != nil {
			return &TaskResultResponse{
				Code:    -1,
				Errmsg:  fmt.Sprintf("请求失败: %v", err),
				ImageURLs: []string{},
			}, nil
		}

		// 读取响应
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return &TaskResultResponse{
				Code:    -1,
				Errmsg:  fmt.Sprintf("读取响应失败: %v", err),
				ImageURLs: []string{},
			}, nil
		}

		// 解析响应
		var response TaskResultResponse
		if err := json.Unmarshal(respBody, &response); err != nil {
			return &TaskResultResponse{
				Code:    -1,
				Errmsg:  fmt.Sprintf("解析响应失败: %v", err),
				ImageURLs: []string{},
			}, nil
		}

		// 检查 code 是否为 0
		if response.Code != 0 {
			return &TaskResultResponse{
				Code:    response.Code,
				Errmsg:  response.Errmsg,
				ImageURLs: []string{},
			}, nil
		}

		// 检查每个任务的 status
		data := response.Data
		imageURLs := []string{}
		allCompleted := true
		
		for _, task := range data {
			if task.Status == 2 {
				// 任务完成，收集 image_url
				if len(task.ImageURL) > 0 {
					imageURLs = append(imageURLs, task.ImageURL...)
				}
			} else {
				// 存在未完成任务，继续轮询
				allCompleted = false
			}
		}

		if allCompleted {
			response.ImageURLs = imageURLs
			return &response, nil
		}

		// 未全部完成，等待后重试
		attempts++
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}

	// 超时
	return &TaskResultResponse{
		Code:    -1,
		Errmsg:  fmt.Sprintf("轮询超时，未在 %d 秒内完成", maxAttempts*pollInterval),
		ImageURLs: []string{},
	}, nil
}

// SuperResolutionResult 超分辨率处理结果
type SuperResolutionResult struct {
	OriginalImage string
	EnhancedImage string
	Success       bool
	ErrorMessage  string
}

// EnhanceImage 完整的图片超分辨率处理流程
func EnhanceImage(imagePath string) (*SuperResolutionResult, error) {
	result := &SuperResolutionResult{
		OriginalImage: imagePath,
		Success:       false,
	}

	// 1. 获取上传URL
	uploadURLResp, err := GetUploadURL()
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("获取上传URL失败: %v", err)
		return result, err
	}
	if uploadURLResp.Code != 0 || uploadURLResp.Data == nil {
		result.ErrorMessage = fmt.Sprintf("获取上传URL失败: %s", uploadURLResp.Errmsg)
		return result, fmt.Errorf(result.ErrorMessage)
	}

	imgURL := uploadURLResp.Data.ImgURL
	uploadURL := uploadURLResp.Data.UploadURL

	// 2. 上传图片
	uploadResp, err := UploadImage(imagePath, uploadURL)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("上传图片失败: %v", err)
		return result, err
	}
	if uploadResp.Code != 0 || uploadResp.StatusCode != 200 {
		result.ErrorMessage = fmt.Sprintf("上传图片失败: %s", uploadResp.Errmsg)
		return result, fmt.Errorf(result.ErrorMessage)
	}

	// 3. 获取任务ID
	taskIDResp, err := GetTaskID(imgURL)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("获取任务ID失败: %v", err)
		return result, err
	}
	if taskIDResp.Code != 0 || taskIDResp.TaskID == "" {
		result.ErrorMessage = fmt.Sprintf("获取任务ID失败: %s", taskIDResp.Errmsg)
		return result, fmt.Errorf(result.ErrorMessage)
	}

	taskID := taskIDResp.TaskID

	// 4. 轮询任务结果
	taskResultResp, err := PollTaskResult([]string{taskID}, 2, 30)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("轮询任务结果失败: %v", err)
		return result, err
	}
	if taskResultResp.Code != 0 || len(taskResultResp.ImageURLs) == 0 {
		result.ErrorMessage = fmt.Sprintf("轮询任务结果失败: %s", taskResultResp.Errmsg)
		return result, fmt.Errorf(result.ErrorMessage)
	}

	// 5. 获取处理后的图片URL
	result.EnhancedImage = taskResultResp.ImageURLs[0]
	result.Success = true
	return result, nil
} 