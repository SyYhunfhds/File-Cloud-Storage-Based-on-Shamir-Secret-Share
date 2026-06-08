package config

import "github.com/gogf/gf/v2/util/grand"

// Item
//
// 条目存储目录
type Item struct {
	ItemStorage
	ItemSeal
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
