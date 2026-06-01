package v2

import (
	"backend/internal/dao"
	"backend/internal/model/do"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"time"

	"github.com/gogf/gf/v2/container/gqueue"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/grand"
)

var (
	ErrCannotReadDir = gerror.NewCodef(
		gcode.CodeInternalError,
		"无法打开目录",
	)
	ErrDirNotfound = gerror.New(
		"目录未添加",
	)
	ErrFileNotfound = gerror.New(
		"文件未找到",
	)
)

type FileInfo struct {
	Filename string `json:"filename" dc:"文件名"`
	Filesize string `json:"size" dc:"文件大小"` // 由gfile库的SizeFormat一步到位拿到可读的文件大小

	UpdatedAt time.Time `json:"updated_at" dc:"创建时间"`
}

// Watcher
//
// 封装gfsnotify方法, 用于监控上传文件、加密文件和解密文件
type Watcher struct {
	option WatcherOption

	ctx    context.Context
	cancel context.CancelFunc
	log    glog.ILogger
}

func New(options ...WatcherOptionFunc) *Watcher {
	watcher := &Watcher{
		option: DefaultOption(),
		log:    glog.New(),
	}

	for _, option := range options {
		option(watcher)
	}

	return watcher
}

func (w *Watcher) Run() error {

	return nil
}

func (w *Watcher) Export(page, size int) (list []FileInfo, err error) {
	filepaths, err := gfile.ScanDirFile(w.option.path, w.option.pattern, false)
	if err != nil {
		w.log.Errorf(w.ctx, "无法检索目录[%v]下的文件: %v", w.option.path, err)
		return nil, err
	}

	list = make([]FileInfo, 0, size)
	offset := (page - 1) * size

	for i := offset; i < len(filepaths) && i < offset+size; i++ {
		list = append(list, FileInfo{
			Filename: gfile.Basename(filepaths[i]),
			Filesize: gfile.ReadableSize(filepaths[i]),

			UpdatedAt: gfile.MTime(filepaths[i]),
		})
	}
	return
}

func (w *Watcher) ExportAll() (list []FileInfo, err error) {
	filepaths, err := gfile.ScanDirFile(w.option.path, w.option.pattern, false)
	if err != nil {
		w.log.Errorf(w.ctx, "无法检索目录下的文件: %v", err)
		return nil, err
	}

	list = make([]FileInfo, 0, len(filepaths))
	for _, path := range filepaths {
		list = append(list, FileInfo{
			Filename: gfile.Basename(path),
			Filesize: gfile.ReadableSize(path),

			UpdatedAt: gfile.MTime(path),
		})
	}

	return
}

func (w *Watcher) GetFileBytes(ctx context.Context, path string) (string, []byte, error) {
	path = gfile.Join(w.option.path, path)
	absPath := gfile.RealPath(path)
	if !gfile.Exists(absPath) {
		return "", nil, gerror.Wrapf(ErrFileNotfound, "文件[%v]不存在", path)
	}

	// glog.Debugf(ctx, "从[%v]路径加载文件", absPath)
	return absPath, gfile.GetBytes(absPath), nil
}

func (w *Watcher) WriteBytes(ctx context.Context, path string, data []byte) error {
	path = gfile.Join(w.option.path, path)
	// absPath := gfile.RealPath(path) // TODO: 验证写入加密文件时的路径
	// glog.Debugf(ctx, "将向[%v]路径写入文件", path)

	return gfile.PutBytes(path, data)
}

func (w *Watcher) DeleteFile(ctx context.Context, path string) error {
	path = gfile.Join(w.option.path, path)
	absPath := gfile.RealPath(path)
	if !gfile.Exists(absPath) {
		return gerror.Wrapf(ErrFileNotfound, "文件[%v]不存在", path)
	}

	return gfile.RemoveFile(absPath)
}

func (w *Watcher) Stop() error {
	return nil
}

type path = string
type WatcherList struct {
	watchers       map[path]*Watcher
	watchersByType map[WatcherType]*Watcher // 按类别区分, 一类模块只会有一个

	ch *gqueue.TQueue[FileProcessTask] // 任务通道
}

var Watchers = WatcherList{
	watchers:       make(map[path]*Watcher, 3),
	watchersByType: make(map[WatcherType]*Watcher, 2),

	ch: gqueue.NewTQueue[FileProcessTask](),
}

func (list *WatcherList) AddWatcher(w *Watcher) {
	list.watchers[w.option.path] = w // 会覆盖前面的监控模块
	list.watchersByType[w.option.wtype] = w
}

func DefaultInit(
	uploadDir, encFileDir, unlockFileDir string,
	keySize int,
) *WatcherList {
	if keySize < 16 {
		keySize = 16
	}

	Watchers.AddBatchWatcher(
		New(
			WithWatchDir(uploadDir),
			WithLogger(glog.New()),
			WithWatcherType(WatcherForPlain),
			WithAESKeySize(keySize),
		),
		New(
			WithWatchDir(encFileDir),
			WithLogger(glog.New()),
			WithWatcherType(WatcherForEnc),
		),
		New(
			WithWatchDir(unlockFileDir),
			WithLogger(glog.New()),
			WithWatcherType(WatcherForDec),
		),
	)

	return &Watchers
}

