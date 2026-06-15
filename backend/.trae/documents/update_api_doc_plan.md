# API文档更新计划（修订版）

## 背景
`api/2026-06-09.md` 目前包含3个API的文档（审计列表、审计操作、份额刷新）。
Item相关的API已在 `api/2026-06-01.md` 和 `api/2026-05-30.md` 中完整记录，无需重复。

## 需要更新的内容

### 变更1：`LessDetailedAudit` 新增字段

**文件**: `api/audit/v1/audit.go`
**位置**: 文档中"审计列表API"的 `LessDetailedAudit` 结构体定义

| 新增字段 | 类型 | JSON | 说明 |
|----------|------|------|------|
| `AuditId` | `[]int` | `audit_id` | 审核ID列表（用于批量操作定位） |
| `ItemId` | `int` | `item_id` | 条目ID（用于份额刷新等后续操作） |

**原因**: 原文档缺少这两个字段，但代码中已实现。

**影响**: 
- Go结构体代码块需补充两个字段
- 响应示例JSON需补充 `audit_id` 和 `item_id` 字段

### 变更2：`Item/delete` 新增实现

**文件**: `api/item/v1/item.go` + `internal/controller/item/item_v1_item_delete.go`
**状态**: 在 `2026-05-30.md` 中标注为"未实现"，现已实现。
**位置**: 文档末尾新增"条目删除"的说明。

当前代码的 `ItemDeleteReq/Res` 定义：
```go
type ItemDeleteReq struct {
    g.Meta `method:"get,post" path:"/item/delete" tags:"Item" summary:"删除条目"`
    ItemIds []int `json:"item_ids" form:"item_ids" v:"required" dc:"指定至少一个条目ID"`
}
type ItemDeleteRes struct {
    Deleted int64 `json:"total_deleted" dc:"成功删除的条目的个数"`
}
```

**注意**: 与旧文档中的 `Filename []string` 不同，现已改为 `ItemIds []int`。

## 不变的内容
- **审计操作API**: 文档已与代码一致（`Operations map[int]string` + `Affected int64`）
- **份额刷新API**: DTO与文档一致，SSE流程描述准确

## 计划步骤

### Step 1: 更新审计列表 `LessDetailedAudit` 
- 补充 `AuditId []int` 和 `ItemId int` 字段到结构体定义
- 更新 `ItemDownloadReq` 路径从 `/item/download/placeholder` 改为 `/item/download`
- 更新 JSON 响应示例

### Step 2: 新增 `Item/delete` 条目
- 在文档末尾新增"条目删除"API描述
- 提供请求DTO、响应DTO和响应示例

### Step 3: 验证
- 对照 `api/audit/v1/audit.go` 确认 `LessDetailedAudit` 完整
- 对照 `api/item/v1/item.go` 确认 `ItemDeleteReq/Res` 准确
