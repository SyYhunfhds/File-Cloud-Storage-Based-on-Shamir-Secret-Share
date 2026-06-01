package v2

// WatcherType 监控模块的类型
type WatcherType uint

const (
	CommonWatcher   = iota // 一般文件监控模块
	WatcherForPlain        // 用于待加密文件的模块
	WatcherForEnc          // 用于存储加密文件的模块
	WatcherForDec          // 用于存储解密文件的模块
)

var watcherTypeStr = map[WatcherType]string{
	CommonWatcher:   "一般模块",
	WatcherForPlain: "存储待加密文件模块",
	WatcherForEnc:   "存储待解密文件模块",
	WatcherForDec:   "存储已解密文件模块",
}

func (t WatcherType) String() string {
	return watcherTypeStr[t]
}

type EncryptionTask struct {
	// Path 待加密文件的路径
	Path string
	// Nonce AES-GCM初始向量
	Nonce []byte `json:"-" yaml:"-"` // 以防万一被序列化
	// DeleteOriginalFile 是否删除源文件
	DeleteOriginalFile bool

	// FailedForWhy 文件处理失败的原因
	FailedForWhy error
}
type DecryptionTask struct {
	// Path 待加密文件的路径
	Path string
	// Nonce AES-GCM初始向量
	Nonce []byte `json:"-" yaml:"-"` // 以防万一被序列化
	// DeleteOriginalFile 是否删除源文件
	DeleteOriginalFile bool

	// FailedForWhy 文件处理失败的原因
	FailedForWhy error
}

type TaskType uint

const (
	TaskTypeUnknown    TaskType = iota
	TaskTypeEncryption          = iota
	TaskTypeDecryption
)

var taskTypeStr = map[TaskType]string{
	TaskTypeUnknown:    "未知任务",
	TaskTypeEncryption: "加密任务",
	TaskTypeDecryption: "解密任务",
}

func (t TaskType) String() string {
	return taskTypeStr[t]
}

// FileProcessTask 文件处理任务
type FileProcessTask struct {
	ID   int // 任务标识
	Type TaskType

	// Path 待加密文件的路径
	Path string
	// Key 待加密文件的密钥
	key []byte
	// Nonce AES-GCM初始向量
	Nonce []byte `json:"-" yaml:"-"` // 以防万一被序列化
	// DeleteOriginalFile 是否删除源文件
	DeleteOriginalFile bool

	// FailedForWhy 文件处理失败的原因
	FailedForWhy error
}
