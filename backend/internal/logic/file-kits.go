package logic

import (
	"backend/internal/config"
	"crypto/aes"
	"crypto/cipher"
	"strings"
	"sync"
	"unicode"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/util/grand"
	"github.com/google/uuid"
)

// TODO: 使用sync.Pool复用文件数据缓冲区
var fileBufferPool = sync.Pool{New: func() any { return new([4096]byte) }}

type FileUtils struct {
	uploadDir      string
	rawUploadDir   string // 原始上传目录
	encryptDir     string // 加密文件存储目录
	tempDir        string // 临时文件存储目录 (缓存解密文件)
	unlockedDir    string // 解密文件的暂存目录
	rawUnlockedDir string // (原始目录)解密文件的暂存目录

	// 加密配置

	filetail []byte // 文件前导魔术字符
	keySize  int    // AES-GCM密钥长度
	nonce    []byte // AES-GCM初始向量
}

func (fu *FileUtils) ReadBytes(file *ghttp.UploadFile) (p []byte, err error) {
	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	p = make([]byte, file.Size)
	_, err = f.Read(p)
	return p, err
}

func (fu *FileUtils) SaveAndReadBytes(file *ghttp.UploadFile, path string) (p []byte, saveName string, err error) {
	// 先读取文件
	p, err = fu.ReadBytes(file)
	if err != nil {
		return nil, "", err
	}
	// 然后保存文件
	if saveName, err = file.Save(path, false); err != nil {
		return nil, "", err
	}
	return p, saveName, nil
}

func (fu *FileUtils) GenGCMCipher(setZeroAfterUse ...bool) (gcm cipher.AEAD, key []byte, err error) {
	var clearKey bool // 自动置零key
	if len(setZeroAfterUse) > 0 {
		clearKey = setZeroAfterUse[0]
	} else {
		clearKey = true
	}
	_ = clearKey

	// 生成密钥
	key = grand.B(fu.keySize)
	// key = []byte(grand.S(fu.keySize)) // 生成字符串密钥 (仅作调试)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	/*
		if clearKey {
				for i := 0; i < fu.keySize; i++ {
					key[i] = 0
				}
			}
	*/

	gcm, err = cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	return gcm, key, nil
}
func (fu *FileUtils) GenGCMCipherWithKey(key []byte) (gcm cipher.AEAD, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err = cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm, nil
}

func (fu *FileUtils) EncryptAndSaveFile(file *ghttp.UploadFile) (ciphertext []byte, key []byte, name string, err error) {
	// 读取文件
	plain, err := fu.ReadBytes(file)
	if err != nil {
		return nil, nil, "", err
	}

	// 获取加密器
	gcm, key, err := fu.GenGCMCipher()
	if err != nil {
		return nil, nil, "", err
	}

	// 加密数据
	tailCopy := make([]byte, len(fu.filetail)) // 附加数据
	copy(tailCopy, fu.filetail)
	ciphertext = gcm.Seal(nil, fu.nonce, plain, tailCopy)

	name = uuid.New().String() + ".enc"
	savePath := gfile.Join(fu.encryptDir, name)
	repeatCount := 1
	for gfile.Exists(savePath) { // 防止重名
		name = uuid.New().String() + ".enc"
		savePath = gfile.Join(fu.encryptDir, name)
		repeatCount++
	}

	err = gfile.PutBytes(savePath, ciphertext)
	return ciphertext, key, savePath, err
}

func (fu *FileUtils) EncryptBytes(plaintext []byte, key []byte, autoSetZero ...bool) (ciphertext []byte, randKey []byte, err error) {
	var enableSetZero bool // 自动置空明文字节
	if len(autoSetZero) > 0 {
		enableSetZero = autoSetZero[0]
	} else {
		enableSetZero = true
	}

	if len(key) < 16 { // 生成新的密钥
		key = grand.B(fu.keySize)
	}
	randKey = key

	// 生成加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	tailCopy := make([]byte, len(fu.filetail)) // 附加数据
	copy(tailCopy, fu.filetail)
	ciphertext = gcm.Seal(nil, fu.nonce, plaintext, tailCopy) // 加密数据

	if enableSetZero {
		Memclr(plaintext) // 置空明文字节
	}

	return ciphertext, randKey, nil
}

func (fu *FileUtils) DecryptBytes(bytes []byte, key []byte, autoMemclr ...bool) (plaintext []byte, err error) {
	var enableMemclr bool
	if len(autoMemclr) > 0 {
		enableMemclr = autoMemclr[0]
	} else {
		enableMemclr = true
	}

	// 获取加密器
	gcm, err := fu.GenGCMCipherWithKey(key)
	if err != nil {
		return nil, err
	}
	tailCopy := make([]byte, len(fu.filetail)) // 附加数据
	copy(tailCopy, fu.filetail)
	// 解密数据
	plaintext, err = gcm.Open(nil, fu.nonce, bytes, tailCopy)

	if enableMemclr {
		Memclr(bytes) // 置空密文数据
		Memclr(key)   // 置空密钥字节
	}
	if err != nil {
		return nil, err
	}
	return plaintext, err
}

