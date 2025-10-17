package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
)

// ResourceStatus 资源状态枚举
type ResourceStatus string

// 资源状态常量
const (
	ResourceStatusPending  ResourceStatus = "PENDING"
	ResourceStatusApproved ResourceStatus = "APPROVED"
	ResourceStatusRejected ResourceStatus = "REJECTED"
)

// Value 实现database/sql/driver.Valuer接口
func (rs ResourceStatus) Value() (driver.Value, error) {
	return string(rs), nil
}

// Scan 实现sql.Scanner接口
func (rs *ResourceStatus) Scan(value interface{}) error {
	if value == nil {
		*rs = ResourceStatusPending
		return nil
	}

	strVal, ok := value.(string)
	if !ok {
		return errors.New("无法将值扫描为ResourceStatus")
	}

	*rs = ResourceStatus(strVal)
	return nil
}

// JsonList 自定义类型，用于存储JSON数组
type JsonList []string

// Value 实现database/sql/driver.Valuer接口
func (j JsonList) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	bytes, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	// 返回JSON字符串而不是字节数组
	return string(bytes), nil
}

// Scan 实现sql.Scanner接口
func (j *JsonList) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("无法将值扫描为JsonList：不支持的类型")
	}

	// 检查空字符串
	if len(b) == 0 {
		*j = JsonList{}
		return nil
	}

	// 检查为NULL的情况
	if string(b) == "null" {
		*j = nil
		return nil
	}

	// 先尝试解析为字符串数组
	var result []string
	if err := json.Unmarshal(b, &result); err == nil {
		*j = result
		return nil
	}

	// 如果无法解析为字符串数组，尝试解析为通用数组
	var rawData interface{}
	if err := json.Unmarshal(b, &rawData); err != nil {
		return fmt.Errorf("解析JsonList失败: %w", err)
	}

	if array, ok := rawData.([]interface{}); ok {
		result := make([]string, len(array))
		for i, item := range array {
			if str, ok := item.(string); ok {
				result[i] = str
			} else {
				itemBytes, err := json.Marshal(item)
				if err != nil {
					return err
				}
				result[i] = string(itemBytes)
			}
		}
		*j = result
		return nil
	}

	// 如果都不成功，返回错误
	return errors.New("无法将值转换为JsonList")
}

// JsonMap 自定义类型，用于存储JSON对象
type JsonMap map[string]interface{}

// Value 实现database/sql/driver.Valuer接口
func (j JsonMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	bytes, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	// 返回JSON字符串而不是字节数组
	return string(bytes), nil
}

// Scan 实现sql.Scanner接口
func (j *JsonMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("无法将值扫描为JsonMap：不支持的类型")
	}

	// 检查空字符串
	if len(b) == 0 {
		*j = JsonMap{}
		return nil
	}

	// 检查为NULL的情况
	if string(b) == "null" {
		*j = nil
		return nil
	}

	// 尝试解析JSON
	var result map[string]interface{}
	if err := json.Unmarshal(b, &result); err != nil {
		// 如果解析失败，记录错误并返回空map
		log.Printf("解析JsonMap失败: %v, 原始数据: %s", err, string(b))
		*j = JsonMap{}
		return nil
	}
	*j = result
	return nil
}

// Resource 资源模型
type Resource struct {
	ID                 int            `db:"id" json:"id"`
	Title              string         `db:"title" json:"title"`
	TitleEn            string         `db:"title_en" json:"title_en"`
	Description        string         `db:"description" json:"description"`
	Images             JsonList       `db:"images" json:"images"`
	PosterImage        *string        `db:"poster_image" json:"poster_image"`
	ResourceType       string         `db:"resource_type" json:"resource_type"`
	Status             ResourceStatus `db:"status" json:"status"`
	HiddenFromAdmin    *bool          `db:"hidden_from_admin" json:"hidden_from_admin"`
	Links              JsonMap        `db:"links" json:"links"`
	OriginalResourceID *int           `db:"original_resource_id" json:"original_resource_id"`
	Supplement         JsonMap        `db:"supplement" json:"supplement"`
	ApprovalHistory    JsonMap        `db:"approval_history" json:"approval_history"`
	IsSupplementApproval bool         `db:"is_supplement_approval" json:"is_supplement_approval"`
	LikesCount         int            `db:"likes_count" json:"likes_count"`
	TmdbID             *int           `db:"tmdb_id" json:"tmdb_id"`
	MediaType          *string        `db:"media_type" json:"media_type"`
	Stickers           JsonMap        `db:"stickers" json:"stickers"`
	CreatedAt          time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at" json:"updated_at"`
	TotalCount         *int           `db:"-" json:"total_count,omitempty"` // 不存储在数据库中，用于分页
	HasPendingSupplement bool         `db:"-" json:"has_pending_supplement,omitempty"` // 不存储在数据库中
}

