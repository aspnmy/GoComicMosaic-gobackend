package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aspnmy/GoComicMosaic-gobackend/gobackend/internal/config"
	"github.com/aspnmy/GoComicMosaic-gobackend/gobackend/internal/models"
	"github.com/aspnmy/GoComicMosaic-gobackend/gobackend/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetSiteSettings 获取指定key的网站设置
func GetSiteSettings(c *gin.Context) {
	settingKey := c.Param("key")
	if settingKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少setting_key参数"})
		return
	}

	var settings models.SiteSettings
	err := models.GetDB().Get(&settings, "SELECT * FROM site_settings WHERE setting_key = ?", settingKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "未找到指定设置",
			"key":   settingKey,
		})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// GetAllSiteSettings 获取所有网站设置
func GetAllSiteSettings(c *gin.Context) {
	var settings []models.SiteSettings
	err := models.GetDB().Select(&settings, "SELECT * FROM site_settings")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取设置失败"})
		return
	}

	if len(settings) == 0 {
		c.JSON(http.StatusOK, []models.SiteSettings{})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateSiteSettings 更新网站设置
func UpdateSiteSettings(c *gin.Context) {
	// 打印请求头信息
	log.Printf("===== 更新设置请求 =====")
	log.Printf("请求方法: %s", c.Request.Method)
	log.Printf("请求路径: %s", c.Request.URL.Path)
	log.Printf("认证头: %s", c.GetHeader("Authorization"))
	log.Printf("Content-Type: %s", c.GetHeader("Content-Type"))

	// 获取原始请求体
	data, err := c.GetRawData()
	if err != nil {
		log.Printf("获取请求体失败: %v", err)
	} else {
		log.Printf("原始请求数据: %s", string(data))
	}
	// 重新设置请求体，否则后续ShouldBindJSON会读取失败
	c.Request.Body = originalBody(data)

	settingKey := c.Param("key")
	log.Printf("设置键名: %v", settingKey)
	if settingKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少setting_key参数"})
		return
	}

	var update models.SiteSettingsUpdate
	if bindErr := c.ShouldBindJSON(&update); bindErr != nil {
		log.Printf("解析JSON数据失败: %v", bindErr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据", "details": bindErr.Error()})
		return
	}

	// 打印解析后的数据
	settingValue, err := update.SettingValue.Value()
	if err != nil {
		log.Printf("获取设置值失败: %v", err)
	} else {
		// 根据实际类型进行处理
		switch v := settingValue.(type) {
		case []byte:
			log.Printf("设置值([]byte): %v", string(v))
		case string:
			log.Printf("设置值(string): %v", v)
		default:
			log.Printf("设置值(其他类型): %v", v)
		}
	}

	db := models.GetDB()

	// 检查设置是否存在
	var settingExists int
	err = db.Get(&settingExists, "SELECT COUNT(*) FROM site_settings WHERE setting_key = ?", settingKey)
	if err != nil {
		log.Printf("查询设置是否存在失败: %v", err)
	}
	log.Printf("设置是否存在: %v (count=%d)", settingExists > 0, settingExists)

	if err != nil || settingExists == 0 {
		// 创建新设置
		settingValue, valueErr := update.SettingValue.Value()
		if valueErr != nil {
			log.Printf("序列化设置值失败: %v", valueErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "序列化设置值失败"})
			return
		}

		result, execErr := db.Exec(
			"INSERT INTO site_settings (setting_key, setting_value, created_at, updated_at) VALUES (?, ?, ?, ?)",
			settingKey, settingValue, time.Now(), time.Now(),
		)
		if execErr != nil {
			log.Printf("保存设置失败: %v", execErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存设置失败", "details": execErr.Error()})
			return
		}
		id, _ := result.LastInsertId()
		log.Printf("创建新设置成功，ID: %d", id)
	} else {
		// 更新现有设置
		settingValue, valueErr := update.SettingValue.Value()
		if valueErr != nil {
			log.Printf("序列化设置值失败: %v", valueErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "序列化设置值失败"})
			return
		}

		result, updateErr := db.Exec(
			"UPDATE site_settings SET setting_value = ?, updated_at = ? WHERE setting_key = ?",
			settingValue, time.Now(), settingKey,
		)
		if updateErr != nil {
			log.Printf("更新设置失败: %v", updateErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新设置失败", "details": updateErr.Error()})
			return
		}
		rows, _ := result.RowsAffected()
		log.Printf("更新设置成功，影响行数: %d", rows)
	}

	// 返回更新后的设置
	var settings models.SiteSettings
	err = models.GetDB().Get(&settings, "SELECT * FROM site_settings WHERE setting_key = ?", settingKey)
	if err != nil {
		log.Printf("读取更新后的设置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取更新后的设置失败"})
		return
	}

	// 特殊处理图片格式设置
	if settingKey == "image_format" {
		// 将设置值应用到配置中
		// 正确处理models.JsonMap类型
		var formatValue string
		
		// 由于SettingValue是JsonMap(map[string]interface{})，我们需要适当处理
		// 尝试获取map中的值作为字符串
		if value, ok := settings.SettingValue["value"]; ok {
			// 尝试不同类型的转换
			switch v := value.(type) {
			case string:
				formatValue = v
			case []byte:
				formatValue = string(v)
			default:
				// 尝试将其他类型转换为字符串
				formatValue = fmt.Sprintf("%v", v)
			}
		} else {
			// 如果没有特定的键，尝试将整个map转换为字符串
			formatValue = fmt.Sprintf("%v", settings.SettingValue)
		}
		
		if config.SetImageFormat(formatValue) {
			log.Printf("图片格式设置已成功应用: %s", formatValue)
		} else {
			log.Printf("图片格式设置应用失败: %s", formatValue)
		}
	}

	log.Printf("更新设置完成，返回结果: %+v", settings)
	c.JSON(http.StatusOK, settings)
}

