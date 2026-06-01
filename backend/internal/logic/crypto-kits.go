package logic

import (
	"backend/internal/config"
	"backend/pkg/shamir/v3"
	"crypto/aes"
	"crypto/cipher"
	"math/rand/v2"

	"github.com/gogf/gf/v2/util/grand"
)

type ShamirOption struct {
	// 总共需要分割多少份额; 除了用户坐标之外的其他坐标都将使用随机数
	//
	// 暂时没有被使用
	TotalShares int
	// 份额上限
	Threshold int
}
type AESOption struct {
	// (AES-GCM)密钥长度, 默认为32
	key []byte
	// AES-GCM随机密钥长度, 固定为32字节
	keyLen          int
	additionalBytes []byte // 附加在密文后面的数据
}

// DefaultShamirOption 默认使用2-of-3模型
func DefaultShamirOption() ShamirOption {
	return ShamirOption{
		TotalShares: 3,
		Threshold:   2,
	}
}
func DefaultAESOption() AESOption {
	return AESOption{
		additionalBytes: []byte("GCM"),
		keyLen:          32,
	}
}

type CryptoUtils struct {
	so ShamirOption
	ao AESOption
}

func NewCryptoUtils() *CryptoUtils {
	return &CryptoUtils{
		so: DefaultShamirOption(),
		ao: DefaultAESOption(),
	}
}

func (cu *CryptoUtils) BuildWithConfig(ic *config.Item) {
	cu.ao.key = []byte(ic.ShareKey)
}

func (cu *CryptoUtils) SplitShare(secret []byte, userID uint32) (
	deviceShare string, authShare []byte, recoveryShare []byte, err error,
) {
	coordinates := []uint32{
		userID, rand.Uint32N(shamir.Prime), rand.Uint32N(shamir.Prime),
	}
	shares, err := shamir.Split(secret, cu.so.Threshold, coordinates)
	if err != nil {
		return
	}

	{
		deviceShare, _ = shares[0].ToBase64()
		authShare, _ = shares[1].ToBase64Bytes()
		recoveryShare, _ = shares[2].ToBase64Bytes()
	}

	return
}
func (cu *CryptoUtils) RecoverShare(shares ...shamir.Share) []byte {
	return shamir.Unpad(shamir.Recover(shares))
}

// SymmetricEncrypt
//
// 使用AES-GCM进行加密; 可接受外部密钥, 默认启用字节数组清空; 若传入的密钥为空, 则使用配置文件的密钥
func (cu *CryptoUtils) SymmetricEncrypt(plaintext []byte, key []byte, autoMemclr ...bool) (ciphertext []byte, err error) {
	if key == nil {
		key = cu.ao.key
	}
	var enableMemclr bool
	if len(autoMemclr) > 0 {
		enableMemclr = autoMemclr[0]
	} else {
		enableMemclr = true
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	nonce := grand.B(gcm.NonceSize())
	ciphertext = gcm.Seal(nonce, nonce, plaintext, cu.ao.additionalBytes)
	// cipher = nonce + encrypted + addition

	if enableMemclr {
		Memclr(plaintext)
	}

	return
}

func (cu *CryptoUtils) SymEncryptWithRandKey(plaintext []byte) (ciphertext []byte, key string, err error) {
	key = grand.S(cu.ao.keyLen)
	ciphertext, err = cu.SymmetricEncrypt(plaintext, []byte(key))
	if err != nil {
		return nil, "", err
	}

	return
}
func (cu *CryptoUtils) SymmetricDecrypt(ciphertext []byte, key []byte, autoMemclr ...bool) (plaintext []byte, err error) {
	if key == nil {
		key = cu.ao.key
	}
	var enableMemclr bool
	if len(autoMemclr) > 0 {
		enableMemclr = autoMemclr[0]
	} else {
		enableMemclr = true
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	nonce := ciphertext[:gcm.NonceSize()]
	plaintext, err = gcm.Open(nil, nonce, ciphertext[gcm.NonceSize():], cu.ao.additionalBytes)
	if enableMemclr {
		Memclr(ciphertext)
	}
	if err != nil {
		return
	}

	return
}
