package crypto

import (
	"backend/internal/logic"
	v2 "backend/internal/logic/crypto/v2"
	"backend/pkg/shamir/v3"
	"hash/crc32"
	"math/rand/v2"
	"testing"

	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/util/grand"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
)

func testEncryptionAndDecryption(t *testing.T, n int) {
	cu := v2.NewCryptoUtils()
	content := grand.B(n)
	checkcum1 := crc32.ChecksumIEEE(content)
	t.Logf("测试数据大小: %s, 校验和: %d", gfile.FormatSize(int64(n)), checkcum1)

	ciphertext, _ := cu.Encrypt(gctx.New(), nil, content)
	plaintext, err := cu.Decrypt(gctx.New(), nil, ciphertext)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	checkcum2 := crc32.ChecksumIEEE(plaintext)

	t.Logf("解密后数据大小: %s, 校验和: %d", gfile.FormatSize(int64(len(plaintext))), checkcum2)
	if checkcum1 != checkcum2 {
		t.Fatalf("校验和验证失败, 原始数据: %d, 解密后数据: %d", checkcum1, checkcum2)
	}
}

func testShareSplitAndRecover(t *testing.T) {
	var (
		cu        = v2.NewCryptoUtils()
		content   = grand.B(grand.Intn(16 * B))
		key       = grand.B(32)
		checksum1 = crc32.ChecksumIEEE(content)
	)
	t.Logf("测试数据大小: %s, 密钥长度: %d bytes", gfile.FormatSize(int64(len(content))), len(key))
	t.Logf("原始文件校验和为 %d", checksum1)

	// 加密文件
	ciphertext, _ := cu.Encrypt(gctx.New(), key, content)

	shares, _ := cu.SplitToJson(gctx.New(), key,
		rand.Uint32N(shamir.Prime), rand.Uint32N(shamir.Prime), rand.Uint32N(shamir.Prime),
	)
	var (
		ds   = shares[0]   // 只进行Base64编码
		as   = shares[1]   // AES-GCM (主密钥) + Base64编码
		rs   = shares[2]   // AES-GCM (随机密钥) + Base64编码
		code = grand.B(32) // rs的恢复码
	)
	// 编码环节
	ds, _ = logic.ChainEncode(gctx.New(), ds, logic.PEFBase64Encode)
	as, _ = cu.ChainBeforeEncrypt(gctx.New(), nil, as) // logic.PEFBase64Encode,
	// logic.PEFBase64Encode

	rs, _ = cu.ChainAfterEncrypt(gctx.New(), code, rs) // logic.PEFBase64Encode,
	// logic.PEFBase64Encode

	// 然后还原
	decryptedDs, err := logic.ChainDecode(gctx.New(), ds, logic.PDFBase64Decode)
	if err != nil {
		t.Errorf("设备份额解码失败: %v", err)
		return
	} // 先AES后Base64会导致密文失效
	// 先Base64后AES, 那解码的时候Base64就会不识别AES明文字节
	decryptedAs, err := cu.ChainAfterDecrypt(gctx.New(), nil, as) // logic.PDFBase64Decode,
	// logic.PDFBase64Decode,

	if err != nil {
		t.Errorf("Auth份额解密失败: %v", err)
		return
	}
	decryptedRs, err := cu.ChainBeforeDecrypt(gctx.New(), code, rs) // logic.PDFBase64Decode,
	// logic.PDFBase64Decode,

	if err != nil {
		t.Errorf("Recovery份额解密失败: %v", err)
		return
	}

	secret, err := cu.RecoverFromJson(gctx.New(), decryptedDs, decryptedAs, decryptedRs)
	if err != nil {
		t.Fatalf("份额还原失败: %v", err)
	}
	if string(key) != string(secret) {
		t.Fatalf("份额还原失败, 密钥不匹配, 期待 %s, 实际 %s", gbase64.EncodeToString(key), gbase64.EncodeToString(secret))
	}

	// 解密文件
	plaintext, err := cu.Decrypt(gctx.New(), secret, ciphertext)
	checksum2 := crc32.ChecksumIEEE(plaintext)
	t.Logf("解密后数据大小: %s, 校验和: %d", gfile.FormatSize(int64(len(plaintext))), checksum2)
	if checksum1 != checksum2 {
		t.Fatalf("解密文件验证失败, 原始校验和: %d, 解密后的检验和: %d", checksum1, checksum2)
	}
}

func testOnlyGCM(t *testing.T) {
	var (
		cu      = v2.NewCryptoUtils()
		content = grand.B(grand.Intn(1 * MB))
	)

	ciphertext, _ := cu.ChainBeforeEncrypt(gctx.New(), nil, content)
	plaintext, err := cu.ChainAfterDecrypt(gctx.New(), nil, ciphertext)
	if err != nil {
		t.Errorf("解密失败: %v", err)
		return
	}

	var (
		sum1 = crc32.ChecksumIEEE(content)
		sum2 = crc32.ChecksumIEEE(plaintext)
	)
	if sum1 != sum2 {
		t.Errorf("校验和验证失败, 原始数据校验和: %d, 解密后数据校验和: %d", sum1, sum2)
	}
}

func testBase64ThenGCM(t *testing.T) {
	var (
		cu      = v2.NewCryptoUtils()
		content = grand.B(grand.Intn(1 * MB))
	)

	ciphertext, _ := cu.ChainBeforeEncrypt(gctx.New(), nil, content, logic.PEFBase64Encode)
	plaintext, err := cu.ChainAfterDecrypt(gctx.New(), nil, ciphertext, logic.PDFBase64Decode)
	if err != nil {
		t.Errorf("解密失败: %v", err)
		return
	}

	var (
		sum1 = crc32.ChecksumIEEE(content)
		sum2 = crc32.ChecksumIEEE(plaintext)
	)
	if sum1 != sum2 {
		t.Errorf("校验和验证失败, 原始数据校验和: %d, 解密后数据校验和: %d", sum1, sum2)
	}
}

