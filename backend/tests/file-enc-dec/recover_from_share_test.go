package file_enc_dec

import (
	"backend/internal/config"
	"backend/internal/logic"
	"backend/pkg/shamir/v3"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/gogf/gf/v2/os/gfile"
)

// 测试从前端拿到的Device Share和后端拿到的Auth Share能否用于恢复文件密钥
const (
	itemStorageDir = "./business/item/encrypted"
)

var (
	cu = logic.NewCryptoUtils()
	fu = logic.NewFileUtils()
)

func testRecoverSecretFromShare(t *testing.T, cu *logic.CryptoUtils, fu *logic.FileUtils, base64AuthShare string, deviceShare []byte, path string) {
	path = "【P7】僵王-泽野弘之.mp3.enc" // 文件路径

	// 编码 json + base64
	deviceShare = []byte(`eyJJbmRleCI6MzYxMDgxMTIxMywiVmFsdWVzIjpbNDA0NTA5MTcxNiw0MDgwODk5NTA1LDIzOTIyNTEwMjAsNDI1OTA2NjYwMywzODIyMDMzMDgxLDIxMjEwMjY1Nyw4Nzk2NTYwNDEsMzY1Mjk5MzM3MF19`)
	// 加密 json + base64 + AES-GCM主密钥加密
	base64AuthShare = `hg2Q6vglWY45/ypg5BfKBeXOP2euMck9In273OY9Th/LJDjQbg2fHCj0sLOhh6SfzRDOnZ9hacZK
NkyWb/alb6Tde8Y2zeH6X6NfSKdvehaTq53u9I4dZCdGKhotEjNP1ynXes9NQq0ubFc4WiUmEVUS
4EKVSEyrgjISKAP9d4k0GgV5EegMnoervnHhuoxYO4+qDLQJCrU19yOz/4OP1m61/3LDlg4n4s/P
tf9BIBnxoVgwjbUf9D2E1mQ=` // 最后这层显示还套了一层base64

	// base64 + json解码Device Share
	deviceShares, err := shamir.ShareFromBase64(string(deviceShare))
	if err != nil {
		t.Errorf("无法对设备份额进行 Base64 解码: %v", err)
		t.SkipNow()
	}
	fmt.Printf("设备份额的Base64解码结果为: \n%#v", deviceShares)

	authShare, err := base64.StdEncoding.DecodeString(base64AuthShare)
	if err != nil {
		t.Errorf("无法对认证份额进行 Base64 解码: %v", err)
		t.SkipNow()
	}
	fmt.Printf("认证份额的Base64解码结果为: \n%s\n", string(authShare))
	deAESGCMAuthShare, err := cu.SymmetricDecrypt(authShare, nil)
	if err != nil {
		t.Errorf("无法对认证份额进行 AES-GCM 解密: %v", err)
		t.SkipNow()
	}
	fmt.Printf("认证份额的AES-GCM解密结果为: \n%s", string(deAESGCMAuthShare))
	authShares, err := shamir.ShareFromBase64(string(deAESGCMAuthShare))
	if err != nil {
		t.Errorf("无法对认证份额进行 Json 解码: %v", err)
		t.SkipNow()
	}
	fmt.Printf("认证份额的Json解码结果为: \n%#v", authShares)

	secret := shamir.Recover([]shamir.Share{
		authShares, deviceShares,
	})
	fmt.Printf("恢复的密钥为: \n%x", secret)

	// 打开文件内容
	cipherdata := gfile.GetBytes(gfile.Join(itemStorageDir, path))
	if cipherdata == nil || len(cipherdata) == 0 {
		t.Errorf("无法打开文件: %s", path)
		t.SkipNow()
	}
	plaindata, err := fu.DecryptBytes(cipherdata, secret)
	if err != nil {
		t.Errorf("无法对文件进行解密: %v", err)
		t.SkipNow()
	}
	err = gfile.PutBytes("草稿.pdf", plaindata)
	if err != nil {
		t.Errorf("无法保存解密后的文件: %v", err)
	}
}

func TestRecoverSecretFromShare(t *testing.T) {
	// 先列出目录下的文件, 看能不能匹配出来
	absDir := gfile.Abs(itemStorageDir)
	fmt.Printf("扫描目录: %s\n", absDir)

	files, err := gfile.ScanDirFile(itemStorageDir, "*.enc", false)
	if err != nil {
		t.Errorf("无法扫描目录: %v", err)
		t.FailNow()
	}
	fmt.Printf("目录[%s]下有%d个加密文件\n", itemStorageDir, len(files))
	for i, file := range files {
		fmt.Printf("\t[%d]: %s\n", i+1, gfile.Basename(file))
	}

	/*
	  key: "dG0@cA0/aC6{aA6(jA0*bF0`rC1+jA4+" # AES-GCM密钥 (32字节) # 不再使用
	  key_size: 32
	  nonce: "dV1.hD0@bD2&" # AES-GCM需要的Nonce值
	*/
	cu.BuildWithConfig(&config.Item{
		ItemSeal: config.ItemSeal{
			ShareKey: "dG0@cA0/aC6{aA6(jA0*bF0`rC1+jA4+",
			KeySize:  32,
			Nonce:    "dV1.hD0@bD2&",
		},
	})
	fu.BuildWithConfig(&config.Item{
		ItemSeal: config.ItemSeal{
			ShareKey: "dG0@cA0/aC6{aA6(jA0*bF0`rC1+jA4+",
			KeySize:  32,
			Nonce:    "dV1.hD0@bD2&",
		},
	})

	testRecoverSecretFromShare(t, cu, fu, "", nil, "")
}
