package v2

import "github.com/gogf/gf/v2/os/glog"

type WatcherOption struct {
	path    string
	pattern string // 文件匹配规则, 直接影响什么到能处理什么文件

	wtype WatcherType

	// 加密条目配置
	keySize int // 密钥长度
}

func DefaultOption() WatcherOption {
	return WatcherOption{
		path:    "./uploads",
		pattern: "*.*",
		wtype:   CommonWatcher,
	}
}

type WatcherOptionFunc func(*Watcher)

func WithWatchDir(dir string) WatcherOptionFunc {
	return func(w *Watcher) {
		w.option.path = dir
	}
}
func WithSearchPattern(pattern string) WatcherOptionFunc {
	return func(w *Watcher) {
		w.option.pattern = pattern
	}
}

func WithLogger(logger *glog.Logger) WatcherOptionFunc {
	return func(w *Watcher) {
		w.log = logger
	}
}
func WithWatcherType(wtype WatcherType) WatcherOptionFunc {
	return func(w *Watcher) {
		w.option.wtype = wtype
	}
}
func WithAESKeySize(size int) WatcherOptionFunc {
	return func(w *Watcher) {
		w.option.keySize = size
	}
}