// User 用户模型
type User struct {
	ID             int       `db:"id" json:"id"`
	Username       string    `db:"username" json:"username"`
	HashedPassword string    `db:"hashed_password" json:"-"` // 不返回给客户端
	IsAdmin        bool      `db:"is_admin" json:"is_admin"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// ResourceCreate 用于创建新资源的请求结构
type ResourceCreate struct {
	Title        string    `json:"title" binding:"required"`
	TitleEn      string    `json:"title_en"`
	Description  string    `json:"description" binding:"required"`
	ResourceType string    `json:"resource_type" binding:"required"`
	Images       []string  `json:"images"`
	PosterImage  string    `json:"poster_image"`
	Links        JsonMap   `json:"links"`
}

// ResourceUpdate 资源更新请求
type ResourceUpdate struct {
	Title        *string   `json:"title"`
	TitleEn      *string   `json:"title_en"`
	Description  *string   `json:"description"`
	ResourceType *string   `json:"resource_type"`
	Images       []string  `json:"images"`
	PosterImage  *string   `json:"poster_image"`
	Links        JsonMap   `json:"links"`
	TmdbID       *int      `json:"tmdb_id"`
	MediaType    *string   `json:"media_type"`
	Stickers     JsonMap   `json:"stickers"`
}

// ResourceApproval 资源审批请求
type ResourceApproval struct {
	Status          ResourceStatus          `json:"status" binding:"required"`
	FieldApprovals  map[string]bool         `json:"field_approvals"`
	FieldRejections map[string]bool         `json:"field_rejections"`
	ApprovedImages  []string                `json:"approved_images"`
	RejectedImages  []string                `json:"rejected_images"`
	PosterImage     string                 `json:"poster_image"`
	Notes           string                 `json:"notes"`
	ApprovedLinks   []map[string]interface{} `json:"approved_links"`
	RejectedLinks   []map[string]interface{} `json:"rejected_links"`
}

// ApprovalRecord 审批记录模型
type ApprovalRecord struct {
	ID               int            `db:"id" json:"id"`
	ResourceID       int            `db:"resource_id" json:"resource_id"`
	Status           ResourceStatus `db:"status" json:"status"`
	FieldApprovals   JsonMap        `db:"field_approvals" json:"field_approvals"`
	FieldRejections  JsonMap        `db:"field_rejections" json:"field_rejections"`
	ApprovedImages   JsonList       `db:"approved_images" json:"approved_images"`
	RejectedImages   JsonList       `db:"rejected_images" json:"rejected_images"`
	PosterImage      string         `db:"poster_image" json:"poster_image"`
	Notes            string         `db:"notes" json:"notes"`
	ApprovedLinks    JsonMap        `db:"approved_links" json:"approved_links"`
	RejectedLinks    JsonMap        `db:"rejected_links" json:"rejected_links"`
	IsSupplementApproval bool       `db:"is_supplement_approval" json:"is_supplement_approval"`
	CreatedAt        time.Time      `db:"created_at" json:"created_at"`
}

// UserCreate 用户创建请求
type UserCreate struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

// PasswordUpdate 密码更新请求
type PasswordUpdate struct {
	CurrentPassword string `json:"current_password" form:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" form:"new_password" binding:"required"`
}

// SupplementCreate 资源补充内容创建请求
type SupplementCreate struct {
	Images []string  `json:"images"`
	Links  JsonMap   `json:"links"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// SiteSettings 网站设置模型
type SiteSettings struct {
	ID           int       `db:"id" json:"id"`
	SettingKey   string    `db:"setting_key" json:"setting_key"`
	SettingValue JsonMap   `db:"setting_value" json:"setting_value"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// SiteSettingsUpdate 网站设置更新请求
type SiteSettingsUpdate struct {
	SettingValue JsonMap `json:"setting_value" binding:"required"`
}