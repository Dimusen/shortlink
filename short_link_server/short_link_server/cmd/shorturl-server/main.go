package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/juju/ratelimit"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"shortlink/internal/config"
	"shortlink/internal/controller"
	"shortlink/internal/model"
	"shortlink/internal/repository"
	"shortlink/internal/service"
)

func main() {
	// 1. 加载配置
	cfg := config.NewDefaultConfig()

	// 2. 初始化数据库
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 3. 初始化Redis
	redisClient, err := initRedis(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// 4. 创建仓库、服务和控制器
	shortLinkRepo := repository.NewShortLinkRepository(db)
	shortLinkService := service.NewShortLinkService(shortLinkRepo, redisClient, fmt.Sprintf("http://localhost:%d", cfg.Server.Port))

	// 5. 创建令牌桶限流器
	rateLimiter := ratelimit.NewBucketWithRate(float64(cfg.RateLimit.Rate), int64(cfg.RateLimit.Capacity))

	// 6. 创建控制器
	shortURLController := controller.NewShortURLController(shortLinkService, rateLimiter, cfg.Auth.APIKeys)

	// 7. 设置路由
	r := gin.Default()

	// 8. API路由组（需要认证）
	apiV1 := r.Group("/api/v1")
	apiV1.Use(shortURLController.APIKeyAuth())
	{
		apiV1.POST("/shorten", shortURLController.CreateShortLink)
	}

	// 9. 短链接重定向（需要限流）
	r.GET("/:shortCode", shortURLController.RateLimitMiddleware(), shortURLController.RedirectToOriginalURL)

	// 10. 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initDatabase 初始化数据库
func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=%t&loc=%s&timeout=%s&readTimeout=%s&writeTimeout=%s",
		cfg.MySQL.User,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.Database,
		cfg.MySQL.ParseTime,
		cfg.MySQL.Loc,
		cfg.MySQL.Timeout,
		cfg.MySQL.ReadTimeout,
		cfg.MySQL.WriteTimeout,
	)
	//db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取数据库连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection pool: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// 自动迁移表结构
	err = db.AutoMigrate(&model.ShortLink{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// initRedis 初始化Redis
func initRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
		PoolSize:     cfg.Redis.PoolSize,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}
