# Frontend API Doc Update Plan

## Summary
同步 `frontend/api.md` 与后端最新 API 实现，涉及 ItemUpdate 请求 DTO 变更、新增 ItemDelete 文档、修正已失效的测试数据引用路径。

## Changes

### Change 1: ItemUpdate 请求 DTO 更新 (line 428-439 + line 847-859)
- **What**: `Filename string` → `ItemId int`
- **Why**: 后端改用 `ItemId` 定位条目（[item.go#L90-L101](file:///f:/GolangWorkspace/CryptographyDesign/backend/api/item/v1/item.go#L90-L101)），`Filename` 不再唯一
- **How**: 
  - 删除 `Filename string`，替换为 `ItemId int`
  - 删除注释 "注意: **文件名重命名**仍不可用, 请不要对接这个字段"
  - 更新"请求参数为空"描述中的说明文本
  - 同步附录中 `ItemUpdateReq` 的相同修改

### Change 2: 新增 ItemDelete 文档 (after line 479)
- **What**: 在 ItemUpdate 和 Apply 之间插入 ItemDelete API 文档
- **Why**: 后端已实现水平越权检测的批量删除（[item.go#L106-L115](file:///f:/GolangWorkspace/CryptographyDesign/backend/api/item/v1/item.go#L106-L115)、[controller](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/controller/item/item_v1_item_delete.go)）
- **How**: 
  - 标题: `条目删除功能 GET POST item/delete`
  - 请求 DTO: `{ "item_ids": [int] }`
  - 响应 DTO: `{ "total_deleted": int64 }`
  - 业务逻辑说明: 批量删除、水平越权检查、只删自己条目

### Change 3: 修正 ItemDownload 测试数据引用路径 (line 514-515)
- **What**: `test\testdata\` → `tests\data\`
- **Why**: 实际测试数据目录为 `tests/data/`，路径错误
- **How**: 替换两个文件路径

## Verification
- 最终文档检查: 无残留的 `Filename string` ItemUpdate 引用
- ItemDelete DTO 与 [item.go#L106-L115](file:///f:/GolangWorkspace/CryptographyDesign/backend/api/item/v1/item.go#L106-L115) 一致
