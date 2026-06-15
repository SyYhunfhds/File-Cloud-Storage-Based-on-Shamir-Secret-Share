package config

import (
	"time"

	"github.com/gogf/gf/v2/util/grand"
)

// Item
//
// 条目存储目录
type Item struct {
	ItemStorage
	ItemSeal
	ItemEtc
}

func DefaultItemConfig() Item {
	return Item{
		ItemStorage: ItemStorage{
			UploadDir:        ".",
			EncryptedFileDir: ".",
			UnlockedFileDir:  ".",
		},
		ItemSeal: ItemSeal{
			ShareKey: grand.S(32),
			KeySize:  32,
			Nonce:    grand.S(12),
		},
		ItemEtc: ItemEtc{
			RawShareExpire: int64((5 * time.Minute).Seconds()), // 默认5分钟过期时间
		},
	}
}

type ItemStorage struct {
	UploadDir        string `yaml:"upload" json:"upload"`
	EncryptedFileDir string `yaml:"encrypted" json:"encrypted"`
	UnlockedFileDir  string `yaml:"unlocked" json:"unlocked"`
}

// ItemSeal 条目加密配置
type ItemSeal struct {
	// ShareKey 用于加密份额的对称密钥
	ShareKey string `yaml:"key" json:"key"`

	KeySize int    `yaml:"key_size" json:"key_size"` // AES密钥长度
	Nonce   string `yaml:"nonce" json:"nonce"`
}

// ItemEtc 杂项配置
type ItemEtc struct {
	RawShareExpire int64 `yaml:"share_expire" json:"share_expire"` // 份额有效时间, 单位为秒
}
