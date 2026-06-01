package file_enc_dec

import (
	"backend/internal/config"
	"backend/internal/logic"
	"backend/pkg/shamir/v3"
	"math/rand/v2"
	"testing"

	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/util/grand"
)

var ic = config.Item{
	ItemSeal: config.ItemSeal{
		ShareKey: "THE-KEY-IS--HERE",
		KeySize:  32,
		Nonce:    "ABCDABCDABCD", // 12字节nonce
	},
	ItemStorage: config.ItemStorage{},
}

func testEncryptAndDecryptFile(t *testing.T, filesize int, cu *logic.CryptoUtils, fu *logic.FileUtils) {
	t.Logf("测试文件大小为: %s | 测试内容: 认证份额不变", gfile.FormatSize(int64(filesize)))

	plaintext := grand.B(filesize)
	cipherdata, enckey, err := fu.EncryptBytes(plaintext, nil, false)
	if err != nil {
		t.Errorf("加密失败: %v", err)
		t.Skip()
	}

	deviceShare, authShare, _, err := cu.SplitShare(enckey, rand.Uint32N(shamir.Prime))
	// 模拟 DeviceShare先AES-GCM加密再AES-GCM解密
	var deviceKey = grand.B(32)
	encryptedDeviceShare, _ := cu.SymmetricEncrypt([]byte(deviceShare), deviceKey, false)
	decryptedDeviceShare, _ := cu.SymmetricDecrypt(encryptedDeviceShare, deviceKey)
	if deviceShare != string(decryptedDeviceShare) {
		t.Errorf("设备份额AES-GCM加密解密还原失败")
	}
	rawDeviceShare, err := shamir.ShareFromBase64Bytes(decryptedDeviceShare)
	if err != nil {
		t.Errorf("设备份额还原失败: %v", err)
		t.Skip()
	}
	// 模拟AuthShare按同样方式再来一遍
	encryptedAuthShare, _ := cu.SymmetricEncrypt(authShare, nil, false)
	decryptedAuthShare, _ := cu.SymmetricDecrypt(encryptedAuthShare, nil)
	if deviceShare != string(decryptedDeviceShare) {
		t.Errorf("认证份额AES-GCM加密解密还原失败")
	}
	rawAuthShare, err := shamir.ShareFromBase64Bytes(decryptedAuthShare)
	if err != nil {
		t.Errorf("认证份额序列化失败: %v", err)
		t.Skip()
	}

	// 模拟恢复份额
	recoveredKey := cu.RecoverShare(rawDeviceShare, rawAuthShare)
	if string(recoveredKey) != string(enckey) {
		t.Errorf("密钥还原失败")
		t.Skip()
	}

	// 模拟校验并解密文件
	recoveredPlaintext, err := fu.DecryptBytes(cipherdata, recoveredKey)
	if err != nil {
		t.Errorf("解密失败: %v", err)
	}
	if string(plaintext) != string(recoveredPlaintext) {
		t.Errorf("解密内容不匹配")
	}
}

func TestEncryptAndDecryptFile(t *testing.T) {
	var cu = logic.NewCryptoUtils()
	var fu = logic.NewFileUtils()

	cu.BuildWithConfig(&ic)
	fu.BuildWithConfig(&ic)

	for exp := 1; exp < 24; exp++ { // 至多到16MB
		filesize := 1 << exp
		testEncryptAndDecryptFile(t, filesize, cu, fu)
	}
}
