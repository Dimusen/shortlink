package repository

import (
    "errors"
    "time"
    "gorm.io/gorm"
    "shortlink/internal/model"
)

// ShortLinkRepository 短链接仓库接口
type ShortLinkRepository interface {
    Create(link *model.ShortLink) error
    GetByShortCode(shortCode string) (*model.ShortLink, error)
    UpdateClickCount(shortCode string) error
    Delete(shortCode string) error
}

// shortLinkRepositoryImpl 短链接仓库实现
 type shortLinkRepositoryImpl struct {
    db *gorm.DB
}

// NewShortLinkRepository 创建短链接仓库实例
 func NewShortLinkRepository(db *gorm.DB) ShortLinkRepository {
    return &shortLinkRepositoryImpl{db: db}
}

// Create 创建短链接
 func (r *shortLinkRepositoryImpl) Create(link *model.ShortLink) error {
    return r.db.Create(link).Error
}

// GetByShortCode 根据短码获取短链接
 func (r *shortLinkRepositoryImpl) GetByShortCode(shortCode string) (*model.ShortLink, error) {
    var link model.ShortLink
    err := r.db.Where("short_code = ?", shortCode).First(&link).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    // 检查是否过期
    if !link.ExpiresAt.IsZero() && link.ExpiresAt.Before(time.Now()) {
        return nil, nil
    }
    return &link, nil
}

// UpdateClickCount 更新点击次数
 func (r *shortLinkRepositoryImpl) UpdateClickCount(shortCode string) error {
    return r.db.Model(&model.ShortLink{}).Where("short_code = ?", shortCode).Update("click_count", gorm.Expr("click_count + ?", 1)).Error
}

// Delete 删除短链接
 func (r *shortLinkRepositoryImpl) Delete(shortCode string) error {
    return r.db.Where("short_code = ?", shortCode).Delete(&model.ShortLink{}).Error
}