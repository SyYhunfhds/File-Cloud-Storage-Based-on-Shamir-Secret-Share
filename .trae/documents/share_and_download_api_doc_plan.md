# Share List / Share Pull / ItemDownload 文档更新计划

## Summary
1. 新建 `backend/api/2026-06-15.md`：收录 Share List、Share Pull 两个新API + 更新 ItemDownload 文档
2. 更新 `frontend/api.md`：新增 Share List + Share Pull，更新 ItemDownload（支持成员下载、失败响应补充）


## Change 1: 新建 backend 文档 — `backend/api/2026-06-15.md`

### 1a. 份额查看 API `share/list`
- 来源：[share.go#L34-L43](file:///f:/GolangWorkspace/CryptographyDesign/backend/api/share/v1/share.go#L34-L43) + [share_v1_share_list.go](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/controller/share/share_v1_share_list.go)
- 业务逻辑：只返回当前用户的、条目公开的、审核通过的 device share；RLS 策略保护；按分页返回
- 请求 DTO: `{ "page": int, "size": int }`
- 响应 DTO: `ShareInfo`（share_id, item_id, filename, owner, is_expired, updated_at, expire_at）
- 成功/失败响应示例

### 1b. 份额拉取 API `share/pull`
- 来源：[share.go#L50-L57](file:///f:/GolangWorkspace/CryptographyDesign/backend/api/share/v1/share.go#L50-L57) + [share_v1_share_pull.go](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/controller/share/share_v1_share_pull.go)
- 业务逻辑：三步走（查未过期 → 检索份额 → 删除己取份额），取后即删
- 请求 DTO: `{ "share_ids": [int] }`
- 响应 DTO: `Share`（item_id, filename, device_share）
- 成功/全部过期/查询失败响应示例

### 1c. 条目下载 API `item/download`（更新版）
- 关键变化：
  1. 权限放宽：从"仅所有者" → "所有者或审核通过的member"
  2. 物理查找改用 Savename（不再依赖 Filename）
  3. Auth Share 现在直接存 Base64 到 DB（去掉AES-GCM二次加密）
- 补充完整的失败响应 JSON 列表（参考已有 tests/data 文件）


## Change 2: 更新前端文档 — `frontend/api.md`

### 2a. 更新 ItemDownload 部分（line ~409-424）
- 补充权限说明：所有者或已审核通过的成员均可下载
- 列出主要失败场景的 JSON 响应（至少含：未登录401、无权限403、份额无效400）
- 修正响应描述（成功时直接返回文件字节流；失败时走 JSON）

### 2b. 新增 Share List API（插入 `份额链路 share/` 下，在 Share Refresh 之后）
- 标题：份额查看功能 `GET POST share/list`
- 业务逻辑：只查自己的、条目公开的、审核通过的、未过期的 device share
- 请求/响应 DTO + 成功示例
- `share/list` 用于前端确认哪些份额可以拉取

### 2c. 新增 Share Pull API（接在 Share List 之后）
- 标题：份额拉取功能 `GET POST share/pull`
- 业务逻辑：拉取 → 传输份额 → 后端删除缓存，取后即焚
- 请求/响应 DTO + 成功/失败示例

## Verification
- 确认 `backend/api/2026-06-15.md` 包含 `share/list`、`share/pull`、`item/download` 三节
- 确认 `frontend/api.md` 的下载部分补充了失败响应
- 确认 `frontend/api.md` 份额部分有 Share List 和 Share Pull
- DTO 字段与 [share.go](file:///f:/GolangWorkspace/CryptographyDesign/backend/api/share/v1/share.go) 一致
