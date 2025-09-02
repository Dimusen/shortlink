package utils

import (
    "net/url"
    "regexp"
    "strings"
    "time"
)

// IsValidURL 验证URL是否有效
func IsValidURL(u string) bool {
    parsedURL, err := url.ParseRequestURI(u)
    if err != nil {
        return false
    }
    // 检查scheme是否为http或https
    if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
        return false
    }
    // 检查host是否有效
    if parsedURL.Host == "" {
        return false
    }
    return true
}

// IsValidDateTime 验证日期时间格式是否有效
func IsValidDateTime(dt string) bool {
    if dt == "" {
        return true
    }
    // 尝试多种常见的日期时间格式
    formats := []string{
        time.RFC3339,
        "2006-01-02T15:04:05",
        "2006-01-02 15:04:05",
        "2006-01-02",
    }
    
    for _, format := range formats {
        _, err := time.Parse(format, dt)
        if err == nil {
            return true
        }
    }
    return false
}

// SanitizeInput 清理输入字符串，防止XSS攻击
func SanitizeInput(s string) string {
    // 移除HTML标签
    re := regexp.MustCompile(`<[^>]*>`)
    s = re.ReplaceAllString(s, "")
    // 转义特殊字符
    s = strings.ReplaceAll(s, "&", "&amp;")
    s = strings.ReplaceAll(s, "<", "&lt;")
    s = strings.ReplaceAll(s, ">", "&gt;")
    s = strings.ReplaceAll(s, "'", "&#39;")
    s = strings.ReplaceAll(s, "\"", "&quot;")
    return s
}