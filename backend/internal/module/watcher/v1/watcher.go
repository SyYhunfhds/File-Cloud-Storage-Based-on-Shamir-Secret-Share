// 基于gfsnotify模块构建的监听待审计条目变动的模块

package v1

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfsnotify"
	"github.com/gogf/gf/v2/os/glog"
)

type Status int

const (
	FileIsUnknown Status = iota
	FileIsPlain
	FileIsProcessing // 文件正在处理中
	FileIsEncrypted

	FileIsLocked   // 文件被锁定
	FileIsNotfound // 文件不存在
)

var FileStatus = map[Status]string{
	FileIsUnknown: "未知文件", // 指的是不会被处理的文件

	FileIsPlain:      "普通文件",
	FileIsProcessing: "正在处理中",
	FileIsEncrypted:  "已加密文件",

	FileIsLocked:   "文件被锁定",
	FileIsNotfound: "文件不存在",
}

type WrappedFile struct {
	Filename string // 文件名
	Filepath string // 绝对路径
	FileSize int64  // 文件大小

	Status
}

type Files struct {
	mux   sync.Mutex
	files map[string]WrappedFile
}

func NewFiles() *Files {
	return &Files{
		mux:   sync.Mutex{},
		files: map[string]WrappedFile{},
	}
}

func (f *Files) Len() int {
	f.mux.Lock()
	defer f.mux.Unlock()

	return len(f.files)
}

func (f *Files) CountWithFlag(flag Status) (n int) {
	f.mux.Lock()
	defer f.mux.Unlock()

	for _, v := range f.files {
		if v.Status == flag {
			n++
		}
	}
	return n
}
func (f *Files) CountEncrypted() (n int) {
	return f.CountWithFlag(FileIsEncrypted)
}
func (f *Files) CountPlain() (n int) {
	return f.CountWithFlag(FileIsPlain)
}

func (f *Files) AddFiles(files []WrappedFile) {
	f.mux.Lock()
	defer f.mux.Unlock()

	for _, file := range files {
		if _, exists := f.files[file.Filename]; !exists {
			f.files[file.Filename] = file
			continue
		}

	}
}

// AddSingleFile
//
// deprecated, 不建议使用, 不会自动比较文件状态
func (f *Files) AddSingleFile(file WrappedFile) {
	f.mux.Lock()
	defer f.mux.Unlock()

	f.files[file.Filename] = file
}

// Get 获取浅拷贝
func (f *Files) Get(k string) (*WrappedFile, bool) {
	f.mux.Lock()
	defer f.mux.Unlock()

	if file, exists := f.files[k]; exists {
		return &file, true
	}
	return nil, false
}

func (f *Files) String() string {
	f.mux.Lock()
	defer f.mux.Unlock()

	var sb = strings.Builder{}
	for filename, info := range f.files {
		sb.WriteString(filename)
		sb.WriteString(": ")
		sb.WriteString(info.Filename)
		sb.WriteString(", ")
		sb.WriteString(FileStatus[info.Status])
		sb.WriteString("\n")
	}
	return sb.String()
}

type Watcher struct {
	*Files

	workers        int32
	wg             sync.WaitGroup
	runningWorkers atomic.Int32
	cancel         context.CancelFunc

	// 其他配置项
	callback func(event *gfsnotify.Event) // 传递给gfsnotify模块
	watchDir string                       // TODO: 增加 目录穿越漏洞防护

	// 辅助字段
	log glog.ILogger

	// 校验函数
	isToBeEncrypted func(ext string) bool
	isHasEncrypted  func(ext string) bool
}

func New(options ...OptionFunc) *Watcher {
	watcher := &Watcher{
		Files: NewFiles(),

		runningWorkers: atomic.Int32{},
		wg:             sync.WaitGroup{},
	}
	watcher.callback = func(event *gfsnotify.Event) {

	}

	WithDefaultConfig()(watcher)
	for _, option := range options {
		option(watcher)
	}

	return watcher
}

func (w *Watcher) scanFiles() (n int, err error) {
	entries, err := os.ReadDir(w.watchDir)
	if err != nil {
		w.log.Errorf(gctx.New(), "无法检索审计条目目录: %v", err)
		return
	}

	wfiles := make([]WrappedFile, 0, len(entries))

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}

		entryNameParts := strings.Split(info.Name(), ".")
		if len(entryNameParts) <= 1 {
			continue
		}

		// basename := entryNameParts[0]
		extName := strings.ToLower(entryNameParts[len(entryNameParts)-1])
		var fileStatus Status
		if w.isToBeEncrypted(extName) {
			fileStatus = FileIsPlain
		} else if w.isHasEncrypted(extName) {
			fileStatus = FileIsEncrypted
		}

		absFilepath, err := filepath.Abs(entry.Name()) // 获取完整路径
		if err != nil {
			continue
		}

		wfiles = append(wfiles, WrappedFile{
			Filename: info.Name(), // 这会导致abc.pdf和abc.enc被识别为不同的文件
			Filepath: absFilepath,
			FileSize: info.Size(),

			Status: fileStatus,
		})
	}
	w.Files.AddFiles(wfiles)

	return
}

func (w *Watcher) worker(ctx context.Context) {
	defer w.wg.Done()
	defer w.runningWorkers.Add(-1)

	<-ctx.Done()
}

func (w *Watcher) Run(ctx context.Context) (context.CancelFunc, error) {
	if _, err := w.scanFiles(); err != nil { // 扫描文件进入缓存
		return nil, err
	}

	innerCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	// 启动监控
	_, err := gfsnotify.Add(w.watchDir, w.callback)
	if err != nil {
		w.log.Errorf(gctx.New(), "无法启动文件监听: %v", err)
		return nil, err
	}

	for range w.workers - w.runningWorkers.Load() {
		w.wg.Add(1)
		w.runningWorkers.Add(1)

		go w.worker(innerCtx)
	}

	return cancel, nil
}

func (w *Watcher) PrintFiles() {
	println(w.String())
}

func (w *Watcher) Shutdown(ctx context.Context) error {
	if w.cancel != nil {
		w.cancel()
	} else {
		return errors.New("watcher is not running")
	}

	done := make(chan struct{}, 1)
	defer close(done)
	go func() {
		w.wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
