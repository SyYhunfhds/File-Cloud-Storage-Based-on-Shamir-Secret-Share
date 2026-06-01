package logic

import (
	"backend/internal/config"
	"crypto/aes"
	"crypto/cipher"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/grand"
)

// TODO: 使用sync.Pool复用文件数据缓冲区
var fileBufferPool = sync.Pool{New: func() any { return new([4096]byte) }}

type FileUtils struct {
	uploadDir    string
	rawUploadDir string // 原始上传目录
	encryptDir   string // 加密文件存储目录

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
func (fu *FileUtils) EncryptAndSaveFile(file *ghttp.UploadFile, path string) (key []byte, name string, err error) {
	// 读取文件
	plain, err := fu.ReadBytes(file)
	if err != nil {
		return nil, "", err
	}

	// 获取加密器
	gcm, key, err := fu.GenGCMCipher()
	if err != nil {
		return nil, "", err
	}

	// 加密数据
	tailCopy := make([]byte, len(fu.filetail)) // 附加数据
	copy(tailCopy, fu.filetail)
	ciphertext := gcm.Seal(plain[:0], fu.nonce, plain, tailCopy) // 存在BUG: 会把原始字节也写进文件里
	// glog.Debugf(gctx.New(), "encrypted data: \n%s", string(ciphertext))

	name = file.Filename + ".enc"
	savePath := gfile.Join(path, name)
	repeatCount := 1
	for gfile.Exists(savePath) { // 防止重名
		name = file.Filename + "." + gconv.String(repeatCount) + ".enc"
		savePath = gfile.Join(path, name)
		repeatCount++
	}

	err = gfile.PutBytes(savePath, ciphertext)
	return key, name, err
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

func NewFileUtils() *FileUtils {
	return &FileUtils{}
}
func (fu *FileUtils) BuildWithConfig(cfg *config.Item) {
	// 配置设置
	{
		fu.uploadDir = gfile.Join("business", "item", cfg.UploadDir)
		fu.rawUploadDir = cfg.UploadDir

		fu.filetail = []byte("ITEM") // 硬编码文件头
		fu.keySize = cfg.KeySize
		fu.nonce = []byte(cfg.Nonce)
	}
}
