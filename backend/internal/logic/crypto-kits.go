package logic

import (
	"backend/internal/config"
	"backend/pkg/shamir/v3"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"math/rand/v2"

	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/grand"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

	AADSize int // 附加数据长度, 默认为32
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
		key:             grand.B(32),
		additionalBytes: grand.B(32),
		keyLen:          32,
		AADSize:         32,
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

func (cu *CryptoUtils) ResplitShare(secret []byte, userID uint32, otherUsers []uint32, autoSetZero ...bool) (
	deviceShare string, authShare []byte, recoveryShare []byte, otherShares [][]byte, err error,
) {
	var enableClearKey bool
	if len(autoSetZero) > 0 {
		enableClearKey = autoSetZero[0]
	} else {
		enableClearKey = true
	}

	coordinates := []uint32{
		userID, rand.Uint32N(shamir.Prime), rand.Uint32N(shamir.Prime),
	}
	coordinates = append(coordinates, otherUsers...)

	shares, err := shamir.Split(secret, cu.so.Threshold, coordinates)
	if err != nil {
		return "", nil, nil, nil, err
	}

	{
		deviceShare, _ = shares[0].ToBase64()
		authShare, _ = shares[1].ToBase64Bytes()
		recoveryShare, _ = shares[2].ToBase64Bytes()
	}
	// 处理其他用户的份额
	otherShares = make([][]byte, len(otherUsers))
	for i := 3; i < len(shares); i++ {
		otherShares[i-3], _ = shares[i].ToBase64Bytes()
	}

	if enableClearKey {
		Memclr(secret)
	}
	return
}

func (cu *CryptoUtils) RecoverShare(shares ...shamir.Share) []byte {
	return shamir.Unpad(shamir.Recover(shares))
}

// SymmetricEncrypt
//
// 使用AES-GCM进行加密; 可接受外部密钥, 默认启用字节数组清空; 若传入的密钥为空, 则使用配置文件的密钥

/*
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
	// glog.Debugf(gctx.New(), "[Encryption] encrypted data: \n%x", ciphertext)

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

	// glog.Debugf(gctx.New(), "[Decryption] encrypted data: \n%x", ciphertext)

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

*/

func (cu *CryptoUtils) EncryptAuthShare(ctx context.Context, share []byte, autoClear ...bool) (ciphertext []byte, err error) {
	// 先给注销了
	return share, nil

	ctx, span := gtrace.NewSpan(ctx, "crypto_utils/encrypt_auth_share")
	defer span.End()

	var enableClear bool
	if len(autoClear) > 0 {
		enableClear = autoClear[0]
	} else {
		enableClear = true
	}
	span.SetAttributes(
		attribute.Bool("auth_share.auto_clear.enable", enableClear),
		attribute.Int("auth_share.ciphertext_length", len(ciphertext)),
		attribute.String("auth_share.ciphertext.base64encode", gbase64.EncodeToString(ciphertext)),
		attribute.Int("crypto_utils.master_key.length", len(cu.ao.key)),
		// TODO: 移除高危Span (服务器主密钥)
		attribute.String("crypto_utils.master_key.base64encode", gbase64.EncodeToString(cu.ao.key)),
	)

	block, err := aes.NewCipher(cu.ao.key)
	if err != nil {
		span.SetStatus(codes.Error, "无法生成AES密钥块")
		span.RecordError(err)

		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		span.SetStatus(codes.Error, "无法生成GCM")
		span.RecordError(err)

		return nil, err
	}

	// 生成Nonce
	nonce := grand.B(gcm.NonceSize())
	copiedNonce := make([]byte, len(nonce))
	copy(copiedNonce, nonce)
	// 生成AAD
	aad := grand.B(cu.ao.AADSize) // 32位长度字节
	span.SetAttributes(
		attribute.String("auth_share.nonce.base64encode", gbase64.EncodeToString(copiedNonce)),
		attribute.Int("auth_share.nonce.length", len(copiedNonce)),
		attribute.String("auth_share.aad.base64encode", gbase64.EncodeToString(aad)),
		attribute.Int("auth_share.aad.length", len(aad)),
	)
	ciphertext = gcm.Seal(nil, nonce, share, aad)
	ciphertext = append(copiedNonce, ciphertext...) // 将加密数据追加到 nonce 后面
	ciphertext = append(ciphertext, aad...)         // 将AAD追加到密文尾部, 供解密时取用

	// TODO: 移除敏感Span
	span.SetAttributes(
		attribute.String("auth_share.ciphertext.base64encode", gbase64.EncodeToString(ciphertext)),
		attribute.Int("auth_share.ciphertext.length", len(ciphertext)),
	)

	if enableClear {
		Memclr(share)
		Memclr(aad)
		Memclr(nonce)
	}
	span.SetStatus(codes.Ok, "加密成功")
	return
}
func (cu *CryptoUtils) DecryptAuthShare(ctx context.Context, ciphertext []byte, autoClear ...bool) (share []byte, err error) {
	// TODO: 暂时移除加解密流程
	return share, nil

	ctx, span := gtrace.NewSpan(ctx, "crypto_utils/decrypt_auth_share")
	defer span.End()

	var enableClear bool
	if len(autoClear) > 0 {
		enableClear = autoClear[0]
	} else {
		enableClear = true
	}

	span.SetAttributes(
		attribute.Bool("auth_share.auto_clear.enable", enableClear),
		attribute.Int("auth_share.ciphertext_length", len(ciphertext)),
		attribute.String("auth_share.ciphertext.base64encode", gbase64.EncodeToString(ciphertext)),
		attribute.Int("crypto_utils.master_key.length", len(cu.ao.key)),
		// TODO: 移除高危Span (服务器主密钥)
		attribute.String("crypto_utils.master_key.base64encode", gbase64.EncodeToString(cu.ao.key)),
	)

	block, err := aes.NewCipher(cu.ao.key)
	if err != nil {
		span.SetStatus(codes.Error, "无法生成AES密钥块")
		span.RecordError(err)

		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		span.SetStatus(codes.Error, "无法生成GCM")
		span.RecordError(err)

		return nil, err
	}

	// 校验长度
	if len(ciphertext) < gcm.NonceSize()+cu.ao.AADSize {
		span.SetStatus(codes.Error, "ciphertext length is not enough")

		return nil, gerror.New("ciphertext length is not enough")
	}
	// 计算Nonce、密文数据和aad部分的起始位置
	var (
		nonceBegin = 0
		nonceEnd   = gcm.NonceSize()
		dataBegin  = nonceEnd
		dataEnd    = len(ciphertext) - cu.ao.AADSize
		aadBegin   = len(ciphertext) - cu.ao.AADSize
		aadEnd     = len(ciphertext)
	)

	// 提取Nonce
	nonce := ciphertext[nonceBegin:nonceEnd]
	// 提取尾部的aad
	aad := ciphertext[aadBegin:aadEnd]

	span.SetAttributes(
		attribute.String("auth_share.nonce.base64encode", gbase64.EncodeToString(nonce)),
		attribute.Int("auth_share.nonce.length", len(nonce)),
		attribute.String("auth_share.aad.base64encode", gbase64.EncodeToString(aad)),
		attribute.Int("auth_share.aad.length", len(aad)),
		attribute.String("auth_share.ciphertext.base64encode", gbase64.EncodeToString(ciphertext)),
		attribute.Int("auth_share.ciphertext.length", len(ciphertext)),
	)

	share, err = gcm.Open(nil, nonce, ciphertext[dataBegin:dataEnd], aad)
	if err != nil {
		span.SetStatus(codes.Error, "无法解密AuthShare")
		span.SetAttributes(attribute.String("auth_share.decryption.error", err.Error()))

		return nil, err
	}

	span.SetAttributes(
		attribute.String("auth_share.base64encode", gbase64.EncodeToString(share)),
		attribute.Int("auth_share.length", len(share)),
	)

	if enableClear {
		Memclr(ciphertext)
	}
	return
}

func (cu *CryptoUtils) EncryptRecoveryShare(share []byte, key []byte, autoClear ...bool) (ciphertext []byte, nk string, err error) {
	if len(key) < cu.ao.keyLen {
		nk = grand.S(cu.ao.keyLen)
		key = []byte(nk)
	} else {
		nk = string(key)
	}

	// TODO 暂时移除加解密流程
	return share, nk, nil

	_, span := gtrace.NewSpan(context.Background(), "crypto_utils/encrypt_recovery_share")
	defer span.End()

	var enableClear bool
	if len(autoClear) > 0 {
		enableClear = autoClear[0]
	} else {
		enableClear = true
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", err
	}

	// 生成Nonce
	nonce := grand.B(gcm.NonceSize())
	copiedNonce := make([]byte, len(nonce))
	copy(copiedNonce, nonce)
	// 生成AAD
	aad := grand.B(cu.ao.AADSize) // 32位长度字节
	ciphertext = gcm.Seal(nil, nonce, share, aad)
	ciphertext = append(copiedNonce, ciphertext...) // 将加密数据追加到 nonce 后面
	ciphertext = append(ciphertext, aad...)         // 将AAD追加到密文尾部, 供解密时取用

	if enableClear {
		Memclr(share)
	}
	return
}
func (cu *CryptoUtils) DecryptRecoveryShare(ciphertext []byte, key []byte, autoClear ...bool) (share []byte, err error) {
	if len(key) < cu.ao.keyLen {
		return nil, gerror.Newf("key length is not enough, should be %d bytes", cu.ao.keyLen)
	}

	var enableClear bool
	if len(autoClear) > 0 {
		enableClear = autoClear[0]
	} else {
		enableClear = true
	}

	return share, nil

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 计算Nonce、密文数据和aad部分的起始位置
	var (
		nonceBegin = 0
		nonceEnd   = gcm.NonceSize()
		dataBegin  = nonceEnd
		dataEnd    = len(ciphertext) - cu.ao.AADSize
		aadBegin   = len(ciphertext) - cu.ao.AADSize
		aadEnd     = len(ciphertext)
	)

	// 提取Nonce
	nonce := ciphertext[nonceBegin:nonceEnd]
	// 提取尾部的aad
	aad := ciphertext[aadBegin:aadEnd]
	share, err = gcm.Open(nil, nonce, ciphertext[dataBegin:dataEnd], aad)
	if err != nil {
		return nil, err
	}

	if enableClear {
		Memclr(ciphertext)
		Memclr(key)
	}
	return
}

func (cu *CryptoUtils) EncryptMemberShare(shares [][]byte, autoClear ...bool) (res [][]byte, err error) {
	var enableClear bool
	if len(autoClear) > 0 {
		enableClear = autoClear[0]
	} else {
		enableClear = true
	}

	res = make([][]byte, 0, len(shares))
	for _, share := range shares {
		ciphertext, err := cu.EncryptAuthShare(gctx.New(), share, enableClear)
		if err != nil {
			return nil, err
		}
		res = append(res, ciphertext)
	}

	return
}
func (cu *CryptoUtils) DecryptMemberShare(ciphertexts [][]byte, autoClear ...bool) (shares [][]byte, err error) {
	var enableClear bool
	if len(autoClear) > 0 {
		enableClear = autoClear[0]
	} else {
		enableClear = true
	}

	shares = make([][]byte, 0, len(ciphertexts))
	for _, ciphertext := range ciphertexts {
		share, err := cu.DecryptAuthShare(gctx.New(), ciphertext, enableClear)
		if err != nil {
			return nil, err
		}
		shares = append(shares, share)
	}

	return
}
