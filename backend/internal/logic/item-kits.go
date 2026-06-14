package logic

import (
	"backend/internal/config"
	"time"
)

type ItemUtils struct {
	shareExpire time.Duration
}
type ItemUtilOptionFunc func(iu *ItemUtils)

func NewItemUtils(options ...ItemUtilOptionFunc) *ItemUtils {
	utils := &ItemUtils{
		shareExpire: 5 * time.Minute, // 份额默认只有5分钟有效时间
	}

	for _, opt := range options {
		opt(utils)
	}
	return utils
}
func (iu *ItemUtils) BuildWithConfig(cfg config.Item) {
	iu.shareExpire = time.Duration(cfg.RawShareExpire) * time.Second
}

func (iu *ItemUtils) ExpireAt() time.Time {
	return time.Now().Add(iu.shareExpire)
}