func (list *WatcherList) AddBatchWatcher(watchers ...*Watcher) {
	for _, w := range watchers {
		list.watchers[w.option.path] = w // 会覆盖前面的监控模块
		list.watchersByType[w.option.wtype] = w
	}
}

func (list *WatcherList) GetWatcherByType(wtype WatcherType) (*Watcher, error) {
	if w, exists := list.watchersByType[wtype]; exists {
		return w, nil
	}
	return nil, gerror.NewCodef(
		gcode.CodeInternalError,
		"未找到类型[%v]的监控模块", wtype.String(),
	)
}

func (list *WatcherList) encryptWorker(ctx context.Context, task FileProcessTask) {
	var (
		err error
		key []byte

		path  = task.Path
		nonce = task.Nonce
	)
	defer func() {
		task.FailedForWhy = err
		if err == nil {
			task.key = key // 更新加密密钥
		}
	}()

	plainW, err := Watchers.GetWatcherByType(WatcherForPlain) // 找到管理待加密文件的模块
	if err != nil {
		return
	}
	// glog.Debugf(ctx, "检索到用于存储待加密文件的模块")
	encW, err := Watchers.GetWatcherByType(WatcherForEnc) // 找到管理已加密文件的模块
	if err != nil {
		return
	}
	// glog.Debugf(ctx, "检索到用于存储已加密文件的模块")

	// 生成AES-GCM加密器
	key = grand.B(plainW.option.keySize)
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	// glog.Debugf(ctx, "AES-GCM加密器 初始化完毕")

	// 加载数据
	_, plainData, err := plainW.GetFileBytes(ctx, path)
	if err != nil {
		return
	}
	// glog.Debugf(ctx, "待加密文件已加载完毕, 大小: %s", gfile.ReadableSize(abspath))

	// 加密数据
	cipherData := gcm.Seal(nil, nonce, plainData, nil)
	// glog.Debugf(ctx, "加密数据已生成, 大小: %s", gfile.FormatSize(int64(len(cipherData))))
	err = encW.WriteBytes(ctx, path, cipherData)
	if err != nil {
		return
	}
	// glog.Debugf(ctx, "已保存加密文件")

	if !task.DeleteOriginalFile {
		return
	}

	// TODO: 删除文件的分支
	err = plainW.DeleteFile(ctx, path)
	if err != nil {
		_ = encW.DeleteFile(ctx, path) // 删除加密好的文件
		return
	}
}

func (list *WatcherList) worker(ctx context.Context, task FileProcessTask) {
	switch task.Type {
	case TaskTypeEncryption:
		list.encryptWorker(ctx, task)
	default:
		glog.Debugf(ctx, "接收到未知的任务类型: %v", task.Type.String())
		task.FailedForWhy = gerror.Newf("未知任务类型: %v", task.Type.String())
	}
}

func (list *WatcherList) record(ctx context.Context, task FileProcessTask) {
	_, err := dao.Tasks.
		Ctx(ctx).
		Data(do.Tasks{
			IsSucceed:    task.FailedForWhy == nil,
			FailedForWhy: task.FailedForWhy,

			FinishedAt: gtime.Now(),
		}).
		WherePri(task.ID).
		Update()
	if err != nil {
		glog.Errorf(ctx, "无法更新任务[%d]: %V", task.ID, err)
	}
}

func (list *WatcherList) Run(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			default:
				select {
				case task := <-list.ch.C:
					list.worker(ctx, task)
					glog.Debugf(ctx, "接收到任务: %+v 并已处理", task)
					list.record(ctx, task) // 用回调函数更新数据库记录
				default:
					// 进入下一个循环
				}
			}
		}
	}()
}

// SearchFiles 检索指定目录下的文件; 给定目录需要提前被注册
func (list *WatcherList) SearchFiles(ctx context.Context, path string, page, size int) (res []FileInfo, err error) {
	w, exists := list.watchers[path]
	if !exists {
		w.log.Errorf(ctx, "检测到对未知目录[%v]的访问, 请注意非法目录穿越行为", path)
		return nil, ErrDirNotfound
	}

	return w.Export(page, size)
}
func SearchFiles(ctx context.Context, path string, page, size int) (res []FileInfo, err error) {
	return Watchers.SearchFiles(ctx, path, page, size)
}

func (list *WatcherList) EncryptFile(ctx context.Context, taskId int, path string, nonce []byte, deleteOriginFile ...bool) (err error) {
	var shouldDelete bool // 是否删除原始文件, 默认是删除的
	if len(deleteOriginFile) == 0 {
		shouldDelete = true
	}
	shouldDelete = deleteOriginFile[0]
	_ = shouldDelete

	list.ch.Push(FileProcessTask{
		ID:                 taskId,
		Path:               path,
		Nonce:              nonce,
		Type:               TaskTypeEncryption,
		DeleteOriginalFile: shouldDelete,
	})
	return nil
}
func EncryptFile(ctx context.Context, taskId int, path string, nonce []byte) (err error) {
	return Watchers.EncryptFile(ctx, taskId, path, nonce, true)
}

func (list *WatcherList) Stop() {
	for _, w := range list.watchers {
		_ = w.Stop()
	}
}
