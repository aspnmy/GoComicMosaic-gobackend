package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ProxyHandler 处理跨域API代理请求
// 支持GET、POST等各种HTTP方法，并保持请求方法的一致性
func ProxyHandler(c *gin.Context) {
	// 从查询参数中获取目标URL
	targetURL := c.Query("url")
	if targetURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "缺少url参数",
		})
		return
	}
	
	// 解析可能的headers参数
	customHeaders := make(map[string]string)
	headersParam := c.Query("headers")
	if headersParam != "" {
		// 尝试解析JSON格式的headers
		var headers map[string]string
		err := json.Unmarshal([]byte(headersParam), &headers)
		if err != nil {
			log.Printf("解析headers参数失败: %v, 参数值: %s", err, headersParam)
		} else {
			customHeaders = headers
			log.Printf("成功解析自定义headers: %v", customHeaders)
		}
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	var err error
	var req *http.Request

	// 重要：确保代理请求使用与原始请求相同的HTTP方法
	// 原始请求方法直接从c.Request.Method获取
	method := c.Request.Method
	log.Printf("代理请求: 使用原始请求方法 %s", method)

	// 根据HTTP方法处理请求体
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		// 这些方法通常不带请求体
		req, err = http.NewRequest(method, targetURL, nil)
	default:
		// POST、PUT、DELETE等可能带有请求体的方法
		bodyData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("读取请求体失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "读取请求体失败",
			})
			return
		}
		// 使用与原始请求相同的方法创建新请求
		log.Printf("4444444请求体: %s", string(bodyData))
		log.Printf("5555555请URL: %s", string(targetURL))
		req, err = http.NewRequest(method, targetURL, bytes.NewReader(bodyData))
		
		// 如果原始请求有Content-Type，则保留它
		contentType := c.GetHeader("Content-Type")
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
	}

	if err != nil {
		log.Printf("创建请求失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建请求失败",
		})
		return
	}

	// 转发原始请求的所有头信息
	for key, values := range c.Request.Header {
		// 跳过某些特定的头，这些头可能会干扰代理
		if key == "Host" || key == "Connection" || key == "Content-Length" {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 应用自定义headers参数中的头信息（优先级高于原始请求的头）
	for key, value := range customHeaders {
		if key != "" && value != "" {
			req.Header.Set(key, value)
			log.Printf("设置自定义请求头: %s = %s", key, value)
		}
	}

	// 确保设置了基本的请求头（如果原始请求中没有且自定义头中也没有）
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36")
	}
	
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "*/*")
	}
	
	if req.Header.Get("Accept-Language") == "" {
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	}

	// 执行请求
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("请求目标URL失败: %v, URL: %s", err, targetURL)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "请求目标URL失败",
		})
		return
	}
	defer resp.Body.Close()

	// 转发响应头
	for key, values := range resp.Header {
		// 跳过某些特定的头
		if key == "Connection" || key == "Transfer-Encoding" {
			continue
		}
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 设置CORS头，允许跨域访问
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	// 设置响应状态码
	c.Status(resp.StatusCode)

	// 记录代理请求信息（调试用）
	log.Printf("代理请求完成: %s %s -> %d", method, targetURL, resp.StatusCode)
	
	// 复制响应体到客户端
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		log.Printf("复制响应内容失败: %v", err)
	}
} 