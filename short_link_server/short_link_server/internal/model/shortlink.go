package model

import "time"

// ShortLink 短链接数据模型
type ShortLink struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ShortCode   string    `gorm:"unique;size:20;not null;index:idx_short_code" json:"short_code"`
	OriginalURL string    `gorm:"size:2000;not null" json:"original_url"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	ModifyAt    time.Time `json:"modify_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	ClickCount  int64     `gorm:"default:0" json:"click_count"`
}

// CreateShortLinkRequest 创建短链接请求
type CreateShortLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	ExpiresAt   string `json:"expires_at" binding:"omitempty"`
}

// CreateShortLinkResponse 创建短链接响应
type CreateShortLinkResponse struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	ExpiresAt   string `json:"expires_at"`
}