func testGCMThenBase64(t *testing.T) {
	var (
		cu      = v2.NewCryptoUtils()
		content = grand.B(grand.Intn(1 * MB))
	)

	ciphertext, _ := cu.ChainAfterEncrypt(gctx.New(), nil, content, logic.PEFBase64Encode)
	plaintext, err := cu.ChainBeforeDecrypt(gctx.New(), nil, ciphertext, logic.PDFBase64Decode)
	if err != nil {
		t.Errorf("解密失败: %v", err)
		return
	}

	var (
		sum1 = crc32.ChecksumIEEE(content)
		sum2 = crc32.ChecksumIEEE(plaintext)
	)
	if sum1 != sum2 {
		t.Errorf("校验和验证失败, 原始数据校验和: %d, 解密后数据校验和: %d", sum1, sum2)
	}
}

func testNRoundBase64(t *testing.T, rounds int) {
	encfns := make([]logic.EncodingFunc, 0, rounds)
	for range rounds {
		encfns = append(encfns, logic.PEFBase64Encode)
	}
	decfns := make([]logic.DecodingFunc, 0, rounds)
	for range rounds {
		decfns = append(decfns, logic.PDFBase64Decode)
	}

	var (
		content = grand.B(grand.Intn(1 * MB))
	)

	ciphertext, _ := logic.ChainEncode(nil, content, encfns...)
	plaintext, err := logic.ChainDecode(nil, ciphertext, decfns...)
	if err != nil {
		t.Errorf("解密失败: %v", err)
		return
	}

	var (
		sum1 = crc32.ChecksumIEEE(content)
		sum2 = crc32.ChecksumIEEE(plaintext)
	)
	if sum1 != sum2 {
		t.Errorf("校验和验证失败, 原始数据校验和: %d, 解密后数据校验和: %d", sum1, sum2)
	}
}

func testNRoundsGCM(t *testing.T, rounds int) {
	var (
		cu      = v2.NewCryptoUtils()
		content = grand.B(grand.Intn(1 * MB))
	)

	var ciphertext []byte
	var plaintext []byte
	var err error
	for range rounds {
		ciphertext, _ = cu.ChainBeforeEncrypt(gctx.New(), nil, content)
	}
	for range rounds {
		plaintext, err = cu.ChainAfterDecrypt(gctx.New(), nil, ciphertext)
		if err != nil {
			t.Errorf("解密失败: %v", err)
			return
		}
	}

	var (
		sum1 = crc32.ChecksumIEEE(content)
		sum2 = crc32.ChecksumIEEE(plaintext)
	)
	if sum1 != sum2 {
		t.Errorf("校验和验证失败, 原始数据校验和: %d, 解密后数据校验和: %d", sum1, sum2)
	}
}

func randCoordinate() uint32 {
	return rand.Uint32N(shamir.Prime)
}

// 测试多轮份额重建, 每次都多一个坐标
func testNRoundsResplit(t *testing.T, rounds int) {
	var (
		userCoordinate = randCoordinate()
		coordinates    = []uint32{userCoordinate, randCoordinate(), randCoordinate()}
	)

	var (
		cu            = v2.NewCryptoUtils()
		secret1       = cu.Key()
		lastTwoShares = func(input [][]byte) (output [][]byte) {
			return input[len(input)-2:]
		}
		shares  [][]byte
		secret2 []byte
		err     error
	)

	for i := range rounds {
		shares, _ = cu.SplitToJson(gctx.New(), secret1, coordinates...)
		secret2, err = cu.RecoverFromJson(gctx.New(), lastTwoShares(shares)...)
		if err != nil {
			t.Errorf("第%d轮 份额还原失败: %v", err, i+1)
			return
		}

		if string(secret1) != string(secret2) {
			t.Errorf("第%d轮 还原得到的密钥与原始密钥不一致, 期待 %s, 得到 %s", i+1, gbase64.EncodeToString(secret1), gbase64.EncodeToString(secret2))
			continue
		} else {
			t.Logf("第%d轮 密钥还原正确", i+1)
			secret1 = secret2 // 轮换密钥
			coordinates = append(coordinates, randCoordinate())
		}
	}

}

func TestCryptoUtilsV2(t *testing.T) {
	t.Logf("测试加密封装套具V2的加解密链路")

	// testEncryptionAndDecryption(t, 1*MB)
	// t.Run("份额加解密与还原 业务测试", testShareSplitAndRecover)

	// 没问题
	t.Run("纯AES-GCM加解密", testOnlyGCM)
	// base64.StdEncoding.Decode failed: illegal base64 data at input byte 0
	// t.Run("先Base64后AES-GCM加密, 先AES-GCM后Base64解密", testBase64ThenGCM)
	// 解密失败: cipher: message authentication failed
	// t.Run("先AES-GCM后Base64加密, 先Base64后AES-GCM解密", testGCMThenBase64)
	// 没问题
	t.Run("Base64编码3轮", func(t *testing.T) {
		testNRoundBase64(t, 3)
	})
	// 没问题
	t.Run("GCM加密3轮", func(t *testing.T) {
		testNRoundsGCM(t, 3)
	})

	// 多轮份额重建测试
	t.Run("份额重建3轮", func(t *testing.T) {
		testNRoundsResplit(t, 3)
	})
	t.Run("份额重建10轮", func(t *testing.T) {
		testNRoundsResplit(t, 10)
	})
}
