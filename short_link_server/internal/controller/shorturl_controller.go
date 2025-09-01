package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
	"net/http"
	"shortlink/internal/model"
	"shortlink/internal/service"
)

// ShortURLController 短链接控制器
type ShortURLController struct {
	shortLinkService service.ShortLinkService
	rateLimiter      *ratelimit.Bucket
	apiKeys          map[string]bool
}

// NewShortURLController 创建短链接控制器
func NewShortURLController(shortLinkService service.ShortLinkService, rateLimiter *ratelimit.Bucket, apiKeys []string) *ShortURLController {
	// 将API Keys转换为map以提高查询效率
	apiKeyMap := make(map[string]bool)
	for _, key := range apiKeys {
		apiKeyMap[key] = true
	}
	return &ShortURLController{
		shortLinkService: shortLinkService,
		rateLimiter:      rateLimiter,
		apiKeys:          apiKeyMap,
	}
}

// APIKeyAuth 验证API Key中间件
func (c *ShortURLController) APIKeyAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 从请求头获取API Key
		apiKey := ctx.GetHeader("X-API-Key")
		if apiKey == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			ctx.Abort()
			return
		}

		// 验证API Key
		if !c.apiKeys[apiKey] {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			ctx.Abort()
			return
		}

		// API Key验证通过，继续处理请求
		ctx.Next()
	}
}

// RateLimitMiddleware 限流中间件
func (c *ShortURLController) RateLimitMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 尝试从令牌桶获取令牌
		if c.rateLimiter.TakeAvailable(1) == 0 {
			ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// CreateShortLink 创建短链接
func (c *ShortURLController) CreateShortLink(ctx *gin.Context) {
	var req model.CreateShortLinkRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.shortLinkService.CreateShortLink(req.OriginalURL, req.ExpiresAt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// RedirectToOriginalURL 重定向到原始链接
func (c *ShortURLController) RedirectToOriginalURL(ctx *gin.Context) {
	shortCode := ctx.Param("shortCode")
	if shortCode == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Short code required"})
		return
	}

	originalURL, err := c.shortLinkService.GetOriginalURL(shortCode)
	if err != nil {
		if err.Error() == "short link not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Short link not found"})
		} else if err.Error() == "short link has expired" {
			ctx.JSON(http.StatusGone, gin.H{"error": "Short link has expired"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// 重定向到原始URL
	ctx.Redirect(http.StatusFound, originalURL)
}
