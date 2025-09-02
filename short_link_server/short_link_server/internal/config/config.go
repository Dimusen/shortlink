package config

import "time"

// Config 应用配置
type Config struct {
	Server    ServerConfig    `json:"server"`
	Database  DatabaseConfig  `json:"database"`
	MySQL     MySQLConfig     `json:"mysql"`
	Redis     RedisConfig     `json:"redis"`
	RateLimit RateLimitConfig `json:"rate_limit"`
	Auth      AuthConfig      `json:"auth"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int `json:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string        `json:"driver"`
	DSN             string        `json:"dsn"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
}
type MySQLConfig struct {
	Host            string        `json:"host" yaml:"host"`
	Port            int           `json:"port" yaml:"port"`
	User            string        `json:"user" yaml:"user"`
	Password        string        `json:"password" yaml:"password"`
	Database        string        `json:"database" yaml:"database"`
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" yaml:"conn_max_idle_time"`
	ParseTime       bool          `json:"parse_time" yaml:"parse_time"` // 是否解析时间字段
	Loc             string        `json:"loc" yaml:"loc"`               // 时区设置
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`       // 连接超时
	ReadTimeout     time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout" yaml:"write_timeout"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr         string        `json:"addr"`
	Password     string        `json:"password"`
	DB           int           `json:"db"`
	DialTimeout  time.Duration `json:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	PoolSize     int           `json:"pool_size"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Capacity int   `json:"capacity"` // 令牌桶容量
	Rate     int64 `json:"rate"`     // 每秒生成的令牌数
}

// AuthConfig 认证配置
type AuthConfig struct {
	APIKeys []string `json:"api_keys"`
}

// NewDefaultConfig 创建默认配置
func NewDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
		},
		Database: DatabaseConfig{
			Driver:          "sqlite3",
			DSN:             "./shortlink.db",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: time.Hour,
		},
		MySQL: MySQLConfig{
			Host:            "localhost",
			Port:            3306,
			User:            "root",
			Password:        "123456",
			Database:        "jike",
			MaxOpenConns:    20, // 比SQLite更大的连接池
			MaxIdleConns:    10,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: 30 * time.Minute,
			ParseTime:       true,    // 重要：解析MySQL的datetime为Go的time.Time
			Loc:             "Local", // 使用本地时区
			Timeout:         10 * time.Second,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
		},
		Redis: RedisConfig{
			Addr:         "localhost:6379",
			Password:     "",
			DB:           0,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     10,
		},
		RateLimit: RateLimitConfig{
			Capacity: 1000,
			Rate:     100,
		},
		Auth: AuthConfig{
			APIKeys: []string{"test-api-key"},
		},
	}
}
