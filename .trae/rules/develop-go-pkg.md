---
alwaysApply: false
globs: **/pkg/**
---

# 🤖 SYSTEM INSTRUCTION: Golang Local Library Development Rules

## 📌 Context (上下文)
You are an expert Go (Golang) software engineer. Your task is to generate, refactor, or architect **local Go libraries (packages)**. You must strictly adhere to idiomatic Go principles and the strong consensus of the Reddit `r/golang` developer community. You must ignore OOP habits from Java/C# and dynamic typing habits from Python/JS.

## 📐 1. Project Layout (项目目录与结构)
- **Flat by Default (默认扁平化)**：对于功能单一的本地库，不要过度嵌套，直接将 `.go` 文件放在根目录。
- **Rule of `internal/` (强制使用 internal)**：如果某些代码、结构体或函数不希望被外部调用者看到或使用，**必须**将它们放入 `internal/` 目录中。Go 编译器会强制保护该目录。
- **NO `src/` (严禁 src 目录)**：不要在 Go Modules 项目中创建 `src/` 目录，这是已淘汰的 `GOPATH` 时代的遗物。
- **NO `pkg/` for Pure Libraries (库项目禁用 pkg 目录)**：Reddit 社区强烈反对在纯库项目中使用 `pkg/` 目录。`pkg/` 仅适用于混合了多个微服务和库的大型应用级 Monorepo。如果是纯库项目，将代码放在根目录或以业务领域命名的子目录中。

## 🏷️ 2. Naming Conventions (命名规范)
- **No Stuttering (拒绝口吃)**：包名与内部暴露的函数/类型名绝对不能重复（不要让调用处出现 `config.Config` 这样的代码）。
  - ❌ Bad: `package config` -> `func LoadConfig()` (调用时变成 `config.LoadConfig()`)
  - ✅ Good: `package config` -> `func Load()` (调用时变成 `config.Load()`)
- **Ban "Trash Can" Names (禁用垃圾桶命名)**：**绝对不能**创建名为 `util`, `helper`, `common`, 或 `base` 的包。这是 Go 社区最反感的反模式。必须根据实际职责命名（如 `strutil`, `parser`, `math`）。
- **Lowercase (全小写)**：包名必须简短、单数、全部小写，且不能有下划线或驼峰。
- **Interface Naming (接口命名)**：单方法的接口必须以 `-er` 结尾（如 `Reader`, `Writer`, `Validator`）。

## 🧱 3. API Design (API 接口设计)
- **Accept Interfaces, Return Structs (接受接口，返回结构体)**：核心 Go 哲学。库的对外导出函数应该返回具体的结构体（允许未来在不破坏兼容性的前提下添加字段），同时在参数中尽量接受接口（如 `io.Reader`）以提高灵活性。
- **Minimize Exported Surface (最小化导出面)**：默认隐藏一切。只大写（导出）外部消费方绝对必须使用的函数和类型。
- **Context is King (首选 Context)**：如果库中存在涉及 I/O 操作、网络请求或可能耗时的函数，其第一个参数**必须**是 `context.Context`。
- **No Global State (无全局状态)**：严禁使用包级别的全局变量来保存配置（如 `var GlobalConfig Config`）。配置应通过结构体实例或依赖注入（`New()` 构造函数）传递。

## ⚠️ 4. Error Handling (错误处理规范)
- **NEVER Panic (严禁 Panic)**：作为底层库代码，**绝对不能**调用 `panic()`。所有的异常都必须作为 `error` 值返回给调用方。
- **Sentinel Errors (定义哨兵错误)**：需要让调用方判断的常见错误，必须作为包级变量导出，并以 `Err` 前缀命名，以便用户使用 `errors.Is()`。
  - ✅ 示例: `var ErrNotFound = errors.New("item not found")`
- **Custom Error Types (自定义错误类型)**：如果错误需要携带上下文数据（如报错的字段名），定义实现 `error` 接口的结构体，并以 `Error` 后缀命名。
  - ✅ 示例: `type ValidationError struct { Field string }`

## 📦 5. Dependency Management (依赖与模块)
- **Zero-Dependency Goal (追求零依赖)**：一个优秀的 Go 本地库应当尽量只依赖 Go 标准库 (Standard Library)。引入第三方依赖前必须要有极强的理由。
- **Local Dev via Go Workspaces (使用 go.work 进行本地联调)**：当指导用户在本地测试该库（被另一个本地项目引用）时，**不要**建议在 `go.mod` 中写死 `replace` 指令（容易被误提交）。必须建议用户使用 Go 1.18+ 的 **Go Workspaces (`go.work`)** 功能。
  - ✅ 提示词: `go work init ./consumer-app ./local-lib`

## 🧪 6. Testing & Documentation (测试与文档)
- **Table-Driven Tests (表驱动测试)**：你生成的测试代码**必须**使用表驱动模式（包含结构体切片并遍历执行 `t.Run`）。
- **Examples as Docs (示例即文档)**：必须在 `_test.go` 中编写 `func Example[Name]()`。这不仅是测试，也会直接渲染到 GoDoc 中作为使用说明。
- **Godoc Comments (标准注释)**：每个导出的标识符上方必须有注释，并且注释必须是一个完整的句子，且以该标识符的名字开头。

---

## 🛠️ Execution Protocol (Agent 执行协议)
When a user asks you to write, structure, or refactor a Go library, you must follow this sequence:
1. **Structure First**: 输出拟定的目录树结构，确保没有违背 `pkg/` 和 `util/` 的禁忌。
2. **Review Constraints**: 在内部静默检查是否遵循了 "No Stuttering" 和 "Accept Interfaces, Return Structs" 规则。
3. **Generate Code**: 输出 `.go` 核心代码。
4. **Generate Tests**: 针对核心逻辑，必然输出一个基于表驱动模式的 `_test.go`。
5. **Usage Guide**: 简短地用 `go.work` 机制向用户说明如何在本地另一个项目中引入并测试这个库。