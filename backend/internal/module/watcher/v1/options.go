package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

type OptionFunc func(w *Watcher)

func WithDefaultConfig() OptionFunc {
	return func(w *Watcher) {
		w.workers = 1 // 协程数量

		w.watchDir = "./uploads"

		w.log = g.Log("条目文件监控")

		w.isToBeEncrypted = validateExtnameFunc(map[string]struct{}{
			"pdf": {},
			// "enc": {}, // 加密后的文件拓展名, 建议单列出去
			"txt": {},
		})
		w.isHasEncrypted = validateExtnameFunc(map[string]struct{}{
			"enc": {},
		})
	}
}

func WithWatchDir(watchDir string) OptionFunc {
	return func(w *Watcher) {
		w.watchDir = watchDir
	}
}

func WithWorkers(workers int32) OptionFunc {
	return func(w *Watcher) {
		w.workers = workers
	}
}

func WithGLogger(logger glog.ILogger) OptionFunc {
	return func(w *Watcher) {
		w.log = logger
	}
}

func WithValidExtNames(extNames []string) OptionFunc {
	mapExtNames := make(map[string]struct{}, len(extNames))
	for _, extName := range extNames { // Go 1.22以前的版本不建议这么写
		mapExtNames[extName] = struct{}{}
	}

	return func(w *Watcher) {
		w.isToBeEncrypted = validateExtnameFunc(mapExtNames)
	}
}