func (fu *FileUtils) Delete(file *ghttp.UploadFile, path string) (err error) {
	name := file.Filename + ".enc"
	deletePath := gfile.Join(path, name)
	if gfile.Exists(deletePath) {
		err = gfile.RemoveFile(deletePath)
	} else {
		err = gerror.Newf("%s 文件不存在", name)
	}

	return
}
func (fu *FileUtils) DeleteItem(filename string) (err error) {
	deletePath := gfile.Join(fu.encryptDir, filename)
	if gfile.Exists(deletePath) {
		err = gfile.RemoveFile(deletePath)
	} else {
		err = gerror.Newf("%s 文件不存在", deletePath)
	}

	return
}

// isDigitOnly 检查字符串是否只包含数字
func (fu *FileUtils) isDigitOnly(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// ParseEncFilename 解析加密文件名，返回基本名和原始扩展名
// 文件名格式: <基本名>.<原始扩展名>.1.1.1.enc
// .1 部分是防重名标记，可能出现多次
// .enc 是固定后缀
func (fu *FileUtils) ParseEncFilename(filename string) (basename string, ext string, matched bool) {
	if !strings.HasSuffix(filename, ".enc") {
		return "", "", false
	}

	nameWithoutEnc := filename[:len(filename)-4]
	parts := strings.Split(nameWithoutEnc, ".")
	if len(parts) == 0 {
		return "", "", false
	}

	i := len(parts) - 1
	for i >= 0 && fu.isDigitOnly(parts[i]) {
		i--
	}

	if i < 0 {
		return "", "", false
	}

	remainingParts := parts[:i+1]

	if len(remainingParts) == 1 {
		return remainingParts[0], "", true
	}

	basename = remainingParts[0]
	ext = strings.Join(remainingParts[1:], ".")

	return basename, ext, true
}

func (fu *FileUtils) ItemExits(filename string) (expectPath string, err error) {
	expectPath = gfile.Join(fu.encryptDir, filename)
	if !gfile.Exists(expectPath) {
		return expectPath, gerror.Newf("文件 %s 不存在", gfile.Abs(expectPath))
	}
	return expectPath, nil
}

func (fu *FileUtils) ReadItemBytes(filename string) (bytes []byte, err error) {
	if _, err = fu.ItemExits(filename); err != nil {
		return nil, err
	}

	path := gfile.Join(fu.encryptDir, filename)
	bytes = gfile.GetBytes(path)
	return bytes, nil
}
func (fu *FileUtils) ReadAndDecryptItem(filename string, key []byte) (plaintext []byte, err error) {
	if _, err = fu.ItemExits(filename); err != nil {
		return nil, err
	}
	path := gfile.Join(fu.encryptDir, filename)
	bytes := gfile.GetBytes(path)
	plaintext, err = fu.DecryptBytes(bytes, key)
	Memclr(bytes) // 置空密文数据
	return plaintext, err
}
func (fu *FileUtils) DecryptAndSaveItem(filename string, key []byte) (del func() error, err error) {
	if _, err = fu.ItemExits(filename); err != nil {
		return del, err
	}
	basename, ext, valid := fu.ParseEncFilename(filename)
	if !valid {
		return del, gerror.Newf("文件名 %s 格式不正确", filename)
	}
	savePath := gfile.Join(fu.tempDir, basename+"."+ext)

	plainbytes, err := fu.ReadAndDecryptItem(filename, key)
	if err != nil {
		return del, err
	}

	err = gfile.PutBytes(savePath, plainbytes)
	if err != nil {
		return del, err
	}

	del = func() error {
		return gfile.RemoveFile(savePath)
	}
	return del, nil
}
func (fu *FileUtils) ItemDownload(r *ghttp.Response, filename string) (err error) {
	basename, ext, valid := fu.ParseEncFilename(filename)
	if !valid {
		return gerror.Newf("文件名 %s 格式不正确", filename)
	}
	temPath := gfile.Join(fu.tempDir, basename+"."+ext)

	r.ServeFileDownload(temPath)
	return nil
}

func NewFileUtils() *FileUtils {
	return &FileUtils{}
}
func (fu *FileUtils) BuildWithConfig(cfg *config.Item) {
	// 配置设置
	{
		fu.uploadDir = gfile.Join("business", "item", cfg.UploadDir)
		fu.rawUploadDir = cfg.UploadDir
		fu.encryptDir = gfile.Join("business", "item", cfg.EncryptedFileDir)
		fu.unlockedDir = gfile.Join("business", "item", cfg.UnlockedFileDir)
		fu.rawUnlockedDir = cfg.UnlockedFileDir
		fu.tempDir = gfile.Temp("business", "item", "decryption")

		// glog.Debugf(gctx.New(), "解密文件暂存目录为: %s", gfile.Abs(fu.unlockedDir))

		fu.filetail = []byte("ITEM") // 硬编码文件头
		fu.keySize = cfg.KeySize
		fu.nonce = []byte(cfg.Nonce)
	}
}
