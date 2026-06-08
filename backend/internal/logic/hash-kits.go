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
	secret string
	params *argon2.Params
}

func (hu *HashUtils) HashGen(password string) (string, error) {
	return argon2.CreateHash(password, hu.secret, hu.params)
}
func (hu *HashUtils) HashVerify(password string, hash string) (bool, error) {
	return argon2.ComparePasswordAndHash(password, hu.secret, hash)
}

func NewHashUtils() *HashUtils {
	return &HashUtils{}
}
func (hu *HashUtils) BuildWithConfig(cfg *config.ArgonConfig) {
	hu.secret = cfg.Secret
	hu.params = &argon2.Params{
		Memory:      cfg.Memory,
		Iterations:  cfg.Iterations,
		Parallelism: cfg.Parallelism,
		SaltLength:  cfg.SaltLength,
		KeyLength:   cfg.KeyLength,
	}

}
