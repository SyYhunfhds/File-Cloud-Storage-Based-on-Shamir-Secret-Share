package file_enc_dec

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"os"
	"testing"
)

const testPdfInputPath = "testdata.pdf"

func randKeyAndIV() (gcm cipher.AEAD, key []byte, iv []byte) {
	key = make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, key) // 生成随机密钥

	block, _ := aes.NewCipher(key)
	gcm, _ = cipher.NewGCM(block)

	iv = make([]byte, gcm.NonceSize())  // 生成随机Nonce
	_, _ = io.ReadFull(rand.Reader, iv) // 填充Nonce
	return
}

func TestFileEncAndDecCycle(t *testing.T) {
	filedata, err := os.ReadFile(testPdfInputPath)
	if err != nil {
		t.Errorf("无法读取文件输入, 输入的路径为: %v", testPdfInputPath)
		return
	}

	gcm, _, encNonce := randKeyAndIV()
	// 加密
	ciphertext := gcm.Seal(encNonce, encNonce, filedata, nil)

	// 解密
	// 拆分出nonce和密文
	decNonce := ciphertext[:len(encNonce)]
	todecCiphertext := ciphertext[len(encNonce):]
	plaintext, err := gcm.Open(nil, decNonce, todecCiphertext, nil)
	if err != nil {
		t.Errorf("无法解密数据: %v", err)
		return
	}

	// 看中间16个字节就可以了
	if string(filedata[16:32]) != string(plaintext[16:32]) {
		t.Errorf("解密后的数据与原始数据不一致")
	}
}