// originalBody 创建一个可重复读取的请求体
func originalBody(data []byte) io.ReadCloser {
	return io.NopCloser(bytes.NewBuffer(data))
}

// UploadFavicon 处理网站图标上传
func UploadFavicon(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("favicon")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件上传"})
		return
	}

	// 确保目录存在
	publicDir := filepath.Join(config.AssetPath, "public")
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目录失败"})
		return
	}

	// 保存favicon.ico
	faviconPath := filepath.Join(publicDir, "favicon.ico")

	// 保存上传的文件
	if err := c.SaveUploadedFile(file, faviconPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	// 增加日志输出
	log.Printf("网站图标已更新，保存路径: %s", faviconPath)

	// 确保返回正确的路径
	c.JSON(http.StatusOK, gin.H{
		"message":     "网站图标已更新",
		"faviconPath": "/assets/public/favicon.ico",
	})
}

// GetTMDBConfig 获取TMDB配置
func GetTMDBConfig(c *gin.Context) {
	var settings models.SiteSettings
	err := models.GetDB().Get(&settings, "SELECT * FROM site_settings WHERE setting_key = ?", utils.TMDB_SETTINGS_KEY)

	if err != nil {
		// 如果找不到配置，返回一个空的配置
		c.JSON(http.StatusOK, gin.H{
			"setting_key": utils.TMDB_SETTINGS_KEY,
			"setting_value": gin.H{
				"api_key": "",
				"enabled": false,
			},
		})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateTMDBConfig 更新TMDB配置
func UpdateTMDBConfig(c *gin.Context) {
	var update struct {
		APIKey  string `json:"api_key" binding:"required"`
		Enabled bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据", "details": err.Error()})
		return
	}

	// 检查配置是否已存在
	var settingsUpdate models.SiteSettingsUpdate
	settingsUpdate.SettingValue = models.JsonMap{
		"api_key": update.APIKey,
		"enabled": update.Enabled,
	}

	db := models.GetDB()

	// 检查设置是否存在
	var settingExists int
	err := db.Get(&settingExists, "SELECT COUNT(*) FROM site_settings WHERE setting_key = ?", utils.TMDB_SETTINGS_KEY)

	if err != nil || settingExists == 0 {
		// 创建新设置
		settingValue, valueErr := settingsUpdate.SettingValue.Value()
		if valueErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "序列化设置值失败"})
			return
		}

		var execErr error
		_, execErr = db.Exec(
			"INSERT INTO site_settings (setting_key, setting_value, created_at, updated_at) VALUES (?, ?, ?, ?)",
			utils.TMDB_SETTINGS_KEY, settingValue, time.Now(), time.Now(),
		)
		if execErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存设置失败", "details": execErr.Error()})
			return
		}
	} else {
		// 更新现有设置
		settingValue, valueErr := settingsUpdate.SettingValue.Value()
		if valueErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "序列化设置值失败"})
			return
		}

		_, err = db.Exec(
			"UPDATE site_settings SET setting_value = ?, updated_at = ? WHERE setting_key = ?",
			settingValue, time.Now(), utils.TMDB_SETTINGS_KEY,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新设置失败", "details": err.Error()})
			return
		}
	}

	// 更新TMDB API密钥缓存
	utils.SetTMDBAPIKey(update.APIKey)

	// 返回更新后的设置
	var settings models.SiteSettings
	err = models.GetDB().Get(&settings, "SELECT * FROM site_settings WHERE setting_key = ?", utils.TMDB_SETTINGS_KEY)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取更新后的设置失败"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// SaveTMDBAPIKeyToDB 将TMDB API密钥保存到数据库
