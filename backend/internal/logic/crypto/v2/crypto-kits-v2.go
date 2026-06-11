package v2

import (
	"backend/internal/config"
	"backend/internal/logic"
	"backend/pkg/shamir/v3"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"errors"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/grand"
)

var (
	debugNonce = []byte("ABCDABCDABCD") // 凑够12字节
)

type RegularOption struct {
	MasterKey string // 主密钥
	KeySize   int    // (随机)密钥长度
	IVSize    int    // IV长度 / Nonce长度
	AADSize   int    // 附加字节长度
}

func DefaultRegularOption() RegularOption {
	opt := RegularOption{
		MasterKey: grand.S(32),
		KeySize:   32,
		IVSize:    12, // GCM Nonce长度
		AADSize:   32,
	}
	return opt
}

type ShamirOption struct {
	Threshold int // 份额上限
}

func DefaultShamirOption() ShamirOption {
	return ShamirOption{
		Threshold: 2,
	}
}

type Option struct {
	RegularOption
	ShamirOption
}

type CryptoUtils struct {
	option Option
}

type OptionFunc func(o *Option)

func NewICryptoUtils(opts ...OptionFunc) logic.ICryptoUtils {
	return NewCryptoUtils(opts...)
}
func NewCryptoUtils(opts ...OptionFunc) *CryptoUtils {
	utils := &CryptoUtils{
		option: Option{
			RegularOption: DefaultRegularOption(),
			ShamirOption:  DefaultShamirOption(),
		},
	}

	for _, opt := range opts {
		opt(&utils.option)
	}
	return utils
}

func (c *CryptoUtils) BuildWithConfig(cfg config.Item) {
	c.option.MasterKey = cfg.ItemSeal.ShareKey
}

func (c *CryptoUtils) StringKey(n ...int) string {
	var size int
	if len(n) > 0 {
		size = n[0]
	} else {
		size = c.option.KeySize
	}

	return grand.S(size)
}

func (c *CryptoUtils) Key(n ...int) []byte {
	var size int
	if len(n) > 0 {
		size = n[0]
	} else {
		size = c.option.KeySize
	}

	return grand.B(size)
}

// Encrypt 使用GCM算法进行加密, 附加随机Nonce和随机AAD
func (c *CryptoUtils) Encrypt(ctx context.Context, key []byte, plaintext []byte) (ciphertext []byte, err error) {
	if len(key) < c.option.KeySize {
		key = []byte(c.option.MasterKey)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 生成Nonce并附加到密文前部
	n1 := grand.B(gcm.NonceSize())
	head := make([]byte, len(n1))
	copy(head, n1)
	// 随机aad
	aad1 := grand.B(c.option.AADSize)
	ciphertext = gcm.Seal(head, n1, plaintext, aad1) // AAD不会自动附加到尾部
	// 附加aad到尾部
	ciphertext = append(ciphertext, aad1...)

	return ciphertext, nil
}

func (c *CryptoUtils) ChainAfterEncrypt(ctx context.Context, key []byte, plaintext []byte, processes ...logic.EncodingFunc) (ciphertext []byte, err error) {
	ciphertext, err = c.Encrypt(ctx, key, plaintext)
	if err != nil {
		return nil, err
	}

	for _, fn := range processes {
		ciphertext, err = fn(ctx, ciphertext)
		if err != nil {
			return nil, err
		}
	}

	return
}

func (c *CryptoUtils) ChainBeforeEncrypt(ctx context.Context, key []byte, plaintext []byte, processes ...logic.EncodingFunc) (ciphertext []byte, err error) {
	for _, fn := range processes {
		ciphertext, err = fn(ctx, ciphertext)
		if err != nil {
			return nil, err
		}
	}

	ciphertext, err = c.Encrypt(ctx, key, plaintext)
	if err != nil {
		return nil, err
	}

	return
}

// Decrypt 使用AES-GCM算法和32位字节主密钥(如果没有额外提供密钥)进行解密, 会自动提取Nonce和AAD
func (c *CryptoUtils) Decrypt(ctx context.Context, key []byte, ciphertext []byte) (plaintext []byte, err error) {
	if len(key) < c.option.KeySize {
		key = []byte(c.option.MasterKey)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Nonce + Payload [+ Tag] + AAD
	// 提取nonce
	nonce := ciphertext[:gcm.NonceSize()]
	payload := ciphertext[gcm.NonceSize() : len(ciphertext)-c.option.AADSize]
	// 提取aad
	aad := ciphertext[len(ciphertext)-c.option.AADSize:]

	plaintext, err = gcm.Open(nil, nonce, payload, aad)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (c *CryptoUtils) ChainDecrypt(ctx context.Context, key []byte, ciphertext []byte, processes ...logic.DecodingFunc) (plaintext []byte, err error) {
	plaintext, err = c.Decrypt(ctx, key, ciphertext)
	if err != nil {
		return nil, err
	}

	for _, fn := range processes {
		plaintext, err = fn(ctx, plaintext)
		if err != nil {
			return nil, err
		}
	}

	return
}

func (c *CryptoUtils) ChainBeforeDecrypt(ctx context.Context, key []byte, ciphertext []byte, processes ...logic.DecodingFunc) (plaintext []byte, err error) {
	for _, fn := range processes {
		plaintext, err = fn(ctx, plaintext)
		if err != nil {
			return nil, err
		}
	}

	plaintext, err = c.Decrypt(ctx, key, ciphertext)
	if err != nil {
		return nil, err
	}

	return
}

func (c *CryptoUtils) ChainAfterDecrypt(ctx context.Context, key []byte, ciphertext []byte, processes ...logic.DecodingFunc) (plaintext []byte, err error) {
	plaintext, err = c.Decrypt(ctx, key, ciphertext)
	if err != nil {
		return nil, err
	}

	for _, fn := range processes {
		plaintext, err = fn(ctx, plaintext)
		if err != nil {
			return nil, err
		}
	}

	return
}

func (c *CryptoUtils) Recover(ctx context.Context, shares ...shamir.Share) ([]byte, error) {
	if len(shares) < c.option.Threshold {
		return nil, errors.New("份额未达到门限, 无法还原秘密")
	}

	return shamir.Recover(shares), nil
}
func (c *CryptoUtils) RecoverFromList(ctx context.Context, shares []shamir.Share) ([]byte, error) {
	if len(shares) < c.option.Threshold {
		return nil, errors.New("份额未达到门限, 无法还原秘密")
	}

	return shamir.Recover(shares), nil
}

func (c *CryptoUtils) Split(ctx context.Context, secret []byte, coordinate ...uint32) ([]shamir.Share, error) {
	return shamir.Split(secret, c.option.Threshold, coordinate)
}
func (c *CryptoUtils) SplitToJson(ctx context.Context, secret []byte, coordinate ...uint32) ([][]byte, error) {
	shares, err := c.Split(ctx, secret, coordinate...)
	if err != nil {
		return nil, err
	}

	output := make([][]byte, len(shares))
	for i := range shares {
		output[i], _ = gjson.Encode(shares[i])
	}

	return output, nil
}
func (c *CryptoUtils) RecoverFromJson(ctx context.Context, data ...[]byte) ([]byte, error) {
	shares := make([]shamir.Share, len(data))
	for i := range data {
		err := gjson.DecodeTo(data[i], &shares[i])
		if err != nil {
			return nil, gerror.Wrapf(err, "第 %d 个份反序列化失败", i+1)
		}
	}

	return c.RecoverFromList(ctx, shares)
}
