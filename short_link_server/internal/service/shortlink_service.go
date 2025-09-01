package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/go-redis/redis/v8"
	"shortlink/internal/model"
	"shortlink/internal/repository"
	"time"
)

// ShortLinkService 短链接服务接口
type ShortLinkService interface {
	CreateShortLink(originalURL string, expiresAt string) (*model.CreateShortLinkResponse, error)
	GetOriginalURL(shortCode string) (string, error)
	GetShortLinkInfo(shortCode string) (*model.ShortLink, error)
	DeleteShortLink(shortCode string) error
}

// shortLinkServiceImpl 短链接服务实现
type shortLinkServiceImpl struct {
	repo        repository.ShortLinkRepository
	redisClient *redis.Client
	baseURL     string
}

// NewShortLinkService 创建短链接服务实例
func NewShortLinkService(repo repository.ShortLinkRepository, redisClient *redis.Client, baseURL string) ShortLinkService {
	return &shortLinkServiceImpl{repo: repo, redisClient: redisClient, baseURL: baseURL}
}

// generateShortCode 生成短码
func (s *shortLinkServiceImpl) generateShortCode() (string, error) {
	b := make([]byte, 6) // 6字节的随机数据
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// 使用base64url编码，避免特殊字符
	shortCode := base64.URLEncoding.EncodeToString(b)[:6] // 取前6个字符
	return shortCode, nil
}

// CreateShortLink 创建短链接
func (s *shortLinkServiceImpl) CreateShortLink(originalURL string, expiresAt string) (*model.CreateShortLinkResponse, error) {
	// 生成唯一短码
	var shortCode string
	var err error
	for i := 0; i < 10; i++ { // 最多尝试10次
		shortCode, err = s.generateShortCode()
		if err != nil {
			continue
		}
		// 检查短码是否已存在
		existingLink, _ := s.repo.GetByShortCode(shortCode)
		if existingLink == nil {
			break
		}
	}
	if err != nil {
		return nil, errors.New("failed to generate short code")
	}

	// 解析过期时间
	var expiresTime time.Time
	if expiresAt != "" {
		expiresTime, err = time.Parse(time.DateTime, expiresAt)
		if err != nil {
			return nil, errors.New("invalid expires_at format")
		}
	}

	// 创建短链接记录
	link := &model.ShortLink{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		ExpiresAt:   expiresTime,
		ModifyAt:    time.Now(),
	}
	err = s.repo.Create(link)
	if err != nil {
		return nil, err
	}

	// 构造响应
	shortURL := s.baseURL + "/" + shortCode
	response := &model.CreateShortLinkResponse{
		ShortCode:   shortCode,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		ExpiresAt:   expiresAt,
	}

	return response, nil
}

// GetOriginalURL 获取原始链接
func (s *shortLinkServiceImpl) GetOriginalURL(shortCode string) (string, error) {
	// 1. 尝试从Redis获取
	ctx := context.Background()
	originalURL, err := s.redisClient.Get(ctx, "shortlink:"+shortCode).Result()
	if err == nil {
		// 异步更新点击次数
		go s.repo.UpdateClickCount(shortCode)
		return originalURL, nil
	}

	// 2. Redis未命中，从数据库获取
	link, err := s.repo.GetByShortCode(shortCode)
	if err != nil {
		return "", err
	}
	if link == nil {
		return "", errors.New("short link not found")
	}

	// 3. 将结果存入Redis
	var ttl time.Duration
	if !link.ExpiresAt.IsZero() {
		ttl = time.Until(link.ExpiresAt)
		if ttl <= 0 {
			return "", errors.New("short link has expired")
		}
	} else {
		ttl = 24 * time.Hour // 默认缓存24小时
	}
	err = s.redisClient.Set(ctx, "shortlink:"+shortCode, link.OriginalURL, ttl).Err()
	if err != nil {
		// 缓存失败不影响主流程
	}

	// 4. 更新点击次数
	go s.repo.UpdateClickCount(shortCode)

	return link.OriginalURL, nil
}

// GetShortLinkInfo 获取短链接信息
func (s *shortLinkServiceImpl) GetShortLinkInfo(shortCode string) (*model.ShortLink, error) {
	return s.repo.GetByShortCode(shortCode)
}

// DeleteShortLink 删除短链接
func (s *shortLinkServiceImpl) DeleteShortLink(shortCode string) error {
	// 删除数据库记录
	err := s.repo.Delete(shortCode)
	if err != nil {
		return err
	}

	// 删除Redis缓存
	ctx := context.Background()
	s.redisClient.Del(ctx, "shortlink:"+shortCode)

	return nil
}
