package logic

import (
	"backend/internal/config"

	"github.com/pilinux/argon2"
)

// Memclr 遍历置零字节数组
func Memclr(b []byte) {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}

type HashUtils struct {
	HashGen    func(password string) (string, error)
	HashVerify func(password string, hash string) (bool, error)
}

func NewHashUtils() *HashUtils {
	return &HashUtils{}
}
func (u *HashUtils) BuildWithConfig(cfg *config.ArgonConfig) {
	params := &argon2.Params{ // 这个参数将会被分配到堆上
		Iterations:  cfg.Iterations,
		KeyLength:   cfg.KeyLength,
		Memory:      cfg.Memory,
		Parallelism: cfg.Parallelism,
		SaltLength:  cfg.SaltLength,
	}
	u.HashGen = func(password string) (string, error) {
		return argon2.CreateHash(password, cfg.Secret, params)
	}
	u.HashVerify = func(password string, hash string) (bool, error) {
		return argon2.ComparePasswordAndHash(password, cfg.Secret, hash)
	}
}