func SaveTMDBAPIKeyToDB(apiKey string) {
	// 确保数据库连接已初始化
	db := models.GetDB()
	if db == nil {
		log.Printf("数据库连接未初始化，无法保存TMDB API密钥")
		return
	}

	// 检查是否已存在TMDB配置
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM site_settings WHERE setting_key = ?", utils.TMDB_SETTINGS_KEY)
	if err != nil {
		log.Printf("查询TMDB配置失败: %v", err)
		return
	}

	// 准备保存的数据
	settingValue := map[string]interface{}{
		"api_key": apiKey,
		"enabled": true, // 默认启用TMDB功能
	}
	valueJSON, err := json.Marshal(settingValue)
	if err != nil {
		log.Printf("序列化TMDB配置失败: %v", err)
		return
	}

	// 插入或更新配置
	if count == 0 {
		// 创建新配置
		_, err = db.Exec(
			"INSERT INTO site_settings (setting_key, setting_value, created_at, updated_at) VALUES (?, ?, ?, ?)",
			utils.TMDB_SETTINGS_KEY, string(valueJSON), time.Now(), time.Now(),
		)
	} else {
		// 更新现有配置
		_, err = db.Exec(
			"UPDATE site_settings SET setting_value = JSON_PATCH(setting_value, ?), updated_at = ? WHERE setting_key = ?",
			fmt.Sprintf(`{"api_key": "%s"}`, apiKey), time.Now(), utils.TMDB_SETTINGS_KEY,
		)
	}

	if err != nil {
		log.Printf("保存TMDB配置到数据库失败: %v", err)
		return
	}

	log.Printf("已成功将环境变量中的TMDB API密钥保存到数据库")
}

// LoadTMDBConfig 从数据库加载TMDB配置并设置API密钥
// 此函数应在服务启动时被调用
func LoadTMDBConfig() {
	// 尝试从数据库加载TMDB配置
	var settings models.SiteSettings
	err := models.GetDB().Get(&settings, "SELECT * FROM site_settings WHERE setting_key = ?", utils.TMDB_SETTINGS_KEY)

	if err == nil && settings.SettingValue != nil {
		// 检查是否有api_key字段
		if apiKey, ok := settings.SettingValue["api_key"].(string); ok && apiKey != "" {
			// 检查是否启用了TMDB功能
			enabled := true // 默认启用
			if val, ok := settings.SettingValue["enabled"].(bool); ok {
				enabled = val
			}

			// 如果启用了TMDB功能，设置TMDB API密钥
			if enabled {
				utils.SetTMDBAPIKey(apiKey)
				log.Printf("已从数据库加载TMDB API密钥，TMDB功能已启用")
			} else {
				// 如果未启用，设置为空字符串
				utils.SetTMDBAPIKey("")
				log.Printf("TMDB功能已禁用")
			}
			return
		}
	}

	// 如果数据库中没有配置或加载失败，尝试从环境变量获取
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey != "" {
		utils.SetTMDBAPIKey(apiKey)
		log.Printf("已从环境变量加载TMDB API密钥")

		// 异步保存API密钥到数据库
		go func() {
			// 等待一段时间确保数据库已初始化
			time.Sleep(2 * time.Second)
			SaveTMDBAPIKeyToDB(apiKey)
		}()

		return
	}

}

// GetTMDBStatus 获取TMDB功能是否启用，只返回enabled状态，不返回API密钥
func GetTMDBStatus(c *gin.Context) {
	var settings models.SiteSettings
	err := models.GetDB().Get(&settings, "SELECT * FROM site_settings WHERE setting_key = ?", utils.TMDB_SETTINGS_KEY)

	if err != nil {
		// 如果找不到配置，返回未启用状态
		c.JSON(http.StatusOK, gin.H{
			"enabled": false,
		})
		return
	}

	// 从配置中只提取enabled字段
	enabled := false
	if val, ok := settings.SettingValue["enabled"].(bool); ok {
		enabled = val
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled": enabled,
	})
}
