package config

import "github.com/pilinux/argon2"

type ArgonConfig struct {
	// Secret Argon2d 额外密钥
	Secret string `json:"secret" yaml:"secret"`

	// Memory Argon2d 内存消耗
	Memory uint32 `json:"memory" yaml:"memory"`
	// Iterations Argon2d 迭代次数
	Iterations uint32 `json:"iterations" yaml:"iterations"`
	// Parallelism Argon2d 并行度
	Parallelism uint8 `json:"parallelism" yaml:"parallelism"`
	// SaltLength Argon2d 盐长度
	SaltLength uint32 `json:"salt_len" yaml:"salt_len"`
	// KeyLength Argon2d 密钥长度
	KeyLength uint32 `json:"key_len" yaml:"key_len"`
}

func (cfg *ArgonConfig) AsArgonParams() argon2.Params {
	return argon2.Params{
		Memory:      cfg.Memory,
		Iterations:  cfg.Iterations,
		Parallelism: cfg.Parallelism,
		SaltLength:  cfg.SaltLength,
		KeyLength:   cfg.KeyLength,
	}
}

type ServerConfig struct {
	// 映射已有配置
	Address     string `json:"address" yaml:"address"`
	OpenapiPath string `json:"openapiPath" yaml:"openapiPath"`
	SwaggerPath string `json:"swaggerPath" yaml:"swaggerPath"`

	// 映射额外配置
	Argon  ArgonConfig  `json:"argon" yaml:"argon"`
	Cookie CookieConfig `json:"cookie" yaml:"cookie"`
}
